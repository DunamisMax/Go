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
)

// =================== Shared & Main Page Code ====================

// Template for the main page.
var mainTmpl = template.Must(template.New("page").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>dunamismax.com</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
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
            flex-direction: column;
        }
        nav {
            margin-bottom: 40px;
        }
        nav button {
            background: #ae92f0;
            color: #000;
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            margin: 0 10px;
            font-size: 16px;
        }
        nav button:hover {
            background: #c3a6f3;
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
</head>
<body>
    <nav>
        <button onclick="window.location.href='/todo'">Todo App</button>
        <button onclick="alert('Future App Coming Soon!')">Another App</button>
    </nav>
    <main>
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

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://unpkg.com; style-src 'self' https://fonts.googleapis.com 'unsafe-inline'; font-src 'self' https://fonts.gstatic.com; img-src 'self';")
	if err := mainTmpl.Execute(w, nil); err != nil {
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

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		// Allow scripts from self and https://unpkg.com
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://unpkg.com; style-src 'self' https://fonts.googleapis.com 'unsafe-inline'; font-src 'self' https://fonts.gstatic.com; img-src 'self';")
		next.ServeHTTP(w, r)
	})
}

// =================== Todo App Code ====================

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

var todoTmpl = template.Must(template.New("todoPage").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>Go + htmx Todo App</title>
<script src="https://unpkg.com/htmx.org@1.9.2"></script>
<link href="https://fonts.googleapis.com/css2?family=Open+Sans:ital,wght@0,300..800;1,300..800&display=swap" rel="stylesheet">
<style>
    body {
        background: #000;
        color: #ae92f0;
        font-family: 'Open Sans', sans-serif;
        margin: 0;
        padding: 20px;
    }

    .back-link {
        display: inline-block;
        margin-bottom: 20px;
        color: #ae92f0;
        text-decoration: none;
        font-size: 16px;
        font-weight: bold;
    }
    .back-link:hover {
        text-decoration: underline;
    }

    h1 {
        text-align: center;
        margin: 20px 0;
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
<a class="back-link" href="/">← Back to Main</a>
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
`))


func handleTodoPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := todoTmpl.Execute(w, template.HTML(store.RenderTodosHTML())); err != nil {
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

// =================== main ====================

func main() {
	// Initialize some example todos
	store.Add("Learn Go")
	store.Add("Build a webapp with htmx")
	store.Add("Deploy to production")

	mux := http.NewServeMux()

	// Main page and toggle
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/toggle", handleToggle)

	// Todo app handlers
	mux.HandleFunc("/todo", handleTodoPage)
	mux.HandleFunc("/todo/add", handleTodoAdd)
	mux.HandleFunc("/todo/toggle", handleTodoToggle)
	mux.HandleFunc("/todo/delete", handleTodoDelete)

	handler := securityHeaders(mux)

	log.Println("Serving on http://localhost:42069")
	if err := http.ListenAndServe(":42069", handler); err != nil {
		log.Fatal(err)
	}
}
