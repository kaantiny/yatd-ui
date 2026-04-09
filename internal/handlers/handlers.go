package handlers

import (
	"encoding/json"
	"net/http"

	"yatd-ui/internal/services"
)

func APIProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := services.GetProjects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func APITasks(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	tasks, err := services.GetTasks(project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func APITask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	project := r.URL.Query().Get("project")

	switch r.Method {
	case "GET":
		task, err := services.GetTask(project, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)

	case "POST":
		var req struct {
			Status string `json:"status,omitempty"`
			Log    string `json:"log,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		args := []string{"update", id}
		if req.Status != "" {
			args = append(args, "-s", req.Status)
		}
		if req.Log != "" {
			args = append(args, "-l", req.Log)
		}
		_, err := services.RunTD(args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		_, err := services.RunTD("delete", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func APINext(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	tasks, err := services.GetNextTasks(project)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}


