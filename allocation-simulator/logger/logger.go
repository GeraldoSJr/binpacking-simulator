package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var Logger *slog.Logger

// Init creates a logger that writes to a file under <project-root>/logs.
func Init(level slog.Leveler) error {
	// resolve <project-root>/logs using the source path of this file
	var _, thisFile, _, ok = runtime.Caller(0)
	if !ok {
		return fmt.Errorf("logger: cannot resolve source file path")
	}
	// thisFile -> .../<project-root>/allocation-simulator/logger/logger.go
	var moduleDir = filepath.Dir(filepath.Dir(thisFile)) // .../allocation-simulator
	var projectRoot = filepath.Dir(moduleDir)            // .../
	var logDir = filepath.Join(projectRoot, "logs")

	// ensure log directory exists
	var err = os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		return err
	}

	// create a new file for each execution (timestamp + pid + nanos)
	var filename = fmt.Sprintf("%s/%s-%d-%d.log",
		logDir,
		time.Now().UTC().Format("20060102-150405"),
		os.Getpid(),
		time.Now().UTC().Nanosecond(),
	)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	// structured logs (JSON).
	var handler slog.Handler = slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: level,
	})

	Logger = slog.New(handler)
	return nil
}

// Module returns a logger with a module tag
func Module(name string) *slog.Logger {
	if Logger == nil {
		fmt.Fprintf(os.Stderr, "Logger has not been initialized yet!")
		os.Exit(1)
	}
	return Logger.With(slog.String("module", name))
}