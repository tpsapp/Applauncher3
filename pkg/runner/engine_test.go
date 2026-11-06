package runner

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/tpsapp/applauncher3/pkg/models"
)

type MockSpawner struct {
	mu           sync.Mutex
	spawnedItems []string
	failOnItem   string
	waitFuncs    map[string]func() error
}

func (m *MockSpawner) Spawn(item *models.LaunchItem) (*SpawnResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.spawnedItems = append(m.spawnedItems, item.ID)
	if m.failOnItem != "" && item.ID == m.failOnItem {
		return nil, errors.New("synthetic spawn failure")
	}

	waitFn := func() error { return nil }
	if m.waitFuncs != nil && m.waitFuncs[item.ID] != nil {
		waitFn = m.waitFuncs[item.ID]
	}

	return &SpawnResult{
		PID:  1000 + len(m.spawnedItems),
		Wait: waitFn,
	}, nil
}

func TestTC_QUE_01_HappyPath_SequentialExecution(t *testing.T) {
	mock := &MockSpawner{}
	engine := NewEngine(mock)

	profile := models.Profile{
		Name: "TestProfile",
		Items: []models.LaunchItem{
			{ID: "item1", Name: "Item 1", Path: "/bin/echo", DelaySeconds: 0.05, Enabled: true},
			{ID: "item2", Name: "Item 2", Path: "/bin/echo", DelaySeconds: 0.05, Enabled: true},
		},
	}

	var events []ProgressEvent
	err := engine.Start(context.Background(), profile, false, func(evt ProgressEvent) {
		events = append(events, evt)
	})

	if err != nil {
		t.Fatalf("Engine Start failed: %v", err)
	}

	if engine.GetState() != StateCompleted {
		t.Errorf("Expected StateCompleted, got %s", engine.GetState())
	}

	if len(mock.spawnedItems) != 2 {
		t.Errorf("Expected 2 spawned items, got %d", len(mock.spawnedItems))
	}
	if mock.spawnedItems[0] != "item1" || mock.spawnedItems[1] != "item2" {
		t.Errorf("Spawn order incorrect: %+v", mock.spawnedItems)
	}
}

func TestTC_QUE_02_FaultPolicy_ContinueWithWarning(t *testing.T) {
	mock := &MockSpawner{failOnItem: "item2"}
	engine := NewEngine(mock)

	profile := models.Profile{
		Name: "TestProfile",
		Items: []models.LaunchItem{
			{ID: "item1", Name: "Item 1", Path: "/bin/echo", DelaySeconds: 0.02, Enabled: true},
			{ID: "item2", Name: "Failing Item", Path: "/bin/invalid", DelaySeconds: 0.02, Enabled: true},
			{ID: "item3", Name: "Item 3", Path: "/bin/echo", DelaySeconds: 0.02, Enabled: true},
		},
	}

	failedEventRecorded := false
	err := engine.Start(context.Background(), profile, false, func(evt ProgressEvent) {
		if evt.CurrentItem.ID == "item2" && evt.ItemStatus == "Failed" {
			failedEventRecorded = true
		}
	})

	if err != nil {
		t.Fatalf("Engine should not return fatal error on single item failure: %v", err)
	}

	if !failedEventRecorded {
		t.Errorf("Expected Failed event for item2")
	}

	// Verify Item 3 still executed after Item 2 failed (Continue with Warning)
	if len(mock.spawnedItems) != 3 {
		t.Errorf("Expected all 3 items attempted, got %d", len(mock.spawnedItems))
	}
	if mock.spawnedItems[2] != "item3" {
		t.Errorf("Expected item3 spawned as 3rd item, got %s", mock.spawnedItems[2])
	}
}

func TestTC_QUE_03_PauseAndResume(t *testing.T) {
	mock := &MockSpawner{}
	engine := NewEngine(mock)

	profile := models.Profile{
		Name: "TestProfile",
		Items: []models.LaunchItem{
			{ID: "item1", Name: "Item 1", Path: "/bin/echo", DelaySeconds: 0.20, Enabled: true},
			{ID: "item2", Name: "Item 2", Path: "/bin/echo", DelaySeconds: 0.02, Enabled: true},
		},
	}

	start := time.Now()
	go func() {
		// Pause after 50ms while item1 is delaying
		time.Sleep(50 * time.Millisecond)
		engine.Pause()
		if engine.GetState() != StatePaused {
			t.Errorf("Expected StatePaused, got %s", engine.GetState())
		}
		// Resume after 150ms pause
		time.Sleep(150 * time.Millisecond)
		engine.Resume()
	}()

	err := engine.Start(context.Background(), profile, false, nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 300*time.Millisecond {
		t.Errorf("Expected total elapsed time >= 300ms with pause, got %v", elapsed)
	}
}

func TestTC_QUE_04_Cancellation(t *testing.T) {
	mock := &MockSpawner{}
	engine := NewEngine(mock)

	profile := models.Profile{
		Name: "TestProfile",
		Items: []models.LaunchItem{
			{ID: "item1", Name: "Item 1", Path: "/bin/echo", DelaySeconds: 1.0, Enabled: true},
			{ID: "item2", Name: "Item 2", Path: "/bin/echo", DelaySeconds: 1.0, Enabled: true},
		},
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		engine.Cancel()
	}()

	_ = engine.Start(context.Background(), profile, false, nil)

	if engine.GetState() != StateCancelled {
		t.Errorf("Expected StateCancelled, got %s", engine.GetState())
	}
	if len(mock.spawnedItems) > 1 {
		t.Errorf("Item 2 should not have spawned after cancel, got %d spawned items", len(mock.spawnedItems))
	}
}

func TestTC_QUE_05_DryRunMode(t *testing.T) {
	mock := &MockSpawner{}
	engine := NewEngine(mock)

	profile := models.Profile{
		Name: "TestProfile",
		Items: []models.LaunchItem{
			{ID: "item1", Name: "Item 1", Path: "/bin/echo", DelaySeconds: 0.01, Enabled: true},
			{ID: "item2", Name: "Item 2", Path: "/bin/echo", DelaySeconds: 0.01, Enabled: true},
		},
	}

	err := engine.Start(context.Background(), profile, true, nil)
	if err != nil {
		t.Fatalf("Dry run failed: %v", err)
	}

	if len(mock.spawnedItems) != 0 {
		t.Errorf("Dry run must spawn zero processes, spawned: %d", len(mock.spawnedItems))
	}
}

func TestTC_QUE_06_ConcurrencyUnderRaceDetector(t *testing.T) {
	mock := &MockSpawner{}
	engine := NewEngine(mock)

	profile := models.Profile{
		Name: "TestProfile",
		Items: []models.LaunchItem{
			{ID: "item1", Name: "Item 1", Path: "/bin/echo", DelaySeconds: 0.05, Enabled: true},
			{ID: "item2", Name: "Item 2", Path: "/bin/echo", DelaySeconds: 0.05, Enabled: true},
		},
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_ = engine.Start(context.Background(), profile, false, nil)
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			time.Sleep(15 * time.Millisecond)
			engine.Pause()
			time.Sleep(10 * time.Millisecond)
			engine.Resume()
			_ = engine.GetState()
		}
	}()

	wg.Wait()
}

