package main

import (
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type inputMode int

const (
	modeNormal inputMode = iota
	modeAdd
	modeEdit
	modeDelete
)

type statusClearMsg struct{}

type Board struct {
	tasks     []Task
	filePath  string
	colIndex  int
	rowIndex  [3]int
	width     int
	height    int
	mode      inputMode
	textInput textinput.Model
	statusMsg string
	showHelp  bool
}

func NewBoard(tasks []Task, path string) Board {
	ti := textinput.New()
	ti.CharLimit = 256

	return Board{
		tasks:     tasks,
		filePath:  path,
		textInput: ti,
	}
}

func (b Board) Init() tea.Cmd {
	return nil
}

func (b Board) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width
		b.height = msg.Height
		return b, nil

	case statusClearMsg:
		b.statusMsg = ""
		return b, nil

	case tea.KeyMsg:
		if b.mode == modeAdd || b.mode == modeEdit {
			return b.updateInput(msg)
		}
		if b.mode == modeDelete {
			return b.updateDelete(msg)
		}
		return b.updateNormal(msg)
	}

	return b, nil
}

func (b Board) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if b.showHelp {
		b.showHelp = false
		return b, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return b, tea.Quit

	case "down":
		col := b.tasksInColumn(b.colIndex)
		if b.rowIndex[b.colIndex] < len(col)-1 {
			b.rowIndex[b.colIndex]++
		} else {
			for next := b.colIndex + 1; next < 3; next++ {
				if len(b.tasksInColumn(next)) > 0 {
					b.colIndex = next
					b.rowIndex[next] = 0
					break
				}
			}
		}
	case "up":
		if b.rowIndex[b.colIndex] > 0 {
			b.rowIndex[b.colIndex]--
		} else {
			for prev := b.colIndex - 1; prev >= 0; prev-- {
				col := b.tasksInColumn(prev)
				if len(col) > 0 {
					b.colIndex = prev
					b.rowIndex[prev] = len(col) - 1
					break
				}
			}
		}

	case "right":
		b.advanceTask()
	case "left":
		b.retreatTask()

	case "b":
		b.toggleBlocked()

	case "n":
		b.mode = modeAdd
		b.textInput.SetValue("")
		b.textInput.Placeholder = "[project] description"
		b.textInput.Focus()
		return b, b.textInput.Cursor.BlinkCmd()

	case "enter":
		col := b.tasksInColumn(b.colIndex)
		if len(col) == 0 {
			return b, nil
		}
		task := col[b.rowIndex[b.colIndex]]
		b.mode = modeEdit
		b.textInput.SetValue("[" + task.Project + "] " + task.Description)
		b.textInput.Focus()
		return b, b.textInput.Cursor.BlinkCmd()

	case "x":
		col := b.tasksInColumn(b.colIndex)
		if len(col) > 0 {
			b.mode = modeDelete
			b.statusMsg = "Delete this task? (y/n)"
		}

	case "?":
		b.showHelp = !b.showHelp
	}

	if b.statusMsg != "" {
		return b, clearStatusAfter(2 * time.Second)
	}
	return b, nil
}

func (b Board) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		val := b.textInput.Value()
		if b.mode == modeAdd {
			task, err := ParseAddInput(val)
			if err != nil {
				b.statusMsg = "Error: " + err.Error()
				b.mode = modeNormal
				b.textInput.Blur()
				return b, clearStatusAfter(3 * time.Second)
			}
			b.tasks = append(b.tasks, task)
			si := stateIndex(task.State)
			b.rowIndex[si] = len(b.tasksInColumn(si)) - 1
			b.saveNow()
			b.statusMsg = "Added: " + task.String()
		} else {
			// edit mode
			parsed, err := ParseAddInput(val)
			if err != nil {
				b.statusMsg = "Error: " + err.Error()
				b.mode = modeNormal
				b.textInput.Blur()
				return b, clearStatusAfter(3 * time.Second)
			}
			idx := b.globalIndex()
			if idx >= 0 {
				parsed.State = b.tasks[idx].State
				b.tasks[idx] = parsed
				b.saveNow()
				b.statusMsg = "Updated: " + parsed.String()
			}
		}
		b.mode = modeNormal
		b.textInput.Blur()
		return b, clearStatusAfter(2 * time.Second)

	case "esc":
		b.mode = modeNormal
		b.textInput.Blur()
		b.statusMsg = ""
		return b, nil
	}

	var cmd tea.Cmd
	b.textInput, cmd = b.textInput.Update(msg)
	return b, cmd
}

