package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// Simple in-memory user store
var users = map[string]string{
	"u1": "Alice",
	"u2": "Bob",
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("GET /users", http.HandlerFunc(listUsersHandler))
	mux.Handle("POST /users", http.HandlerFunc(createUserHandler))

	srv := &http.Server{
		Addr:         ":8081",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Println("[usersvc] Starting on :8081")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[usersvc] ListenAndServe error: %v", err)
		}
	}()

	<-stop
	log.Println("[usersvc] Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[usersvc] Shutdown error: %v", err)
	}
	log.Println("[usersvc] Stopped.")
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	list := []User{}
	for id, name := range users {
		list = append(list, User{ID: id, Name: name})
	}
	writeJSON(w, http.StatusOK, list)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if u.ID == "" || u.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing fields"})
		return
	}
	users[u.ID] = u.Name
	writeJSON(w, http.StatusCreated, u)
}

func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
