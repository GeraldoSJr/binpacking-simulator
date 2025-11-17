package repacking_test

import (
	heuristic "allocation-simulator/allocation_heuristics"
	repacking "allocation-simulator/repacking_heuristics"
	"allocation-simulator/input"
	"allocation-simulator/item"
	"allocation-simulator/logger"
	"allocation-simulator/operation"
	"allocation-simulator/simulator"
	"log/slog"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

/*
Test file for the Optimal Fit algorithm.

This file covers three main areas:
1. Validation of basic functioning.
2. Validation of integration with the main simulator.
3. Performance testing to evaluate algorrithm limits under stress.

Note: Some performance tests are designed near the upper limit of the algorithm’s capacity.
Depending on the hardware and environment, they may trigger timeout exceptions.
*/

// Basic functioning
func TestOptimalFit_1Item_1Bin(t *testing.T) {
	var of = &repacking.OptimalFit{}

	var (
		CPU_Capacity          int64   = 16
		Memory_Capacity       int64   = 16
		MinOccupationLimit    float64 = 0
		RepackingEnabled      bool    = true
		EventBasedRepacking   bool    = true
		RunOptimal            bool    = true
	)
	var sim = simulator.NewSimulator([]input.Event{}, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, EventBasedRepacking, RunOptimal)

	var it = item.Item{ID: "0", CPU: 2, Memory: 2}
	var items []item.Item = make([]item.Item, 0)
	items = append(items, it)

	var ItemToBinMap, err = of.Solve(sim, items)
	if err != nil {
		t.Fatalf("Solve returned error: %v", err)
	}
	if ItemToBinMap == nil || len(ItemToBinMap) != 1 {
		t.Fatalf("Invalid 'ItemToBinMap' format: %#v", ItemToBinMap)
	}
	if ItemToBinMap["0"] != 0 {
		t.Fatalf("Incorrect bin for item: got=%d want=0", ItemToBinMap["0"])
	}

	t.Logf("%v", ItemToBinMap)
}

func TestOptimalFit_ItemExceedsCapacity(t *testing.T) {
	var of = &repacking.OptimalFit{}

	var (
		CPU_Capacity          int64   = 16
		Memory_Capacity       int64   = 16
		MinOccupationLimit    float64 = 0
		RepackingEnabled      bool    = true
		EventBasedRepacking   bool    = true
		RunOptimal            bool    = true
	)
	var sim = simulator.NewSimulator([]input.Event{}, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, EventBasedRepacking, RunOptimal)

	var it = item.Item{ID: "0", CPU: 17, Memory: 1}
	var items []item.Item = make([]item.Item, 0)
	items = append(items, it)

	var ItemToBinMap, err = of.Solve(sim, items)
	if err == nil {
		t.Fatalf("expected error for exceeding capacity, got ItemToBinMap=%#v", ItemToBinMap)
	}
}

func TestOptimalFit_4Items_2Bins(t *testing.T) {
	var items []item.Item = make([]item.Item, 0)
	items = append(items, item.Item{ID: "0", CPU: 9, Memory: 11})
	items = append(items, item.Item{ID: "1", CPU: 7, Memory: 5})
	items = append(items, item.Item{ID: "2", CPU: 6, Memory: 12})
	items = append(items, item.Item{ID: "3", CPU: 10, Memory: 4})

	var usedBins = genericTest(t, items)

	if usedBins != 2 {
		t.Fatalf("wrong number of bins used: got=%d want=2", usedBins)
	}
}

func TestOptimalFit_12SmallItems_2Bins_PerfectFill(t *testing.T) {
	// Small items that fit in exactly 2 bins.
	// 2×(3,3) + 2×(3,2) + 2×(2,3) = (16,16)

	var items []item.Item = make([]item.Item, 0)
	// Bin A
	items = append(items, item.Item{ID: "0", CPU: 3, Memory: 2})
	items = append(items, item.Item{ID: "1", CPU: 2, Memory: 3})
	items = append(items, item.Item{ID: "2", CPU: 3, Memory: 2})
	items = append(items, item.Item{ID: "3", CPU: 3, Memory: 3})
	items = append(items, item.Item{ID: "4", CPU: 2, Memory: 3})
	items = append(items, item.Item{ID: "5", CPU: 3, Memory: 3})

	// Bin B
	items = append(items, item.Item{ID: "6", CPU: 2, Memory: 3})
	items = append(items, item.Item{ID: "7", CPU: 3, Memory: 2})
	items = append(items, item.Item{ID: "8", CPU: 3, Memory: 3})
	items = append(items, item.Item{ID: "9", CPU: 2, Memory: 3})
	items = append(items, item.Item{ID: "10", CPU: 3, Memory: 3})
	items = append(items, item.Item{ID: "11", CPU: 3, Memory: 2})

	var usedBins = genericTest(t, items)

	if usedBins != 2 {
		t.Fatalf("wrong number of bins used: got=%d want=2", usedBins)
	}
}

// Simulator integration
func TestOptimalFit_InsertionChangesBinMapping(t *testing.T) {
	logger.Init(slog.LevelDebug)
	var ff = heuristic.NewFirstFit()
	var of = repacking.NewOptimal()

	// first insert 4 items that completely fills 2 bins.
	// then, remove 1 item from each bin.
	// optimal fit should find a better solution and update the simulator
	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 3, Memory: 3}},
		{Time: 1, Operation: operation.Insertion, Item: item.Item{ID: "1", CPU: 13, Memory: 13}},
		{Time: 2, Operation: operation.Insertion, Item: item.Item{ID: "2", CPU: 3, Memory: 3}},
		{Time: 3, Operation: operation.Insertion, Item: item.Item{ID: "3", CPU: 13, Memory: 13}},

		{Time: 4, Operation: operation.Deletion, Item: item.Item{ID: "1", CPU: 13, Memory: 13}},
		{Time: 5, Operation: operation.Deletion, Item: item.Item{ID: "3", CPU: 13, Memory: 13}},

		{Time: 6, Operation: operation.Insertion, Item: item.Item{ID: "4", CPU: 3, Memory: 3}},
	}

	var (
		CPU_Capacity          int64   = 16
		Memory_Capacity       int64   = 16
		MinOccupationLimit    float64 = 0
		RepackingEnabled      bool    = true
		EventBasedRepacking   bool    = true
		RunOptimal            bool    = true
	)
	var sim = simulator.NewSimulator(events, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, EventBasedRepacking, RunOptimal)

	// First Fit handle events
	var err = ff.HandleEvent(sim, 0)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	err = ff.HandleEvent(sim, 1)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	err = ff.HandleEvent(sim, 2)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	err = ff.HandleEvent(sim, 3)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	err = ff.HandleEvent(sim, 4)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	err = ff.HandleEvent(sim, 5)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}

	// running optimal fit
	err = of.HandleEvent(sim, 6, nil, 1)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	var usedBins = sim.NumberOfBinsUsed()
	if usedBins != 1 {
		t.Fatalf("wrong number of bins used: got=%d want=1", usedBins)
	}
	var itemsAllocated = len(sim.ItemToBinMap)
	if itemsAllocated != 3 {
		t.Fatalf("wrong number of items allocated: got=%d want=3", itemsAllocated)
	}
}

