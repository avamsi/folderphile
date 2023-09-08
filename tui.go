package main

import (
	"fmt"
	"strings"

	"github.com/avamsi/ergo/check"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	files  []file
	cursor int
	editor editor

	// model operates in two modes: normal and quit.
	// In normal mode, the user can navigate between and edit files, and quit.
	// In the quit mode, the user is prompted to save before quitMode.
	quitMode bool
	aborted  bool // true if the user quit without saving
}

func (m model) Init() tea.Cmd {
	return nil
}

// Key map for the model in normal mode.
var keyMap = struct {
	edit, down, up, confirm, quit key.Binding
}{
	edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e:", "edit"),
	),
	down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("down / j:", "down"),
	),
	up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("  up / k:", "up"),
	),
	confirm: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("         c:", "confirm"),
	),
	quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("ctrl+c / q:", "quit"),
	),
}

func (m model) View() string {
	if m.quitMode {
		return "Save before quiting (y / n / esc)?"
	}
	var b strings.Builder
	for i, f := range m.files {
		cursor := ' '
		if i == m.cursor {
			cursor = '*'
		}
		content := '?'
		switch f.state {
		case fileUnedited:
			content = ' '
		case fileSameAsBase:
			content = 'B'
		case fileSameAsLeft:
			content = 'L'
		case fileSameAsRight:
			content = 'R'
		case fileEditedOther:
			content = 'E'
		}
		fmt.Fprintf(&b, "%c [%c] %s\n", cursor, content, f.name)
	}
	var (
		h      = help.New()
		sep    = h.Styles.FullSeparator.Render(h.FullSeparator)
		header = lipgloss.JoinHorizontal(
			lipgloss.Top,
			h.Styles.FullDesc.Render("B: base"),
			sep,
			h.Styles.FullDesc.Render("L: left"),
			sep,
			h.Styles.FullDesc.Render("R: right"),
			sep,
			h.Styles.FullDesc.Render("E: edited (other)"),
			"\n",
		)
		footer = h.FullHelpView([][]key.Binding{
			{keyMap.edit},
			{keyMap.down, keyMap.up},
			{keyMap.confirm, keyMap.quit},
		})
	)
	return lipgloss.JoinVertical(lipgloss.Left, header, b.String(), footer)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y":
				return m, tea.Quit
			case "n":
				m.aborted = true
				return m, tea.Quit
			case "esc":
				m.quitMode = false
			}
		}
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keyMap.edit):
			m.editor.edit(&m.files[m.cursor])
		case key.Matches(msg, keyMap.down):
			m.cursor = (m.cursor + 1) % len(m.files)
		case key.Matches(msg, keyMap.up):
			m.cursor = (m.cursor + len(m.files) - 1) % len(m.files)
		case key.Matches(msg, keyMap.confirm):
			return m, tea.Quit
		case key.Matches(msg, keyMap.quit):
			for _, f := range m.files {
				if f.state != fileUnedited {
					m.quitMode = true
					return m, nil
				}
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

var _ tea.Model = model{}

func tuiEditFiles(files []file, e editor) bool {
	var (
		p      = tea.NewProgram(model{files: files, editor: e}, tea.WithAltScreen())
		m, err = p.Run()
	)
	check.Nil(err)
	return !m.(model).aborted
}
