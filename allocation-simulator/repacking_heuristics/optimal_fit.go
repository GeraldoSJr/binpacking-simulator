package repacking

import (
	"allocation-simulator/bin"
	"allocation-simulator/item"
	"allocation-simulator/logger"
	"allocation-simulator/operation"
	"allocation-simulator/simulator"
	"fmt"
	"log/slog"
	"sort"
)

// Simulator state used to: compare cost, save best state, and prune the search tree.
type SimState struct {
    BinCount     int            // number of non-empty bins
    FreeCPU      int64          // total CPU slack across all bins
    FreeMEM      int64          // total memory slack across all bins
	ItemToBinMap map[string]int // maps item IDs to bin IDs
}

// Search context for the recursive optimal fit.
type searchCtx struct {
    sim     *simulator.Simulator
    items   []item.Item
    bins    *[]*bin.Bin // workspace bins to explore combinations without touching the main simulation

    // Lower-bound helpers: sum of remaining demand from i..end.
    necCPU  []int64   // necCPU[i] = sum(items[i:].CPU)
    necMEM  []int64   // necMEM[i] = sum(items[i:].Memory)
    minBins int       // theoretical minimum number of bins

    currentState SimState // current algorithm state
    bestSoFar    SimState // best state found so far (fewest bins)
}

type OptimalFit struct{
	log *slog.Logger
}

func NewOptimal() *OptimalFit {
    return &OptimalFit{ log: logger.Module("optimal_fit") }
}

// h and occupationLimit are ignored. These parameters are here only in order to implement the Repacking interface.
func (of *OptimalFit) HandleEvent(s *simulator.Simulator, index int, h simulator.Heuristic, occupationLimit float64) error {
	var log = of.log
	var event = s.Events[index]

	if (event.Operation == operation.Deletion) {
		var itemID = event.Item.ID
		log.Debug("removing item", "item_id", itemID)

		var binID, found = s.ItemToBinMap[itemID]
		if !found {
			log.Error("item not found in ItemToBinMap")
			return fmt.Errorf("item %s not found in ItemToBinMap", itemID)
		}
		var item, itemFound = s.Items[itemID]
		if !itemFound {
			log.Error("item not found in Items map")
			return fmt.Errorf("item %s found in ItemToBinMap but not in Items map", itemID)
		}

		var usedBin = &s.Bins[binID]
		usedBin.RemoveItem(item)
		delete(s.ItemToBinMap, item.ID)
		delete(s.Items, item.ID)
		log.Debug("item removed from bin", "item_cpu", item.CPU, "item_mem", item.Memory, "bin_id", binID, "remaining_cpu", usedBin.RemainingCPUCapacity(), "remaining_mem", usedBin.RemainingMemoryCapacity())

		if usedBin.Empty() {
			log.Debug("bin became empty, so it was closed.", "bin_id", binID)
			usedBin.Closed = true
		}

		fmt.Printf("%d,%s,%s,%d,%d,%d,%d,%d\n", event.Time, event.Operation, itemID, event.Item.CPU, event.Item.Memory, binID, usedBin.CPU_Capacity, usedBin.Memory_Capacity)
	}

	// make slice of currently used items
	var items = make([]item.Item, 0)
	for itemID := range s.ItemToBinMap {
		items = append(items, s.Items[itemID])
	}

	if (event.Operation == operation.Insertion) {
		var newItem = event.Item
		// add new item
		items = append(items, newItem)
		s.Items[newItem.ID] = newItem
		log.Debug("adding new item to simulation. optimal fit allocation still must be found.", "item_cpu", newItem.CPU, "item_mem", newItem.Memory)
	}

	// find the optimal allocation for the given bins
	log.Debug("finding optimal allocation")
	var newItemToBinMap, err = of.Solve(s, items)
	if err != nil {
		return err
	}
	
	// calculate how many bins were used and validate mapping
	var usedBinIDs = make(map[int]struct{})
	for _, it := range items {
		var binID, ok = newItemToBinMap[it.ID]
		if !ok {
			log.Error("item not found in newItemToBinMap")
			return fmt.Errorf("item %s not mapped in newItemToBinMap", it.ID)
		}
		usedBinIDs[binID] = struct{}{}
	}
	log.Debug("optimal allocation found", "used_bins", len(usedBinIDs))

	// initialize new bins
	log.Debug("initializing new bins. all bins are going to be reseted.")
	var newBins []bin.Bin = make([]bin.Bin, len(usedBinIDs))
	for id := range newBins {
		var newBin = *bin.NewBin(id, s.CPU_Capacity, s.Memory_Capacity)
		newBins[id] = newBin
	}

	// add items to new bins
	log.Debug("adding items to new bins and computing migrations")
	for _, item := range items {
		var fromBinID, existingItem = s.ItemToBinMap[item.ID]
		var toBinID = newItemToBinMap[item.ID]
		var toBin = &newBins[toBinID]

		if !toBin.TryAdd(item) {
			log.Error("failed to add item to bin", "item_id", item.ID, "bin_id", toBinID)
			return fmt.Errorf("failed to add item %s to bin %d", item.ID, toBinID)
		}
		
		if (!existingItem) {
			log.Debug("new item added to bin", "item_id", item.ID, "bin_id", toBinID)
			fmt.Printf("%d,%s,%s,%d,%d,%d,%d,%d\n", event.Time, event.Operation, event.Item.ID, event.Item.CPU, event.Item.Memory, toBinID, toBin.CPU_Capacity, toBin.Memory_Capacity)
		}
		if (existingItem && fromBinID != toBinID) {
			log.Debug("item migrated", "item_id", item.ID, "from_bin_id", fromBinID, "to_bin_id", toBinID)
			fmt.Printf("%d,migration,%s,%d,%d,%d,%d,%d\n", s.Events[index].Time, item.ID, item.CPU, item.Memory, toBinID, toBin.CPU_Capacity, toBin.Memory_Capacity)
		}
	}

	// update simulator
	s.Bins = newBins
	s.ItemToBinMap = newItemToBinMap
	
	log.Debug("optimal fit completed")

	return nil
}

