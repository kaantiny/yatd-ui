package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
)

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

var tdPath string
var tdHome string

func init() {
	// Try to find td binary
	tdPath = "/home/exedev/td"
	if _, err := os.Stat(tdPath); err != nil {
		tdPath = "td"
	}

	// Resolve td home directory
	if home := os.Getenv("TD_HOME"); home != "" {
		tdHome = home
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "."
		}
		tdHome = homeDir
	}
}

func runTD(args ...string) ([]byte, error) {
	cmd := exec.Command(tdPath, args...)
	cmd.Dir = tdHome
	return cmd.Output()
}

func getProjects() ([]string, error) {
	out, err := runTD("project", "list", "-j")
	if err != nil {
		return nil, err
	}
	var projects []string
	if err := json.Unmarshal(out, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func getTasks(project string) ([]Task, error) {
	args := []string{"-j", "list"}
	if project != "" {
		args = append(args, "--project", project)
	}
	out, err := runTD(args...)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	if err := json.Unmarshal(out, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func getTask(project, id string) (*Task, error) {
	args := []string{"-j", "show", id}
	if project != "" {
		args = append(args, "--project", project)
	}
	out, err := runTD(args...)
	if err != nil {
		return nil, err
	}
	var task Task
	if err := json.Unmarshal(out, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func apiProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := getProjects()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func apiTasks(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	tasks, err := getTasks(project)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func apiTask(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id", 400)
		return
	}
	task, err := getTask(project, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

type NextTask struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Priority       string  `json:"priority"`
	Effort         string  `json:"effort"`
	Rank           int     `json:"rank"`
	Score          float64 `json:"score"`
	DirectUnblocked int    `json:"direct_unblocked"`
}

func getNextTasks(project string) ([]NextTask, error) {
	args := []string{"-j", "next"}
	if project != "" {
		args = append(args, "--project", project)
	}
	out, err := runTD(args...)
	if err != nil {
		return nil, err
	}
	var tasks []NextTask
	if err := json.Unmarshal(out, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func apiNext(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	tasks, err := getNextTasks(project)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/api/projects", apiProjects)
	http.HandleFunc("/api/tasks", apiTasks)
	http.HandleFunc("/api/task", apiTask)
	http.HandleFunc("/api/next", apiNext)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		tmpl.Execute(w, nil)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	fmt.Println("Server starting on :" + port)
	http.ListenAndServe(":"+port, nil)
}

var tmpl = template.Must(template.New("main").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>yatd UI</title>
    <style>
        :root {
            --bg: #0d1117;
            --bg-secondary: #161b22;
            --bg-tertiary: #21262d;
            --border: #30363d;
            --text: #e6edf3;
            --text-muted: #8b949e;
            --accent: #58a6ff;
            --accent-hover: #79b8ff;
            --success: #56d364;
            --warning: #e3b341;
            --danger: #f85149;
            --info: #79c0ff;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'JetBrains Mono', 'Fira Code', 'SF Mono', Monaco, monospace;
            background: var(--bg);
            color: var(--text);
            height: 100vh;
            overflow: hidden;
            font-size: 13px;
            line-height: 1.5;
        }

        .app {
            display: flex;
            height: 100vh;
        }

        .sidebar {
            width: 240px;
            background: var(--bg-secondary);
            border-right: 1px solid var(--border);
            display: flex;
            flex-direction: column;
        }

        .header {
            padding: 12px 16px;
            border-bottom: 1px solid var(--border);
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .header h1 {
            font-size: 14px;
            font-weight: 600;
            color: var(--accent);
        }

        .projects-list {
            flex: 1;
            overflow-y: auto;
        }

        .project-item {
            padding: 8px 16px;
            cursor: pointer;
            display: flex;
            align-items: center;
            gap: 8px;
            transition: background 0.1s;
        }

        .project-item:hover {
            background: var(--bg-tertiary);
        }

        .project-item.active {
            background: var(--bg-tertiary);
            border-left: 2px solid var(--accent);
            padding-left: 14px;
        }

        .project-item::before {
            content: "📁";
            font-size: 12px;
        }

        .next-bar {
            display: flex;
            align-items: center;
            gap: 12px;
            padding: 12px 16px;
            background: linear-gradient(90deg, rgba(210, 153, 34, 0.15) 0%, rgba(210, 153, 34, 0.05) 100%);
            border-bottom: 1px solid var(--border);
            border-left: 3px solid #d29922;
        }

        .next-bar-label {
            font-size: 11px;
            font-weight: 600;
            text-transform: uppercase;
            color: #e3b341;
            white-space: nowrap;
            letter-spacing: 0.05em;
        }

        .next-bar-content {
            display: flex;
            align-items: center;
            gap: 12px;
            cursor: pointer;
            flex: 1;
            min-width: 0;
        }

        .next-bar-content:hover .next-bar-title {
            color: var(--accent-hover);
        }

        .next-bar-title {
            font-weight: 500;
            color: var(--text);
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .next-bar-id {
            font-size: 11px;
            color: var(--text-muted);
            font-family: inherit;
            background: var(--bg-tertiary);
            padding: 2px 8px;
            border-radius: 4px;
            white-space: nowrap;
        }

        .next-bar-meta {
            display: flex;
            gap: 8px;
            font-size: 11px;
            color: var(--text-muted);
            white-space: nowrap;
        }

        .next-bar-score {
            font-size: 11px;
            color: #e3b341;
            font-weight: 500;
        }

        .next-bar-empty {
            color: var(--text-muted);
            font-style: italic;
            font-size: 13px;
        }

        @media (max-width: 768px) {
            .next-bar {
                padding: 12px;
                flex-wrap: wrap;
                gap: 8px;
            }
            .next-bar-content {
                width: 100%;
                flex-wrap: wrap;
            }
            .next-bar-title {
                font-size: 14px;
                width: 100%;
            }
            .next-bar-meta {
                width: 100%;
            }
        }

        .help-bar {
            padding: 12px 16px;
            border-top: 1px solid var(--border);
            font-size: 11px;
            color: var(--text-muted);
        }

        .help-bar kbd {
            background: var(--bg-tertiary);
            padding: 2px 6px;
            border-radius: 3px;
            border: 1px solid var(--border);
            font-family: inherit;
        }

        .main {
            flex: 1;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }

        .toolbar {
            padding: 12px 16px;
            border-bottom: 1px solid var(--border);
            display: flex;
            gap: 12px;
            align-items: center;
        }

        .search-box {
            background: var(--bg-tertiary);
            border: 1px solid var(--border);
            border-radius: 4px;
            padding: 6px 12px;
            color: var(--text);
            font-family: inherit;
            font-size: 12px;
            width: 300px;
            outline: none;
        }

        .search-box:focus {
            border-color: var(--accent);
        }

        .filter-bar {
            display: flex;
            gap: 8px;
        }

        .filter-btn {
            background: var(--bg-tertiary);
            border: 1px solid var(--border);
            color: var(--text-muted);
            padding: 4px 10px;
            border-radius: 4px;
            cursor: pointer;
            font-family: inherit;
            font-size: 11px;
            transition: all 0.1s;
        }

        .filter-btn:hover, .filter-btn.active {
            background: var(--accent);
            color: var(--bg);
            border-color: var(--accent);
        }

        .content {
            flex: 1;
            display: flex;
            overflow: hidden;
        }

        .task-list {
            flex: 1;
            overflow-y: auto;
            padding: 8px;
        }

        .task-item {
            background: var(--bg-secondary);
            border: 1px solid var(--border);
            border-radius: 6px;
            padding: 12px;
            margin-bottom: 8px;
            cursor: pointer;
            transition: all 0.1s;
            display: flex;
            gap: 12px;
            align-items: flex-start;
        }

        .task-item:hover {
            border-color: var(--accent);
        }

        .task-item.selected {
            border-color: var(--accent);
            background: var(--bg-tertiary);
        }

        .next-task-item {
            background: linear-gradient(90deg, rgba(210, 153, 34, 0.15) 0%, rgba(210, 153, 34, 0.05) 100%);
            border: 1px solid #d29922;
            border-radius: 6px;
            padding: 12px;
            margin-bottom: 8px;
            cursor: pointer;
            transition: all 0.1s;
            display: flex;
            gap: 12px;
            align-items: flex-start;
            border-left: 3px solid #d29922;
        }

        .next-task-item:hover {
            border-color: #e3b341;
            background: linear-gradient(90deg, rgba(210, 153, 34, 0.2) 0%, rgba(210, 153, 34, 0.08) 100%);
        }

        .next-task-label {
            font-size: 10px;
            font-weight: 600;
            text-transform: uppercase;
            color: #e3b341;
            margin-right: 4px;
        }

        .next-task-title {
            font-weight: 500;
            margin-bottom: 4px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .next-task-meta {
            display: flex;
            gap: 8px;
            font-size: 11px;
            color: var(--text-muted);
            flex-wrap: wrap;
        }

        .task-status {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            margin-top: 5px;
            flex-shrink: 0;
        }

        .task-status.open { background: var(--success); }
        .task-status.closed { background: var(--text-muted); }

        .task-content {
            flex: 1;
            min-width: 0;
        }

        .task-title {
            font-weight: 500;
            margin-bottom: 4px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .task-meta {
            display: flex;
            gap: 8px;
            font-size: 11px;
            color: var(--text-muted);
            flex-wrap: wrap;
        }

        .task-tag {
            padding: 1px 6px;
            border-radius: 3px;
            background: var(--bg-tertiary);
        }

        .task-tag.priority-high { color: var(--danger); }
        .task-tag.priority-medium { color: var(--warning); }
        .task-tag.priority-low { color: var(--info); }

        .task-detail {
            width: 400px;
            background: var(--bg-secondary);
            border-left: 1px solid var(--border);
            overflow-y: auto;
            display: none;
        }

        .task-detail.visible {
            display: block;
        }

        .detail-header {
            padding: 16px;
            border-bottom: 1px solid var(--border);
        }

        .detail-id {
            font-size: 11px;
            color: var(--text-muted);
            font-family: inherit;
            margin-bottom: 8px;
        }

        .detail-title {
            font-size: 16px;
            font-weight: 600;
            line-height: 1.4;
            margin-bottom: 12px;
        }

        .detail-props {
            display: flex;
            flex-wrap: wrap;
            gap: 8px;
        }

        .detail-prop {
            display: flex;
            gap: 6px;
            font-size: 11px;
        }

        .detail-prop-label {
            color: var(--text-muted);
        }

        .detail-prop-value {
            color: var(--accent);
        }

        .detail-section {
            padding: 16px;
            border-bottom: 1px solid var(--border);
        }

        .detail-section-title {
            font-size: 11px;
            text-transform: uppercase;
            color: var(--text-muted);
            letter-spacing: 0.05em;
            margin-bottom: 12px;
        }

        .detail-description {
            line-height: 1.6;
            color: var(--text);
            white-space: pre-wrap;
        }

        .detail-list {
            list-style: none;
        }

        .detail-list-item {
            padding: 6px 0;
            font-size: 12px;
            color: var(--text-muted);
            border-bottom: 1px solid var(--bg-tertiary);
        }

        .detail-list-item:last-child {
            border-bottom: none;
        }

        .dep-link {
            color: var(--accent);
            text-decoration: none;
            cursor: pointer;
        }

        .dep-link:hover {
            text-decoration: underline;
            color: var(--accent-hover);
        }

        .dep-section {
            border-left: 3px solid var(--border);
            margin-bottom: 16px;
            background: var(--bg-secondary);
        }

        .dep-incoming {
            border-left-color: #d29922;
            background: rgba(210, 153, 34, 0.08);
        }

        .dep-outgoing {
            border-left-color: #3fb950;
            background: rgba(63, 185, 80, 0.08);
        }

        .dep-incoming .detail-section-title {
            color: #e3b341;
        }

        .dep-outgoing .detail-section-title {
            color: #56d364;
        }

        .dep-empty {
            color: var(--text-muted);
            font-style: italic;
            padding: 8px 0;
        }

        .empty-state {
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            height: 100%;
            color: var(--text-muted);
            gap: 12px;
        }

        .empty-state-icon {
            font-size: 48px;
            opacity: 0.5;
        }

        .shortcut-hint {
            position: fixed;
            bottom: 16px;
            right: 16px;
            background: var(--bg-secondary);
            border: 1px solid var(--border);
            border-radius: 6px;
            padding: 12px 16px;
            font-size: 11px;
            color: var(--text-muted);
            display: none;
            max-width: 90vw;
            max-height: 70vh;
            overflow-y: auto;
            z-index: 1000;
        }

        .shortcut-hint.visible {
            display: block;
        }

        .mobile-menu-btn {
            display: none;
            background: none;
            border: none;
            color: var(--text);
            font-size: 20px;
            padding: 8px;
            cursor: pointer;
        }

        .detail-back-btn {
            display: none;
            background: var(--bg-tertiary);
            border: 1px solid var(--border);
            color: var(--text);
            padding: 8px 16px;
            border-radius: 4px;
            margin-bottom: 16px;
            cursor: pointer;
            font-family: inherit;
            font-size: 13px;
        }

        /* Mobile Responsive */
        @media (max-width: 768px) {
            body {
                font-size: 14px;
            }

            .sidebar {
                position: fixed;
                left: 0;
                top: 0;
                height: 100vh;
                width: 260px;
                z-index: 100;
                transform: translateX(-100%);
                transition: transform 0.2s ease;
            }

            .sidebar.open {
                transform: translateX(0);
            }

            .sidebar-overlay {
                display: none;
                position: fixed;
                top: 0;
                left: 0;
                right: 0;
                bottom: 0;
                background: rgba(0, 0, 0, 0.7);
                z-index: 99;
            }

            .sidebar-overlay.open {
                display: block;
            }

            .main {
                width: 100%;
            }

            .header {
                padding: 8px 12px;
            }

            .mobile-menu-btn {
                display: block;
            }

            .toolbar {
                padding: 8px 12px;
                flex-wrap: wrap;
                gap: 8px;
            }

            .search-box {
                width: 100%;
                order: 1;
            }

            .filter-bar {
                order: 2;
                width: 100%;
            }

            .filter-btn {
                flex: 1;
                padding: 8px;
            }

            .content {
                position: relative;
            }

            .task-list {
                padding: 8px;
                width: 100%;
            }

            .task-item {
                padding: 16px;
                margin-bottom: 8px;
            }

            .task-title {
                font-size: 15px;
                white-space: normal;
                overflow: visible;
            }

            .task-meta {
                margin-top: 4px;
            }

            .task-detail {
                position: fixed;
                top: 0;
                right: 0;
                width: 100%;
                height: 100vh;
                z-index: 50;
                border-left: none;
            }

            .task-detail.visible {
                display: block;
            }

            .detail-back-btn {
                display: block;
            }

            .detail-header {
                padding: 12px;
            }

            .detail-title {
                font-size: 18px;
            }

            .detail-section {
                padding: 12px;
            }

            .help-bar {
                display: none;
            }

            .shortcut-hint {
                left: 16px;
                right: 16px;
                bottom: 16px;
                font-size: 12px;
            }
        }

        @media (max-width: 480px) {
            .task-tag {
                font-size: 11px;
                padding: 2px 8px;
            }

            .detail-props {
                flex-direction: column;
                gap: 4px;
            }
        }

        ::-webkit-scrollbar {
            width: 8px;
            height: 8px;
        }

        ::-webkit-scrollbar-track {
            background: var(--bg);
        }

        ::-webkit-scrollbar-thumb {
            background: var(--border);
            border-radius: 4px;
        }

        ::-webkit-scrollbar-thumb:hover {
            background: var(--text-muted);
        }
    </style>
</head>
<body>
    <div class="sidebar-overlay" id="sidebarOverlay"></div>
    <div class="app">
        <aside class="sidebar" id="sidebar">
            <div class="header">
                <h1>◈ yatd UI</h1>
            </div>
            <div class="projects-list" id="projects"></div>
            <div class="help-bar">
                <div><kbd>↑</kbd><kbd>↓</kbd> Navigate</div>
                <div><kbd>Enter</kbd> View task</div>
                <div><kbd>/</kbd> Search</div>
                <div><kbd>Esc</kbd> Close/Back</div>
                <div><kbd>?</kbd> Help</div>
            </div>
        </aside>
        <main class="main">
            <div class="toolbar">
                <button class="mobile-menu-btn" id="mobileMenuBtn">☰</button>
                <input type="text" class="search-box" id="search" placeholder="Search tasks... (/)" />
                <div class="filter-bar">
                    <button class="filter-btn active" data-filter="all">all</button>
                    <button class="filter-btn" data-filter="open">open</button>
                    <button class="filter-btn" data-filter="closed">closed</button>
                </div>
            </div>
            <div class="content">
                <div class="task-list" id="taskList"></div>
                <div class="task-detail" id="taskDetail">
                    <div class="empty-state">
                        <div class="empty-state-icon">📋</div>
                        <div>Select a task to view details</div>
                    </div>
                </div>
            </div>
        </main>
    </div>

    <div class="shortcut-hint" id="shortcutHint">
        <strong>Keyboard Shortcuts</strong><br>
        <kbd>j</kbd>/<kbd>k</kbd> or <kbd>↑</kbd>/<kbd>↓</kbd> - Navigate<br>
        <kbd>Enter</kbd> - View task details<br>
        <kbd>Esc</kbd> - Close detail panel<br>
        <kbd>/</kbd> - Focus search<br>
        <kbd>o</kbd> - Show open only<br>
        <kbd>c</kbd> - Show closed only<br>
        <kbd>a</kbd> - Show all<br>
        <kbd>r</kbd> - Refresh<br>
        <kbd>?</kbd> - Toggle this help
    </div>

    <script>
        // State
        let projects = [];
        let currentProject = '';
        let tasks = [];
        let filteredTasks = [];
        let selectedIndex = 0;
        let currentFilter = 'all';
        let searchQuery = '';
        let showHelp = false;
        let nextTasks = [];

        // DOM elements
        const projectsEl = document.getElementById('projects');
        const taskListEl = document.getElementById('taskList');
        const taskDetailEl = document.getElementById('taskDetail');
        const searchEl = document.getElementById('search');
        const shortcutHintEl = document.getElementById('shortcutHint');
        const sidebarEl = document.getElementById('sidebar');
        const sidebarOverlayEl = document.getElementById('sidebarOverlay');
        const mobileMenuBtn = document.getElementById('mobileMenuBtn');
        const nextBarContentEl = document.getElementById('nextBarContent');

        // Load projects
        async function loadProjects() {
            const res = await fetch('/api/projects');
            projects = await res.json();
            renderProjects();
            if (projects.length > 0) {
                selectProject(projects[0]);
            }
        }

        // Load next tasks
        async function loadNext(project) {
            const url = project ? '/api/next?project=' + encodeURIComponent(project) : '/api/next';
            const res = await fetch(url);
            nextTasks = await res.json();
            // Next task is rendered inline in renderTasks()
        }

        // Load tasks for project
        async function loadTasks(project) {
            const url = project ? '/api/tasks?project=' + encodeURIComponent(project) : '/api/tasks';
            const res = await fetch(url);
            tasks = await res.json();
            applyFilter();
            // Also load next tasks
            loadNext(project);
        }

        // Load task details
        async function loadTaskDetail(id) {
            const url = '/api/task?project=' + encodeURIComponent(currentProject) + '&id=' + encodeURIComponent(id);
            const res = await fetch(url);
            const task = await res.json();
            renderTaskDetail(task);
        }

        // Render projects
        function renderProjects() {
            projectsEl.innerHTML = projects.map(p =>
                '<div class="project-item ' + (p === currentProject ? 'active' : '') + '" data-project="' + p + '">' +
                p + '</div>'
            ).join('');

            projectsEl.querySelectorAll('.project-item').forEach(el => {
                el.addEventListener('click', () => selectProject(el.dataset.project));
            });
        }


        // Render next task in the next bar

        // Render task list
        // Render task list
        function renderTasks() {
            if (filteredTasks.length === 0) {
                taskListEl.innerHTML = '<div class="empty-state"><div class="empty-state-icon">📝</div><div>No tasks found</div></div>';
                return;
            }

            let html = '';

            // Render the next task first (if available and not filtered out)
            if (nextTasks && nextTasks.length > 0) {
                const nextTask = nextTasks[0];
                // Find the full task data from filteredTasks
                const nextTaskFull = filteredTasks.find(t => t.id === nextTask.id);
                if (nextTaskFull) {
                    html += '<div class="next-task-item" data-id="' + nextTask.id + '">' +
                        '<div class="task-status ' + nextTaskFull.status + '"></div>' +
                        '<div class="task-content">' +
                        '<div class="next-task-title"><span class="next-task-label">NEXT:</span>' + escapeHtml(nextTaskFull.title) + '</div>' +
                        '<div class="next-task-meta">' +
                        '<span class="task-tag">' + nextTask.id + '</span>' +
                        (nextTaskFull.priority ? '<span class="task-tag priority-' + nextTaskFull.priority + '">' + nextTaskFull.priority + '</span>' : '') +
                        (nextTaskFull.effort ? '<span class="task-tag">' + nextTaskFull.effort + '</span>' : '') +
                        '<span class="task-tag" style="color: #e3b341;">#' + nextTask.rank + ' · Score: ' + nextTask.score.toFixed(1) + '</span>' +
                        '</div>' +
                        '</div>' +
                        '</div>';
                }
            }

            // Render the rest of the tasks
            html += filteredTasks.map((t, i) =>
                '<div class="task-item ' + (i === selectedIndex ? 'selected' : '') + '" data-index="' + i + '" data-id="' + t.id + '">' +
                '<div class="task-status ' + t.status + '"></div>' +
                '<div class="task-content">' +
                '<div class="task-title">' + escapeHtml(t.title) + '</div>' +
                '<div class="task-meta">' +
                '<span class="task-tag">' + t.id + '</span>' +
                (t.priority ? '<span class="task-tag priority-' + t.priority + '">' + t.priority + '</span>' : '') +
                (t.labels ? t.labels.map(l => '<span class="task-tag">' + l + '</span>').join('') : '') +
                (t.blockers && t.blockers.length ? '<span class="task-tag">⛓ ' + t.blockers.length + '</span>' : '') +
                '</div>' +
                '</div>' +
                '</div>'
            ).join('');

            taskListEl.innerHTML = html;

            // Add click handlers
            taskListEl.querySelectorAll('.task-item').forEach(el => {
                el.addEventListener('click', () => {
                    selectedIndex = parseInt(el.dataset.index);
                    renderTasks();
                    loadTaskDetail(el.dataset.id);
                });
            });

            // Add click handler for next task
            const nextTaskEl = taskListEl.querySelector('.next-task-item');
            if (nextTaskEl) {
                nextTaskEl.addEventListener('click', () => {
                    loadTaskDetail(nextTaskEl.dataset.id);
                });
            }

            // Auto-scroll to selected
            const selected = taskListEl.querySelector('.task-item.selected');
            if (selected) {
                selected.scrollIntoView({ block: 'nearest' });
            }
        }

        // Render task detail
        function renderTaskDetail(task) {
            taskDetailEl.classList.add('visible');

            let html = '<button class="detail-back-btn" id="detailBackBtn">← Back to tasks</button>';
            html += '<div class="detail-header">';
            html += '<div class="detail-id">' + task.id + '</div>';
            html += '<div class="detail-title">' + escapeHtml(task.title) + '</div>';
            html += '<div class="detail-props">';
            html += '<div class="detail-prop"><span class="detail-prop-label">status:</span><span class="detail-prop-value">' + task.status + '</span></div>';
            html += '<div class="detail-prop"><span class="detail-prop-label">priority:</span><span class="detail-prop-value">' + task.priority + '</span></div>';
            html += '<div class="detail-prop"><span class="detail-prop-label">effort:</span><span class="detail-prop-value">' + task.effort + '</span></div>';
            html += '</div>';
            html += '</div>';

            if (task.description) {
                html += '<div class="detail-section">';
                html += '<div class="detail-section-title">Description</div>';
                html += '<div class="detail-description">' + escapeHtml(task.description) + '</div>';
                html += '</div>';
            }

            if (task.labels && task.labels.length) {
                html += '<div class="detail-section">';
                html += '<div class="detail-section-title">Labels</div>';
                html += '<div class="detail-props">';
                task.labels.forEach(l => {
                    html += '<span class="task-tag">' + l + '</span>';
                });
                html += '</div>';
                html += '</div>';
            }

            html += '<div class="detail-section dep-section dep-incoming">';
            html += '<div class="detail-section-title">⬇ Blocked By</div>';
            if (task.blockers && task.blockers.length) {
                html += '<ul class="detail-list">';
                task.blockers.forEach(b => {
                    // Find the task to get its title
                    const depTask = tasks.find(t => t.id === b);
                    const title = depTask ? depTask.title : 'Unknown';
                    html += '<li class="detail-list-item"><a href="#" class="dep-link" data-id="' + b + '">' + b + ' - ' + escapeHtml(title) + '</a></li>';
                });
                html += '</ul>';
            } else {
                html += '<div class="dep-empty">None</div>';
            }
            html += '</div>';

            // Find tasks that depend on this one (reverse dependencies)
            const dependents = tasks.filter(t => t.blockers && t.blockers.includes(task.id));
            html += '<div class="detail-section dep-section dep-outgoing">';
            html += '<div class="detail-section-title">⬆ Blocks</div>';
            if (dependents.length) {
                html += '<ul class="detail-list">';
                dependents.forEach(dep => {
                    html += '<li class="detail-list-item"><a href="#" class="dep-link" data-id="' + dep.id + '">' + dep.id + ' - ' + escapeHtml(dep.title) + '</a></li>';
                });
                html += '</ul>';
            } else {
                html += '<div class="dep-empty">None</div>';
            }
            html += '</div>';

            if (task.logs && task.logs.length) {
                html += '<div class="detail-section">';
                html += '<div class="detail-section-title">Work Log</div>';
                html += '<ul class="detail-list">';
                task.logs.forEach(entry => {
                    html += '<li class="detail-list-item">';
                    html += '<div style="color: var(--text-muted); margin-bottom: 2px;">' + formatDate(entry.timestamp) + '</div>';
                    html += '<div>' + escapeHtml(entry.message) + '</div>';
                    html += '</li>';
                });
                html += '</ul>';
                html += '</div>';
            }

            html += '<div class="detail-section">';
            html += '<div class="detail-section-title">Timestamps</div>';
            html += '<div class="detail-props">';
            html += '<div class="detail-prop"><span class="detail-prop-label">created:</span><span class="detail-prop-value">' + formatDate(task.created_at) + '</span></div>';
            html += '<div class="detail-prop"><span class="detail-prop-label">updated:</span><span class="detail-prop-value">' + formatDate(task.updated_at) + '</span></div>';
            html += '</div>';
            html += '</div>';

            taskDetailEl.innerHTML = html;
        }

        // Apply filters
        function applyFilter() {
            filteredTasks = tasks.filter(t => {
                if (currentFilter === 'open' && t.status !== 'open') return false;
                if (currentFilter === 'closed' && t.status !== 'closed') return false;
                if (searchQuery) {
                    const q = searchQuery.toLowerCase();
                    return (t.title && t.title.toLowerCase().includes(q)) ||
                           (t.description && t.description.toLowerCase().includes(q)) ||
                           (t.id && t.id.toLowerCase().includes(q));
                }
                return true;
            });
            selectedIndex = 0;
            renderTasks();
        }

        // Select project
        function selectProject(name) {
            currentProject = name;
            renderProjects();
            loadTasks(name);
            taskDetailEl.classList.remove('visible');
        }

        // Utils
        function escapeHtml(text) {
            if (!text) return '';
            return text
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;')
                .replace(/"/g, '&quot;');
        }

        function formatDate(d) {
            if (!d) return '-';
            return new Date(d).toLocaleString();
        }

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.target === searchEl) {
                if (e.key === 'Escape') {
                    searchEl.blur();
                    searchQuery = '';
                    searchEl.value = '';
                    applyFilter();
                    return;
                }
                return;
            }

            switch (e.key) {
                case 'j':
                case 'ArrowDown':
                    e.preventDefault();
                    if (selectedIndex < filteredTasks.length - 1) {
                        selectedIndex++;
                        renderTasks();
                    }
                    break;
                case 'k':
                case 'ArrowUp':
                    e.preventDefault();
                    if (selectedIndex > 0) {
                        selectedIndex--;
                        renderTasks();
                    }
                    break;
                case 'Enter':
                    e.preventDefault();
                    if (filteredTasks[selectedIndex]) {
                        loadTaskDetail(filteredTasks[selectedIndex].id);
                    }
                    break;
                case 'Escape':
                    taskDetailEl.classList.remove('visible');
                    break;
                case '/':
                    e.preventDefault();
                    searchEl.focus();
                    break;
                case 'o':
                    currentFilter = 'open';
                    updateFilterButtons();
                    applyFilter();
                    break;
                case 'c':
                    currentFilter = 'closed';
                    updateFilterButtons();
                    applyFilter();
                    break;
                case 'a':
                    currentFilter = 'all';
                    updateFilterButtons();
                    applyFilter();
                    break;
                case 'r':
                    loadTasks(currentProject);
                    break;
                case '?':
                    showHelp = !showHelp;
                    shortcutHintEl.classList.toggle('visible', showHelp);
                    break;
            }
        });

        // Search input
        searchEl.addEventListener('input', (e) => {
            searchQuery = e.target.value;
            applyFilter();
        });

        // Filter buttons
        function updateFilterButtons() {
            document.querySelectorAll('.filter-btn').forEach(btn => {
                btn.classList.toggle('active', btn.dataset.filter === currentFilter);
            });
        }

        document.querySelectorAll('.filter-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                currentFilter = btn.dataset.filter;
                updateFilterButtons();
                applyFilter();
            });
        });

        // Mobile menu toggle
        function openSidebar() {
            sidebarEl.classList.add('open');
            sidebarOverlayEl.classList.add('open');
        }

        function closeSidebar() {
            sidebarEl.classList.remove('open');
            sidebarOverlayEl.classList.remove('open');
        }

        mobileMenuBtn.addEventListener('click', openSidebar);
        sidebarOverlayEl.addEventListener('click', closeSidebar);

        // Close sidebar when project selected on mobile
        const originalSelectProject = selectProject;
        selectProject = function(name) {
            currentProject = name;
            renderProjects();
            loadTasks(name);
            taskDetailEl.classList.remove('visible');
            closeSidebar();
        };

        // Delegate click for back button and dependency links in detail panel
        taskDetailEl.addEventListener('click', (e) => {
            if (e.target.id === 'detailBackBtn') {
                taskDetailEl.classList.remove('visible');
                return;
            }
            if (e.target.classList.contains('dep-link')) {
                e.preventDefault();
                const depId = e.target.dataset.id;
                // Find the task in current tasks list
                const depTask = tasks.find(t => t.id === depId);
                if (depTask) {
                    loadTaskDetail(depId);
                } else {
                    // Task not in current list, fetch it
                    fetch('/api/task?project=' + encodeURIComponent(currentProject) + '&id=' + encodeURIComponent(depId))
                        .then(r => r.json())
                        .then(t => renderTaskDetail(t));
                }
                return;
            }
        });

        // Touch swipe support for closing detail panel
        let touchStartX = 0;
        let touchEndX = 0;

        taskDetailEl.addEventListener('touchstart', (e) => {
            touchStartX = e.changedTouches[0].screenX;
        }, {passive: true});

        taskDetailEl.addEventListener('touchend', (e) => {
            touchEndX = e.changedTouches[0].screenX;
            handleSwipe();
        }, {passive: true});

        function handleSwipe() {
            // Swipe right to close detail panel
            if (touchEndX > touchStartX + 50) {
                taskDetailEl.classList.remove('visible');
            }
        }

        // Initial load
        loadProjects();
    </script>
</body>
</html>
`))
