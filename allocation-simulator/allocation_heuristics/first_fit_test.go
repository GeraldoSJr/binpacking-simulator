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

func TestFirstFit_InsertThreeItems(t *testing.T) {
	logger.Init(slog.LevelDebug)

	var h = heuristic.NewFirstFit()

	// input events
	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 13, Memory: 7}},
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

func TestFirstFit_ItemExceedsCapacity(t *testing.T) {
	logger.Init(slog.LevelDebug)

	var h = heuristic.NewFirstFit()

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

func TestFirstFit_InsertElementWhenNoBins(t *testing.T) {
	logger.Init(slog.LevelDebug)

	var h = heuristic.NewFirstFit()

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

func TestFirstFit_ExactFitAtBoundary(t *testing.T) {
	logger.Init(slog.LevelDebug)

	var h = heuristic.NewFirstFit()
	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 8, Memory: 8}},
		{Time: 1, Operation: operation.Insertion, Item: item.Item{ID: "1", CPU: 8, Memory: 8}},
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
	if usedID != 0 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=0", usedID)
	}
}

func TestFirstFit_DeletionClosesEmptyBin(t *testing.T) {
	logger.Init(slog.LevelDebug)

	var h = heuristic.NewFirstFit()

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

func TestFirstFit_DeletionItemNotFound_ReturnsError(t *testing.T) {
	logger.Init(slog.LevelDebug)

	var h = heuristic.NewFirstFit()

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

func TestFirstFit_SkipsClosedBin(t *testing.T) {
	logger.Init(slog.LevelDebug)

	var h = heuristic.NewFirstFit()

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

	// step 2 -> must use new bin
	err = h.HandleEvent(sim, 2)
	usedID = sim.ItemToBinMap["1"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 1 {
		t.Fatalf("Incorrect bin was chosen: got=%d want=1", usedID)
	}
}

func TestFirstFit_DeletionDoesNotCloseNonEmptyBin(t *testing.T) {
	logger.Init(slog.LevelDebug)

	var h = heuristic.NewFirstFit()
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
		t.Fatalf("bin should remain open; it still has item 1")
	}
}

func TestFirstFit_InsertDeleteSequence(t *testing.T) {
	logger.Init(slog.LevelDebug)

	var h = heuristic.NewFirstFit()

	var events = []input.Event{
		// Step 0: create b0 with item 0
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 8, Memory: 6}},
		// Step 1: fits on b0 -> still b0
		{Time: 1, Operation: operation.Insertion, Item: item.Item{ID: "1", CPU: 5, Memory: 5}},
		// Step 2: doesn't fit on b0 -> create b1
		{Time: 2, Operation: operation.Insertion, Item: item.Item{ID: "2", CPU: 4, Memory: 6}},
		// Step 3: fits b0 and b1 -> first-fit picks b0
		{Time: 3, Operation: operation.Insertion, Item: item.Item{ID: "3", CPU: 2, Memory: 2}},
		// Step 4: delete item 1 (free space on b0)
		{Time: 4, Operation: operation.Deletion, Item: item.Item{ID: "1", CPU: 5, Memory: 5}},
		// Step 5: now item 4 should fit on b0
		{Time: 5, Operation: operation.Insertion, Item: item.Item{ID: "4", CPU: 6, Memory: 6}},
		// Step 6: delete item 0; b0 remains with 3 and 4
		{Time: 6, Operation: operation.Deletion, Item: item.Item{ID: "0", CPU: 8, Memory: 6}},
		// Step 7: item 5 won't fit on b0, should go to b1
		{Time: 7, Operation: operation.Insertion, Item: item.Item{ID: "5", CPU: 10, Memory: 3}},
		// Step 8: item 6 fits exactly to fill b0 to (16,16)
		{Time: 8, Operation: operation.Insertion, Item: item.Item{ID: "6", CPU: 8, Memory: 8}},
		// Step 9: delete item 3, b0 goes down to (14,14)
		{Time: 9, Operation: operation.Deletion, Item: item.Item{ID: "3", CPU: 2, Memory: 2}},
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

	// Step 0: insert 0 -> b0
	var err = h.HandleEvent(sim, 0)
	var usedID = sim.ItemToBinMap["0"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 0 {
		t.Fatalf("step 0 wrong bin: got=%d want=0", usedID)
	}

	// Step 1: insert 1 -> still b0
	err = h.HandleEvent(sim, 1)
	usedID = sim.ItemToBinMap["1"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 0 {
		t.Fatalf("step 1 wrong bin: got=%d want=0", usedID)
	}

	// Step 2: insert 2 -> doesn't fit on b0 -> b1
	err = h.HandleEvent(sim, 2)
	usedID = sim.ItemToBinMap["2"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 1 {
		t.Fatalf("step 2 wrong bin: got=%d want=1", usedID)
	}

	// Step 3: insert 3 -> fits b0 and b1 -> first-fit picks b0
	err = h.HandleEvent(sim, 3)
	usedID = sim.ItemToBinMap["3"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 0 {
		t.Fatalf("step 3 wrong bin: got=%d want=0", usedID)
	}

	// Step 4: delete 1 from b0
	err = h.HandleEvent(sim, 4)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	var ok bool
	_, ok = sim.ItemToBinMap["1"]
	if ok {
		t.Fatalf("step 4 map should not contain 1 anymore")
	}

	// Step 5: insert 4 -> should fit b0 now
	err = h.HandleEvent(sim, 5)
	usedID = sim.ItemToBinMap["4"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 0 {
		t.Fatalf("step 5 wrong bin: got=%d want=0", usedID)
	}

	// Step 6: delete 0 from b0
	err = h.HandleEvent(sim, 6)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	_, ok = sim.ItemToBinMap["0"]
	if ok {
		t.Fatalf("step 6 map should not contain 0 anymore")
	}

	// Step 7: insert 5 -> should go to b1 (b0 would exceed CPU)
	err = h.HandleEvent(sim, 7)
	usedID = sim.ItemToBinMap["5"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 1 {
		t.Fatalf("step 7 wrong bin: got=%d want=1", usedID)
	}

	// Step 8: insert 6 -> exactly fills b0 to (16,16)
	err = h.HandleEvent(sim, 8)
	usedID = sim.ItemToBinMap["6"]
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	if usedID != 0 {
		t.Fatalf("step 8 wrong bin: got=%d want=0", usedID)
	}

	// Step 9: delete 3 from b0 -> b0 usage decreases
	err = h.HandleEvent(sim, 9)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	_, ok = sim.ItemToBinMap["3"]
	if ok {
		t.Fatalf("step 9 map should not contain 3 anymore")
	}

	// Final sanity check
	var got = sim.NumberOfBinsUsed()
	if got != 2 {
		t.Fatalf("final bins used: got=%d want=2", got)
	}
}
