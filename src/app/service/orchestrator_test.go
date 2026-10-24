package service

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

// fakePID is never a real process; fake kills only record it
const fakePID = 4242424

// newTestOrchestrator returns an orchestrator with a temp state dir and fake
// process operations; kills are recorded in the returned slice
func newTestOrchestrator(t *testing.T) (*Orchestrator, *[]int) {
	t.Helper()
	sm, err := newStateManagerAt(t.TempDir())
	if err != nil {
		t.Fatalf("state manager: %v", err)
	}
	o := newOrchestrator(sm)
	killed := &[]int{}
	o.spawn = func(string, string, []string, int) (int, error) { return fakePID, nil }
	o.kill = func(pid int) { *killed = append(*killed, pid) }
	return o, killed
}

func readTestState(t *testing.T, o *Orchestrator) *ProcessState {
	t.Helper()
	state, err := o.stateManager.loadState()
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	return state
}

func TestStartService_RecordsStartingEntryBeforeHealthWait(t *testing.T) {
	o, _ := newTestOrchestrator(t)

	var seen *Instance
	o.waitHealthy = func(ctx context.Context, port int) ([]string, error) {
		// Read the file as another process would see it after a SIGKILL
		for _, inst := range readTestState(t, o).Instances {
			inst := inst
			seen = &inst
		}
		return []string{"test.Service"}, nil
	}

	result, err := o.StartService(context.Background(), "demo", "/bin/demo", "", nil, ModeShared)
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	if seen == nil || seen.PID != fakePID || seen.Status != StatusStarting {
		t.Fatalf("expected a starting entry for pid %d during the wait, got %+v", fakePID, seen)
	}
	if len(seen.Sessions) != 0 {
		t.Errorf("starting entry must have no session, got %d", len(seen.Sessions))
	}

	inst := readTestState(t, o).Instances[InstanceKey("demo", result.Port)]
	if inst.Status != "" || len(inst.Sessions) != 1 || inst.Sessions[0].ID != result.SessionID {
		t.Errorf("expected a ready entry with the returned session, got %+v", inst)
	}
	if !result.NewProcess || len(result.Services) != 1 {
		t.Errorf("unexpected result %+v", result)
	}
}

func TestStartService_InterruptedWaitStopsProcess(t *testing.T) {
	o, killed := newTestOrchestrator(t)

	ctx, cancel := context.WithCancel(context.Background())
	o.waitHealthy = func(ctx context.Context, port int) ([]string, error) {
		cancel() // SIGTERM arrives while the plugin starts
		<-ctx.Done()
		return nil, ctx.Err()
	}

	if _, err := o.StartService(ctx, "demo", "/bin/demo", "", nil, ModeShared); err == nil {
		t.Fatal("expected an error for an interrupted start")
	}

	if len(*killed) != 1 || (*killed)[0] != fakePID {
		t.Errorf("expected the spawned process to be stopped, kills: %v", *killed)
	}
	if n := len(readTestState(t, o).Instances); n != 0 {
		t.Errorf("expected no entries after an interrupted start, got %d", n)
	}
}

func TestStartService_HealthFailureStopsProcess(t *testing.T) {
	o, killed := newTestOrchestrator(t)
	o.waitHealthy = func(context.Context, int) ([]string, error) {
		return nil, errors.New("health check timeout")
	}

	if _, err := o.StartService(context.Background(), "demo", "/bin/demo", "", nil, ModeShared); err == nil {
		t.Fatal("expected an error for a failed health check")
	}

	if len(*killed) != 1 {
		t.Errorf("expected the spawned process to be stopped, kills: %v", *killed)
	}
	if n := len(readTestState(t, o).Instances); n != 0 {
		t.Errorf("expected no entries after a failed start, got %d", n)
	}
}

func TestStartService_CancelledBeforeLockDoesNotSpawn(t *testing.T) {
	o, _ := newTestOrchestrator(t)
	spawned := false
	o.spawn = func(string, string, []string, int) (int, error) {
		spawned = true
		return fakePID, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := o.StartService(ctx, "demo", "/bin/demo", "", nil, ModeShared); err == nil {
		t.Fatal("expected an error for a cancelled start")
	}
	if spawned {
		t.Error("a cancelled start must not spawn a process")
	}
}

func TestCleanStaleEntries_StopsLeftoverStartingEntry(t *testing.T) {
	o, killed := newTestOrchestrator(t)

	// The test process stands in for a live plugin left by a killed start
	pid := os.Getpid()
	key := InstanceKey("demo", 50100)
	state := &ProcessState{Instances: map[string]Instance{
		key: {PluginName: "demo", PID: pid, Port: 50100, Mode: string(ModeShared), Status: StatusStarting},
	}}

	o.cleanStaleEntries(state)

	if _, exists := state.Instances[key]; exists {
		t.Error("expected the starting entry to be removed")
	}
	if len(*killed) != 1 || (*killed)[0] != pid {
		t.Errorf("expected the leftover process to be stopped, kills: %v", *killed)
	}
}

func TestStopAll_StopsStartingEntries(t *testing.T) {
	o, killed := newTestOrchestrator(t)

	err := o.stateManager.WithLock(func(state *ProcessState) error {
		state.Instances[InstanceKey("demo", 50100)] = Instance{
			PluginName: "demo", PID: fakePID, Port: 50100, Status: StatusStarting,
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed state: %v", err)
	}

	count, err := o.StopAll()
	if err != nil || count != 1 || len(*killed) != 1 {
		t.Errorf("expected one stopped instance, got count=%d kills=%v err=%v", count, *killed, err)
	}
}

func TestWaitForHealthy_ReturnsOnCancel(t *testing.T) {
	sm, err := newStateManagerAt(t.TempDir())
	if err != nil {
		t.Fatalf("state manager: %v", err)
	}
	o := newOrchestrator(sm)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	// Nothing listens on this port; each health probe fails
	if _, err := o.waitForHealthy(ctx, 1); err == nil {
		t.Fatal("expected an error")
	}
	if elapsed := time.Since(start); elapsed >= o.healthTimeout {
		t.Errorf("expected an early return on cancel, took %s", elapsed)
	}
}