func (b Board) updateDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		idx := b.globalIndex()
		if idx >= 0 {
			b.tasks = append(b.tasks[:idx], b.tasks[idx+1:]...)
			b.saveNow()
			// adjust cursor
			col := b.tasksInColumn(b.colIndex)
			if b.rowIndex[b.colIndex] >= len(col) && len(col) > 0 {
				b.rowIndex[b.colIndex] = len(col) - 1
			}
			b.statusMsg = "Deleted"
		}
		b.mode = modeNormal
		return b, clearStatusAfter(2 * time.Second)
	default:
		b.mode = modeNormal
		b.statusMsg = ""
		return b, nil
	}
}

func (b *Board) moveTask(dir int) {
	col := b.tasksInColumn(b.colIndex)
	if len(col) == 0 {
		return
	}

	newStateIdx := b.colIndex + dir
	if newStateIdx < 0 || newStateIdx > 2 {
		return
	}

	idx := b.globalIndex()
	if idx < 0 {
		return
	}

	oldState := b.tasks[idx].State
	b.tasks[idx].State = validStates[newStateIdx]
	b.saveNow()

	// adjust cursor in old column
	oldCol := b.tasksInColumn(b.colIndex)
	if b.rowIndex[b.colIndex] >= len(oldCol) && len(oldCol) > 0 {
		b.rowIndex[b.colIndex] = len(oldCol) - 1
	} else if len(oldCol) == 0 {
		b.rowIndex[b.colIndex] = 0
	}

	// move focus to new column and select the moved task
	b.colIndex = newStateIdx
	newCol := b.tasksInColumn(newStateIdx)
	row := 0
	for i, t := range newCol {
		if t == b.tasks[idx] {
			row = i
			break
		}
	}
	b.rowIndex[newStateIdx] = row

	b.statusMsg = "Moved: [" + oldState + "] → [" + validStates[newStateIdx] + "]"
}

func (b *Board) advanceTask() {
	col := b.tasksInColumn(b.colIndex)
	if len(col) == 0 {
		return
	}

	idx := b.globalIndex()
	if idx < 0 {
		return
	}

	si := stateIndex(b.tasks[idx].State)
	if si >= 2 {
		return
	}

	b.moveTask(1)
}

func (b *Board) retreatTask() {
	col := b.tasksInColumn(b.colIndex)
	if len(col) == 0 {
		return
	}

	idx := b.globalIndex()
	if idx < 0 {
		return
	}

	si := stateIndex(b.tasks[idx].State)
	if si <= 0 {
		return
	}

	b.moveTask(-1)
}

func (b *Board) toggleBlocked() {
	col := b.tasksInColumn(b.colIndex)
	if len(col) == 0 {
		return
	}

	idx := b.globalIndex()
	if idx < 0 {
		return
	}

	desc := b.tasks[idx].Description
	if strings.Contains(desc, "#blocked") {
		desc = strings.TrimSpace(strings.ReplaceAll(desc, "#blocked", ""))
		b.tasks[idx].Description = desc
		b.statusMsg = "Unblocked"
	} else {
		b.tasks[idx].Description = desc + " #blocked"
		b.statusMsg = "Blocked"
	}
	b.saveNow()
}

func (b Board) tasksInColumn(col int) []Task {
	state := validStates[col]
	var result []Task
	for _, t := range b.tasks {
		if t.State == state {
			result = append(result, t)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		pi, pj := result[i].Project, result[j].Project
		bi := strings.Contains(result[i].Description, "#blocked")
		bj := strings.Contains(result[j].Description, "#blocked")
		if pi != pj {
			return pi < pj
		}
		if bi != bj {
			return !bi
		}
		return false
	})
	return result
}

