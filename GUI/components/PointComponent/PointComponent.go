package pointcomponent

import (
	"fmt"
	equationcomponent "puntosCurvaEliptica/GUI/components/EquationComponent"
	"strconv"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	inputX = iota
	inputY
)

var (
	focusedStyle = lipgloss.NewStyle().
			Bold(true)
	notFocusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
	notFocusedInput = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(4)
	focusedInput = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Align(lipgloss.Center).
			Width(4).
			Italic(true)
)

type Model struct {
	X, Y         int
	Focused      bool
	focusedInput int
	Err          error
	Inputs       []textinput.Model
}

func isAllowedRune(r rune) bool {
	if unicode.IsDigit(r) {
		return true
	}
	if r == 0 {
		return true
	}
	if r == '+' || r == '-' {
		return true
	}
	return false
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, 2)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.Focused {
			break
		}
		switch msg.Type {
		case tea.KeyEnter:
			m.nextInput()
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP, tea.KeyLeft:
			if m.Inputs[m.focusedInput].Value() != "" {
				m.prevInput()
			}
		case tea.KeyTab, tea.KeyCtrlN, tea.KeyRight:
			if m.Inputs[m.focusedInput].Value() != "" {
				m.nextInput()
			}
		}
		for i := range m.Inputs {
			m.Inputs[i].Blur()
		}
		m.Inputs[m.focusedInput].Focus()

		if len(msg.Runes) > 0 {
			if !isAllowedRune(msg.Runes[0]) {
				return m, nil
			}
		}
	}
	var err bool
	for _, inp := range m.Inputs {
		if inp.Err != nil {
			err = true
			m.Err = inp.Err
		}
	}
	if !err {
		m.Err = nil
	}
	for i := range m.Inputs {
		m.Inputs[i], cmds[i] = m.Inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var styleRender = focusedStyle.Render
	focusedInputRender := focusedInput.Render
	if !m.Focused {
		styleRender = notFocusedStyle.Render
		focusedInputRender = notFocusedInput.Render
	}
	var inputViews [2]string

	for i := range m.Inputs {
		inputViews[i] = notFocusedInput.Render(m.Inputs[i].View())
		if m.focusedInput == i {
			inputViews[i] = focusedInputRender(m.Inputs[i].View())
		}
	}

	return fmt.Sprintf(
		`%s %s %s %s %s`,
		styleRender("("),
		inputViews[inputX],
		styleRender(","),
		inputViews[inputY],
		styleRender(")"),
	)
}

func NewPointModel() Model {
	inputs := make([]textinput.Model, 2)

	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].CharLimit = 3
		inputs[i].Width = 4
		inputs[i].Prompt = ""
		inputs[i].Validate = pointValidator
	}
	inputs[inputX].Placeholder = "x"
	inputs[inputY].Placeholder = "y"

	return Model{
		Inputs:       inputs,
		Focused:      false,
		focusedInput: inputX,
		X:            -1,
		Y:            -1,
	}
}

func pointValidator(s string) error {
	pVal := equationcomponent.GlobalPValue
	if pVal == 0 {
		return nil
	}
	n, _ := strconv.Atoi(s)
	if n >= pVal {
		return fmt.Errorf("%d should be (mod %d)", n, pVal)
	}
	return nil
}

func (m *Model) nextInput() {
	m.focusedInput = (m.focusedInput + 1) % len(m.Inputs)
}

// prevInput focuses the previous input field
func (m *Model) prevInput() {
	m.focusedInput--
	// Wrap around
	if m.focusedInput < 0 {
		m.focusedInput = len(m.Inputs) - 1
	}
}

func (m *Model) Focus() {
	m.Inputs[m.focusedInput].Focus()
	m.Focused = true
}

func (m *Model) Blur() {
	for i := range m.Inputs {
		m.Inputs[i].Blur()
	}
	X, _ := strconv.Atoi(m.Inputs[inputX].Value())
	Y, _ := strconv.Atoi(m.Inputs[inputY].Value())
	m.X = X
	m.Y = Y
	m.Focused = false
}
