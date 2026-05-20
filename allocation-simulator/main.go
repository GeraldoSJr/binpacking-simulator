package main

import (
	heuristic "allocation-simulator/allocation_heuristics"
	"allocation-simulator/input"
	"allocation-simulator/logger"
	repacking "allocation-simulator/repacking_heuristics"
	"allocation-simulator/simulator"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func main() {
	// Default parameter values (CPU and memory defaults are derived from input)
	var (
		defBinMin     float64 = 0.70       // defines the bin occupation limit lower bound (bins less than x% are selected for repacking)
		defRepack     bool    = true        // toggles repacking
		defEventBased bool    = true        // toggles event-based repacking: true -> event based repacking, false -> timestamp based repacking
		defHeuristic  string  = "firstfit" // options: 'firstfit' or 'bestfit'
		defRepackHeur string  = "r1"       // options: 'r1' or 'optimal'
	)

	// Initialize logger
	var err = logger.Init(slog.LevelInfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL ERROR: could not init logger: %v\n", err)
		os.Exit(1)
	}
	var log = logger.Module("main")

	// Require <input.csv> as first positional argument
	if len(os.Args) < 2 || os.Args[1][0] == '-' {
		fmt.Fprintf(os.Stderr, "Usage: go run . <input.csv> [optional flags]\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fmt.Fprintf(os.Stderr, "  -CPUCapacity int64            (default: max CPU value found in input)\n")
		fmt.Fprintf(os.Stderr, "  -MemCapacity int64            (default: max memory value found in input)\n")
		fmt.Fprintf(os.Stderr, "  -BinMinLimit float64          (default 0.70)\n")
		fmt.Fprintf(os.Stderr, "  -RepackingEnabled bool        (default true)\n")
		fmt.Fprintf(os.Stderr, "  -EventBased bool              (default true)\n")
		fmt.Fprintf(os.Stderr, "  -Heuristic string             'bestfit'|'firstfit' (default 'firstfit')\n")
		fmt.Fprintf(os.Stderr, "  -RepackingHeuristic string    'r1'|'optimal' (default 'r1')\n")
		log.Error("invalid parameters", "args", os.Args)
		os.Exit(2)
	}
	var csvFilePath string = os.Args[1]

	// Load events first so we can derive default CPU/memory capacities from the input
	log.Info("loading CSV events", "file", csvFilePath)
	var events []input.Event
	events, err = input.LoadEventsFromCSV(csvFilePath)
	if err != nil {
		log.Error("could not load events", "file", csvFilePath)
		fmt.Fprintf(os.Stderr, "FATAL ERROR: Could not process event file: %v. Check logs for more details.\n", err)
		os.Exit(1)
	}

	// Derive default bin capacities from the maximum CPU and memory values across all events
	var defCPU, defMem int64
	for _, e := range events {
		if e.CPU > defCPU {
			defCPU = e.CPU
		}
		if e.Memory > defMem {
			defMem = e.Memory
		}
	}
	log.Info("derived default capacities from input", "default_cpu", defCPU, "default_mem", defMem)

	// Define FlagSet for arguments appearing after <input.csv>
	var fs = flag.NewFlagSet("alloc-sim", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var cpuCap = fs.Int64("CPUCapacity", defCPU, "CPU capacity (int64)")
	var memCap = fs.Int64("MemCapacity", defMem, "Memory capacity (int64)")
	var binMin = fs.Float64("BinMinLimit", defBinMin, "Minimum occupation limit (float64)")
	var repackEnabled = fs.Bool("RepackingEnabled", defRepack, "Enable repacking (bool)")
	var eventBased = fs.Bool("EventBased", defEventBased, "Event-based repacking (bool)")
	var heuristicName = fs.String("Heuristic", defHeuristic, "Allocation heuristic: 'bestfit'|'firstfit'")
	var repackHeuristicName = fs.String("RepackingHeuristic", defRepackHeur, "Repacking heuristic: 'r1'|'optimal'")

	err = fs.Parse(os.Args[2:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Usage: go run . <input.csv> [optional flags]\n")
		log.Error("parameters parsing error", "error", err)
		os.Exit(2)
	}

	// Resolved values (flags or defaults)
	var CPU_Capacity int64 = *cpuCap
	var Memory_Capacity int64 = *memCap
	var MinOccupationLimit float64 = *binMin
	var RepackingEnabled bool = *repackEnabled
	var EventBasedRepacking bool = *eventBased

	// select heuristic by string argument
	var h simulator.Heuristic
	var hName string = strings.ToLower(strings.TrimSpace(*heuristicName))
	switch hName {
	case "firstfit":
		h = heuristic.NewFirstFit()
	case "bestfit":
		h = heuristic.NewBestFit()
	default:
		log.Error("unknown allocation heuristic", "value", *heuristicName)
		fmt.Fprintf(os.Stderr, "Unknown heuristic provided.\n")
		os.Exit(2)
	}

	// select repacking heuristic
	var r simulator.Repacking
	var repName string = strings.ToLower(strings.TrimSpace(*repackHeuristicName))
	var RunOptimal bool
	switch repName {
	case "optimal":
		r = repacking.NewOptimal()
		RunOptimal = true
	case "r1":
		r = repacking.NewR1()
		RunOptimal = false
	default:
		log.Error("unknown repacking heuristic", "value", *repackHeuristicName)
		fmt.Fprintf(os.Stderr, "Unknown repacking heuristic provided.\n")
		os.Exit(2)
	}

	var sim = simulator.NewSimulator(
		events,
		CPU_Capacity,
		Memory_Capacity,
		MinOccupationLimit,
		RepackingEnabled,
		EventBasedRepacking,
		RunOptimal,
	)

	// Run simulation
	log.Info("running simulation", 
		"cpu_cap", CPU_Capacity, 
		"mem_cap", Memory_Capacity, 
		"occupation_limit", MinOccupationLimit,
		"repacking_enabled", RepackingEnabled,
		"event_based", EventBasedRepacking,
		"heuristic", hName,
		"repacking_heuristic", repName,
	)

	err = sim.RunSimulation(h, r)
	if err != nil {
		log.Error("simulation failed")
		fmt.Fprintf(os.Stderr, "FATAL ERROR: simulation failed: %v Check logs for more details.\n", err)
		os.Exit(1)
	}

	log.Info("simulation finished successfully", "file", csvFilePath)
}
