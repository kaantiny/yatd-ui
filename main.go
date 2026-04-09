package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"runtime/debug"

	"yatd-ui/internal/handlers"
)

//go:embed templates/*.html
var templateFS embed.FS

var tmpl *template.Template

var version = "dev"

func resolvedVersion() string {
	if version != "dev" {
		return version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}

	return version
}

func init() {
	var err error
	tmpl, err = template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		panic(err)
	}
}

func main() {
	v := resolvedVersion()

	showVersion := flag.Bool("version", false, "print version")
	flag.Parse()
	if *showVersion {
		fmt.Printf("%s\n", v)
		return
	}

	fmt.Printf("yatd UI %s starting...\n", v)

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
