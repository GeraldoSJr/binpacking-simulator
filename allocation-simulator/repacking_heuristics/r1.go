package repacking

import (
	"allocation-simulator/bin"
	"allocation-simulator/item"
	"allocation-simulator/logger"
	"allocation-simulator/simulator"
	"fmt"
	"log/slog"
	"sort"
)

type R1 struct{
	log *slog.Logger
}

func NewR1() *R1 {
    return &R1{ log: logger.Module("r1_repacking") }
}

func (r1 *R1) HandleEvent(s *simulator.Simulator, index int, h simulator.Heuristic, occupationLimit float64) error {
	simCopy := s.DeepCopy()
	
	var log = r1.log

	// identify underutilized bins
	log.Debug("identifying underutilized bins")
	underutilizedBinIDs := make(map[int]struct{}) //a set of underutilized bin IDs
	for i := range simCopy.Bins {
		b := &simCopy.Bins[i]
		if b.Closed {
			continue
		}

		var binOccupation float64
		if b.CPU_Capacity > 0 && b.Memory_Capacity > 0 {
			binOccupation = 1 - (
				float64(b.RemainingCPUCapacity())/float64(b.CPU_Capacity) +
				float64(b.RemainingMemoryCapacity())/float64(b.Memory_Capacity)) / 2
		}
		if binOccupation < occupationLimit {
			log.Debug("underutilized bin identified", "id", b.ID, "remaining_cpu", b.RemainingCPUCapacity(), "remaining_mem", b.RemainingMemoryCapacity())

			underutilizedBinIDs[b.ID] = struct{}{}
			b.Closed = true
		}
	}

	if len(underutilizedBinIDs) == 0 {
		log.Info("no underutilized bin was found. no repacking required.")
		return nil
	}

	// identify items to be migrated
	log.Debug("identifying items to be migrated")
	migratedItems := []item.Item{}
	for itemID, binID := range s.ItemToBinMap {

		var _, isUnderutilized = underutilizedBinIDs[binID]
		if isUnderutilized {
			var item = simCopy.Items[itemID]
			var bin = &simCopy.Bins[binID]
			
			log.Debug("item to be migrated", "item_id", item.ID)

			migratedItems = append(migratedItems, item)
			bin.RemoveItem(item)
			delete(simCopy.ItemToBinMap, item.ID)
		}
	}
	// sort items
	log.Debug("sorting items")
	sort.Sort(item.ByCPUAndMemoryDesc(migratedItems))

	// add items
	log.Debug("adding migration items to simCopy")
	var usedBin *bin.Bin
	for _, itemToMigrate := range migratedItems {
		log.Debug("adding item", "item_ID", itemToMigrate.ID, "cpu", itemToMigrate.CPU, "mem", itemToMigrate.Memory)
		// selecting bin
		var binID = h.RunHeuristic(simCopy, itemToMigrate)

		if binID == len(simCopy.Bins) {
			log.Debug("creating new bin (simcopy)", "bin_ID", binID)
			var newBin = bin.NewBin(len(simCopy.Bins), simCopy.CPU_Capacity, simCopy.Memory_Capacity)
			simCopy.Bins = append(simCopy.Bins, *newBin)
			usedBin = &simCopy.Bins[len(simCopy.Bins)-1]
		} else {
			usedBin = &simCopy.Bins[binID]
		}

		// add item to bin
		if !usedBin.TryAdd(itemToMigrate) {
			return fmt.Errorf("heuristic chose bin %d but item %s does not fit", binID, itemToMigrate.ID)
		}
		simCopy.ItemToBinMap[itemToMigrate.ID] = binID
		log.Debug("item added", "item_cpu", itemToMigrate.CPU, "item_mem", itemToMigrate.Memory, "bin_ID", binID, "remaining_cpu", usedBin.RemainingCPUCapacity(), "remaining_mem", usedBin.RemainingMemoryCapacity())
	}

	// check if the repacked status is better than the current
	var simCopyUsedBins = simCopy.NumberOfBinsUsed()
	var simUsedBins = s.NumberOfBinsUsed()
	if simCopyUsedBins < simUsedBins {
		log.Info("repacking found a better allocation", "current_bin_usage", simUsedBins, "repacking_bin_usage", simCopyUsedBins)
		for _, it := range migratedItems {
			var newBinID = simCopy.ItemToBinMap[it.ID]
			var oldBinID, ok = s.ItemToBinMap[it.ID]

			if !ok || oldBinID != newBinID {
				log.Debug("migrating item", "item_id", it.ID, "from_bin_id", oldBinID, "to_bin_id", newBinID)
				var newBin = simCopy.Bins[newBinID]
				fmt.Printf("%d,migration,%s,%d,%d,%d,%d,%d\n",s.Events[index].Time,it.ID,it.CPU,it.Memory,newBinID,newBin.CPU_Capacity,newBin.Memory_Capacity,)
			}
		}

		s.Bins = simCopy.Bins
		s.ItemToBinMap = simCopy.ItemToBinMap
	} else {
		log.Info("repacking didn't find a better allocation", "current_bin_usage", simUsedBins, "repacking_bin_usage", simCopyUsedBins)
	}

	return nil
}
