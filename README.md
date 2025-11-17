# Dynamic Bin Packing Simulator in Go

This repository contains a Go implementation of a simulator for the dynamic bin packing problem. Its goal is to allow running experiments and analyzing different allocation and consolidation heuristics in an environment where items can be added and removed over time.

## 📦 About Dynamic Bin Packing and Metric Collection

The **Dynamic Bin Packing** problem is a variation of the classic packing problem, where a sequence of items of different sizes must be allocated into a minimum number of fixed-capacity containers (bins). The "dynamic" nature implies that items not only arrive (are added) but also depart (are removed), which can lead to space fragmentation and underutilization of bins.

The simulator also supports items and bins with two capacity dimensions, CPU and memory, as these are some of the most important measures in allocation decisions within the Kubernetes context.

## Input

The simulator's input is a file containing information about item arrival and departure times, as well as their sizes in each dimension and their assigned name (in the Kubernetes context, the name refers to the application the pod represents). The CSV must follow the format:

timestamp,id,cpu,memory,appName,operation

## Output

This simulator is designed to model this scenario and collect essential metrics to evaluate the efficiency of the strategies used. The main output of the program is a CSV file that each line represents a event processed in the format.

timestamp,operation,itemID,itemCPU,itemMemory,binID,binCPU,binMemory

The value of the `operation` column can be one of three types: `insertion`, `deletion`, or `migration`. In this context, migration means moving an already allocated item from one bin to another, usually to improve resource utilization or reduce waste.

## 🧠 Packing Heuristics

The item allocation strategy implemented in this simulator are **First-Fit-Decreasing** and **Best-Fit-Decreasing**.

### First-Fit-Decreasing

When a new set of items needs to be allocated, the heuristic works as follows:

1. Sort the items once at the start in descending order by CPU size, using memory as a tiebreaker.
2. For each item in that sorted list (in order):
    a. Iterate through the existing bins in the order they were created.
    b. Place the item into the first bin that has sufficient remaining capacity.
    c. If no existing bin can accommodate the item, create a new bin and place the item there.

It is a simple, fast heuristic that serves as an excellent baseline for comparison with more complex strategies.

### Best-Fit-Decreasing

When a new set of items needs to be allocated, the heuristic works as follows:

1. Sort the items once at the start in descending order by CPU size, using memory as a tiebreaker.

2. For each item in the sorted list:
    a. Examine all existing bins and identify those that can accommodate the item.
    b. Among these, select the bin with the least remaining capacity after placing the item.
    c. Place the item into that bin.
    d. If no existing bin can accommodate the item, create a new bin and place the item there.

This approach follows a different placement strategy from First-Fit-Decreasing, focusing on selecting the bin that will be left with the least remaining capacity after each allocation. It can lead to different packing patterns and may be preferable in scenarios where tighter local utilization of individual bins is desired.

## 🔄 Repacking Heuristics

### Threshold-based repacking

It uses a bin occupation limit (default: 70%).

For each timestamp with at least one event, after the event is processed, it goes through the bins looking for those below the occupation limit.

Once it finds the underutilized bins, it removes all items from them and tries to reallocate these items into the other bins (using the same allocation heuristic as the rest of the simulation).

When this reallocation is done, it compares the new cluster wastage with the main simulation cluster. If the repacking leads to better results, the migrations are applied to the main cluster.

### Optimal-Fit

Optimal-Fit is an offline allocation strategy that seeks a configuration using the minimum number of bins.

It explores a decision tree while maintaining the best allocation found so far, and returns that solution once the search completes.

To make the process tractable, several optimizations and pruning techniques are applied to reduce the search space and improve performance.

## ⚙️ How to Run Experiments

1. Inside the `allocation-simulator` directory, run:

```bash
go run main.go path/to/input_file.csv > path/to/output_file.csv
```

**What can be changed**
You can modify the bin capacities (CPU and memory), the minimum occupation threshold used by the repacking heuristic, whether repacking is enabled, how often it is triggered (after each event or per timestamp), which allocation heuristic is used (*firstfit* or *bestfit*), and which repacking heuristic is used (*R1* or *Optimal*).

These settings can be adjusted either temporarily through command-line flags or permanently by editing the code.

### Changing configuration using flags

```bash
go run . <input.csv> [flags]
```

All flags are optional (defaults in parentheses):

* **-CPUCapacity** int64 (`16`)
* **-MemCapacity** int64 (`16`)
* **-BinMinLimit** float64 (`0.70`)
* **-RepackingEnabled** bool (`true`)
* **-EventBased** bool (`true`)
* **-Heuristic** string `bestfit` | `firstfit` (`firstfit`)
* **-RepackingHeuristic** string `r1` | `optimal` (`r1`)

Examples:

```bash
# Minimal run
go run . input.csv

# Best-Fit-Decreasing with timestamp-based repacking
go run . input.csv -Heuristic=bestfit -EventBased=false

# Custom capacities and threshold
go run . input.csv -CPUCapacity=6000 -MemCapacity=12884901888 -BinMinLimit=0.6

# Running with Optimal repacking
go run . input.csv -RepackingHeuristic=optimal
```

**Important notes:**

* The `optimal` repacking heuristic **requires** both `-RepackingEnabled=true` **and** `-EventBased=true`.
  If either is disabled, the simulator will stop with an error.
* The `optimal` mode **ignores the occupation limit** (`-BinMinLimit`), since it searches for the mathematically best packing configuration instead of using thresholds.
* For boolean flags, always use `=` when setting `false` (e.g., `-EventBased=false`).
* The `r1` repacking heuristic is the internal name for our threshold-based repacking.

### Changing configuration in code

Default values are defined at the top of `main.go`:

```go
var (
    defCPU          int64   = 16
    defMem          int64   = 16
    defBinMin       float64 = 0.70
    defRepack       bool    = true
    defEventBased   bool    = true
    defHeuristic    string  = "firstfit"
    defRepackHeur   string  = "r1"
)
```

---

## 🔧 Debugging

### Logs

The simulator includes a structured logging system built on top of Go's `slog` package.  
Logs are written in **JSON format** to a file automatically created under the `logs` directory. Each execution generates a new file, with a name based on timestamp, process ID, and nanoseconds for uniqueness:

```bash
logs/20250930-133616-66702-985337235.log
```

#### Log Levels

The log verbosity is controlled when initializing the logger in `main.go`:

```go
var err = logger.Init(slog.LevelInfo)
```

You can switch between different levels (`Debug`, `Info`, `Warn`, `Error`) to increase or decrease the amount of detail recorded.

For example, `Debug` includes detailed messages about item placement and bin usage, while `Info` records high-level events such as simulation start, CSV loading, or repacking results.

By default, logs are always enabled and set to `Info`. If you want to quickly check only the simulation results (CSV output), you can either ignore the log files or set the log level to `Error`. When debugging or analyzing heuristic behavior, the logs provide fine-grained visibility into allocation, repacking, and event handling.

---
