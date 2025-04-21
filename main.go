package main

import (
	"log"

	"github.com/rskv-p/mini/mod/m_bus/bus_core"
	"github.com/rskv-p/mini/mod/m_bus/bus_req"
	"github.com/rskv-p/mini/mod/m_bus/bus_type"
	"github.com/rskv-p/mini/mod/m_cfg/cfg_client"
	"github.com/rskv-p/mini/mod/m_db/db_client"
	"github.com/rskv-p/mini/mod/m_log/log_client"
)

func main() {
	clientFactory := func(id uint64, bus bus_type.IBus) bus_type.IBusClient {
		return db_client.NewClient(id, bus.(*bus_core.Bus)) // Используйте вашу реализацию клиента
	}

	// Initialize the bus with the client factory
	bus := bus_core.NewBus(false, "11111", clientFactory)

	// Create the clients
	dbClient := db_client.NewClient(1, bus)
	logClient := log_client.NewClient(2, bus)
	cfgClient := cfg_client.NewClient(3, bus)

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
	cfgClient.HandleMessage = func(req *bus_req.Request) {
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
