package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

func runCLI(args []string) {
	if len(args) == 0 {
		cmdOpen()
		return
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		cmdHelp()
		return
	}

	command := args[0]
	rest := args[1:]

	switch command {
	case "add":
		cmdAdd(rest)
	case "list":
		cmdList(rest)
	case "state":
		cmdState(rest)
	case "move":
		cmdMove(rest)
	case "done":
		cmdDone(rest)
	case "edit":
		cmdEdit(rest)
	case "copy":
		cmdCopy(rest)
	case "projects":
		cmdProjects()
	case "open":
		cmdOpen()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nRun 'kenban help' for usage.\n", command)
		os.Exit(1)
	}
}

func cmdAdd(args []string) {
	input := strings.Join(args, " ")
	if input == "" {
		fmt.Fprintln(os.Stderr, "Usage: kenban add \"[project] description\"")
		os.Exit(1)
	}

	task, err := ParseAddInput(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	path := ResolvePath()
	if err := AppendTask(path, task); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("Added:", task)
}

func cmdList(args []string) {
	path := ResolvePath()
	tasks, err := ReadTasks(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	filter := ""
	if len(args) > 0 {
		filter = args[0]
	}

	for i, task := range tasks {
		if filter != "" && !task.MatchesProject(filter) {
			continue
		}
		fmt.Printf("%3d. %s\n", i+1, task)
	}
}

func cmdState(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: kenban state <%s>\n", strings.Join(validStates, "|"))
		os.Exit(1)
	}

	stateName := strings.ToLower(args[0])
	if !isValidState(stateName) {
		fmt.Fprintf(os.Stderr, "Unknown state: %s\nValid states: %s\n", args[0], strings.Join(validStates, ", "))
		os.Exit(1)
	}

	path := ResolvePath()
	tasks, err := ReadTasks(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	for i, task := range tasks {
		if task.MatchesState(stateName) {
			fmt.Printf("%3d. %s\n", i+1, task)
		}
	}
}

func cmdMove(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: kenban move <task_number> <state>")
		os.Exit(1)
	}

	index := parseTaskNumber(args[0])
	newState := strings.ToLower(args[1])
	if !isValidState(newState) {
		fmt.Fprintf(os.Stderr, "Unknown state: %s\nValid states: %s\n", args[1], strings.Join(validStates, ", "))
		os.Exit(1)
	}

	path := ResolvePath()
	tasks, err := ReadTasks(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	validateIndex(index, len(tasks))

	oldState := tasks[index].State
	tasks[index].State = newState
	tasks[index].StampDone()
	// blocked items moving out of todo get unblocked
	if newState != "todo" && strings.Contains(tasks[index].Description, "#blocked") {
		tasks[index].Description = strings.TrimSpace(strings.ReplaceAll(tasks[index].Description, "#blocked", ""))
	}
	if err := WriteTasks(path, tasks); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Moved: [%s] -> [%s] [%s] %s\n", oldState, newState, tasks[index].Project, tasks[index].Description)
}

func cmdDone(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: kenban done <task_number>")
		os.Exit(1)
	}
	cmdMove([]string{args[0], "done"})
}

func cmdEdit(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: kenban edit <task_number>")
		os.Exit(1)
	}

	index := parseTaskNumber(args[0])
	path := ResolvePath()
	tasks, err := ReadTasks(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	validateIndex(index, len(tasks))

	task := tasks[index]
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	tmp, err := os.CreateTemp("", "kenban-edit-*.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	tmp.WriteString(task.String())
	tmp.Close()
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Editor error: %s\n", err)
		os.Exit(1)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	newLine := strings.TrimSpace(string(data))
	if newLine == "" {
		fmt.Fprintln(os.Stderr, "Edit cancelled (empty line).")
		os.Exit(1)
	}

	newTask, err := ParseTask(newLine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid task format: %s\nOriginal task preserved.\n", err)
		os.Exit(1)
	}

	tasks[index] = newTask
	if err := WriteTasks(path, tasks); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("Updated:", newTask)
}

func cmdCopy(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: kenban copy <task_number>")
		os.Exit(1)
	}

	index := parseTaskNumber(args[0])
	path := ResolvePath()
	tasks, err := ReadTasks(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	validateIndex(index, len(tasks))

	line := tasks[index].String()
	clipCmd := clipboardCommand()
	if clipCmd == "" {
		fmt.Fprintln(os.Stderr, "No clipboard command found. Install xclip or xsel on Linux.")
		os.Exit(1)
	}

	cmd := exec.Command("sh", "-c", clipCmd)
	cmd.Stdin = strings.NewReader(line)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Clipboard error: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("Copied:", line)
}

func cmdProjects() {
	path := ResolvePath()
	tasks, err := ReadTasks(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	seen := make(map[string]string) // lowercase -> original
	for _, t := range tasks {
		key := strings.ToLower(t.Project)
		if _, ok := seen[key]; !ok {
			seen[key] = t.Project
		}
	}

	var projects []string
	for _, v := range seen {
		projects = append(projects, v)
	}
	sort.Slice(projects, func(i, j int) bool {
		return strings.ToLower(projects[i]) < strings.ToLower(projects[j])
	})

	for _, p := range projects {
		fmt.Println(p)
	}
}

func cmdOpen() {
	path := ResolvePath()
	tasks, err := ReadTasks(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading tasks: %s\n", err)
		os.Exit(1)
	}

	board := NewBoard(tasks, path)
	runTUI(board)
}

func cmdHelp() {
	fmt.Print(`kenban - plain-text kanban for solo developers

Usage:
  kenban add "[project] description"   Add a new task (state: todo)
  kenban list [project]                List tasks, optionally filtered by project
  kenban state <state>                 List tasks filtered by state
  kenban move <number> <state>         Move a task to a new state
  kenban done <number>                 Shortcut: move task to done
  kenban edit <number>                 Edit a task in $EDITOR
  kenban copy <number>                 Copy a task to clipboard
  kenban projects                      List all project names
  kenban open                          Open the interactive kanban board

States: todo, doing, blocked, done

Task format:
  [state] [project] description @tag +label

File location:
  ./tasks.txt (if present) or ~/.kenban/tasks.txt
`)
}

func clipboardCommand() string {
	if runtime.GOOS == "darwin" {
		return "pbcopy"
	}
	if _, err := exec.LookPath("xclip"); err == nil {
		return "xclip -selection clipboard"
	}
	if _, err := exec.LookPath("xsel"); err == nil {
		return "xsel --clipboard --input"
	}
	return ""
}

func parseTaskNumber(s string) int {
	num, err := strconv.Atoi(s)
	if err != nil || num < 1 {
		fmt.Fprintf(os.Stderr, "Invalid task number: %s\n", s)
		os.Exit(1)
	}
	return num - 1
}

func validateIndex(index, size int) {
	if index < 0 || index >= size {
		fmt.Fprintf(os.Stderr, "Task number out of range. You have %d task(s).\n", size)
		os.Exit(1)
	}
}
