package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

// We will serve on port 42069
// The site name is dunamismax.com (we'll just use it as a title and domain reference).
// We'll create multiple pages:
//  - / (index) : A landing page with buttons to go to other pages
//  - /blog : A static blog page listing posts
//  - /blog/{id} : Individual blog post pages
//  - /weather : "Weather App Coming Soon"
//  - /todo : "To-do app coming soon"
//  - /portfolio : Show links to github.com/dunamismax
//  - /contact : A contact form (non-functional for now)
//  - /chat : "Chat app coming soon"
//
// We'll incorporate htmx for progressive enhancement (e.g., loading individual blog posts via an hx-get call).
// We'll serve simple minimal vanilla CSS inline and load Open Sans via a CDN.
// We won't specify versions for external assets (like htmx), using the latest by pointing to the stable CDN.
// The design: clean and minimal.
// We'll implement a small in-memory set of blog posts and load them dynamically on the blog page using htmx.
//
// Directory structure (at runtime) is all in-memory for this example. Just run `go run main.go`.
// Open your browser to http://localhost:42069
//
// Note: Since the user requested minimal styling and no version numbers, we will pull Open Sans and htmx without specifying exact versions.

type BlogPost struct {
	ID      string
	Title   string
	Content string
	Date    time.Time
}

var blogPosts = []BlogPost{
	{
		ID:      "1",
		Title:   "Welcome to My Blog",
		Content: "This is the first post on my blog! Stay tuned for more content.",
		Date:    time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC),
	},
	{
		ID:      "2",
		Title:   "Another Post",
		Content: "Here's another sample post to show off the static blog functionality.",
		Date:    time.Date(2024, time.January, 2, 12, 0, 0, 0, time.UTC),
	},
	{
		ID:      "3",
		Title:   "Golang and htmx",
		Content: "Combining Go backends with htmx front-ends can produce dynamic user experiences without heavy JavaScript frameworks.",
		Date:    time.Date(2024, time.January, 3, 8, 30, 0, 0, time.UTC),
	},
}

var tpl = template.Must(template.New("").Funcs(template.FuncMap{
	"FormatDate": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
}).ParseFS(templateFS, "templates/*.html"))

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/blog", blogHandler)
	http.HandleFunc("/blog/", blogPostHandler)
	http.HandleFunc("/weather", weatherHandler)
	http.HandleFunc("/todo", todoHandler)
	http.HandleFunc("/portfolio", portfolioHandler)
	http.HandleFunc("/contact", contactHandler)
	http.HandleFunc("/chat", chatHandler)

	// Endpoint for partial blog post load via htmx if desired:
	http.HandleFunc("/partials/blogposts", blogPostsPartialHandler)

	log.Println("Starting server on :42069")
	if err := http.ListenAndServe(":42069", nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "home.html", nil)
}

func blogHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/blog" {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "blog.html", blogPosts)
}

func blogPostHandler(w http.ResponseWriter, r *http.Request) {
	// URL format: /blog/{id}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(pathParts) != 2 {
		http.NotFound(w, r)
		return
	}
	id := pathParts[1]
	var found *BlogPost
	for i := range blogPosts {
		if blogPosts[i].ID == id {
			found = &blogPosts[i]
			break
		}
	}
	if found == nil {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "blogpost.html", found)
}

func blogPostsPartialHandler(w http.ResponseWriter, r *http.Request) {
	// For demonstration: load blog posts dynamically.
	// If requested via htmx, we return just the list items snippet.
	// If not htmx, fallback gracefully.
	// We'll assume htmx request by checking HX-Request header
	if r.Header.Get("HX-Request") == "true" {
		renderTemplate(w, "blogpostspartial.html", blogPosts)
		return
	}
	// If not htmx, just redirect to full blog page
	http.Redirect(w, r, "/blog", http.StatusSeeOther)
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/weather" {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "comingsoon.html", "Weather App Coming Soon")
}

func todoHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/todo" {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "comingsoon.html", "To-do app coming soon")
}

func portfolioHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/portfolio" {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "portfolio.html", nil)
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/contact" {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "contact.html", nil)
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/chat" {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "comingsoon.html", "Chat app coming soon")
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tpl.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

// Embed templates
// We will provide a minimal CSS and load Open Sans from Google Fonts and htmx from CDN.
// Minimal styling and a clean layout.

import (
	"embed"
)

//go:embed templates
var templateFS embed.FS

// ----------------------- templates/home.html -----------------------
/*
{{define "home.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link href="https://fonts.googleapis.com/css2?family=Open+Sans&display=swap" rel="stylesheet">
<script src="https://unpkg.com/htmx.org"></script>
<style>
body {
  font-family: 'Open Sans', sans-serif;
  margin: 20px;
  background: #f9f9f9;
  color: #333;
}
.container {
  max-width: 600px;
  margin: 0 auto;
  text-align: center;
}
a {
  display: block;
  margin: 10px 0;
  text-decoration: none;
  background: #333;
  color: #fff;
  padding: 10px;
  border-radius: 4px;
}
a:hover {
  background: #555;
}
h1 {
  margin-bottom: 40px;
}
</style>
<title>dunamismax.com</title>
</head>
<body>
<div class="container">
<h1>Welcome to dunamismax.com</h1>
<a href="/blog">Blog</a>
<a href="/weather">Weather App</a>
<a href="/todo">To-Do</a>
<a href="/portfolio">Portfolio</a>
<a href="/contact">Contact</a>
<a href="/chat">Chat</a>
</div>
</body>
</html>
{{end}}
*/

// ----------------------- templates/blog.html -----------------------
/*
{{define "blog.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link href="https://fonts.googleapis.com/css2?family=Open+Sans&display=swap" rel="stylesheet">
<script src="https://unpkg.com/htmx.org"></script>
<style>
body {
  font-family: 'Open Sans', sans-serif;
  margin: 20px;
  background: #f9f9f9;
  color: #333;
}
.container {
  max-width: 600px;
  margin: 0 auto;
}
a {
  text-decoration: none;
  color: #333;
}
a:hover {
  text-decoration: underline;
}
h1 {
  margin-bottom: 20px;
}
.post-list {
  list-style-type: none;
  padding: 0;
}
.post-list li {
  margin: 10px 0;
}
.nav {
  margin-bottom: 20px;
}
.nav a {
  margin-right: 10px;
}
</style>
<title>dunamismax.com - Blog</title>
</head>
<body>
<div class="container">
<div class="nav"><a href="/">Home</a></div>
<h1>Blog</h1>
<ul class="post-list" id="post-container">
  <!-- We'll load posts directly here since we have them. Or we could htmx-load them. -->
  {{range .}}
  <li>
    <a href="/blog/{{.ID}}"><strong>{{.Title}}</strong></a> <small>({{FormatDate .Date}})</small>
  </li>
  {{end}}
</ul>
</div>
</body>
</html>
{{end}}
*/

// ----------------------- templates/blogpost.html -----------------------
/*
{{define "blogpost.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link href="https://fonts.googleapis.com/css2?family=Open+Sans&display=swap" rel="stylesheet">
<script src="https://unpkg.com/htmx.org"></script>
<style>
body {
  font-family: 'Open Sans', sans-serif;
  margin: 20px;
  background: #f9f9f9;
  color: #333;
}
.container {
  max-width: 600px;
  margin: 0 auto;
}
.nav {
  margin-bottom: 20px;
}
.nav a {
  margin-right: 10px;
  text-decoration: none;
  color: #333;
}
.nav a:hover {
  text-decoration: underline;
}
h1 {
  margin-bottom: 10px;
}
.date {
  font-size: 0.9em;
  color: #666;
  margin-bottom: 20px;
}
.content {
  margin-bottom: 40px;
}
</style>
<title>{{.Title}} - dunamismax.com</title>
</head>
<body>
<div class="container">
<div class="nav"><a href="/blog">Back to Blog</a> <a href="/">Home</a></div>
<h1>{{.Title}}</h1>
<div class="date">Published on: {{FormatDate .Date}}</div>
<div class="content">
  <p>{{.Content}}</p>
</div>
</div>
</body>
</html>
{{end}}
*/

