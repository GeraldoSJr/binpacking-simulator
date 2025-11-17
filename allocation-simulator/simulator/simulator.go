package simulator

import (
	"allocation-simulator/bin"
	"allocation-simulator/input"
	"allocation-simulator/item"
	"allocation-simulator/logger"
	"fmt"
)

type Heuristic interface {
	HandleEvent(s *Simulator, eventIndex int) error
	RunHeuristic(s *Simulator, item item.Item) int
}

type Repacking interface {
	HandleEvent(s *Simulator, eventIndex int, h Heuristic, occupationLimit float64) error
}

// In this version of the simulator, the bins are all the same size.
type Simulator struct {
	Events          []input.Event
	Bins            []bin.Bin
	Items           map[string]item.Item
	ItemToBinMap    map[string]int

	CPU_Capacity         int64
	Memory_Capacity      int64
	MinOccupationLimit   float64
	RepackingEnabled     bool
	EventBasedRepacking  bool // true -> event based repacking, false -> timestamp based repacking
	RunOptimal           bool
}

func NewSimulator(events []input.Event, cpuCap int64, memCap int64, occLimit float64, repEnabled bool, eventRep bool, runOptimal bool) *Simulator {
	return &Simulator{
		Events:          events,
		Bins:            []bin.Bin{},
		Items:           make(map[string]item.Item),
		ItemToBinMap:    make(map[string]int),

		CPU_Capacity:        cpuCap,
		Memory_Capacity:     memCap,
		MinOccupationLimit:  occLimit,
		RepackingEnabled:    repEnabled,
		EventBasedRepacking: eventRep,
		RunOptimal:          runOptimal,
	}
}

// The simulator receives a events slice and iterates over it, calling a heuristic to process each event.
// Repacking can be executed with two different frequencies:
// - mode 1: repacking at every event.
// - mode 2: repacking at the end of every received timestamp.
func (s *Simulator) RunSimulation(h Heuristic, r Repacking) error {
	var log = logger.Module("simulator")

	if len(s.Events) == 0 {
		log.Error("no events available for simulation")
		return fmt.Errorf("no events available for simulation")
	}

	if s.RunOptimal {
		if !s.RepackingEnabled {
			log.Error("optimal couldn't run because repacking was disabled")
			return fmt.Errorf("optimal couldn't run because repacking was disabled")
		}

		if !s.EventBasedRepacking {
			log.Error("optimal couldn't run because event-based repacking was disabled")
			return fmt.Errorf("optimal couldn't run because event-based repacking was disabled")
		}
	}

	for i := range s.Events {
		var event = s.Events[i]
		log.Info("handling event", "tm", event.Time, "op", event.Operation.String(), "item_id", event.Item.ID)
		
		var err error

		if s.RunOptimal {
			// if optimal solution is required, heuristic and further repacking is not needed.
			err = r.HandleEvent(s, i, h, s.MinOccupationLimit)
			if err != nil {
				return err
			}
			continue
		}

		// runs heuristic
		err = h.HandleEvent(s, i)
		if err != nil {
			return err
		}

		if s.RepackingEnabled {
			// mode 1: event based repacking
			if s.EventBasedRepacking {
				log.Info("calling repacking")
				err = r.HandleEvent(s, i, h, s.MinOccupationLimit)
				if (err != nil) {
					return err
				}

				continue
			}

			// mode 2: timestamp based repacking
			var isLastEvent = (i == len(s.Events)-1)
			if isLastEvent || s.Events[i+1].Time != event.Time {
				log.Info("calling repacking")
				err = r.HandleEvent(s, i, h, s.MinOccupationLimit)
				if (err != nil) {
					return err
				}
			}
		}
	}

	log.Info("simulation finished successfully")

	return nil
}

func (s *Simulator) NumberOfBinsUsed() int {
	count := 0
	for _, b := range s.Bins {
		if !b.Closed && !b.Empty() {
			count++
		}
	}
	return count
}

// Used by repacking to create a copy of the simulator to simulate the migration operations
func (s *Simulator) DeepCopy() *Simulator {
	var simCopy = &Simulator{
		CPU_Capacity:        s.CPU_Capacity,
		Memory_Capacity:     s.Memory_Capacity,
		MinOccupationLimit:  s.MinOccupationLimit,
		RepackingEnabled:    s.RepackingEnabled,
		EventBasedRepacking: s.EventBasedRepacking,
		RunOptimal:          s.RunOptimal,
	}

	simCopy.Events = make([]input.Event, len(s.Events))
	copy(simCopy.Events, s.Events)

	simCopy.Bins = make([]bin.Bin, len(s.Bins))
	copy(simCopy.Bins, s.Bins)

	simCopy.Items = make(map[string]item.Item, len(s.Items))
	for k, v := range s.Items {
		simCopy.Items[k] = v
	}

	simCopy.ItemToBinMap = make(map[string]int, len(s.ItemToBinMap))
	for k, v := range s.ItemToBinMap {
		simCopy.ItemToBinMap[k] = v
	}

	return simCopy
}