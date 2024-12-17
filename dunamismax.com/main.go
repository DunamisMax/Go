package main

import (
	"log"
	"net/http"
)

// A simple HTML and CSS page capturing a 90's retro hacker vibe.
// The page uses a black background, green "console" text, and centers "Hello, world."
// vertically and horizontally, reminiscent of an old-school terminal.
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Inline CSS for simplicity and full control:
		// - Black background
		// - Green monospaced text
		// - Centered content both vertically and horizontally
		// - Single line of text
		page := `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
    body {
        background: #000;
        color: #0f0;
        font-family: monospace;
        display: flex;
        justify-content: center;
        align-items: center;
        height: 100vh;
        margin: 0;
        font-size: 2rem;
    }
</style>
<title>Retro Hacker Vibes</title>
</head>
<body>
    Hello, world.
</body>
</html>
`
		_, _ = w.Write([]byte(page))
	})

	log.Println("Serving on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
