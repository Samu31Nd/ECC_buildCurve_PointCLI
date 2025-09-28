package minicheckcomponent

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().
			Bold(true)
	notFocusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
	notFocusedInput = lipgloss.NewStyle()
	focusedInput    = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Italic(true)
)

type Model struct {
	option  string
	Check   bool
	focused bool
}

func NewCheckModel(text string) Model {
	return Model{
		option:  text,
		Check:   false,
		focused: false,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused {
				m.Check = !m.Check
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	var styleRender = focusedStyle.Render
	focusedInputRender := focusedInput.Render
	if !m.focused {
		styleRender = notFocusedStyle.Render
		focusedInputRender = notFocusedInput.Render
	}
	check := "[   ]"
	if m.Check {
		check = "[ X ]"
	}
	return fmt.Sprintf("%s %s", focusedInputRender(check), styleRender(m.option))
}

func (m *Model) Focus() {
	m.focused = true
}

func (m *Model) Blur() {
	m.focused = false
}