func TestTC_QUE_07_WaitForExit_BlockingMode(t *testing.T) {
	exitHeld := true
	mock := &MockSpawner{
		waitFuncs: map[string]func() error{
			"blocking-item": func() error {
				time.Sleep(100 * time.Millisecond)
				exitHeld = false
				return nil
			},
		},
	}
	engine := NewEngine(mock)

	profile := models.Profile{
		Name: "BlockingProfile",
		Items: []models.LaunchItem{
			{ID: "blocking-item", Name: "Blocking Item", Path: "/bin/sleep", WaitForExit: true, DelaySeconds: 0.01, Enabled: true},
			{ID: "next-item", Name: "Next Item", Path: "/bin/echo", DelaySeconds: 0.01, Enabled: true},
		},
	}

	waitingEventSeen := false
	err := engine.Start(context.Background(), profile, false, func(evt ProgressEvent) {
		if evt.CurrentItem.ID == "blocking-item" && evt.ItemStatus == "WaitingForExit" {
			waitingEventSeen = true
		}
	})

	if err != nil {
		t.Fatalf("Engine start failed: %v", err)
	}

	if !waitingEventSeen {
		t.Errorf("Expected 'WaitingForExit' status event")
	}
	if exitHeld {
		t.Errorf("Expected blocking wait to complete before queue finished")
	}
}

func TestTC_QUE_08_TCPPortProbe_Success(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port

	mock := &MockSpawner{}
	engine := NewEngine(mock)

	profile := models.Profile{
		Name: "ProbeProfile",
		Items: []models.LaunchItem{
			{
				ID:                  "probe-item",
				Name:                "Backend Service",
				Path:                "/bin/echo",
				TCPPortProbe:        port,
				ProbeTimeoutSeconds: 2.0,
				DelaySeconds:        0.01,
				Enabled:             true,
			},
		},
	}

	probingEventSeen := false
	err = engine.Start(context.Background(), profile, false, func(evt ProgressEvent) {
		if evt.CurrentItem.ID == "probe-item" && evt.ItemStatus == "ProbingPort" {
			probingEventSeen = true
		}
	})

	if err != nil {
		t.Fatalf("Engine start failed: %v", err)
	}
	if !probingEventSeen {
		t.Errorf("Expected ProbingPort event")
	}
}

func TestTC_QUE_09_TCPPortProbe_FailureContinuesWithWarning(t *testing.T) {
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close() // Port closed -> probe will time out

	mock := &MockSpawner{}
	engine := NewEngine(mock)

	profile := models.Profile{
		Name: "ProbeFailProfile",
		Items: []models.LaunchItem{
			{
				ID:                  "failing-probe",
				Name:                "Closed Port",
				Path:                "/bin/echo",
				TCPPortProbe:        port,
				ProbeTimeoutSeconds: 0.1, // Fast timeout
				DelaySeconds:        0.01,
				Enabled:             true,
			},
			{
				ID:           "subsequent-item",
				Name:         "Subsequent Item",
				Path:         "/bin/echo",
				DelaySeconds: 0.01,
				Enabled:      true,
			},
		},
	}

	failedEventRecorded := false
	err := engine.Start(context.Background(), profile, false, func(evt ProgressEvent) {
		if evt.CurrentItem.ID == "failing-probe" && evt.ItemStatus == "Failed" {
			failedEventRecorded = true
		}
	})

	if err != nil {
		t.Fatalf("Engine should not fail fatally: %v", err)
	}
	if !failedEventRecorded {
		t.Errorf("Expected Failed event on probe timeout")
	}
	// Verify subsequent item executed
	if len(mock.spawnedItems) != 2 || mock.spawnedItems[1] != "subsequent-item" {
		t.Errorf("Expected subsequent item to execute after probe failure, spawned: %+v", mock.spawnedItems)
	}
}
