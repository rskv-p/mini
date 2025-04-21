package main

import (
	"log"

	"github.com/rskv-p/mini/pkg/x_bus"
	"github.com/rskv-p/mini/pkg/x_cfg"
	"github.com/rskv-p/mini/pkg/x_db"
	"github.com/rskv-p/mini/pkg/x_log"
	"github.com/rskv-p/mini/pkg/x_req"
)

func main() {
	// Initialize the bus
	bus := x_bus.NewBus(false, "11111")

	// Create the clients
	dbClient := x_db.NewClient(1, bus)
	logClient := x_log.NewClient(2, bus)
	cfgClient := x_cfg.NewClient(3, bus)

	// Set up message handlers for each client to process messages
	// Subscribing to the correct subjects only
	err := dbClient.Subscribe("db.create") // DB client only subscribes to db.create
	if err != nil {
		log.Printf("Error subscribing DB client: %v", err)
	}

	err = logClient.Subscribe("log.message") // Log client only subscribes to log.message
	if err != nil {
		log.Printf("Error subscribing Log client: %v", err)
	}

	err = cfgClient.Subscribe("config.update") // Config client only subscribes to config.update
	if err != nil {
		log.Printf("Error subscribing Config client: %v", err)
	}

	// Define custom message handlers
	cfgClient.HandleMessage = func(req *x_req.Request) {
		log.Printf("Custom handler for Config Client - Subject: %s, Data: %s", req.Subject, string(req.Data))
	}

	// Publish messages to simulate actions
	log.Println("Publishing messages...")

	// Simulate DB create operation
	err = dbClient.Create("some_model", "model1")
	if err != nil {
		log.Printf("Error in DB create operation: %v", err)
	}

	// Simulate Log message publishing
	logClient.Log("INFO", "Service started")

	// Simulate Config update operation
	err = cfgClient.SetConfig("log_level", "DEBUG")
	if err != nil {
		log.Printf("Error in Config update operation: %v", err)
	}

	// Call HandleSubscribe to simulate the processing of messages
	dbClient.HandleSubscribe()
	logClient.HandleSubscribe()
	cfgClient.HandleSubscribe()
}
