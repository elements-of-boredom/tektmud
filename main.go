package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"tektmud/internal/server"
)

func main() {

	slog.SetLogLoggerLevel(slog.LevelDebug)
	//create our server
	server, err := server.NewMudServer("_data/config.yaml")
	if err != nil {
		slog.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	//Initialize the server
	if err := server.Initialize(); err != nil {
		slog.Error("Failed to initialize server", "error", err)
		os.Exit(1)
	}

	//Start the server to begin accepting connections
	if err := server.Start(); err != nil {
		slog.Error("Failed to start listners", "error", err)
		os.Exit(1)
	}

	//Listen for any manual shutdown events to our service
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	//Block how until shutdown issued
	<-sigChan
	slog.Warn("Shutdown request received")

	//Give our pieces time to shutdown gracefully
	if err := server.Shutdown(); err != nil {
		slog.Error("Error while attempting to shutdown", "error", err)
		os.Exit(1)
	}
	os.Exit(0)

}
