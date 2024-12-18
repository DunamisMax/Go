package main

import (
	"context"
	"encoding/json"
	"errors"
	"html"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

// User represents a simple data model for demonstration purposes.
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// In-memory user store protected by a mutex for thread-safe access.
var (
	users = map[string]User{
		"1": {ID: "1", Name: "Alice"},
		"2": {ID: "2", Name: "Bob"},
	}
	usersMu sync.RWMutex
)

func main() {
	mux := http.NewServeMux()

	// Users endpoints
	mux.Handle("GET /users", http.HandlerFunc(listUsersHandler))
	mux.Handle("POST /users", http.HandlerFunc(createUserHandler))
	mux.Handle("GET /users/{id}", http.HandlerFunc(getUserHandler))
	mux.Handle("PUT /users/{id}", http.HandlerFunc(updateUserHandler))
	mux.Handle("DELETE /users/{id}", http.HandlerFunc(deleteUserHandler))

	// A simple home endpoint
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "Welcome to the API!", "path": html.EscapeString(r.URL.Path)})
	}))

	// Wrap with logging middleware
	handler := loggingMiddleware(mux)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Println("Server started on :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	log.Println("Server stopped.")
}

// listUsersHandler returns all users as JSON.
func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	usersMu.RLock()
	defer usersMu.RUnlock()

	list := make([]User, 0, len(users))
	for _, u := range users {
		list = append(list, u)
	}
	writeJSON(w, http.StatusOK, list)
}

// createUserHandler adds a new user.
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	if u.ID == "" || u.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing fields"})
		return
	}

	usersMu.Lock()
	defer usersMu.Unlock()
	if _, exists := users[u.ID]; exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "User ID already exists"})
		return
	}
	users[u.ID] = u
	writeJSON(w, http.StatusCreated, u)
}

// getUserHandler returns a single user by ID.
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	usersMu.RLock()
	defer usersMu.RUnlock()
	u, ok := users[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// updateUserHandler modifies an existing user.
func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	if u.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing name field"})
		return
	}

	usersMu.Lock()
	defer usersMu.Unlock()
	existing, ok := users[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}
	existing.Name = u.Name
	users[id] = existing
	writeJSON(w, http.StatusOK, existing)
}

// deleteUserHandler removes a user by ID.
func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	usersMu.Lock()
	defer usersMu.Unlock()
	if _, ok := users[id]; !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}
	delete(users, id)
	writeJSON(w, http.StatusNoContent, nil)
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// loggingMiddleware logs incoming HTTP requests and their durations.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
