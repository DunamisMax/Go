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

// The gateway service fetches data from both the user and order services.
// GET /all -> returns a combined JSON: { "users": [...], "orders": [...] }

func main() {
	mux := http.NewServeMux()
	mux.Handle("GET /all", http.HandlerFunc(allHandler))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Println("[gateway] Starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[gateway] ListenAndServe error: %v", err)
		}
	}()

	<-stop
	log.Println("[gateway] Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[gateway] Shutdown error: %v", err)
	}
	log.Println("[gateway] Stopped.")
}

func allHandler(w http.ResponseWriter, r *http.Request) {
	users, err := fetchJSON("http://localhost:8081/users")
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to fetch users"})
		return
	}
	orders, err := fetchJSON("http://localhost:8082/orders")
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to fetch orders"})
		return
	}

	resp := map[string]any{
		"users":  users,
		"orders": orders,
	}
	writeJSON(w, http.StatusOK, resp)
}

func fetchJSON(url string) (any, error) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var v any
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
