package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"yatd-ui/internal/handlers"
)

//go:embed templates/*.html
var templateFS embed.FS

var tmpl *template.Template

var version = "dev"

func init() {
	var err error
	tmpl, err = template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Printf("yatd UI %s starting...\n", version)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "index.html", nil)
	})
	http.HandleFunc("/api/projects", handlers.APIProjects)
	http.HandleFunc("/api/tasks", handlers.APITasks)
	http.HandleFunc("/api/task", handlers.APITask)
	http.HandleFunc("/api/next", handlers.APINext)

	fmt.Printf("Server listening on :%s\n", port)
	http.ListenAndServe(":"+port, nil)
}
