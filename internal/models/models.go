package models

type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	Effort      string    `json:"effort"`
	Labels      []string  `json:"labels"`
	Blockers    []string  `json:"blockers"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	Parent      *string   `json:"parent"`
	Logs        []LogEntry `json:"logs,omitempty"`
}

type LogEntry struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

type Project struct {
	Name  string `json:"name"`
	Tasks []Task `json:"tasks"`
}

type NextTask struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Priority string `json:"priority"`
	Effort   string `json:"effort"`
}
