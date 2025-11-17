package simulator_test

import (
	heuristic "allocation-simulator/allocation_heuristics"
	"allocation-simulator/input"
	"allocation-simulator/item"
	"allocation-simulator/logger"
	"allocation-simulator/operation"
	repacking "allocation-simulator/repacking_heuristics"
	"allocation-simulator/simulator"
	"log/slog"
	"testing"
)

// Integration tests for the Simulator's orchestration logic
//
// Intent
//   These tests validate that the Simulator correctly orchestrates heuristics
//   and repacking according to the selected modes and preconditions.
//   The focus is on end-to-end behavior of RunSimulation rather than unit-testing the internal
//   details of allocation or repacking strategies.

type fakeHeuristic struct {
	called bool
}

// HandleEvent (fake): mark that it was invoked.
func (f *fakeHeuristic) HandleEvent(s *simulator.Simulator, eventIndex int) error {
	f.called = true
	return nil
}

// RunHeuristic is unused in these tests but required by the interface.
func (f *fakeHeuristic) RunHeuristic(s *simulator.Simulator, it item.Item) int {
	return -1
}

// guarantees optimal heuristic handles everything and no other heuristic is called
func TestRunSimulation_Optimal_SkipsInsertionHeuristic(t *testing.T) {
	// Initialize logger at a verbose level for debugging
	logger.Init(slog.LevelDebug)

	// Arrange test events
	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "a", CPU: 8, Memory: 8}},
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "b", CPU: 8, Memory: 8}},
	}

	// Select repacking = Optimal; insertion heuristic will be ignored by the simulator in this mode
	var h simulator.Heuristic = &fakeHeuristic{}
	var r simulator.Repacking = repacking.NewOptimal()

	// Build simulator with explicit var block (as in your style)
	var (
		CPU_Capacity        int64   = 16
		Memory_Capacity     int64   = 16
		MinOccupationLimit  float64 = 0.70
		RepackingEnabled    bool    = true
		EventBasedRepacking bool    = true
		RunOptimal          bool    = true
	)
	var sim = simulator.NewSimulator(
		events,
		CPU_Capacity,
		Memory_Capacity,
		MinOccupationLimit,
		RepackingEnabled,
		EventBasedRepacking,
		RunOptimal,
	)

	// Act
	var err error = sim.RunSimulation(h, r)

	// Assert
	if err != nil {
		t.Fatalf("RunSimulation (optimal) returned error: %v", err)
	}
	if fh, ok := h.(*fakeHeuristic); ok && fh.called {
		t.Fatalf("Insertion heuristic must NOT be called in optimal mode")
	}
}

// guarantees optimal fit only runs when repacking is enabled
func TestRunSimulation_Optimal_FailsIfRepackingDisabled(t *testing.T) {
	logger.Init(slog.LevelDebug)

	// Minimal event set
	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "x", CPU: 4, Memory: 4}},
	}

	// Use any heuristic (won't be called in optimal mode)
	var h simulator.Heuristic = &fakeHeuristic{}
	var r simulator.Repacking = repacking.NewOptimal()

	// Build simulator with repacking disabled (violates Optimal requirements)
	var (
		CPU_Capacity        int64   = 16
		Memory_Capacity     int64   = 16
		MinOccupationLimit  float64 = 0.70
		RepackingEnabled    bool    = false // <- invalid for optimal
		EventBasedRepacking bool    = true
		RunOptimal          bool    = true
	)
	var sim = simulator.NewSimulator(
		events,
		CPU_Capacity,
		Memory_Capacity,
		MinOccupationLimit,
		RepackingEnabled,
		EventBasedRepacking,
		RunOptimal,
	)

	var err error = sim.RunSimulation(h, r)
	if err == nil {
		t.Fatalf("expected error when RunOptimal=true but RepackingEnabled=false")
	}
}

// guarantees optimal fit only runs when event-based repacking is enabled
func TestRunSimulation_Optimal_FailsIfNotEventBased(t *testing.T) {
	logger.Init(slog.LevelDebug)

	// Minimal event set
	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "x", CPU: 4, Memory: 4}},
	}

	var h simulator.Heuristic = &fakeHeuristic{}
	var r simulator.Repacking = repacking.NewOptimal()

	// Build simulator with timestamp-based cadence (violates Optimal requirements)
	var (
		CPU_Capacity        int64   = 16
		Memory_Capacity     int64   = 16
		MinOccupationLimit  float64 = 0.70
		RepackingEnabled    bool    = true
		EventBasedRepacking bool    = false // <- invalid for optimal
		RunOptimal          bool    = true
	)
	var sim = simulator.NewSimulator(
		events,
		CPU_Capacity,
		Memory_Capacity,
		MinOccupationLimit,
		RepackingEnabled,
		EventBasedRepacking,
		RunOptimal,
	)

	var err error = sim.RunSimulation(h, r)
	if err == nil {
		t.Fatalf("expected error when RunOptimal=true but EventBasedRepacking=false")
	}
}

