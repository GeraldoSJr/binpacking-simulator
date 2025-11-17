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
	"testing"
)

func TestRepacking_2Items_RepackingNeeded_firstFit(t *testing.T) {
	logger.Init(slog.LevelDebug)

	var h = heuristic.NewFirstFit()
	var r = repacking.NewR1()

	// step 1: insert items
	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 6, Memory: 6}},
		{Time: 1, Operation: operation.Insertion, Item: item.Item{ID: "1", CPU: 6, Memory: 6}},
		{Time: 2, Operation: operation.Insertion, Item: item.Item{ID: "2", CPU: 6, Memory: 6}},
		{Time: 3, Operation: operation.Deletion, Item: item.Item{ID: "1", CPU: 6, Memory: 6}},
	}

	// despite repacking being disabled bellow, repacking still occurs because
	// this test runs it manually when required.
	var (
		CPU_Capacity        int64   = 16
		Memory_Capacity     int64   = 16
		MinOccupationLimit  float64 = 0
		RepackingEnabled    bool    = false
		EventBasedRepacking bool    = false
		RunOptimal          bool    = false
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

	err = h.HandleEvent(sim, 3)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}

	// step 2: run repacking
	err = r.HandleEvent(sim, 3, h, 0.7)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	// bin 0 and 1 are underutilized. They'll be closed and items added to a new bin.
	if !sim.Bins[0].Closed {
		t.Fatalf("Bin 0 should have been closed after repacking.")
	}
	if !sim.Bins[1].Closed {
		t.Fatalf("Bin 1 should have been closed after repacking.")
	}
	if sim.ItemToBinMap["0"] != 2 {
		t.Fatalf("Item 0 should have been allocated to bin 2.")
	}
	if sim.ItemToBinMap["2"] != 2 {
		t.Fatalf("Item 2 should have been allocated to bin 2.")
	}
}

func TestRepacking_2Items_RepackingNeeded_bestFit(t *testing.T) {
	logger.Init(slog.LevelDebug)

	var h = heuristic.NewBestFit()
	var r = repacking.NewR1()

	// step 1: insert items
	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 12, Memory: 12}},
		{Time: 1, Operation: operation.Insertion, Item: item.Item{ID: "1", CPU: 14, Memory: 14}},
		{Time: 2, Operation: operation.Insertion, Item: item.Item{ID: "2", CPU: 4, Memory: 4}},
		{Time: 3, Operation: operation.Insertion, Item: item.Item{ID: "3", CPU: 2, Memory: 2}},
		{Time: 4, Operation: operation.Insertion, Item: item.Item{ID: "4", CPU: 2, Memory: 2}},

		{Time: 5, Operation: operation.Deletion, Item: item.Item{ID: "2", CPU: 4, Memory: 4}},
		{Time: 6, Operation: operation.Deletion, Item: item.Item{ID: "3", CPU: 2, Memory: 2}},
	}

	// despite repacking being disabled bellow, repacking still occurs because
	// this test runs it manually when required.
	var (
		CPU_Capacity        int64   = 16
		Memory_Capacity     int64   = 16
		MinOccupationLimit  float64 = 0
		RepackingEnabled    bool    = false
		EventBasedRepacking bool    = false
		RunOptimal          bool    = false
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

	err = h.HandleEvent(sim, 3)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}

	err = h.HandleEvent(sim, 4)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}

	err = h.HandleEvent(sim, 5)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}

	err = h.HandleEvent(sim, 6)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}

	// step 2: run repacking
	err = r.HandleEvent(sim, 6, h, 0.7)
	if err != nil {
		t.Fatalf("HandleEvent returned error: %v", err)
	}
	// bin 2 is underutilized. It'll be closed and its item will go to bin 1.
	if !sim.Bins[2].Closed {
		t.Fatalf("Bin 0 should have been closed after repacking.")
	}
	if sim.ItemToBinMap["4"] != 1 {
		t.Fatalf("Item 0 should have been allocated to bin 1.")
	}
}

func TestRepacking_3items_NoRepackingEvenWithUnderusage(t *testing.T) {
	logger.Init(slog.LevelDebug)

	var h = heuristic.NewFirstFit()
	var r = repacking.NewR1()

	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 6, Memory: 6}},
		{Time: 1, Operation: operation.Insertion, Item: item.Item{ID: "1", CPU: 6, Memory: 6}},
		{Time: 2, Operation: operation.Insertion, Item: item.Item{ID: "2", CPU: 6, Memory: 6}},
	}

	// despite repacking being disabled bellow, repacking still occurs because
	// this test runs it manually when required.
	var (
		CPU_Capacity        int64   = 16
		Memory_Capacity     int64   = 16
		MinOccupationLimit  float64 = 0
		RepackingEnabled    bool    = false
		EventBasedRepacking bool    = false
		RunOptimal          bool    = false
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

	// Repacking should not change anything (bin 0 is full; bin 1 cannot be merged).
	err = r.HandleEvent(sim, 2, h, 0.7)
	if err != nil {
		t.Fatalf("Repacking returned error: %v", err)
	}
	if sim.Bins[0].Closed {
		t.Fatalf("Bin 0 should remain open after repacking.")
	}
	if sim.Bins[1].Closed {
		t.Fatalf("Bin 1 should remain open after repacking.")
	}
	if sim.ItemToBinMap["0"] != 0 || sim.ItemToBinMap["1"] != 0 || sim.ItemToBinMap["2"] != 1 {
		t.Fatalf("Item to bin mapping changed unexpectedly: %+v", sim.ItemToBinMap)
	}
}

func TestRepacking_3items_NoRepackingNoUnderusage(t *testing.T) {
	logger.Init(slog.LevelDebug)

	var h = heuristic.NewFirstFit()
	var r = repacking.NewR1()

	var events = []input.Event{
		{Time: 0, Operation: operation.Insertion, Item: item.Item{ID: "0", CPU: 8, Memory: 8}},
		{Time: 1, Operation: operation.Insertion, Item: item.Item{ID: "1", CPU: 8, Memory: 8}},
		{Time: 2, Operation: operation.Insertion, Item: item.Item{ID: "2", CPU: 16, Memory: 16}},
	}

	// despite repacking being disabled bellow, repacking still occurs because
	// this test runs it manually when required.
	var (
		CPU_Capacity        int64   = 16
		Memory_Capacity     int64   = 16
		MinOccupationLimit  float64 = 0
		RepackingEnabled    bool    = false
		EventBasedRepacking bool    = false
		RunOptimal          bool    = false
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

	// Repacking should not change anything (bin 0 and bin 1 are full).
	err = r.HandleEvent(sim, 2, h, 0.7)
	if err != nil {
		t.Fatalf("Repacking returned error: %v", err)
	}
	if sim.Bins[0].Closed {
		t.Fatalf("Bin 0 should remain open after repacking.")
	}
	if sim.Bins[1].Closed {
		t.Fatalf("Bin 1 should remain open after repacking.")
	}
	if sim.ItemToBinMap["0"] != 0 || sim.ItemToBinMap["1"] != 0 || sim.ItemToBinMap["2"] != 1 {
		t.Fatalf("Item to bin mapping changed unexpectedly: %+v", sim.ItemToBinMap)
	}
}
