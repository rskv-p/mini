package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/srv"
)

func main() {
	// Register built-in modules
	// This step will register all necessary modules such as m_api, m_bus, m_cfg, etc.
	mod.RegisterBuiltinModules()

	// Create a new service and add modules to it
	// The service is initialized with a name and a collection of modules.
	service := &srv.Service{
		Name: "MyService", // Name of your service
	}

	// Get all modules using the GetModules function and add them to the service
	modules := mod.GetModules()

	// Add all the retrieved modules to the service
	// This allows the service to dynamically load modules for usage
	for _, module := range modules {
		service.AddModule(module)
	}

	// Start the service
	// This will initialize and start all modules registered in the service
	if err := service.Start(); err != nil {
		// If there is an error during service startup, log the error and stop the program
		log.Fatalf("Error starting service: %v", err)
	}

	// Set up and start an HTTP server to handle incoming requests
	// The server listens on port 8080 and handles requests at the /api/ route
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		// API request handler that responds with service and module information
		fmt.Fprintf(w, "Service: %s, Modules Loaded: %v\n", service.Name, service.Modules)
	})

	// Start the HTTP server on port 8080
	// The server will handle API requests and serve responses accordingly
	log.Println("Starting API server on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		// If there is an error starting the HTTP server, log the error and stop the program
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
