package main

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
)

// Template for the main page.
var tmpl = template.Must(template.New("page").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />

    <title>dunamismax.com</title>

    <!-- Preconnect to Google Fonts -->
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>

    <!-- Open Sans font -->
    <link href="https://fonts.googleapis.com/css2?family=Open+Sans:ital,wght@0,300..800;1,300..800&display=swap" rel="stylesheet">

    <style>
        body {
            background: #000;
            color: #ae92f0;
            font-family: 'Open Sans', sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
        }
        main {
            text-align: center;
        }
        h1 {
            cursor: pointer;
            transition: text-decoration 0.2s ease;
        }
        h1:hover {
            text-decoration: underline;
        }
    </style>

    <!-- Security Headers are set server-side. -->
</head>
<body>
    <main>
        <!-- Start state is "hello" -->
        <h1 id="headline"
            hx-get="/toggle?state=hello"
            hx-trigger="click"
            hx-target="#headline"
            hx-swap="outerHTML">
            Hello, world.
        </h1>
    </main>

    <script src="https://unpkg.com/htmx.org/dist/htmx.min.js"></script>
</body>
</html>
`))

func main() {
	// Set up a custom mux for clarity and possible future expansion.
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/toggle", handleToggle)

	// Wrap our handlers with security headers for all responses.
	handler := securityHeaders(mux)

	log.Println("Serving on http://localhost:42069")
	if err := http.ListenAndServe(":42069", handler); err != nil {
		log.Fatal(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://unpkg.com; style-src 'self' https://fonts.googleapis.com 'unsafe-inline'; font-src 'self' https://fonts.gstatic.com; img-src 'self';")
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleToggle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	q := r.URL.Query()
	currentState := q.Get("state")

	text, newState := toggleText(currentState)

	h1 := buildHeadlineHTML(text, newState)
	if _, err := w.Write([]byte(h1)); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func toggleText(currentState string) (text, newState string) {
	if currentState == "hello" {
		return "Goodbye, world.", "goodbye"
	}
	return "Hello, world.", "hello"
}

func buildHeadlineHTML(text, newState string) string {
	return `<h1 id="headline" ` +
		`hx-get="/toggle?state=` + url.QueryEscape(newState) + `" ` +
		`hx-trigger="click" ` +
		`hx-target="#headline" ` +
		`hx-swap="outerHTML">` +
		text +
		`</h1>`
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Basic security hardening
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		// A basic CSP that allows fonts and the htmx script from known sources, disallows inline scripts
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://unpkg.com; style-src 'self' https://fonts.googleapis.com 'unsafe-inline'; font-src 'self' https://fonts.gstatic.com; img-src 'self';")

		next.ServeHTTP(w, r)
	})
}
