package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

// User is an example data model
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// In-memory data store (replace with a database for production)
var users = map[string]User{
	"1": {ID: "1", Name: "Alice"},
	"2": {ID: "2", Name: "Bob"},
}

// logger is a global structured logger for demonstration.
// In production, configure properly or inject as dependency.
var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

func main() {
	addr := getServerAddress()

	mux := http.NewServeMux()

	// Register a wildcard route to catch everything else for 404s:
	mux.Handle("/", http.HandlerFunc(notFoundHandler))

	// Register routes with method and path patterns (Go 1.22+)
	mux.Handle("GET /", http.HandlerFunc(HomeHandler))
	mux.Handle("GET /users", http.HandlerFunc(ListUsersHandler))
	mux.Handle("GET /users/{id}", http.HandlerFunc(GetUserHandler))
	mux.Handle("POST /users", http.HandlerFunc(CreateUserHandler))

	// Health check endpoint
	mux.Handle("GET /healthz", http.HandlerFunc(HealthHandler))

	// Wrap mux with middleware: logging, security, etc.
	handler := LoggingMiddleware(mux)
	handler = SecurityMiddleware(handler) // Add security headers, CORS, etc.

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		// Adjust IdleTimeout, MaxHeaderBytes, TLS config as needed
	}

	// Graceful shutdown setup
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		logger.Info("Server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("ListenAndServe error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	<-stop
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Finish any ongoing requests
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", "error", err)
	} else {
		logger.Info("Server gracefully stopped.")
	}
}

// HomeHandler: a simple handler for the home page
// Demonstrates a basic HTML page that can integrate with htmx for dynamic updates.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Example: if the request is from htmx (hx-boost, hx-get, hx-post),
	// we could return partial HTML. Otherwise, return full HTML.

	// For now, just return a simple page with a form to create a new user.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>My API Home</title></head>
<body>
<h1>Welcome to my API!</h1>
<p>Your path: %q</p>
<form id="create-user-form" hx-post="/users" hx-target="#result" hx-swap="afterend">
  <input type="text" name="id" placeholder="User ID" required>
  <input type="text" name="name" placeholder="User Name" required>
  <button type="submit">Create User</button>
</form>
<div id="result"></div>
</body>
</html>`, html.EscapeString(r.URL.Path))
}

// HealthHandler: a health check endpoint
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListUsersHandler: returns the list of all users
func ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowedHandler(w)
		return
	}
	userList := make([]User, 0, len(users))
	for _, u := range users {
		userList = append(userList, u)
	}
	// Optionally add caching or ETag headers here.
	jsonResponse(w, http.StatusOK, userList)
}

// GetUserHandler: returns a single user by ID
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowedHandler(w)
		return
	}

	id := r.PathValue("id")
	user, ok := users[id]
	if !ok {
		jsonError(w, http.StatusNotFound, "User not found")
		return
	}
	jsonResponse(w, http.StatusOK, user)
}

// CreateUserHandler: creates a new user
// Supports both JSON and form-encoded input. For form submission (like from HTML forms or htmx), we parse form values.
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowedHandler(w)
		return
	}

	var u User
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			jsonError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
	} else {
		// Handle form data
		if err := r.ParseForm(); err != nil {
			jsonError(w, http.StatusBadRequest, "Unable to parse form")
			return
		}
		u.ID = r.Form.Get("id")
		u.Name = r.Form.Get("name")
	}

	if u.ID == "" || u.Name == "" {
		jsonError(w, http.StatusBadRequest, "Missing fields: both 'id' and 'name' required")
		return
	}

	// Check if user already exists
	if _, exists := users[u.ID]; exists {
		jsonError(w, http.StatusConflict, "User ID already exists")
		return
	}

	users[u.ID] = u
	jsonResponse(w, http.StatusCreated, u)
}

// notFoundHandler: custom handler for 404 Not Found
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	jsonError(w, http.StatusNotFound, "Not found")
}

// methodNotAllowedHandler: custom handler for 405 Method Not Allowed
func methodNotAllowedHandler(w http.ResponseWriter) {
	jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

// LoggingMiddleware: logs requests with structured logging
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rr := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rr, r)
		duration := time.Since(start)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rr.statusCode,
			"duration", duration.String(),
			"client_ip", r.RemoteAddr,
		)
	})
}

// SecurityMiddleware: adds basic security headers and CORS policies.
// Customize as needed for production security requirements.
func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Basic security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// CORS example (adjust as needed)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// jsonResponse is a helper function to send JSON responses with status codes
func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.Error("Failed to encode JSON response", "error", err)
		}
	}
}

// jsonError writes a JSON error message with a given status code
func jsonError(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

// getServerAddress returns the server address from PORT env var or default ":8080"
func getServerAddress() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}

// responseRecorder is used in LoggingMiddleware to capture the status code
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}