// guarantees RunSimulation returns error when there are no events
func TestRunSimulation_NoEvents_ReturnsError(t *testing.T) {
	logger.Init(slog.LevelDebug)

	// No events
	var events []input.Event

	var h simulator.Heuristic = &fakeHeuristic{}
	var r simulator.Repacking = repacking.NewR1()

	var (
		CPU_Capacity        int64   = 16
		Memory_Capacity     int64   = 16
		MinOccupationLimit  float64 = 0.70
		RepackingEnabled    bool    = true
		EventBasedRepacking bool    = true
		RunOptimal          bool    = false
	)
	var sim = simulator.NewSimulator(
		events,
		CPU_Capacity,
		Memory_Capacity,
		MinOccupationLimit,
		RepackingEnabled,
		EventBasedRepacking,
		RunOptimal,
	)

	var err error = sim.RunSimulation(h, r)
	if err == nil {
		t.Fatalf("expected error when events slice is empty")
	}
}

// compares disabled repacking vs enabled event-based repacking on the same sequence
func TestRunSimulation_RepackingEnabledVsDisabled_EventBased_Comparison(t *testing.T) {
	logger.Init(slog.LevelDebug)

	// Construct a sequence where event-based repacking can consolidate into fewer bins.
	// Capacity 16, limit 0.7 -> items of 9 fit two-per-bin only after moves.
	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "a", CPU: 9, Memory: 9}}, // b0
		{Time: 1, Operation: operation.Insertion, Item: item.Item{ID: "b", CPU: 7, Memory: 7}}, // b0 (now full)
		{Time: 2, Operation: operation.Insertion, Item: item.Item{ID: "c", CPU: 7, Memory: 7}}, // new bin
		{Time: 3, Operation: operation.Deletion, Item: item.Item{ID: "b", CPU: 7, Memory: 7}}, // free underutilized space on b0
	}

	// Heuristic: use a simple one; fake is fine for integration
	var h simulator.Heuristic = heuristic.NewFirstFit()
	var r simulator.Repacking = repacking.NewR1()

	// A) Repacking disabled
	var (
		CPU_Capacity_A        int64   = 16
		Memory_Capacity_A     int64   = 16
		MinOccupationLimit_A  float64 = 0.70
		RepackingEnabled_A    bool    = false
		EventBasedRepacking_A bool    = false
		RunOptimal_A          bool    = false
	)
	var simA = simulator.NewSimulator(
		events,
		CPU_Capacity_A,
		Memory_Capacity_A,
		MinOccupationLimit_A,
		RepackingEnabled_A,
		EventBasedRepacking_A,
		RunOptimal_A,
	)
	var err error = simA.RunSimulation(h, r)
	if err != nil {
		t.Fatalf("RunSimulation (A) returned error: %v", err)
	}
	var binsA = simA.NumberOfBinsUsed()

	// B) Repacking enabled (event-based)
	var (
		CPU_Capacity_B        int64   = 16
		Memory_Capacity_B     int64   = 16
		MinOccupationLimit_B  float64 = 0.70
		RepackingEnabled_B    bool    = true
		EventBasedRepacking_B bool    = true
		RunOptimal_B          bool    = false
	)
	var simB = simulator.NewSimulator(
		events,
		CPU_Capacity_B,
		Memory_Capacity_B,
		MinOccupationLimit_B,
		RepackingEnabled_B,
		EventBasedRepacking_B,
		RunOptimal_B,
	)
	err = simB.RunSimulation(h, r)
	if err != nil {
		t.Fatalf("RunSimulation (B) returned error: %v", err)
	}
	var binsB = simB.NumberOfBinsUsed()

	// Expect that enabling repacking does not increase fragmentation; it should be <=.
	if binsB >= binsA {
		t.Fatalf("repacking should not increase bin count: disabled=%d enabled=%d", binsA, binsB)
	}
}


