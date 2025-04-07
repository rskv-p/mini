package runn_api

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rskv-p/mini/servs/s_runn/runn_serv"
)

// ServeREST starts the HTTP API server.
func ServeREST(addr string, manager *runn_serv.Service) {
	// Initialize JWT key for authentication
	InitAuth()

	// Create a new router
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Public endpoints
	r.Post("/auth/login", HandleLogin(manager.DB()))       // Login
	r.Post("/auth/register", HandleRegister(manager.DB())) // Register new user

	// Protected API endpoints
	r.Group(func(r chi.Router) {
		r.Use(JWTMiddleware("admin")) // Apply JWTMiddleware for admin users

		// Process management routes
		r.Route("/api", func(r chi.Router) {
			r.Get("/processes", handleList(manager))                         // List processes
			r.Post("/start/{id}", handleStart(manager))                      // Start a process
			r.Post("/stop/{id}", handleStop(manager))                        // Stop a process
			r.Post("/pause/{id}", handlePause(manager))                      // Pause a process
			r.Post("/resume/{id}", handleResume(manager))                    // Resume a process
			r.Post("/add", handleAdd(manager))                               // Add a new process
			r.Delete("/remove/{id}", handleRemove(manager))                  // Remove a process
			r.Post("/api/start-group/{group_id}", handleStartGroup(manager)) // Start a group of processes
			r.Post("/edit/{id}", handleEdit(manager))                        // Edit a process
		})
	})

	// Protected WebSocket with JWT Middleware
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Wrap HandlerWS with JWTMiddleware
		JWTMiddleware("admin")(http.HandlerFunc(HandleWS)).ServeHTTP(w, r)
	})

	// Start the server
	log.Printf("REST API listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
