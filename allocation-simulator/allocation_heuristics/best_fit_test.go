package heuristic_test

import (
	heuristic "allocation-simulator/allocation_heuristics"
	"allocation-simulator/input"
	"allocation-simulator/item"
	"allocation-simulator/logger"
	"allocation-simulator/operation"
	"allocation-simulator/simulator"
	"log/slog"
	"testing"
)

func TestBestFit_InsertThreeItems(t *testing.T) {
	logger.Init(slog.LevelDebug)
	var h = heuristic.NewBestFit()

	// input events
	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 9, Memory: 4}},
		{Time: 1, Operation: operation.Insertion, Item: item.Item{ID: "1", CPU: 10, Memory: 5}},
		{Time: 2, Operation: operation.Insertion, Item: item.Item{ID: "2", CPU: 6, Memory: 2}},
	}

	var (
		CPU_Capacity          int64   = 16
		Memory_Capacity       int64   = 16
		MinOccupationLimit    float64 = 0
		RepackingEnabled      bool    = false
		EventBasedRepacking   bool    = false
		RunOptimal            bool    = false
	)
	var sim = simulator.NewSimulator(events, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, EventBasedRepacking, RunOptimal)

	// inserting first item
	var err = h.HandleEvent(sim, 0)
	var usedID = sim.ItemToBinMap["0"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 0 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=0", usedID)
	}

	// inserting second item
	err = h.HandleEvent(sim, 1)
	usedID = sim.ItemToBinMap["1"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 1 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=1", usedID)
	}

	// inserting third item
	err = h.HandleEvent(sim, 2)
	usedID = sim.ItemToBinMap["2"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 1 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=1", usedID)
	}
}

func TestBestFit_ItemExceedsCapacity(t *testing.T) {
	logger.Init(slog.LevelDebug)
	var h = heuristic.NewBestFit()

	// input events
	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 17, Memory: 2}},
	}

	var (
		CPU_Capacity          int64   = 16
		Memory_Capacity       int64   = 16
		MinOccupationLimit    float64 = 0
		RepackingEnabled      bool    = false
		EventBasedRepacking   bool    = false
		RunOptimal            bool    = false
	)
	var sim = simulator.NewSimulator(events, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, EventBasedRepacking, RunOptimal)

	// inserting item
	var err = h.HandleEvent(sim, 0)
	if err == nil {
		t.Fatalf("expected error for exceeding capacity: %v", err)
	}
}

func TestBestFit_InsertElementWhenNoBins(t *testing.T) {
	logger.Init(slog.LevelDebug)
	var h = heuristic.NewBestFit()

	// input events
	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 2, Memory: 2}},
	}

	var (
		CPU_Capacity          int64   = 16
		Memory_Capacity       int64   = 16
		MinOccupationLimit    float64 = 0
		RepackingEnabled      bool    = false
		EventBasedRepacking   bool    = false
		RunOptimal            bool    = false
	)
	var sim = simulator.NewSimulator(events, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, EventBasedRepacking, RunOptimal)

	// inserting item
	var err = h.HandleEvent(sim, 0)
	var usedID = sim.ItemToBinMap["0"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 0 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=0", usedID)
	}
}

func TestBestFit_ExactFitAtBoundary(t *testing.T) {
	logger.Init(slog.LevelDebug)
	var h = heuristic.NewBestFit()

	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 10, Memory: 8}},
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "1", CPU: 8, Memory: 8}},
		{Time: 1, Operation: operation.Insertion, Item: item.Item{ID: "2", CPU: 8, Memory: 8}},
	}
	var (
		CPU_Capacity          int64   = 16
		Memory_Capacity       int64   = 16
		MinOccupationLimit    float64 = 0
		RepackingEnabled      bool    = false
		EventBasedRepacking   bool    = false
		RunOptimal            bool    = false
	)
	var sim = simulator.NewSimulator(events, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, EventBasedRepacking, RunOptimal)

	var err = h.HandleEvent(sim, 0)
	var usedID = sim.ItemToBinMap["0"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 0 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=0", usedID)
	}

	err = h.HandleEvent(sim, 1)
	usedID = sim.ItemToBinMap["1"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 1 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=1", usedID)
	}

	err = h.HandleEvent(sim, 2)
	usedID = sim.ItemToBinMap["2"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 1 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=1", usedID)
	}
}