func TestOptimalFit_DeletionChangesBinMapping(t *testing.T) {
	logger.Init(slog.LevelDebug)
	var ff = heuristic.NewFirstFit()
	var of = repacking.NewOptimal()

	// first insert 4 items that completely fills 2 bins.
	// then, remove 1 item from each bin.
	// optimal fit should find a better solution and update the simulator
	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 3, Memory: 3}},
		{Time: 1, Operation: operation.Insertion, Item: item.Item{ID: "1", CPU: 13, Memory: 13}},
		{Time: 2, Operation: operation.Insertion, Item: item.Item{ID: "2", CPU: 3, Memory: 3}},
		{Time: 3, Operation: operation.Insertion, Item: item.Item{ID: "3", CPU: 13, Memory: 13}},

		{Time: 4, Operation: operation.Deletion, Item: item.Item{ID: "1", CPU: 13, Memory: 13}},
		{Time: 5, Operation: operation.Deletion, Item: item.Item{ID: "3", CPU: 13, Memory: 13}},
	}

	var (
		CPU_Capacity          int64   = 16
		Memory_Capacity       int64   = 16
		MinOccupationLimit    float64 = 0
		RepackingEnabled      bool    = true
		EventBasedRepacking   bool    = true
		RunOptimal            bool    = true
	)
	var sim = simulator.NewSimulator(events, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, EventBasedRepacking, RunOptimal)

	// First Fit handle insertion events
	var err = ff.HandleEvent(sim, 0)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	err = ff.HandleEvent(sim, 1)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	err = ff.HandleEvent(sim, 2)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	err = ff.HandleEvent(sim, 3)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}

	// running optimal fit
	err = of.HandleEvent(sim, 4, nil, 1)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	var usedBins = sim.NumberOfBinsUsed()
	if usedBins != 2 {
		t.Fatalf("wrong number of bins used: got=%d want=2", usedBins)
	}
	var itemsAllocated = len(sim.ItemToBinMap)
	if itemsAllocated != 3 {
		t.Fatalf("wrong number of items allocated: got=%d want=3", itemsAllocated)
	}

	err = of.HandleEvent(sim, 5, nil, 1)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	usedBins = sim.NumberOfBinsUsed()
	if usedBins != 1 {
		t.Fatalf("wrong number of bins used: got=%d want=1", usedBins)
	}
	itemsAllocated = len(sim.ItemToBinMap)
	if itemsAllocated != 2 {
		t.Fatalf("wrong number of items allocated: got=%d want=2", itemsAllocated)
	}
}

