package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// User is an example data model
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// In-memory data store
var users = map[string]User{
	"1": {ID: "1", Name: "Alice"},
	"2": {ID: "2", Name: "Bob"},
}

// jsonResponse is a helper function to send JSON responses
func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// HomeHandler: a simple handler for the home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to my API! Your path: %q", html.EscapeString(r.URL.Path))
}

// ListUsersHandler: returns the list of all users
func ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	// Return all users as JSON
	userList := make([]User, 0, len(users))
	for _, u := range users {
		userList = append(userList, u)
	}
	jsonResponse(w, http.StatusOK, userList)
}

// GetUserHandler: returns a single user by ID
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	// Extract path variable (new ServeMux pattern: e.g. GET /users/{id})
	id := r.PathValue("id")
	user, ok := users[id]
	if !ok {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}
	jsonResponse(w, http.StatusOK, user)
}

// CreateUserHandler: creates a new user
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	if u.ID == "" || u.Name == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Missing fields"})
		return
	}
	users[u.ID] = u
	jsonResponse(w, http.StatusCreated, u)
}

// LoggingMiddleware: example middleware that logs requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v\n", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register routes with method and path patterns
	// Patterns (Go 1.22+):
	// - "GET /" matches GET requests on "/"
	// - "/users/" matches all methods on any path starting with "/users/"
	// - "GET /users/{id}" matches GET requests with "/users/<id>"
	// - "POST /users" matches POST requests on "/users"
	mux.Handle("GET /", http.HandlerFunc(HomeHandler))
	mux.Handle("GET /users", http.HandlerFunc(ListUsersHandler))
	mux.Handle("GET /users/{id}", http.HandlerFunc(GetUserHandler))
	mux.Handle("POST /users", http.HandlerFunc(CreateUserHandler))

	// Wrap mux with logging middleware
	loggedMux := LoggingMiddleware(mux)

	// Create a Server with sensible defaults
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      loggedMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		// Idle connections and max header sizes can be tuned as needed
	}

	// Graceful shutdown setup
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Println("Server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on :8080: %v\n", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Finish any ongoing requests
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown failed: %v", err)
	}
	log.Println("Server gracefully stopped.")
}