func TestBestFit_DeletionClosesEmptyBin(t *testing.T) {
	logger.Init(slog.LevelDebug)
	var h = heuristic.NewBestFit()

	// input events
	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 2, Memory: 2}},
		{Time: 0, Operation: operation.Deletion, Item: item.Item{ID: "0", CPU: 2, Memory: 2}},
	}

	var (
		CPU_Capacity          int64   = 16
		Memory_Capacity       int64   = 16
		MinOccupationLimit    float64 = 0
		RepackingEnabled      bool    = false
		EventBasedRepacking   bool    = false
		RunOptimal            bool    = false
	)
	var sim = simulator.NewSimulator(events, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, EventBasedRepacking, RunOptimal)

	// inserting item
	var err = h.HandleEvent(sim, 0)
	var usedID = sim.ItemToBinMap["0"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 0 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=0", usedID)
	}

	// removing item
	err = h.HandleEvent(sim, 1)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if !sim.Bins[0].Closed {
		t.Fatalf("Empty bin should have been closed")
	}
}

func TestBestFit_DeletionItemNotFound_ReturnsError(t *testing.T) {
	logger.Init(slog.LevelDebug)
	var h = heuristic.NewBestFit()

	// input events: deletion of a non-existent item
	var events = []input.Event{
		{Time: 1, Operation: operation.Deletion, Item: item.Item{ID: "0", CPU: 1, Memory: 1}},
	}
	var (
		CPU_Capacity          int64   = 16
		Memory_Capacity       int64   = 16
		MinOccupationLimit    float64 = 0
		RepackingEnabled      bool    = false
		EventBasedRepacking   bool    = false
		RunOptimal            bool    = false
	)
	var sim = simulator.NewSimulator(events, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, EventBasedRepacking, RunOptimal)

	var err = h.HandleEvent(sim, 0)
	if err == nil {
		t.Fatalf("expected error for unmapped item")
	}
}

func TestBestFit_SkipsClosedBin(t *testing.T) {
	logger.Init(slog.LevelDebug)
	var h = heuristic.NewBestFit()

	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 2, Memory: 2}},
		{Time: 1, Operation: operation.Deletion, Item: item.Item{ID: "0", CPU: 2, Memory: 2}},
		{Time: 2, Operation: operation.Insertion, Item: item.Item{ID: "1", CPU: 2, Memory: 2}},
	}
	var (
		CPU_Capacity          int64   = 16
		Memory_Capacity       int64   = 16
		MinOccupationLimit    float64 = 0
		RepackingEnabled      bool    = false
		EventBasedRepacking   bool    = false
		RunOptimal            bool    = false
	)
	var sim = simulator.NewSimulator(events, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, EventBasedRepacking, RunOptimal)

	// step 0
	var err = h.HandleEvent(sim, 0)
	var usedID = sim.ItemToBinMap["0"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 0 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=0", usedID)
	}

	// step 1 -> closes b0
	err = h.HandleEvent(sim, 1)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if !sim.Bins[0].Closed {
		t.Fatalf("bin 0 should be closed after deletion")
	}

	// step 2 -> must use new bin (b1)
	err = h.HandleEvent(sim, 2)
	usedID = sim.ItemToBinMap["1"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 1 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=1", usedID)
	}
}

func TestBestFit_DeletionDoesNotCloseNonEmptyBin(t *testing.T) {
	logger.Init(slog.LevelDebug)
	var h = heuristic.NewBestFit()

	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 4, Memory: 4}},
		{Time: 1, Operation: operation.Insertion, Item: item.Item{ID: "1", CPU: 4, Memory: 4}},
		{Time: 2, Operation: operation.Deletion, Item: item.Item{ID: "0", CPU: 4, Memory: 4}},
	}
	var (
		CPU_Capacity          int64   = 16
		Memory_Capacity       int64   = 16
		MinOccupationLimit    float64 = 0
		RepackingEnabled      bool    = false
		EventBasedRepacking   bool    = false
		RunOptimal            bool    = false
	)
	var sim = simulator.NewSimulator(events, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, EventBasedRepacking, RunOptimal)

	var err = h.HandleEvent(sim, 0)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}

	err = h.HandleEvent(sim, 1)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}

	err = h.HandleEvent(sim, 2)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}

	if sim.Bins[0].Closed {
		t.Fatalf("bin should remain open; it still has item 2")
	}
}