func (of *OptimalFit) Solve(s *simulator.Simulator, items []item.Item) (map[string]int, error) {
	// 0) Validate: each item must fit in an empty bin.
	var n int = len(items)
	for _, it := range items {
		if it.CPU > s.CPU_Capacity || it.Memory > s.Memory_Capacity {
			return nil, fmt.Errorf("item %s exceeds bin capacity (cpu=%d, mem=%d)", it.ID, it.CPU, it.Memory)
		}
	}

	// 1) Make a copy of items to avoid in-place changes.
	itemsSorted := make([]item.Item, n)
	copy(itemsSorted, items)

	// 2) Sort directly by CPU desc, then Memory desc.
	sort.Sort(item.ByCPUAndMemoryDesc(itemsSorted))
	
	// 3) Build suffix sums and derive theoretical lower bound K from them.
	var ctx searchCtx
	ctx.sim = s
	ctx.items = itemsSorted

	ctx.necCPU = make([]int64, n+1)
	ctx.necMEM = make([]int64, n+1)
	for k := n - 1; k >= 0; k-- {
		ctx.necCPU[k] = ctx.necCPU[k+1] + itemsSorted[k].CPU
		ctx.necMEM[k] = ctx.necMEM[k+1] + itemsSorted[k].Memory
	}

	// total CPU and MEM usage are at index 0; use integer ceil-div to compute lower bound
	var totalCPU int64 = ctx.necCPU[0]
	var totalMEM int64 = ctx.necMEM[0]
	var minBinsCPU int = int((totalCPU + s.CPU_Capacity - 1) / s.CPU_Capacity)
	var minBinsMEM int = int((totalMEM + s.Memory_Capacity - 1) / s.Memory_Capacity)
	var K int = max(minBinsCPU, minBinsMEM)

	// 4) Create K empty working bins minimally required (at least one).
	var base = max(1, K)
	var workingBins []*bin.Bin = make([]*bin.Bin, base)
	for id := range workingBins {
		workingBins[id] = bin.NewBin(id, s.CPU_Capacity, s.Memory_Capacity)
	}
	
	// 5) Prepare search context.
	ctx.bins = &workingBins
	ctx.minBins = K

	ctx.currentState.ItemToBinMap = make(map[string]int, n)
	ctx.bestSoFar.ItemToBinMap = make(map[string]int, n)
	ctx.bestSoFar.BinCount = n + 1 // sentinel: worst-case uses n bins

	// Total free capacity across all working bins at the start.
	ctx.currentState.FreeCPU += int64(base) * s.CPU_Capacity
	ctx.currentState.FreeMEM += int64(base) * s.Memory_Capacity

	// 6) Solve recursively.
	of.solveRecursive(&ctx, 0)

	return ctx.bestSoFar.ItemToBinMap, nil
}

