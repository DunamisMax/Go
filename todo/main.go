package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Todo represents a single task with a title and completion status.
type Todo struct {
	ID      int
	Title   string
	Completed bool
}

// Store will hold our in-memory todo list and an incrementing ID counter.
// We'll protect it with a mutex for safe concurrent access.
type Store struct {
	sync.Mutex
	todos []Todo
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
			// Delete this todo from the slice
			s.todos = append(s.todos[:i], s.todos[i+1:]...)
			return true
		}
	}
	return false
}

// Render the entire list of todos as HTML.
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
				<input type="checkbox" hx-post="/toggle?id=%d" hx-swap="outerHTML" hx-trigger="change" %s />
				<span>%s</span>
				<button class="delete-btn" hx-post="/delete?id=%d" hx-target="#todo-%d" hx-swap="outerHTML">✕</button>
			</li>
		`, t.ID, html.EscapeString(completedClass), t.ID, checkedAttr(t.Completed), html.EscapeString(t.Title), t.ID, t.ID))
	}
	return sb.String()
}

// Render a single updated todo item, used after toggling completion.
func renderSingleTodoHTML(t Todo) string {
	completedClass := ""
	if t.Completed {
		completedClass = "completed"
	}
	return fmt.Sprintf(`
	<li id="todo-%d" class="todo-item %s">
		<input type="checkbox" hx-post="/toggle?id=%d" hx-swap="outerHTML" hx-trigger="change" %s />
		<span>%s</span>
		<button class="delete-btn" hx-post="/delete?id=%d" hx-target="#todo-%d" hx-swap="outerHTML">✕</button>
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

func main() {
	// Initial setup: add a few example todos
	store.Add("Learn Go")
	store.Add("Build a webapp with htmx")
	store.Add("Deploy to production")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve the entire page
		fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>Go + htmx Todo App</title>
<script src="https://unpkg.com/htmx.org@1.9.2"></script>
<style>
	body {
		font-family: sans-serif;
		max-width: 600px;
		margin: 50px auto;
		background: #f9f9f9;
		padding: 20px;
		border-radius: 8px;
		border: 1px solid #ccc;
	}
	h1 {
		text-align: center;
	}
	form.add-todo-form {
		display: flex;
		margin-bottom: 20px;
	}
	form.add-todo-form input[type="text"] {
		flex: 1;
		padding: 10px;
		font-size: 16px;
		border: 1px solid #ccc;
		border-radius: 4px 0 0 4px;
	}
	form.add-todo-form button {
		padding: 10px 20px;
		font-size: 16px;
		border: none;
		color: #fff;
		background: #007BFF;
		border-radius: 0 4px 4px 0;
		cursor: pointer;
	}
	form.add-todo-form button:hover {
		background: #0056b3;
	}
	.todo-list {
		list-style: none;
		padding: 0;
	}
	.todo-item {
		display: flex;
		align-items: center;
		padding: 10px;
		border-bottom: 1px solid #eee;
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
	}
	.delete-btn {
		margin-left: auto;
		background: none;
		border: none;
		font-size: 18px;
		cursor: pointer;
		color: #999;
	}
	.delete-btn:hover {
		color: #e00;
	}
</style>
</head>
<body>
<h1>My Todos</h1>
<form class="add-todo-form" hx-post="/add" hx-target="#todo-list" hx-swap="afterbegin">
	<input type="text" name="title" placeholder="What do you need to do?" required />
	<button type="submit">Add</button>
</form>
<ul class="todo-list" id="todo-list">
%s
</ul>
</body>
</html>`, store.RenderTodosHTML())
	})

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		// Add a new todo, return the newly created item as HTML to prepend
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
	})

	http.HandleFunc("/toggle", func(w http.ResponseWriter, r *http.Request) {
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
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
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
		// For htmx: returning an empty response removes the element when using 'hx-swap="outerHTML"' on target
		// We can simply return nothing here.
	})

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
