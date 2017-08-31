package main

import (
	"fmt"
	"net/http"
	"log"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"regexp"
)

// App Container
type App struct{
	db *sql.DB
}

// Routing struct
type route struct {
	pattern *regexp.Regexp
	verb    string
	handler http.Handler
}

type RegexpHandler struct {
	routes []*route
}

// Request Handler

func (h *RegexpHandler) Handler(pattern *regexp.Regexp, verb string, handler http.Handler) {
	h.routes = append(h.routes, &route{pattern, verb, handler})
}

func (h *RegexpHandler) HandleFunc(r string, v string, handler func(http.ResponseWriter, *http.Request)) {
	re := regexp.MustCompile(r)
	h.routes = append(h.routes, &route{re, v, http.HandlerFunc(handler)})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) && route.verb == r.Method {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}


// Response Body
type responseBody struct {
	Status bool `json:"status"`
	Data interface{} `json:"data"`
}

func jsonResponse(w http.ResponseWriter,body interface{}){

	b := &responseBody{true,body}

	js,err := json.Marshal(b)

	if err != nil {
		http.Error(w,err.Error(),http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type","application/json")
	w.Write(js)
}


func main() {
	fmt.Println("Hello Todo API")

	// DB initalization
	db, err := sql.Open("mysql" , "root:root@/todo_go")
	if err != nil {
		panic(err)
	}

	app := &App{db:db}

	reHandler := new(RegexpHandler)

	// Routes
	reHandler.HandleFunc("/todos$","GET",app.IndexHandler)
	reHandler.HandleFunc("/todos/[0-9]+$","GET",app.ShowHandler)


	// Initialize http server
	log.Fatal(http.ListenAndServe(":8080", reHandler))
}

// todo struct
type Todo struct {
	Id string `json:"id"`
	Title string `json:"title"`
	Category string `json:"category"`
	Status bool `json:"status"`
}

// Get all todos
func (app *App) IndexHandler(w http.ResponseWriter, r *http.Request) {
	rows , err := app.db.Query("SELECT * FROM todos")
	if err != nil {
		panic(err)
	}

	var todos []*Todo
	for rows.Next() {
		todo := &Todo{}
		rows.Scan(&todo.Id,&todo.Title,&todo.Category,&todo.Status)
		todos = append(todos,todo)
	}

	jsonResponse(w,todos)
}

// Show specific todo
func (app *App) ShowHandler(w http.ResponseWriter, r *http.Request) {
	reg, _ := regexp.Compile(`\d+$`)
	Id := reg.FindString(r.URL.Path)

	todo := &Todo{}
	app.db.QueryRow("SELECT * FROM todos where id = ?",Id).Scan(&todo.Id,&todo.Title,&todo.Category,&todo.Status)


	jsonResponse(w,todo)
}