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

var orders = []Order{
	{ID: "o1", Item: "Book", UserID: "u1"},
}

type Order struct {
	ID     string `json:"id"`
	Item   string `json:"item"`
	UserID string `json:"user_id"`
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("GET /orders", http.HandlerFunc(listOrdersHandler))
	mux.Handle("POST /orders", http.HandlerFunc(createOrderHandler))

	srv := &http.Server{
		Addr:         ":8082",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Println("[ordersvc] Starting on :8082")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ordersvc] ListenAndServe error: %v", err)
		}
	}()

	<-stop
	log.Println("[ordersvc] Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[ordersvc] Shutdown error: %v", err)
	}
	log.Println("[ordersvc] Stopped.")
}

func listOrdersHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, orders)
}

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	var o Order
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if o.ID == "" || o.Item == "" || o.UserID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing fields"})
		return
	}
	orders = append(orders, o)
	writeJSON(w, http.StatusCreated, o)
}

func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
