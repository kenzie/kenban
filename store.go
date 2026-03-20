package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ResolvePath() string {
	local := filepath.Join(".", "tasks.txt")
	if _, err := os.Stat(local); err == nil {
		return local
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return local
	}
	return filepath.Join(home, ".kenban", "tasks.txt")
}

func ReadTasks(path string) ([]Task, error) {
	ensureFileExists(path)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var tasks []Task
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		task, err := ParseTask(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: line %d: %s\n", lineNum, err)
			continue
		}
		tasks = append(tasks, task)
	}
	return tasks, scanner.Err()
}

func WriteTasks(path string, tasks []Task) error {
	ensureDirExists(path)

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "kenban-*.txt")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	for _, t := range tasks {
		if _, err := fmt.Fprintln(tmp, t.String()); err != nil {
			tmp.Close()
			os.Remove(tmpName)
			return err
		}
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

func AppendTask(path string, task Task) error {
	ensureFileExists(path)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, task.String())
	return err
}

func ensureFileExists(path string) {
	ensureDirExists(path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.WriteFile(path, []byte{}, 0644)
	}
}

func ensureDirExists(path string) {
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)
}