// Performance tests
func TestOptimalFit_badCase1_20items(t *testing.T) {
	var items []item.Item = make([]item.Item, 0)
	var id int = 0
	for range 6 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 5, Memory: 3})
		id++
	}
	for range 7 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 8, Memory: 5})
		id++
	}
	for range 7 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 3, Memory: 8})
		id++
	}

	var usedBins = genericTest(t, items)

	if usedBins != 7 {
		t.Fatalf("wrong number of bins used: got=%d want=7", usedBins)
	}
}

func TestOptimalFit_badCase2_23Items(t *testing.T) {
	var items []item.Item = make([]item.Item, 0)
	var id int = 0
	for range 5 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 12, Memory: 3})
		id++
	}
	for range 4 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 4, Memory: 7})
		id++
	}
	for range 4 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 3, Memory: 6})
		id++
	}
	for range 5 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 5, Memory: 12})
		id++
	}
	for range 5 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 4, Memory: 2})
		id++
	}

	var usedBins = genericTest(t, items)

	if usedBins != 12 {
		t.Fatalf("wrong number of bins used: got=%d want=12", usedBins)
	}
}

func TestOptimalFit_badCase3_21Items(t *testing.T) {
	var items []item.Item = make([]item.Item, 0)
	var id int = 0
	for range 2 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 7, Memory: 5})
		id++
	}
	for range 7 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 5, Memory: 7})
		id++
	}
	for range 3 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 1, Memory: 9})
		id++
	}
	for range 6 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 3, Memory: 6})
		id++
	}
	for range 3 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 12, Memory: 5})
		id++
	}

	var usedBins = genericTest(t, items)

	if usedBins != 11 {
		t.Fatalf("wrong number of bins used: got=%d want=11", usedBins)
	}
}

// This case is interesting because the theoretical minimum amount of bins is 22, but the algorithm achieved
// the real solution of 29 bins extremely quickly. This can only be explained by the great effectiveness of
// the pruning techniques.
func TestOptimalFit_goodCase1_40Items(t *testing.T) {
	var items []item.Item = make([]item.Item, 0)
	var id int = 0
	for range 10 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 12, Memory: 3})
		id++
	}
	for range 6 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 4, Memory: 9})
		id++
	}
	for range 5 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 3, Memory: 9})
		id++
	}
	for range 9 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 9, Memory: 4})
		id++
	}
	for range 10 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 11, Memory: 2})
		id++
	}

	var usedBins = genericTest(t, items)

	if usedBins != 29 {
		t.Fatalf("wrong number of bins used: got=%d want=29", usedBins)
	}
}

// This case is interesting because the theoretical minimum amount of bins is 22, and the algorithm
// also achieved 22 quickly. This can be explained by considering that pre-ordering the items could
// lead to the theoretical minimum solution quite fast, which makes unnecessary to keep searching for better solutions.
func TestOptimalFit_goodCase2_50Items(t *testing.T) {
	var items []item.Item = make([]item.Item, 0)
	var id int = 0
	for range 11 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 12, Memory: 3})
		id++
	}
	for range 6 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 4, Memory: 7})
		id++
	}
	for range 8 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 3, Memory: 6})
		id++
	}
	for range 15 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 8, Memory: 2})
		id++
	}
	for range 10 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 4, Memory: 2})
		id++
	}

	var usedBins = genericTest(t, items)

	if usedBins != 22 {
		t.Fatalf("wrong number of bins used: got=%d want=22", usedBins)
	}
}

func TestOptimalFit_goodCase3_50Items(t *testing.T) {
	var items []item.Item = make([]item.Item, 0)
	var id int = 0
	for range 6 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 12, Memory: 3})
		id++
	}
	for range 9 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 4, Memory: 7})
		id++
	}
	for range 4 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 3, Memory: 6})
		id++
	}
	for range 21 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 8, Memory: 2})
		id++
	}
	for range 10 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 4, Memory: 2})
		id++
	}

	var usedBins = genericTest(t, items)

	if usedBins != 21 {
		t.Fatalf("wrong number of bins used: got=%d want=21", usedBins)
	}
}

