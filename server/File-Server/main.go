package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

func main() {
	// Configuration
	dir := flag.String("dir", "./public", "Directory to serve files from")
	addr := flag.String("addr", ":8080", "Address to listen on (e.g. :8080)")
	flag.Parse()

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatalf("Failed to resolve absolute path: %v", err)
	}

	// Ensure the directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", absDir)
	}

	// Create a ServeMux
	mux := http.NewServeMux()

	// Create the file server
	// The prefix "/static/" maps to the directory absDir.
	// You can change the prefix to "/" if you want all requests to serve files directly.
	fileServer := http.FileServer(http.Dir(absDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Optionally, serve an index page (redirect root to /static/)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/", http.StatusFound)
	})

	// Wrap with logging middleware
	handler := LoggingMiddleware(mux)

	// Create a server with sensible defaults
	srv := &http.Server{
		Addr:         *addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Printf("File server started on %s, serving %s\n", *addr, absDir)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	} else {
		log.Println("Server stopped.")
	}
}

// LoggingMiddleware provides a simple access log for each request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
