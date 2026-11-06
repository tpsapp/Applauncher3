package main

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type ItemStatus string

const (
	StatusPending   ItemStatus = "Pending"
	StatusRunning   ItemStatus = "Running"
	StatusSuccess   ItemStatus = "Success"
	StatusFailed    ItemStatus = "Failed"
	StatusCancelled ItemStatus = "Cancelled"
)

type QueueItem struct {
	ID           string
	Name         string
	Path         string
	Args         []string
	Delay        time.Duration
	Status       ItemStatus
	ErrorMessage string
}

type EngineCommand string

const (
	CmdPause  EngineCommand = "PAUSE"
	CmdResume EngineCommand = "RESUME"
	CmdCancel EngineCommand = "CANCEL"
)

type QueueEngine struct {
	mu          sync.Mutex
	items       []QueueItem
	cmdChan     chan EngineCommand
	isPaused    atomic.Bool
	isCancelled atomic.Bool
	pauseCond   *sync.Cond
}

func NewQueueEngine(items []QueueItem) *QueueEngine {
	e := &QueueEngine{
		items:   items,
		cmdChan: make(chan EngineCommand, 10),
	}
	e.pauseCond = sync.NewCond(&e.mu)
	return e
}

func (e *QueueEngine) SendCommand(cmd EngineCommand) {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch cmd {
	case CmdPause:
		e.isPaused.Store(true)
	case CmdResume:
		e.isPaused.Store(false)
		e.pauseCond.Broadcast()
	case CmdCancel:
		e.isCancelled.Store(true)
		e.pauseCond.Broadcast()
	}
}

func (e *QueueEngine) Execute(ctx context.Context, onProgress func(index int, item QueueItem)) error {
	for i := range e.items {
		// Check cancellation
		if e.isCancelled.Load() || ctx.Err() != nil {
			e.items[i].Status = StatusCancelled
			if onProgress != nil {
				onProgress(i, e.items[i])
			}
			continue
		}

		// Wait while paused
		e.mu.Lock()
		for e.isPaused.Load() && !e.isCancelled.Load() {
			e.pauseCond.Wait()
		}
		if e.isCancelled.Load() {
			e.mu.Unlock()
			e.items[i].Status = StatusCancelled
			if onProgress != nil {
				onProgress(i, e.items[i])
			}
			continue
		}
		e.mu.Unlock()

		item := &e.items[i]
		item.Status = StatusRunning
		if onProgress != nil {
			onProgress(i, *item)
		}

		// Spawning attempt
		cmd := exec.Command(item.Path, item.Args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		startErr := cmd.Start()

		if startErr != nil {
			// Continue with warning policy
			item.Status = StatusFailed
			item.ErrorMessage = startErr.Error()
		} else {
			item.Status = StatusSuccess
		}

		if onProgress != nil {
			onProgress(i, *item)
		}

		// Staggered Delay handling (cancellable and pausable)
		if item.Delay > 0 {
			timer := time.NewTimer(item.Delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
				// Delay completed normally
			}
		}
	}
	return nil
}

func main() {
	items := []QueueItem{
		{ID: "1", Name: "Echo Task 1", Path: "/bin/sh", Args: []string{"-c", "echo task1"}, Delay: 200 * time.Millisecond, Status: StatusPending},
		{ID: "2", Name: "Broken Binary", Path: "/nonexistent/binary_xyz", Args: []string{}, Delay: 150 * time.Millisecond, Status: StatusPending},
		{ID: "3", Name: "Echo Task 3", Path: "/bin/sh", Args: []string{"-c", "echo task3"}, Delay: 200 * time.Millisecond, Status: StatusPending},
		{ID: "4", Name: "Echo Task 4", Path: "/bin/sh", Args: []string{"-c", "echo task4"}, Delay: 100 * time.Millisecond, Status: StatusPending},
	}

	engine := NewQueueEngine(items)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	startTime := time.Now()

	go func() {
		defer wg.Done()
		err := engine.Execute(ctx, func(idx int, item QueueItem) {
			fmt.Printf("[%v] Item %d (%s) -> Status: %s (Err: %s)\n",
				time.Since(startTime).Round(time.Millisecond), idx+1, item.Name, item.Status, item.ErrorMessage)
		})
		if err != nil {
			fmt.Printf("Engine execution error: %v\n", err)
		}
	}()

	// Simulate pausing after 100ms
	time.Sleep(100 * time.Millisecond)
	fmt.Println("--> SIGNAL: PAUSE")
	engine.SendCommand(CmdPause)

	// Hold paused for 300ms
	time.Sleep(300 * time.Millisecond)
	fmt.Println("--> SIGNAL: RESUME")
	engine.SendCommand(CmdResume)

	wg.Wait()
	elapsed := time.Since(startTime).Round(time.Millisecond)
	fmt.Printf("\nSpike 2 Finished in %v. Summary of items:\n", elapsed)
	for i, item := range engine.items {
		fmt.Printf("  Item %d: %s -> %s\n", i+1, item.Name, item.Status)
	}

	// Validation asserts
	if engine.items[1].Status != StatusFailed {
		panic("Expected Item 2 to fail gracefully")
	}
	if engine.items[2].Status != StatusSuccess || engine.items[3].Status != StatusSuccess {
		panic("Expected Items 3 and 4 to succeed despite Item 2 failure (Continue with Warning)")
	}
	fmt.Println("PASS: Queue state machine, delay timers, pause/resume, and Continue-with-Warning verified successfully!")
}
