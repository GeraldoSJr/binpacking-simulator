package heuristic

import (
	"allocation-simulator/bin"
	"allocation-simulator/item"
	"allocation-simulator/logger"
	"allocation-simulator/operation"
	"allocation-simulator/simulator"
	"fmt"
	"log/slog"
)

// The heuristic allocates the item to a bin where it leaves the least leftover space.
type BestFit struct{
    log *slog.Logger
}

func NewBestFit() *BestFit {
    return &BestFit{ log: logger.Module("best_fit") }
}

func (bf *BestFit) HandleEvent(s *simulator.Simulator, index int) error {
	var log = bf.log

	var event = s.Events[index]
	var usedBin *bin.Bin

	if event.Operation == operation.Insertion {
		var item = event.Item

		// checking if item fits
		if item.CPU > s.CPU_Capacity || item.Memory > s.Memory_Capacity {
			log.Error("oversized item", "item_cpu", item.CPU, "item_mem", item.Memory, "cpu_capacity", s.CPU_Capacity, "mem_capacity", s.Memory_Capacity)
			
			return fmt.Errorf("Oversized item. ID: %s, CPU: %d (Capacity: %d), Memory: %d (Capacity: %d)",
				item.ID, item.CPU, s.CPU_Capacity, item.Memory, s.Memory_Capacity,
			)
		}

		// selecting first bin
		var binID = bf.RunHeuristic(s, item)
		if binID == len(s.Bins) {
			log.Debug("creating new bin", "bin_ID", binID)
			var newBin = bin.NewBin(len(s.Bins), s.CPU_Capacity, s.Memory_Capacity)
			s.Bins = append(s.Bins, *newBin)
			usedBin = &s.Bins[binID]
		} else {
			usedBin = &s.Bins[binID]
		}

		// add item to bin
		if !usedBin.TryAdd(item) {
			log.Error("heuristic chose bin but item couldn't fit", "bin_ID", usedBin.ID, "item_ID", item.ID)
			return fmt.Errorf("heuristic chose bin %d but item %s couldn't fit", usedBin.ID, item.ID)
		}
		s.Items[item.ID] = item
		s.ItemToBinMap[item.ID] = binID
		log.Debug("item added", "item_cpu", item.CPU, "item_mem", item.Memory, "bin_ID", binID, "remaining_cpu", usedBin.RemainingCPUCapacity(), "remaining_mem", usedBin.RemainingMemoryCapacity())
	}

	// If a bin becomes empty after an item is removed, it is permanently closed.
	if event.Operation == operation.Deletion {
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

		usedBin = &s.Bins[binID]

		usedBin.RemoveItem(item)
		delete(s.ItemToBinMap, item.ID)
		delete(s.Items, item.ID)
		log.Debug("item removed from bin", "item_cpu", item.CPU, "item_mem", item.Memory, "bin_id", binID, "remaining_cpu", usedBin.RemainingCPUCapacity(), "remaining_mem", usedBin.RemainingMemoryCapacity())

		if usedBin.Empty() {
			log.Debug("bin became empty, so it was closed.", "bin_id", binID)
			usedBin.Closed = true
		}
	}

	fmt.Printf("%d,%s,%s,%d,%d,%d,%d,%d\n", event.Time, event.Operation, event.Item.ID, event.Item.CPU, event.Item.Memory, usedBin.ID, usedBin.CPU_Capacity, usedBin.Memory_Capacity)

	return nil
}

func (bf *BestFit) RunHeuristic(s *simulator.Simulator, item item.Item) int {
	var log = bf.log
	log.Debug("inserting new item. searching best fit bin", "item_id", item.ID, "cpu", item.CPU, "mem", item.Memory)

	var usedBin *bin.Bin
	for i := range s.Bins {
		var currentBin = &s.Bins[i]
		
		if currentBin.Closed || !currentBin.CanFit(item) {
			continue
		}

		if usedBin == nil {
			usedBin = currentBin
			continue
		}
		
		var currCPU = currentBin.RemainingCPUCapacity()
		var currMem = currentBin.RemainingMemoryCapacity()
		var chosenCPU = usedBin.RemainingCPUCapacity()
		var chosenMem = usedBin.RemainingMemoryCapacity()

		if currCPU < chosenCPU ||
		(currCPU == chosenCPU && currMem < chosenMem) {
			usedBin = currentBin
		}
	}

	if usedBin != nil {
		return usedBin.ID
	} else {
		return len(s.Bins)
	}
}