func TestOptimalFit_goodCase4_40Items(t *testing.T) {
	var items []item.Item = make([]item.Item, 0)
	var id int = 0
	for range 10 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 12, Memory: 3})
		id++
	}
	for range 6 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 4, Memory: 9})
		id++
	}
	for range 5 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 3, Memory: 9})
		id++
	}
	for range 9 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 9, Memory: 4})
		id++
	}
	for range 10 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 11, Memory: 2})
		id++
	}

	var usedBins = genericTest(t, items)

	if usedBins != 29 {
		t.Fatalf("wrong number of bins used: got=%d want=29", usedBins)
	}
}

func TestOptimalFit_reallyGoodCase_150Items(t *testing.T) {
	var items []item.Item = make([]item.Item, 0)
	var id int = 0
	for range 30 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 5, Memory: 3})
		id++
	}
	for range 30 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 12, Memory: 7})
		id++
	}
	for range 30 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 3, Memory: 7})
		id++
	}
	for range 30 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 8, Memory: 5})
		id++
	}
	for range 30 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 1, Memory: 2})
		id++
	}

	var usedBins = genericTest(t, items)

	if usedBins != 55 {
		t.Fatalf("wrong number of bins used: got=%d want=55", usedBins)
	}
}

// Performance test template
func TestOptimalFit_playground(t *testing.T) {
	var items []item.Item = make([]item.Item, 0)
	var id int = 0
	for range 2 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 5, Memory: 8})
		id++
	}
	for range 2 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 3, Memory: 5})
		id++
	}
	for range 2 {
		items = append(items, item.Item{ID: strconv.Itoa(id), CPU: 8, Memory: 3})
		id++
	}
	genericTest(t, items)
}

// Generic test for performance tests
func genericTest(t *testing.T, items []item.Item) int {
	var of = &repacking.OptimalFit{}
	var (
		CPU_Capacity          int64   = 16
		Memory_Capacity       int64   = 16
		MinOccupationLimit    float64 = 0
		RepackingEnabled      bool    = true
		EventBasedRepacking   bool    = true
		RunOptimal            bool    = true
	)
	var sim = simulator.NewSimulator([]input.Event{}, CPU_Capacity, Memory_Capacity, MinOccupationLimit, RepackingEnabled, EventBasedRepacking, RunOptimal)

	// deterministic shuffle
	var r = rand.New(rand.NewSource(424242))
	r.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })

	// solve
	var start = time.Now()
	var ItemToBinMap, err = of.Solve(sim, items)
	var duration = time.Since(start)
	t.Logf("Solve() took %s", duration)

	if err != nil {
		t.Fatalf("Solve returned error: %v", err)
	}
	if ItemToBinMap == nil || len(ItemToBinMap) != len(items) {
		t.Fatalf("invalid ItemToBinMap length: got=%d want=%d", len(ItemToBinMap), len(items))
	}

	// validate capacities and compute aggregates
	var used = make(map[int]struct{})
	var loadCPU = make(map[int]int64)
	var loadMEM = make(map[int]int64)
	var sumCPU int64
	var sumMEM int64

	for _, it := range items {
		var binID, ok = ItemToBinMap[it.ID]
		if !ok {
			t.Fatalf("item %s not mapped in ItemToBinMap", it.ID)
		}

		used[binID] = struct{}{}
		loadCPU[binID] += it.CPU
		loadMEM[binID] += it.Memory
		sumCPU += it.CPU
		sumMEM += it.Memory

		if loadCPU[binID] > sim.CPU_Capacity {
			t.Fatalf("bin %d exceeds CPU capacity: load=%d cap=%d", binID, loadCPU[binID], sim.CPU_Capacity)
		}
		if loadMEM[binID] > sim.Memory_Capacity {
			t.Fatalf("bin %d exceeds MEM capacity: load=%d cap=%d", binID, loadMEM[binID], sim.Memory_Capacity)
		}
	}

	var minBinsCPU int = int((sumCPU + sim.CPU_Capacity - 1) / sim.CPU_Capacity)
	var minBinsMEM int = int((sumMEM + sim.CPU_Capacity - 1) / sim.Memory_Capacity)
	var theoreticalMin = max(minBinsCPU, minBinsMEM)

	t.Logf("used bins: %d", len(used))
	t.Logf("theoretical minimum bins: %d", theoreticalMin)
	t.Logf("%v", ItemToBinMap)

	return len(used)
}
