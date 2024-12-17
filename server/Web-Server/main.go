package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// helloHandler responds with a simple greeting.
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// It's often best practice to handle your logic simply and clearly.
	// Here we just write a greeting, but in a real app you might parse
	// query params, perform business logic, etc.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "Hello, world!")
}

func main() {
	// Set up our routes in a clear, centralized manner.
	http.HandleFunc("/", helloHandler)

	// Create a server with sensible timeouts.
	srv := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 5 * time.Second,
		// Other timeouts can be set based on your use case:
		// WriteTimeout, IdleTimeout, ReadTimeout, etc.
	}

	// Start the server in a goroutine so we can gracefully shut it down later.
	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Set up a channel to listen for interrupt signals (Ctrl+C) to shut down gracefully.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	// Block until we receive an interrupt.
	<-quit
	log.Println("Shutting down server...")

	// Create a context with a timeout to ensure the server stops within a given time.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt a graceful shutdown.
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Graceful shutdown failed; forcing exit: %v", err)
	} else {
		log.Println("Server stopped gracefully.")
	}
}
