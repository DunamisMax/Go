package main

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// We will integrate everything into one file, including templates inline.
// We'll define all templates in a single multi-line string and parse them together.
// We will have the following templates defined:
//   mainpage.html - The root page with the headline toggle and navigation
//   todo.html - The todo page
//   blog.html - The blog listing page
//   blogpost.html - Individual blog post page
//   blogposts_partial.html - Partial listing of blog posts for htmx loading
//   comingsoon.html - A generic "coming soon" page
//   portfolio.html - The portfolio page
//   contact.html - The contact page
//
// All pages share a consistent dark theme, same as the original working code.

var templates = `
{{define "mainpage.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>dunamismax.com</title>
    <link href="https://fonts.googleapis.com/css2?family=Open+Sans:ital,wght@0,300..800;1,300..800&display=swap" rel="stylesheet">
    <script src="https://unpkg.com/htmx.org/dist/htmx.min.js"></script>
    <style>
        body {
            background: #000;
            color: #ae92f0;
            font-family: 'Open Sans', sans-serif;
            margin: 0;
            padding: 20px;
        }
        nav {
            margin-bottom: 40px;
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
        }
        nav a {
            background: #ae92f0;
            color: #000;
            padding: 8px 16px;
            border: none;
            border-radius: 4px;
            text-decoration: none;
            font-size: 16px;
        }
        nav a:hover {
            background: #c3a6f3;
        }
        h1 {
            cursor: pointer;
            transition: text-decoration 0.2s ease;
            text-align: center;
        }
        h1:hover {
            text-decoration: underline;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
        }
    </style>
</head>
<body>
    <nav>
        <a href="/todo">Todo App</a>
        <a href="/blog">Blog</a>
        <a href="/weather">Weather App</a>
        <a href="/portfolio">Portfolio</a>
        <a href="/contact">Contact</a>
        <a href="/chat">Chat</a>
    </nav>
    <div class="container">
        <h1 id="headline"
            hx-get="/toggle?state=hello"
            hx-trigger="click"
            hx-target="#headline"
            hx-swap="outerHTML">
            Hello, world.
        </h1>
    </div>
</body>
</html>
{{end}}

{{define "todo.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>dunamismax.com - Todo</title>
<script src="https://unpkg.com/htmx.org/dist/htmx.min.js"></script>
<link href="https://fonts.googleapis.com/css2?family=Open+Sans:ital,wght@0,300..800;1,300..800&display=swap" rel="stylesheet">
<style>
    body {
        background: #000;
        color: #ae92f0;
        font-family: 'Open Sans', sans-serif;
        margin: 0;
        padding: 20px;
    }
    nav {
        margin-bottom: 40px;
        display: flex;
        gap: 10px;
        flex-wrap: wrap;
    }
    nav a {
        background: #ae92f0;
        color: #000;
        padding: 8px 16px;
        border: none;
        border-radius: 4px;
        text-decoration: none;
        font-size: 16px;
    }
    nav a:hover {
        background: #c3a6f3;
    }
    h1 {
        margin-bottom: 20px;
    }
    form.add-todo-form {
        display: flex;
        margin-bottom: 20px;
    }
    form.add-todo-form input[type="text"] {
        flex: 1;
        padding: 10px;
        font-size: 16px;
        border: 1px solid #ae92f0;
        border-radius: 4px 0 0 4px;
        background: #000;
        color: #ae92f0;
    }
    form.add-todo-form input[type="text"]::placeholder {
        color: #c3a6f3;
    }
    form.add-todo-form button {
        padding: 10px 20px;
        font-size: 16px;
        border: none;
        color: #000;
        background: #ae92f0;
        border-radius: 0 4px 4px 0;
        cursor: pointer;
    }
    form.add-todo-form button:hover {
        background: #c3a6f3;
        color: #000;
    }
    .todo-list {
        list-style: none;
        padding: 0;
        margin: 0;
    }
    .todo-item {
        display: flex;
        align-items: center;
        padding: 10px;
        border-bottom: 1px solid #c3a6f3;
    }
    .todo-item:last-child {
        border-bottom: none;
    }
    .todo-item.completed span {
        text-decoration: line-through;
        color: #777;
    }
    .todo-item input[type="checkbox"] {
        margin-right: 10px;
        width: 20px;
        height: 20px;
        accent-color: #ae92f0;
    }
    .delete-btn {
        margin-left: auto;
        background: none;
        border: none;
        font-size: 18px;
        cursor: pointer;
        color: #ae92f0;
    }
    .delete-btn:hover {
        color: #c3a6f3;
    }
</style>
</head>
<body>
<nav>
    <a href="/">Home</a>
    <a href="/blog">Blog</a>
    <a href="/weather">Weather</a>
    <a href="/portfolio">Portfolio</a>
    <a href="/contact">Contact</a>
    <a href="/chat">Chat</a>
</nav>
<h1>My Todos</h1>
<form class="add-todo-form" hx-post="/todo/add" hx-target="#todo-list" hx-swap="afterbegin">
    <input type="text" name="title" placeholder="What do you need to do?" required />
    <button type="submit">Add</button>
</form>
<ul class="todo-list" id="todo-list">
{{ . }}
</ul>
</body>
</html>
{{end}}

{{define "blog.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link href="https://fonts.googleapis.com/css2?family=Open+Sans&display=swap" rel="stylesheet">
<script src="https://unpkg.com/htmx.org/dist/htmx.min.js"></script>
<style>
body {
  font-family: 'Open Sans', sans-serif;
  margin: 0;
  padding:20px;
  background: #000;
  color: #ae92f0;
}
.container {
  max-width: 600px;
  margin: 0 auto;
}
a {
  text-decoration: none;
  color: #ae92f0;
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
nav {
  margin-bottom: 40px;
  display:flex;
  gap:10px;
  flex-wrap:wrap;
}
nav a {
  background:#ae92f0;
  color:#000;
  padding:8px 16px;
  border-radius:4px;
  text-decoration:none;
}
nav a:hover {
  background:#c3a6f3;
}
</style>
<title>dunamismax.com - Blog</title>
</head>
<body>
<nav>
    <a href="/">Home</a>
    <a href="/weather">Weather</a>
    <a href="/todo">Todo</a>
    <a href="/portfolio">Portfolio</a>
    <a href="/contact">Contact</a>
    <a href="/chat">Chat</a>
</nav>
<div class="container">
<h1>Blog</h1>
<ul class="post-list" id="post-container">
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

{{define "blogpost.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link href="https://fonts.googleapis.com/css2?family=Open+Sans&display=swap" rel="stylesheet">
<script src="https://unpkg.com/htmx.org/dist/htmx.min.js"></script>
<style>
body {
  font-family: 'Open Sans', sans-serif;
  margin: 0;
  padding:20px;
  background: #000;
  color: #ae92f0;
}
.container {
  max-width: 600px;
  margin: 0 auto;
}
nav {
  margin-bottom: 40px;
  display:flex;
  gap:10px;
  flex-wrap:wrap;
}
nav a {
  background:#ae92f0;
  color:#000;
  padding:8px 16px;
  border-radius:4px;
  text-decoration:none;
  font-size:16px;
}
nav a:hover {
  background:#c3a6f3;
}
h1 {
  margin-bottom:10px;
}
.date {
  font-size:0.9em;
  color:#c3a6f3;
  margin-bottom:20px;
}
.content {
  margin-bottom:40px;
}
</style>
<title>{{.Title}} - dunamismax.com</title>
</head>
<body>
<nav>
    <a href="/blog">Back to Blog</a>
    <a href="/">Home</a>
    <a href="/weather">Weather</a>
    <a href="/todo">Todo</a>
    <a href="/portfolio">Portfolio</a>
    <a href="/contact">Contact</a>
    <a href="/chat">Chat</a>
</nav>
<div class="container">
<h1>{{.Title}}</h1>
<div class="date">Published on: {{FormatDate .Date}}</div>
<div class="content">
  <p>{{.Content}}</p>
</div>
</div>
</body>
</html>
{{end}}

{{define "blogposts_partial.html"}}
{{range .}}
<li>
  <a href="/blog/{{.ID}}"><strong>{{.Title}}</strong></a> <small>({{FormatDate .Date}})</small>
</li>
{{end}}
{{end}}

{{define "comingsoon.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link href="https://fonts.googleapis.com/css2?family=Open+Sans&display=swap" rel="stylesheet">
<script src="https://unpkg.com/htmx.org/dist/htmx.min.js"></script>
<style>
body {
  font-family: 'Open Sans', sans-serif;
  margin: 0;
  padding:20px;
  background:#000;
  color:#ae92f0;
  text-align:center;
}
.container {
  max-width:600px;
  margin:0 auto;
}
nav {
  margin-bottom:40px;
  display:flex;
  gap:10px;
  flex-wrap:wrap;
  justify-content:center;
}
nav a {
  background:#ae92f0;
  color:#000;
  padding:8px 16px;
  border-radius:4px;
  text-decoration:none;
  font-size:16px;
}
nav a:hover {
  background:#c3a6f3;
}
h1 {
  margin-bottom:40px;
}
</style>
<title>dunamismax.com - Coming Soon</title>
</head>
<body>
<nav>
    <a href="/">Home</a>
    <a href="/blog">Blog</a>
    <a href="/weather">Weather</a>
    <a href="/todo">Todo</a>
    <a href="/portfolio">Portfolio</a>
    <a href="/contact">Contact</a>
    <a href="/chat">Chat</a>
</nav>
<div class="container">
<h1>{{.}}</h1>
</div>
</body>
</html>
{{end}}

{{define "portfolio.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link href="https://fonts.googleapis.com/css2?family=Open+Sans&display=swap" rel="stylesheet">
<script src="https://unpkg.com/htmx.org/dist/htmx.min.js"></script>
<style>
body {
  font-family: 'Open Sans', sans-serif;
  margin: 0;
  padding:20px;
  background:#000;
  color:#ae92f0;
}
.container {
  max-width:600px;
  margin:0 auto;
  text-align:center;
}
nav {
  margin-bottom:40px;
  display:flex;
  gap:10px;
  flex-wrap:wrap;
}
nav a {
  background:#ae92f0;
  color:#000;
  padding:8px 16px;
  border-radius:4px;
  text-decoration:none;
  font-size:16px;
}
nav a:hover {
  background:#c3a6f3;
}
h1 {
  margin-bottom:20px;
}
</style>
<title>dunamismax.com - Portfolio</title>
</head>
<body>
<nav>
    <a href="/">Home</a>
    <a href="/blog">Blog</a>
    <a href="/weather">Weather</a>
    <a href="/todo">Todo</a>
    <a href="/contact">Contact</a>
    <a href="/chat">Chat</a>
</nav>
<div class="container">
<h1>My Portfolio</h1>
<p>Check out my GitHub:</p>
<p><a href="https://github.com/dunamismax" target="_blank" style="color:#ae92f0;text-decoration:underline;">github.com/dunamismax</a></p>
</div>
</body>
</html>
{{end}}

{{define "contact.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link href="https://fonts.googleapis.com/css2?family=Open+Sans&display=swap" rel="stylesheet">
<script src="https://unpkg.com/htmx.org/dist/htmx.min.js"></script>
<style>
body {
  font-family: 'Open Sans', sans-serif;
  margin:0;
  padding:20px;
  background:#000;
  color:#ae92f0;
}
.container {
  max-width:600px;
  margin:0 auto;
}
nav {
  margin-bottom:40px;
  display:flex;
  gap:10px;
  flex-wrap:wrap;
}
nav a {
  background:#ae92f0;
  color:#000;
  padding:8px 16px;
  border-radius:4px;
  text-decoration:none;
  font-size:16px;
}
nav a:hover {
  background:#c3a6f3;
}
h1 {
  margin-bottom:20px;
}
form {
  display:flex;
  flex-direction:column;
}
label {
  margin:10px 0 5px;
}
input[type="text"], input[type="email"], textarea {
  padding:8px;
  border:1px solid #c3a6f3;
  border-radius:4px;
  background:#000;
  color:#ae92f0;
}
textarea {
  min-height:100px;
}
button {
  margin-top:20px;
  padding:10px;
  background:#ae92f0;
  color:#000;
  border:none;
  border-radius:4px;
  cursor:pointer;
}
button:hover {
  background:#c3a6f3;
}
</style>
<title>dunamismax.com - Contact</title>
</head>
<body>
<nav>
    <a href="/">Home</a>
    <a href="/blog">Blog</a>
    <a href="/weather">Weather</a>
    <a href="/todo">Todo</a>
    <a href="/portfolio">Portfolio</a>
    <a href="/chat">Chat</a>
</nav>
<div class="container">
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
`

