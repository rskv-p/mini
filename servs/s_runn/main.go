package main

import (
	"log"

	"github.com/rskv-p/mini/servs/s_runn/runn_api"
	"github.com/rskv-p/mini/servs/s_runn/runn_cfg"
	"github.com/rskv-p/mini/servs/s_runn/runn_serv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	if err := runn_cfg.Load(""); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to the database
	db, err := gorm.Open(sqlite.Open(runn_cfg.C().DBPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Initialize the process manager
	manager, err := runn_serv.New(db)
	if err != nil {
		log.Fatalf("Failed to initialize the process manager: %v", err)
	}

	// Load preconfigured processes into the database
	if err := runn_serv.LoadPreconfiguredProcesses(db); err != nil {
		log.Fatalf("Error loading preconfigured processes: %v", err)
	}

	// Start the REST API in a goroutine to avoid blocking the main thread
	go func() {
		runn_api.ServeREST(runn_cfg.C().HTTPAddress, manager)
	}()

	// Block the main thread to keep the server running
	select {} // Blocks the program while the server is running
}
