package main

import (
	"fmt"
	"regexp"
	"strings"
)

var validStates = []string{"todo", "doing", "done"}

var linePattern = regexp.MustCompile(`^\[(\w+)\]\s+\[([^\]]+)\]\s+(.+)$`)
var addPattern = regexp.MustCompile(`^\[([^\]]+)\]\s+(.+)$`)

type Task struct {
	State       string
	Project     string
	Description string
}

func ParseTask(line string) (Task, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return Task{}, fmt.Errorf("empty line")
	}

	m := linePattern.FindStringSubmatch(line)
	if m == nil {
		return Task{}, fmt.Errorf("malformed task line: %q", line)
	}

	state := strings.ToLower(m[1])
	desc := strings.TrimSpace(m[3])
	if desc == "" {
		return Task{}, fmt.Errorf("task description cannot be empty")
	}

	// migrate old "blocked" state to todo with #blocked tag
	if state == "blocked" {
		state = "todo"
		if !strings.Contains(desc, "#blocked") {
			desc = desc + " #blocked"
		}
	}

	if !isValidState(state) {
		return Task{}, fmt.Errorf("unknown state [%s]. Valid states: %s", m[1], strings.Join(validStates, ", "))
	}

	return Task{State: state, Project: m[2], Description: desc}, nil
}

func ParseAddInput(input string) (Task, error) {
	input = strings.TrimSpace(input)
	m := addPattern.FindStringSubmatch(input)
	if m == nil {
		return Task{}, fmt.Errorf("expected format: [project] description")
	}

	desc := strings.TrimSpace(m[2])
	if desc == "" {
		return Task{}, fmt.Errorf("task description cannot be empty")
	}

	return Task{State: "todo", Project: m[1], Description: desc}, nil
}

func (t Task) String() string {
	return fmt.Sprintf("[%s] [%s] %s", t.State, t.Project, t.Description)
}

func (t Task) MatchesProject(name string) bool {
	return strings.EqualFold(t.Project, name)
}

func (t Task) MatchesState(name string) bool {
	return strings.EqualFold(t.State, name)
}

func isValidState(s string) bool {
	for _, v := range validStates {
		if v == s {
			return true
		}
	}
	return false
}

func stateIndex(s string) int {
	for i, v := range validStates {
		if v == s {
			return i
		}
	}
	return -1
}