func (of *OptimalFit) solveRecursive(ctx *searchCtx, i int) {
	// Base case: all items have been placed.
	if i == len(ctx.items) {
		// Update best solution if the current one is strictly better.
		if ctx.currentState.BinCount < ctx.bestSoFar.BinCount {
			ctx.bestSoFar.BinCount = ctx.currentState.BinCount
			
			ctx.bestSoFar.ItemToBinMap = make(map[string]int, len(ctx.currentState.ItemToBinMap))
			for k, v := range ctx.currentState.ItemToBinMap {
				ctx.bestSoFar.ItemToBinMap[k] = v
			}
		}
		return
	}
	
	// Prune 0: stop if we already matched the theoretical minimum bin count.
	if ctx.bestSoFar.BinCount == ctx.minBins {
		return
	}

	// Prune 1: stop if the current solution already uses
	// at least as many bins as the best solution found so far.
	if ctx.currentState.BinCount >= ctx.bestSoFar.BinCount {
		return
	}
	
	// Prune 2: compute a lower bound on the number of additional bins required
	// to place the remaining items, based on demand vs. free capacity.
	var defCPU int64 = max(0, ctx.necCPU[i] - ctx.currentState.FreeCPU)
	var defMEM int64 = max(0, ctx.necMEM[i] - ctx.currentState.FreeMEM)

	var needCPU int = int((defCPU+ctx.sim.CPU_Capacity-1) / ctx.sim.CPU_Capacity) // ceil division
	var needMEM int = int((defMEM+ctx.sim.Memory_Capacity-1) / ctx.sim.Memory_Capacity) // ceil division

	var lb int = max(needCPU, needMEM)

	// If the lower bound makes the solution worse than bestSoFar, stop.
	if ctx.currentState.BinCount+lb >= ctx.bestSoFar.BinCount {
		return
	}

	// Current item to be placed.
	var it item.Item = ctx.items[i]

	// Symmetry-breaking: only allow placing into one empty bin per level.
	var usedEmptyBin = false

	// Try to place the item into each available bin.
	for _, currentBin := range *ctx.bins {
		// Prune 0 again: currentState.BinCount could have been updated on previous iterations.
		if ctx.bestSoFar.BinCount == ctx.minBins {
			break
		}

		// Prune 1 is not needed again, because Prune 2 already includes Prune 1.
		
		// Prune 2 again: currentState.BinCount could have been updated on previous iterations.
		if ctx.currentState.BinCount+lb >= ctx.bestSoFar.BinCount {
			continue
		}

		// Checks if item can fit.
		if !currentBin.CanFit(it) {
			continue
		}

		var wasEmpty bool = currentBin.Empty()
		if wasEmpty {
			// If bin is empty, only try once to avoid equivalent permutations.
			if usedEmptyBin {
				continue
			}
			usedEmptyBin = true

			// Since we are adding an item to an empty bin, increase the used bin count.
			ctx.currentState.BinCount++
		}

		// Place item in the bin.
		currentBin.TryAdd(it)
		ctx.currentState.ItemToBinMap[it.ID] = currentBin.ID
		ctx.currentState.FreeCPU -= it.CPU
		ctx.currentState.FreeMEM -= it.Memory

		// If that bin was the only empty bin, create a new one.
		var binWasCreated bool = false
		if wasEmpty {
			var emptyBins int = len(*ctx.bins) - ctx.currentState.BinCount
			if emptyBins == 0 {
				var newID int = len(*ctx.bins)
				var newBin = bin.NewBin(newID, ctx.sim.CPU_Capacity, ctx.sim.Memory_Capacity)
				*ctx.bins = append(*ctx.bins, newBin)
				binWasCreated = true

				ctx.currentState.FreeCPU += ctx.sim.CPU_Capacity
				ctx.currentState.FreeMEM += ctx.sim.Memory_Capacity
			}
		}

		// Recurse with next item.
		of.solveRecursive(ctx, i+1)

		// Backtrack: if new bin was created, delete it.
		if binWasCreated {
			var bins = *ctx.bins
			var last = len(bins) - 1
			*ctx.bins = bins[:last]

			ctx.currentState.FreeCPU -= ctx.sim.CPU_Capacity
			ctx.currentState.FreeMEM -= ctx.sim.Memory_Capacity
		}

		// Backtrack: remove item and restore state.
		ctx.currentState.FreeCPU += it.CPU
		ctx.currentState.FreeMEM += it.Memory
		delete(ctx.currentState.ItemToBinMap, it.ID)
		currentBin.RemoveItem(it)
		if wasEmpty {
			ctx.currentState.BinCount--
		}
	}
}