package main

import (
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
	rowIndex  [4]int
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

	case "tab":
		b.colIndex = (b.colIndex + 1) % 4
	case "shift+tab":
		b.colIndex = (b.colIndex + 3) % 4
	case "right":
		// move right within row: 0->1, 2->3
		if b.colIndex%2 == 0 {
			b.colIndex++
		}
	case "left":
		// move left within row: 1->0, 3->2
		if b.colIndex%2 == 1 {
			b.colIndex--
		}
	case "down":
		col := b.tasksInColumn(b.colIndex)
		if b.rowIndex[b.colIndex] < len(col)-1 {
			b.rowIndex[b.colIndex]++
		} else if b.colIndex < 2 {
			// wrap to row below
			b.colIndex += 2
		}
	case "up":
		if b.rowIndex[b.colIndex] > 0 {
			b.rowIndex[b.colIndex]--
		} else if b.colIndex >= 2 {
			// wrap to row above
			b.colIndex -= 2
		}

	case "enter":
		b.advanceTask()
	case "b":
		b.retreatTask()

	case "a":
		b.mode = modeAdd
		b.textInput.SetValue("")
		b.textInput.Placeholder = "[project] description"
		b.textInput.Focus()
		return b, b.textInput.Cursor.BlinkCmd()

	case "e":
		col := b.tasksInColumn(b.colIndex)
		if len(col) == 0 {
			return b, nil
		}
		task := col[b.rowIndex[b.colIndex]]
		b.mode = modeEdit
		b.textInput.SetValue("[" + task.Project + "] " + task.Description)
		b.textInput.Focus()
		return b, b.textInput.Cursor.BlinkCmd()

	case "d":
		col := b.tasksInColumn(b.colIndex)
		if len(col) > 0 {
			b.mode = modeDelete
			b.statusMsg = "Delete this task? (y/n)"
		}

	case "?":
		b.showHelp = !b.showHelp
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
	if newStateIdx < 0 || newStateIdx > 3 {
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
	newCol := b.tasksInColumn(newStateIdx)
	b.colIndex = newStateIdx
	b.rowIndex[newStateIdx] = len(newCol) - 1

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
	if si >= 3 {
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

func (b Board) tasksInColumn(col int) []Task {
	state := validStates[col]
	var result []Task
	for _, t := range b.tasks {
		if t.State == state {
			result = append(result, t)
		}
	}
	return result
}

func (b Board) globalIndex() int {
	col := b.tasksInColumn(b.colIndex)
	if len(col) == 0 || b.rowIndex[b.colIndex] >= len(col) {
		return -1
	}
	target := col[b.rowIndex[b.colIndex]]

	count := 0
	for i, t := range b.tasks {
		if t.State == validStates[b.colIndex] {
			if count == b.rowIndex[b.colIndex] {
				_ = target
				return i
			}
			count++
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

	colWidth := b.width / 2
	colHeight := (b.height - 4) / 2 // leave room for status + hints
	if colWidth < 20 {
		return "Terminal too narrow."
	}

	topRow := lipgloss.JoinHorizontal(lipgloss.Top,
		b.renderColumn(0, colWidth, colHeight),
		b.renderColumn(1, colWidth, colHeight),
	)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top,
		b.renderColumn(2, colWidth, colHeight),
		b.renderColumn(3, colWidth, colHeight),
	)
	board := lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)

	// input bar
	var inputBar string
	if b.mode == modeAdd || b.mode == modeEdit {
		label := "Add"
		if b.mode == modeEdit {
			label = "Edit"
		}
		inputBar = "\n " + label + ": " + b.textInput.View()
	}

	// status bar
	statusBar := statusBarStyle.Render(b.statusMsg)

	// hint bar
	hints := helpStyle.Render("tab:navigate  enter:advance  b:back  a:add  e:edit  d:delete  q:quit  ?:help")

	return board + inputBar + "\n" + statusBar + "\n" + hints
}

func (b Board) renderColumn(col int, width int, height int) string {
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

	// find widest project name in this column
	maxProj := 0
	for _, t := range tasks {
		w := len(t.Project) + 2 // brackets
		if w > maxProj {
			maxProj = w
		}
	}

	// tasks
	maxTasks := height - 2 // leave room for header
	if maxTasks < 1 {
		maxTasks = 1
	}

	var lines []string
	lines = append(lines, header)

	for i, t := range tasks {
		if i >= maxTasks {
			lines = append(lines, taskStyle.Width(width).Faint(true).Render("..."))
			break
		}

		proj := "[" + t.Project + "]"
		padded := proj + strings.Repeat(" ", maxProj-len(proj))
		descWidth := width - maxProj - 3 // padding + gap
		desc := truncate(t.Description, descWidth)

		if focused && i == b.rowIndex[col] {
			content := padded + " " + desc
			lines = append(lines, selectedTaskStyle.Foreground(color).Width(width).MaxHeight(1).Render(content))
		} else {
			content := projectStyle.Render(padded) + " " + desc
			lines = append(lines, taskStyle.Width(width).MaxHeight(1).Render(content))
		}
	}

	// pad to fill height so columns align
	for len(lines) < height {
		lines = append(lines, taskStyle.Width(width).Render(""))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (b Board) renderHelp() string {
	help := `
  kenban — keyboard shortcuts

  Navigation
    Tab / →          Next column
    Shift+Tab / ←    Previous column
    ↑ / ↓            Move cursor up/down

  Actions
    Enter        Advance task (todo→doing→done)
    b            Move task back one column
    a            Add new task
    e            Edit selected task
    d            Delete selected task

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
