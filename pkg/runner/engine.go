package runner

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tpsapp/applauncher3/pkg/models"
)

var (
	ErrQueueAlreadyRunning = errors.New("queue is already running")
	ErrQueueNotRunning     = errors.New("queue is not currently running")
)

type QueueState string

const (
	StateIdle      QueueState = "Idle"
	StateRunning   QueueState = "Running"
	StatePaused    QueueState = "Paused"
	StateCancelled QueueState = "Cancelled"
	StateCompleted QueueState = "Completed"
)

type ProgressEvent struct {
	CurrentIndex   int               `json:"currentIndex"`
	TotalItems     int               `json:"totalItems"`
	CurrentItem    models.LaunchItem `json:"currentItem"`
	ItemStatus     string            `json:"itemStatus"`
	State          QueueState        `json:"state"`
	DelayRemaining float64           `json:"delayRemaining"`
	ErrorMessage   string            `json:"errorMessage,omitempty"`
}

type Engine struct {
	spawner     ProcessSpawner
	mu          sync.Mutex
	pauseCond   *sync.Cond
	state       QueueState
	isPaused    atomic.Bool
	isCancelled atomic.Bool
	cancelFunc  context.CancelFunc
}

func NewEngine(spawner ProcessSpawner) *Engine {
	if spawner == nil {
		spawner = NewOSSpawner()
	}
	e := &Engine{
		spawner: spawner,
		state:   StateIdle,
	}
	e.pauseCond = sync.NewCond(&e.mu)
	return e
}

func (e *Engine) GetState() QueueState {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.state
}

func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.state == StateRunning {
		e.state = StatePaused
		e.isPaused.Store(true)
	}
}

func (e *Engine) Resume() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.state == StatePaused {
		e.state = StateRunning
		e.isPaused.Store(false)
		e.pauseCond.Broadcast()
	}
}

func (e *Engine) Cancel() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.state == StateRunning || e.state == StatePaused {
		e.state = StateCancelled
		e.isCancelled.Store(true)
		e.isPaused.Store(false)
		e.pauseCond.Broadcast()
		if e.cancelFunc != nil {
			e.cancelFunc()
		}
	}
}

func (e *Engine) Start(ctx context.Context, profile models.Profile, dryRun bool, onProgress func(ProgressEvent)) error {
	e.mu.Lock()
	if e.state == StateRunning || e.state == StatePaused {
		e.mu.Unlock()
		return ErrQueueAlreadyRunning
	}
	e.state = StateRunning
	e.isPaused.Store(false)
	e.isCancelled.Store(false)

	runCtx, cancel := context.WithCancel(ctx)
	e.cancelFunc = cancel
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		if e.state != StateCancelled {
			e.state = StateCompleted
		}
		cancel()
		e.mu.Unlock()
	}()

	items := profile.Items
	total := len(items)

	for i := 0; i < total; i++ {
		item := items[i]

		if e.isCancelled.Load() || runCtx.Err() != nil {
			e.emitProgress(onProgress, ProgressEvent{
				CurrentIndex: i,
				TotalItems:   total,
				CurrentItem:  item,
				ItemStatus:   "Cancelled",
				State:        StateCancelled,
			})
			continue
		}

		// Wait while paused
		e.mu.Lock()
		for e.isPaused.Load() && !e.isCancelled.Load() && runCtx.Err() == nil {
			e.pauseCond.Wait()
		}
		if e.isCancelled.Load() || runCtx.Err() != nil {
			e.mu.Unlock()
			e.emitProgress(onProgress, ProgressEvent{
				CurrentIndex: i,
				TotalItems:   total,
				CurrentItem:  item,
				ItemStatus:   "Cancelled",
				State:        StateCancelled,
			})
			continue
		}
		e.mu.Unlock()

		if !item.Enabled {
			e.emitProgress(onProgress, ProgressEvent{
				CurrentIndex: i,
				TotalItems:   total,
				CurrentItem:  item,
				ItemStatus:   "Skipped",
				State:        StateRunning,
			})
			continue
		}

		// Transition to Running
		e.emitProgress(onProgress, ProgressEvent{
			CurrentIndex: i,
			TotalItems:   total,
			CurrentItem:  item,
			ItemStatus:   "Running",
			State:        StateRunning,
		})

		// Spawning attempt
		var spawnErr error
		if !dryRun {
			_, spawnErr = e.spawner.Spawn(&item)
		}

		itemStatus := "Success"
		errMsg := ""
		if spawnErr != nil {
			itemStatus = "Failed"
			errMsg = spawnErr.Error()
		}

		e.emitProgress(onProgress, ProgressEvent{
			CurrentIndex: i,
			TotalItems:   total,
			CurrentItem:  item,
			ItemStatus:   itemStatus,
			State:        StateRunning,
			ErrorMessage: errMsg,
		})

		// Delay handling
		if item.DelaySeconds > 0 {
			if err := e.waitDelay(runCtx, item.DelaySeconds, i, total, item, onProgress); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Engine) waitDelay(ctx context.Context, delaySec float64, idx, total int, item models.LaunchItem, onProgress func(ProgressEvent)) error {
	remaining := time.Duration(delaySec * float64(time.Second))
	tickInterval := 50 * time.Millisecond

	for remaining > 0 {
		if e.isCancelled.Load() || ctx.Err() != nil {
			return ctx.Err()
		}

		// Handle pause during delay
		e.mu.Lock()
		for e.isPaused.Load() && !e.isCancelled.Load() && ctx.Err() == nil {
			e.pauseCond.Wait()
		}
		if e.isCancelled.Load() || ctx.Err() != nil {
			e.mu.Unlock()
			return ctx.Err()
		}
		e.mu.Unlock()

		step := tickInterval
		if step > remaining {
			step = remaining
		}

		timer := time.NewTimer(step)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			remaining -= step
			e.emitProgress(onProgress, ProgressEvent{
				CurrentIndex:   idx,
				TotalItems:     total,
				CurrentItem:    item,
				ItemStatus:     "Delaying",
				State:          e.GetState(),
				DelayRemaining: remaining.Seconds(),
			})
		}
	}
	return nil
}

func (e *Engine) emitProgress(callback func(ProgressEvent), evt ProgressEvent) {
	if callback != nil {
		callback(evt)
	}
}