func (b Board) globalIndex() int {
	col := b.tasksInColumn(b.colIndex)
	if len(col) == 0 || b.rowIndex[b.colIndex] >= len(col) {
		return -1
	}
	target := col[b.rowIndex[b.colIndex]]

	for i, t := range b.tasks {
		if t == target {
			return i
		}
	}
	return -1
}

func (b *Board) saveNow() {
	if err := WriteTasks(b.filePath, b.tasks); err != nil {
		b.statusMsg = "Error saving: " + err.Error()
	}
}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return statusClearMsg{}
	})
}

func (b Board) View() string {
	if b.width == 0 {
		return "Loading..."
	}

	if b.showHelp {
		return b.renderHelp()
	}

	colWidth := b.width

	// compute widest project name across all tasks
	maxProj := 0
	for _, t := range b.tasks {
		w := len(t.Project) + 2 // brackets
		if w > maxProj {
			maxProj = w
		}
	}

	var sections []string
	for i := 0; i < 3; i++ {
		sections = append(sections, b.renderColumn(i, colWidth, 0, maxProj))
		if i < 2 {
			sections = append(sections, "", "")
		}
	}
	board := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// hint/status bar — status temporarily replaces hints
	var bottomBar string
	if b.statusMsg != "" {
		bottomBar = statusBarStyle.Render(b.statusMsg)
	} else {
		bottomBar = helpStyle.Render("↑↓:navigate  ←→:move  b:blocked  n:new  enter:edit  x:delete  q:quit  ?:help")
	}

	return board + "\n\n\n" + bottomBar
}

func (b Board) renderColumn(col int, width int, height int, maxProj int) string {
	state := validStates[col]
	color := stateColors[state]
	focused := col == b.colIndex
	tasks := b.tasksInColumn(col)

	// header
	title := strings.ToUpper(state)
	count := len(tasks)
	if focused {
		headerStyle = headerStyle.Underline(true)
	} else {
		headerStyle = headerStyle.Underline(false)
	}
	header := headerStyle.
		Foreground(color).
		Width(width).
		Render(title + " (" + itoa(count) + ")")

	var lines []string
	lines = append(lines, header)

	for i, t := range tasks {
		// show inline text input for edit mode
		if focused && i == b.rowIndex[col] && b.mode == modeEdit {
			lines = append(lines, taskStyle.Width(width).Render(b.textInput.View()))
			continue
		}

		proj := "[" + t.Project + "]"
		padded := proj + strings.Repeat(" ", maxProj-len(proj))
		descWidth := width - maxProj - 3 // padding + gap
		desc := t.Description
		blocked := strings.Contains(desc, "#blocked")
		if blocked {
			desc = strings.TrimSpace(strings.ReplaceAll(desc, "#blocked", ""))
		}
		desc = truncate(desc, descWidth)

		if focused && i == b.rowIndex[col] {
			content := padded + " " + desc
			if blocked {
				content += " " + blockedTagStyle.Render("#blocked")
			}
			lines = append(lines, selectedTaskStyle.Foreground(color).Width(width).MaxHeight(1).Render(content))
		} else {
			content := projectStyle.Render(padded) + " " + desc
			if blocked {
				content += " " + blockedTagStyle.Render("#blocked")
			}
			lines = append(lines, taskStyle.Width(width).MaxHeight(1).Render(content))
		}
	}

	// show inline text input for add mode at end of focused section
	if focused && b.mode == modeAdd {
		lines = append(lines, taskStyle.Width(width).Render(b.textInput.View()))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (b Board) renderHelp() string {
	help := `
  kenban — keyboard shortcuts

  Navigation
    ↑ / ↓            Move cursor (traverses all sections)

  Actions
    →            Advance task (todo→doing→done)
    ←            Move task back one column
    b            Toggle #blocked tag
    n            Add new task
    Enter        Edit selected task
    x            Delete selected task

  File
    q / Ctrl+C   Quit (all changes saved automatically)

  ?              Toggle this help

  Press any key to close...
`
	return helpStyle.Render(help)
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
