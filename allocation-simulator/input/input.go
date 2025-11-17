package input

import (
	"allocation-simulator/item"
	"allocation-simulator/logger"
	"allocation-simulator/operation"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

// Represents the events of insertion and deletion of items in our simulation
// The simulator's input
type Event struct {
	Item      item.Item
	Time      int
	Operation operation.OperationType
}

// This function receives a CSV file with the format:
// timestamp,id,cpu,memory,appName,operation
// And returns a slice of events, sorted and ready to be used by the simulator.
func LoadEventsFromCSV(filename string) ([]Event, error) {
	var log = logger.Module("input")
	
	file, err := os.Open(filename)
	if err != nil {
		log.Error("error opening file", "file", filename, "error", err)
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	_, err = reader.Read()
	if err != nil {
		log.Error("error reading header", "file", filename, "error", err)
		return nil, fmt.Errorf("error reading header: %v", err)
	}

	var events []Event
	lineNumber := 1

	for {
		lineNumber++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error("error reading CSV line", "line", lineNumber, "error", err)
			return nil, fmt.Errorf("error reading CSV line %d: %v", lineNumber, err)
		}

		time, err := strconv.Atoi(record[0])
		if err != nil {
			log.Error("invalid 'timestamp'", "line", lineNumber, "raw", record[0], "error", err)
			return nil, fmt.Errorf("error parsing 'timestamp' on line %d: %v", lineNumber, err)
		}

		id := record[1]

		cpu, err := strconv.ParseInt(record[2], 10, 64)
		if err != nil {
			log.Error("invalid 'cpu'", "line", lineNumber, "raw", record[2], "error", err)
			return nil, fmt.Errorf("error parsing 'cpu' on line %d: %v", lineNumber, err)
		}

		memory, err := strconv.ParseInt(record[3], 10, 64)
		if err != nil {
			log.Error("invalid 'memory'", "line", lineNumber, "raw", record[3], "error", err)
			return nil, fmt.Errorf("error parsing 'memory' on line %d: %v", lineNumber, err)
		}

		appName := record[4]
		operationStr := record[5]

		var opType operation.OperationType
		if operationStr == "insertion" {
			opType = operation.Insertion
		} else if operationStr == "deletion" {
			opType = operation.Deletion
		} else {
			log.Error("invalid operation type", "line", lineNumber, "raw", operationStr)
			return nil, fmt.Errorf("invalid operation type '%s' on line %d", operationStr, lineNumber)
		}

		event := Event{
			Item: item.Item{
				ID:              id,
				CPU:             cpu,
				Memory:          memory,
				ApplicationName: appName,
			},
			Time:      time,
			Operation: opType,
		}

		events = append(events, event)
	}

	log.Info("CSV loaded", "file", filename, "len_events", len(events))
	log.Info("initializing sorting")

	sort.Slice(events, func(i, j int) bool {
		if events[i].Time != events[j].Time {
			return events[i].Time < events[j].Time
		}

		if events[i].Operation != events[j].Operation {
			return events[i].Operation > events[j].Operation
		}

		if events[i].Item.CPU != events[j].Item.CPU {
			return events[i].Item.CPU > events[j].Item.CPU
		}

		if events[i].Item.Memory != events[j].Item.Memory {
			return events[i].Item.Memory > events[j].Item.Memory
		}

		return events[i].Item.ID > events[j].Item.ID
	})

	log.Info("CSV sorted")

	return events, nil
}