// ============ Data Structures from the second snippet (Blog) ============

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

// ============ Template Parsing ============

var tpl = template.Must(template.New("").Funcs(template.FuncMap{
	"FormatDate": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
}).Parse(templates))

// ============ Toggle Feature (from first snippet) ============

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://unpkg.com; style-src 'self' https://fonts.googleapis.com 'unsafe-inline'; font-src 'self' https://fonts.gstatic.com; img-src 'self';")
	if err := tpl.ExecuteTemplate(w, "mainpage.html", nil); err != nil {
		log.Printf("Error rendering main template: %v", err)
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

// ============ Security Headers ============

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://unpkg.com 'unsafe-inline'; style-src 'self' https://fonts.googleapis.com 'unsafe-inline'; font-src 'self' https://fonts.gstatic.com; img-src 'self';")
		next.ServeHTTP(w, r)
	})
}

// ============ Todo App (from first snippet) ============

type Todo struct {
	ID        int
	Title     string
	Completed bool
}

type Store struct {
	sync.Mutex
	todos  []Todo
	nextID int
}

func (s *Store) Add(title string) Todo {
	s.Lock()
	defer s.Unlock()
	s.nextID++
	t := Todo{ID: s.nextID, Title: title, Completed: false}
	s.todos = append(s.todos, t)
	return t
}

