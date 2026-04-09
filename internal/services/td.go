package services

import (
	"encoding/json"
	"os"
	"os/exec"

	"yatd-ui/internal/models"
)

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

func RunTD(args ...string) ([]byte, error) {
	cmd := exec.Command(tdPath, args...)
	cmd.Dir = tdHome
	return cmd.Output()
}

func GetProjects() ([]string, error) {
	out, err := RunTD("project", "list", "-j")
	if err != nil {
		return nil, err
	}
	var projects []string
	if err := json.Unmarshal(out, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func GetTasks(project string) ([]models.Task, error) {
	args := []string{"-j", "list"}
	if project != "" {
		args = append(args, "--project", project)
	}
	out, err := RunTD(args...)
	if err != nil {
		return nil, err
	}
	var tasks []models.Task
	if err := json.Unmarshal(out, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func GetTask(project, id string) (*models.Task, error) {
	args := []string{"-j", "show", id}
	if project != "" {
		args = append(args, "--project", project)
	}
	out, err := RunTD(args...)
	if err != nil {
		return nil, err
	}
	var task models.Task
	if err := json.Unmarshal(out, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func GetNextTasks(project string) ([]models.NextTask, error) {
	args := []string{"-j", "next"}
	if project != "" {
		args = append(args, "--project", project)
	}
	out, err := RunTD(args...)
	if err != nil {
		return nil, err
	}
	var tasks []models.NextTask
	if err := json.Unmarshal(out, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}