// Proves event-based repacking actually happens within the same timestamp.
// Construction (capacity 16, limit 0.55):
// - At t=10, fill bin0 fully with A(9,9) + E(7,7).
// - Insert B(7,7) -> goes to bin1.
// - Insert C(4,4) -> goes to bin1 (bin1 usage becomes 11/16 ~= 0.6875 >= 0.55).
// - Delete E -> bin0 usage becomes 9/16 ~= 0.5625 > 0.55 (not underutilized).
// - Delete C -> bin1 usage becomes 7/16 = 0.4375 < 0.55 (underutilized).
//   - Event-based: repacking runs now and moves B to bin0 (fits exactly).
//
// - Still t=10: Insert D(7,7).
//   - Event-based: bin0 is full (A+B), so D cannot go there and is placed elsewhere.
//   - Timestamp-based: repacking has not run yet; bin0 has space 7, so D goes into bin0.
//
// At the boundary, timestamp-based repacking cannot move B (bin0 is full), so the final
// allocation differs in a way that only event-based mode produces.
func TestRepacking_EventBased_RunSimulation_RepackRequired(t *testing.T) {
	logger.Init(slog.LevelDebug)

	var h = heuristic.NewFirstFit()
	var r = repacking.NewR1()

	var events = []input.Event{
		{Time: 10, Operation: operation.Insertion, Item: item.Item{ID: "a", CPU: 9, Memory: 9}}, // bin0
		{Time: 10, Operation: operation.Insertion, Item: item.Item{ID: "e", CPU: 7, Memory: 7}}, // bin0 full (16)
		{Time: 10, Operation: operation.Insertion, Item: item.Item{ID: "b", CPU: 7, Memory: 7}}, // bin1
		{Time: 10, Operation: operation.Insertion, Item: item.Item{ID: "c", CPU: 4, Memory: 4}}, // bin1 (usage 11/16 >= 0.55)
		{Time: 10, Operation: operation.Deletion, Item: item.Item{ID: "e", CPU: 7, Memory: 7}},  // bin0 usage 9/16 > 0.55
		{Time: 10, Operation: operation.Deletion, Item: item.Item{ID: "c", CPU: 4, Memory: 4}},  // bin1 usage 7/16 < 0.55 -> underutilized
		{Time: 10, Operation: operation.Insertion, Item: item.Item{ID: "d", CPU: 7, Memory: 7}}, // allocation differs by mode
	}

	var (
		CPU_Capacity        int64   = 16
		Memory_Capacity     int64   = 16
		MinOccupationLimit  float64 = 0.55
		RepackingEnabled    bool    = true
		RunOptimal          bool    = false
	)

	// Event-based repacking
	var simEvent = simulator.NewSimulator(events, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, true, RunOptimal)
	var err = simEvent.RunSimulation(h, r)
	if err != nil {
		t.Fatalf("RunSimulation (event-based) returned error: %v", err)
	}

	// Timestamp-based repacking
	var simTime = simulator.NewSimulator(events, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, false, RunOptimal)
	err = simTime.RunSimulation(h, r)
	if err != nil {
		t.Fatalf("RunSimulation (timestamp-based) returned error: %v", err)
	}

	// Event-based: B must be colocated with A.
	var aBinEvent = simEvent.ItemToBinMap["a"]
	var bBinEvent = simEvent.ItemToBinMap["b"]
	var dBinEvent = simEvent.ItemToBinMap["d"]
	if aBinEvent != bBinEvent {
		t.Fatalf("event-based: expected a and b in the same bin, got a->%d b->%d", aBinEvent, bBinEvent)
	}
	if dBinEvent == aBinEvent {
		t.Fatalf("event-based: expected d NOT in a's bin, got a->%d d->%d", aBinEvent, dBinEvent)
	}

	// Timestamp-based: D must be colocated with A (took the free space before repacking ran).
	var aBinTime = simTime.ItemToBinMap["a"]
	var bBinTime = simTime.ItemToBinMap["b"]
	var dBinTime = simTime.ItemToBinMap["d"]
	if aBinTime != dBinTime {
		t.Fatalf("timestamp-based: expected a and d in the same bin, got a->%d d->%d", aBinTime, dBinTime)
	}
	if bBinTime == aBinTime {
		t.Fatalf("timestamp-based: expected b NOT in a's bin, got a->%d b->%d", aBinTime, bBinTime)
	}

	// Both modes end up with two bins in use; the difference is the allocation caused by event-based repacking.
	if simEvent.NumberOfBinsUsed() != 2 || simTime.NumberOfBinsUsed() != 2 {
		t.Fatalf("expected 2 bins in use in both modes, got event=%d time=%d", simEvent.NumberOfBinsUsed(), simTime.NumberOfBinsUsed())
	}
}