func (s *Store) Toggle(id int) (Todo, bool) {
	s.Lock()
	defer s.Unlock()
	for i := range s.todos {
		if s.todos[i].ID == id {
			s.todos[i].Completed = !s.todos[i].Completed
			return s.todos[i], true
		}
	}
	return Todo{}, false
}

func (s *Store) Delete(id int) bool {
	s.Lock()
	defer s.Unlock()
	for i := range s.todos {
		if s.todos[i].ID == id {
			s.todos = append(s.todos[:i], s.todos[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Store) RenderTodosHTML() string {
	s.Lock()
	defer s.Unlock()

	var sb strings.Builder
	for _, t := range s.todos {
		completedClass := ""
		if t.Completed {
			completedClass = "completed"
		}
		sb.WriteString(fmt.Sprintf(`
			<li id="todo-%d" class="todo-item %s">
				<input type="checkbox" hx-post="/todo/toggle?id=%d" hx-swap="outerHTML" hx-trigger="change" %s />
				<span>%s</span>
				<button class="delete-btn" hx-post="/todo/delete?id=%d" hx-target="#todo-%d" hx-swap="outerHTML">✕</button>
			</li>
		`, t.ID, html.EscapeString(completedClass), t.ID, checkedAttr(t.Completed), html.EscapeString(t.Title), t.ID, t.ID))
	}
	return sb.String()
}

func renderSingleTodoHTML(t Todo) string {
	completedClass := ""
	if t.Completed {
		completedClass = "completed"
	}
	return fmt.Sprintf(`
	<li id="todo-%d" class="todo-item %s">
		<input type="checkbox" hx-post="/todo/toggle?id=%d" hx-swap="outerHTML" hx-trigger="change" %s />
		<span>%s</span>
		<button class="delete-btn" hx-post="/todo/delete?id=%d" hx-target="#todo-%d" hx-swap="outerHTML">✕</button>
	</li>
	`, t.ID, html.EscapeString(completedClass), t.ID, checkedAttr(t.Completed), html.EscapeString(t.Title), t.ID, t.ID)
}

func checkedAttr(completed bool) string {
	if completed {
		return "checked"
	}
	return ""
}

var store = &Store{}

func handleTodoPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/todo" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.ExecuteTemplate(w, "todo.html", template.HTML(store.RenderTodosHTML())); err != nil {
		log.Printf("Error rendering todo template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleTodoAdd(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	title := strings.TrimSpace(r.Form.Get("title"))
	if title == "" {
		http.Error(w, "Title cannot be empty", http.StatusBadRequest)
		return
	}
	t := store.Add(title)
	fmt.Fprint(w, renderSingleTodoHTML(t))
}

func handleTodoToggle(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	t, ok := store.Toggle(id)
	if !ok {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}
	fmt.Fprint(w, renderSingleTodoHTML(t))
}

func handleTodoDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	ok := store.Delete(id)
	if !ok {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}
	// Return nothing, htmx will remove the element.
}

// ============ Additional pages (from second snippet) ============

func blogHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/blog" {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "blog.html", blogPosts)
}

func blogPostHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(pathParts) != 2 || pathParts[0] != "blog" {
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
	if r.Header.Get("HX-Request") == "true" {
		renderTemplate(w, "blogposts_partial.html", blogPosts)
		return
	}
	http.Redirect(w, r, "/blog", http.StatusSeeOther)
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/weather" {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "comingsoon.html", "Weather App Coming Soon")
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
		log.Printf("Error rendering template %q: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ============ main ============

func main() {
	// Initialize some example todos
	store.Add("Learn Go")
	store.Add("Build a webapp with htmx")
	store.Add("Deploy to production")

	mux := http.NewServeMux()

	// Main page and toggle
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/toggle", handleToggle)

	// Todo handlers
	mux.HandleFunc("/todo", handleTodoPage)
	mux.HandleFunc("/todo/add", handleTodoAdd)
	mux.HandleFunc("/todo/toggle", handleTodoToggle)
	mux.HandleFunc("/todo/delete", handleTodoDelete)

	// Additional pages
	mux.HandleFunc("/blog", blogHandler)
	mux.HandleFunc("/blog/", blogPostHandler)
	mux.HandleFunc("/partials/blogposts", blogPostsPartialHandler)
	mux.HandleFunc("/weather", weatherHandler)
	mux.HandleFunc("/portfolio", portfolioHandler)
	mux.HandleFunc("/contact", contactHandler)
	mux.HandleFunc("/chat", chatHandler)

	handler := securityHeaders(mux)

	log.Println("Serving on http://localhost:42069")
	if err := http.ListenAndServe(":42069", handler); err != nil {
		log.Fatal(err)
	}
}
