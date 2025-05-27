package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	configs "tektmud/internal/config"
	"tektmud/internal/logger"
	"tektmud/internal/server"
)

func main() {

	//TODO Pull from ENV Vars
	c, err := configs.LoadConfig("_data/config.yaml")
	if err != nil {
		slog.Error("failed to find config at:", "path", "_data/config.yaml", "error", err)
		os.Exit(1)
	}

	//Initialize Logging system
	if err := logger.InitGlobalLogger(); err != nil {
		slog.Error("failed to initalize logger:", "error", err)
		os.Exit(1)
	}

	// Set up panic recovery
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			logger.GetLogger().LogPanic(r, stack)
			fmt.Printf("PANIC: %v\n", r)
		}
	}()

	version := "1.0.0"
	systemConfig := map[string]any{
		"version":   version,
		"log_level": c.Logging.Level,
		"tick_rate": "150ms",
		"data_path": "./data/areas",
	}
	logger.LogSystemStart(version, systemConfig)

	//create our server
	server, err := server.NewMudServer()
	if err != nil {
		logger.Error("failed to create server", "error", err)
		os.Exit(1)
	}

	//Initialize the server
	if err := server.Initialize(); err != nil {
		logger.Error("Failed to initialize server", "error", err)
		os.Exit(1)
	}

	//Start the server to begin accepting connections
	if err := server.Start(); err != nil {
		logger.Error("Failed to start listners", "error", err)
		os.Exit(1)
	}

	//Listen for any manual shutdown events to our service
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	//Block how until shutdown issued
	<-sigChan
	logger.Warn("Shutdown request received")

	//Give our pieces time to shutdown gracefully
	if err := server.Shutdown(); err != nil {
		logger.Error("Error while attempting to shutdown", "error", err)
		os.Exit(1)
	}
	os.Exit(0)

}