// Single complex scenario exercising Best-Fit decisions
func TestBestFit_InsertDeleteSequence(t *testing.T) {
	logger.Init(slog.LevelDebug)
	var h = heuristic.NewBestFit()

	var events = []input.Event{
		// 0) Insert item 0 into a new bin b0.
		//    b0 usage becomes (6,6); leftovers (10,10).
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 6, Memory: 6}},

		// 1) Insert item 1; it does not fit in b0 (CPU 10 < 11), so create b1.
		//    b1 usage becomes (11,1); leftovers (5,15).
		{Time: 1, Operation: operation.Insertion, Item: item.Item{ID: "1", CPU: 11, Memory: 1}},

		// 2) Insert item 2; it fits both b0 and b1.
		//    Best-Fit compares CPU leftovers after placement:
		//      - in b0: 10-5 = 5
		//      - in b1:  5-5 = 0  -> choose b1 (lower CPU leftover).
		//    b1 usage becomes (16,7); leftovers (0,9).
		{Time: 2, Operation: operation.Insertion, Item: item.Item{ID: "2", CPU: 5, Memory: 6}},

		// 3) Insert item 3; it fits neither b0 nor b1, so create b2.
		//    b2 usage becomes (6,15); leftovers (10,1).
		{Time: 3, Operation: operation.Insertion, Item: item.Item{ID: "3", CPU: 6, Memory: 15}},

		// 4) Insert item 4; it fits b0 and b2 and yields a tie on CPU leftover.
		//    After placement:
		//      - in b0: CPU leftover 10-4 = 6, MEM leftover 10-1 = 9
		//      - in b2: CPU leftover 10-4 = 6, MEM leftover  1-1 = 0 -> tie on CPU; choose smaller MEM leftover -> b2.
		//    b2 usage becomes (10,16); leftovers (6,0).
		{Time: 4, Operation: operation.Insertion, Item: item.Item{ID: "4", CPU: 4, Memory: 1}},

		// 5) Delete item 1 from b1.
		//    b1 usage goes from (16,7) to (5,6); leftovers become (11,10).
		{Time: 5, Operation: operation.Deletion, Item: item.Item{ID: "1", CPU: 11, Memory: 1}},

		// 6) Insert item 5; it fits b0 and b1.
		//    Compare CPU leftovers after placement:
		//      - in b0: 10-10 = 0
		//      - in b1: 11-10 = 1 -> choose b0.
		{Time: 6, Operation: operation.Insertion, Item: item.Item{ID: "5", CPU: 10, Memory: 3}},

		// 7) Delete item 2 from b1.
		//    b1 usage goes from (5,6) to (0,0) and must be closed.
		{Time: 7, Operation: operation.Deletion, Item: item.Item{ID: "2", CPU: 5, Memory: 6}},
	}

	var (
		CPU_Capacity          int64   = 16
		Memory_Capacity       int64   = 16
		MinOccupationLimit    float64 = 0
		RepackingEnabled      bool    = false
		EventBasedRepacking   bool    = false
		RunOptimal            bool    = false
	)
	var sim = simulator.NewSimulator(events, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, EventBasedRepacking, RunOptimal)

	// 0) insert item 0
	var err = h.HandleEvent(sim, 0)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	var usedID = sim.ItemToBinMap["0"]
	if usedID != 0 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=0", usedID)
	}

	// 1) insert item 1
	err = h.HandleEvent(sim, 1)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	usedID = sim.ItemToBinMap["1"]
	if usedID != 1 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=1", usedID)
	}

	// 2) insert item 2 (Best-Fit prefers b1 due to lower CPU leftover)
	err = h.HandleEvent(sim, 2)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	usedID = sim.ItemToBinMap["2"]
	if usedID != 1 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=1", usedID)
	}

	// 3) insert item 3 (creates b2)
	err = h.HandleEvent(sim, 3)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	usedID = sim.ItemToBinMap["3"]
	if usedID != 2 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=2", usedID)
	}

	// 4) insert item 4 (tie on CPU; break on MEM -> choose b2)
	err = h.HandleEvent(sim, 4)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	usedID = sim.ItemToBinMap["4"]
	if usedID != 2 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=2", usedID)
	}

	// 5) delete item 1 from b1
	err = h.HandleEvent(sim, 5)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	var ok bool
	_, ok = sim.ItemToBinMap["1"]
	if ok {
		t.Fatalf("Item should have been removed from map")
	}

	// 6) insert item 5 (Best-Fit chooses b0)
	err = h.HandleEvent(sim, 6)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	usedID = sim.ItemToBinMap["5"]
	if usedID != 0 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=0", usedID)
	}

	// 7) delete item 2 from b1 (bin should close)
	err = h.HandleEvent(sim, 7)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	_, ok = sim.ItemToBinMap["2"]
	if ok {
		t.Fatalf("Item should have been removed from map")
	}
	if !sim.Bins[1].Closed {
		t.Fatalf("Bin should have been closed")
	}

	// final sanity check: b0 and b2 are in use; b1 was closed
	var got = sim.NumberOfBinsUsed()
	if got != 2 {
		t.Fatalf("Incorrect number of bins used: got=%d want=2", got)
	}
}