// ----------------------- templates/blogpostspartial.html -----------------------
/*
{{define "blogpostspartial.html"}}
{{range .}}
<li>
  <a href="/blog/{{.ID}}"><strong>{{.Title}}</strong></a> <small>({{FormatDate .Date}})</small>
</li>
{{end}}
{{end}}
*/

// ----------------------- templates/comingsoon.html -----------------------
/*
{{define "comingsoon.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link href="https://fonts.googleapis.com/css2?family=Open+Sans&display=swap" rel="stylesheet">
<script src="https://unpkg.com/htmx.org"></script>
<style>
body {
  font-family: 'Open Sans', sans-serif;
  margin: 20px;
  background: #f9f9f9;
  color: #333;
  text-align: center;
}
.container {
  max-width: 600px;
  margin: 0 auto;
}
.nav {
  margin-bottom: 20px;
}
.nav a {
  margin-right: 10px;
  text-decoration: none;
  color: #333;
}
.nav a:hover {
  text-decoration: underline;
}
h1 {
  margin-bottom: 40px;
}
</style>
<title>dunamismax.com - Coming Soon</title>
</head>
<body>
<div class="container">
<div class="nav"><a href="/">Home</a></div>
<h1>{{.}}</h1>
</div>
</body>
</html>
{{end}}
*/

// ----------------------- templates/portfolio.html -----------------------
/*
{{define "portfolio.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link href="https://fonts.googleapis.com/css2?family=Open+Sans&display=swap" rel="stylesheet">
<script src="https://unpkg.com/htmx.org"></script>
<style>
body {
  font-family: 'Open Sans', sans-serif;
  margin: 20px;
  background: #f9f9f9;
  color: #333;
}
.container {
  max-width: 600px;
  margin: 0 auto;
}
.nav {
  margin-bottom: 20px;
}
.nav a {
  margin-right: 10px;
  text-decoration: none;
  color: #333;
}
.nav a:hover {
  text-decoration: underline;
}
h1 {
  margin-bottom: 20px;
}
</style>
<title>dunamismax.com - Portfolio</title>
</head>
<body>
<div class="container">
<div class="nav"><a href="/">Home</a></div>
<h1>My Portfolio</h1>
<p>Check out my GitHub:</p>
<p><a href="https://github.com/dunamismax" target="_blank">github.com/dunamismax</a></p>
</div>
</body>
</html>
{{end}}
*/

// ----------------------- templates/contact.html -----------------------
/*
{{define "contact.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link href="https://fonts.googleapis.com/css2?family=Open+Sans&display=swap" rel="stylesheet">
<script src="https://unpkg.com/htmx.org"></script>
<style>
body {
  font-family: 'Open Sans', sans-serif;
  margin: 20px;
  background: #f9f9f9;
  color: #333;
}
.container {
  max-width: 600px;
  margin: 0 auto;
}
.nav {
  margin-bottom: 20px;
}
.nav a {
  margin-right: 10px;
  text-decoration: none;
  color: #333;
}
.nav a:hover {
  text-decoration: underline;
}
h1 {
  margin-bottom: 20px;
}
form {
  display: flex;
  flex-direction: column;
}
label {
  margin: 10px 0 5px;
}
input[type="text"], input[type="email"], textarea {
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
}
textarea {
  min-height: 100px;
}
button {
  margin-top: 20px;
  padding: 10px;
  background: #333;
  color: #fff;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
button:hover {
  background: #555;
}
</style>
<title>dunamismax.com - Contact</title>
</head>
<body>
<div class="container">
<div class="nav"><a href="/">Home</a></div>
<h1>Contact Me</h1>
<form>
  <label for="name">Name</label>
  <input type="text" id="name" name="name" placeholder="Your name">

  <label for="email">Email</label>
  <input type="email" id="email" name="email" placeholder="Your email">

  <label for="message">Message</label>
  <textarea id="message" name="message" placeholder="Your message"></textarea>

  <button type="submit">Send</button>
</form>
</div>
</body>
</html>
{{end}}
*/
