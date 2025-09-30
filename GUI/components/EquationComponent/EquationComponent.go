package equationcomponent

import (
	"fmt"
	"math"
	mathcurve "puntosCurvaEliptica/MathCurve"
	"strconv"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	inputP = iota
	inputA
	inputB
)

func primeValidator(s string) error {
	if s == "" {
		return nil
	}
	n, _ := strconv.Atoi(s)
	if !isPrime(n) {
		return fmt.Errorf("%d (%s) no es un numero primo", n, s)
	}
	// Checar que cumpla p == 3 (mod 4)
	// no siempre aplica
	// if (n % 4) != 3 {
	// 	return fmt.Errorf("el numero primo no cumple p ≡ 3 (mod 4) | p ≡ %d (mod 4)", n%4)
	// }
	GlobalPValue = n
	return nil
}

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func numValidator(s string) error {
	if s == "" || GlobalPValue == 0 {
		return nil
	}
	n, _ := strconv.Atoi(s)
	if n >= GlobalPValue {
		return fmt.Errorf("invalid, %s must be in (mod %d)", s, GlobalPValue)
	}
	return nil
}

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

	GlobalPValue = 0
)

type Model struct {
	A, B, P      int
	negA, negB   bool
	Focused      bool
	focusedInput int
	Err          error
	Inputs       []textinput.Model
	ValidPoints  []mathcurve.Point // only filled after isValidCurve is called and return no error
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

func allInputsAreFill(inputs []textinput.Model) bool {
	filled := true

	for _, v := range inputs {
		if v.Value() == "" {
			filled = false
		}
	}

	return filled
}

func (m *Model) IsValidCurve() error {
	if !allInputsAreFill(m.Inputs) {
		return fmt.Errorf("not all the inputs are filled")
	}
	m.ValidPoints = nil
	if m.Err != nil {
		return nil
	}
	// CHECAR SINGULARIDADES
	a, _ := strconv.Atoi(m.Inputs[inputA].Value())
	if m.negA {
		a = a * (-1)
	}
	m.A = a
	b, _ := strconv.Atoi(m.Inputs[inputB].Value())
	if m.negB {
		b = b * (-1)
	}
	m.B = b
	delta := int((4*math.Pow(float64(a), 3) + 27*math.Pow(float64(b), 2))) % GlobalPValue
	if delta == 0 {
		return fmt.Errorf("la ecuacion tiene singularidades, valor de delta = %d", delta)
	}

	// CONSEGUIR LOS PUNTOS DE LA CURVA Y GUARDARLOS
	m.ValidPoints = GetPoints(a, b, GlobalPValue)
	m.P = GlobalPValue
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, 3)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.Focused {
			break
		}
		switch msg.String() {
		case "-":
			switch m.focusedInput {
			case inputA:
				m.negA = true
				return m, nil
			case inputB:
				m.negB = true
				return m, nil
			}
		case "+":
			switch m.focusedInput {
			case inputA:
				m.negA = false
				return m, nil
			case inputB:
				m.negB = false
				return m, nil
			}
		}
		switch msg.Type {
		case tea.KeyEnter:
			m.nextInput()
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
				// Ignoramos el input
				return m, nil
			}
		}
	}

	for i := range m.Inputs {
		m.Inputs[i], cmds[i] = m.Inputs[i].Update(msg)
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
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var styleRender = focusedStyle.Render
	focusedInputRender := focusedInput.Render
	if !m.Focused {
		styleRender = notFocusedStyle.Render
		focusedInputRender = notFocusedInput.Render
	}
	var inputViews [3]string
	var signs [2]string
	signs[0] = "+"
	signs[1] = "+"
	if m.negA {
		signs[0] = "-"
	}

	if m.negB {
		signs[1] = "-"
	}

	for i := range m.Inputs {
		inputViews[i] = notFocusedInput.Render(m.Inputs[i].View())
		if m.focusedInput == i {
			inputViews[i] = focusedInputRender(m.Inputs[i].View())
		}
	}

	return fmt.Sprintf(
		`%s %s %s %s %s %s%s`,
		styleRender("y2 = x3 "+signs[0]),
		inputViews[inputA],
		styleRender(signs[1]+" x"),
		inputViews[inputB],
		styleRender("(mod"),
		inputViews[inputP],
		styleRender(")"),
	)
}

func NewEqModel() Model {
	inputs := make([]textinput.Model, 3)

	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].CharLimit = 3
		inputs[i].Width = 4
		inputs[i].Prompt = ""
	}
	inputs[inputA].Placeholder = "a"
	inputs[inputA].Validate = numValidator
	inputs[inputB].Placeholder = "b"
	inputs[inputB].Validate = numValidator
	inputs[inputP].Placeholder = "p"
	inputs[inputP].Validate = primeValidator

	inputs[inputP].Focus()

	return Model{
		Inputs:       inputs,
		Focused:      true,
		focusedInput: inputP,
	}
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

func GetPoints(a, b, p int) (stackPoints []mathcurve.Point) {
	critEulerPow := (p - 1) / 2
	raicesPow := (p + 1) / 4
	for i := range p {
		z := ((i*i*i+i*a+b)%p + p) % p
		// caso peculiar de zero
		if z == 0 {
			//TODO: agregar a la pila
			//fmt.Println(i, "0") // ESTO ES EL PUNTO
			stackPoints = append(stackPoints,
				mathcurve.Point{X: i, Y: 0})
			continue
		}
		if int(math.Pow(float64(z), float64(critEulerPow)))%p != 1 {
			continue
		}

		yAbsoluto := int(math.Pow(float64(z), float64(raicesPow)))

		y1 := ((yAbsoluto % p) + p) % p
		y2 := ((-yAbsoluto % p) + p) % p

		stackPoints = append(stackPoints,
			mathcurve.Point{X: i, Y: y2},
			mathcurve.Point{X: i, Y: y1})
	}
	return
}

func (m *Model) Focus() {
	m.Inputs[m.focusedInput].Focus()
	m.Focused = true
}

func (m *Model) Blur() {
	for i := range m.Inputs {
		m.Inputs[i].Blur()
	}
	m.Focused = false
}
