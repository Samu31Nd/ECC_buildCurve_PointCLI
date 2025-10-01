package gui

import (
	"fmt"
	"log"
	equationcomponent "puntosCurvaEliptica/GUI/components/EquationComponent"
	minicheckcomponent "puntosCurvaEliptica/GUI/components/MiniCheckComponent"
	pointcomponent "puntosCurvaEliptica/GUI/components/PointComponent"
	rightview "puntosCurvaEliptica/GUI/components/rightView"
	mathcurve "puntosCurvaEliptica/MathCurve"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	eqView = iota
	p1View
	checkView
	p2View
	proceedView
)

var (
	titleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1).
			Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true)

	noErrorText      = lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Italic(true).Render("No se han encontrado errores (aún...)")
	helpStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpText         = helpStyle.Render("Help: Ctrl + h")
	fullHelpText     = helpStyle.Render("e: Ecuation form mode ● p: Table Points mode ● g: Graph with points mode ● Ctrl+C: exit.")
	leftWindowStyle  = lipgloss.NewStyle()
	rightWindowStyle = lipgloss.NewStyle()
)

type SumPointsModel struct {
	eqComp         equationcomponent.Model
	p1, p2         pointcomponent.Model
	checkInput     minicheckcomponent.Model
	proceed        bool
	finalPoint     mathcurve.Point
	focusedInput   int
	exit           bool
	globalError    error
	rightView      bool
	rightComponent rightview.Model
	width, height  int
	displayHelp    bool
}

func (m SumPointsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SumPointsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, 6)
	var err, customMsgRight bool
	var errorValue error

	if m.checkInput.Check {
		m.p2.Inputs[0].SetValue(m.p1.Inputs[0].Value())
		m.p2.Inputs[1].SetValue(m.p1.Inputs[1].Value())
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		leftWindowStyle = leftWindowStyle.Height(msg.Height / 2).Width(msg.Width / 2)
		rightWindowStyle = rightWindowStyle.Height(msg.Height / 2).Width(msg.Height / 2)
		customMsgRight = true
		rightWindowStyle = rightWindowStyle.Width(m.width/2 - 2)
		m.rightComponent, cmds[4] = m.rightComponent.Update(rightview.CustomSizeMsg{Width: msg.Width/2 - 4, Height: msg.Height - 5})

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.exit = true
			return m, tea.Quit
		case "e":
			m.rightComponent.Blur()
			m.rightView = false
		case "p", "g":
			m.rightComponent.Focus(msg.String())
			m.rightView = true
		case "ctrl+h":
			m.displayHelp = !m.displayHelp
		}

		if m.rightView {
			break
		}
		switch msg.Type {
		case tea.KeyEnter:
			if m.focusedInput == proceedView {
				m.finalPoint, errorValue = mathcurve.AddPoints(
					m.p1.X, m.p1.Y,
					m.p2.X, m.p2.Y,
					m.eqComp.A,
					m.eqComp.P,
				)

				if errorValue != nil {
					err = true
					m.globalError = errorValue
				}
				customMsgRight = true
				m.rightComponent, cmds[4] = m.rightComponent.Update(rightview.PointMsg{
					Point: m.finalPoint,
					Index: 2,
				})
				m.proceed = true
			}
		case tea.KeyUp:
			switch m.focusedInput {
			case eqView:
				if err := m.eqComp.IsValidCurve(); err != nil {
					m.globalError = err
					break
				}
				m.prevInput()
				customMsgRight = true
				m.rightComponent.P = m.eqComp.P
				m.rightComponent, cmds[4] = m.rightComponent.Update(rightview.NewValuesTableMsg{Points: m.eqComp.ValidPoints})
			case p1View:
				m.prevInput()
				customMsgRight = true
				m.rightComponent, cmds[4] = m.rightComponent.Update(rightview.PointMsg{Point: mathcurve.Point{
					X: m.p1.X,
					Y: m.p1.Y,
				},
					IsP1SameAsP2: m.checkInput.Check,
					Index:        0,
				})
				// cmds[5] = tea.Printf("%d.{%d,%d}", 1, m.p1.X, m.p1.Y)
			case p2View:
				m.prevInput()
				customMsgRight = true
				m.rightComponent, cmds[4] = m.rightComponent.Update(rightview.PointMsg{Point: mathcurve.Point{
					X: m.p2.X,
					Y: m.p2.Y,
				},
					IsP1SameAsP2: m.checkInput.Check,
					Index:        1,
				})
				// cmds[5] = tea.Printf("%d.{%d,%d}", 2, m.p2.X, m.p2.Y)
			default:
				m.prevInput()

			}

		case tea.KeyDown:
			switch m.focusedInput {
			case eqView:
				if err := m.eqComp.IsValidCurve(); err != nil {
					m.globalError = err
					break
				}
				m.nextInput()
				customMsgRight = true
				m.rightComponent.P = m.eqComp.P
				m.rightComponent, cmds[4] = m.rightComponent.Update(rightview.NewValuesTableMsg{Points: m.eqComp.ValidPoints})
			case p1View:
				m.nextInput()
				customMsgRight = true
				m.rightComponent, cmds[4] = m.rightComponent.Update(rightview.PointMsg{Point: mathcurve.Point{
					X: m.p1.X,
					Y: m.p1.Y,
				}, IsP1SameAsP2: m.checkInput.Check, Index: 0})
				// cmds[5] = tea.Printf("%d.{%d,%d}", 1, m.p1.X, m.p1.Y)
			case p2View:
				m.nextInput()
				customMsgRight = true
				m.rightComponent, cmds[4] = m.rightComponent.Update(rightview.PointMsg{Point: mathcurve.Point{
					X: m.p2.X,
					Y: m.p2.Y,
				}, IsP1SameAsP2: m.checkInput.Check, Index: 1})
				// cmds[5] = tea.Printf("%d.{%d,%d}", 2, m.p2.X, m.p2.Y)
			default:
				m.nextInput()
			}
		}
		m.FocusActualInput()
	}

	for inputs := range 4 {
		switch inputs {
		case eqView:
			if m.eqComp.Err != nil {
				err = true
				m.globalError = m.eqComp.Err
			}
		case p1View:
			if m.p1.Err != nil {
				err = true
				m.globalError = m.p1.Err
			}
		}
	}

	if !err {
		m.globalError = nil
	}

	m.eqComp, cmds[0] = m.eqComp.Update(msg)
	m.p1, cmds[1] = m.p1.Update(msg)
	m.checkInput, cmds[2] = m.checkInput.Update(msg)
	m.p2, cmds[3] = m.p2.Update(msg)
	if !customMsgRight {
		m.rightComponent, cmds[4] = m.rightComponent.Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m SumPointsModel) View() string {
	if m.exit {
		return ""
	}

	var proceedText string = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("[ Calcular suma de puntos ]")

	if !(m.focusedInput == proceedView) {
		proceedText = "[ Calcular suma de puntos ]"
	}

	if m.proceed {
		proceedText += fmt.Sprintf("\n  Punto R: {%d, %d}", m.finalPoint.X, m.finalPoint.Y)
	}

	var err string = noErrorText
	if m.globalError != nil {
		err = "Error: " + errorStyle.Render(m.globalError.Error())
	}

	fullWindow := lipgloss.JoinHorizontal(lipgloss.Left,
		leftWindowStyle.Render(fmt.Sprintf(
			`
  %s

  %s

  Punto 1: %s
    %s

  Punto 2: %s

  %s

  %s
		`,
			titleStyle.Render("Practica 2 - Suma de puntos"),
			m.eqComp.View(),
			m.p1.View(),
			m.checkInput.View(),
			m.p2.View(),
			proceedText,
			err,
		)),
		rightWindowStyle.Render(m.rightComponent.View()),
	)

	if m.displayHelp {
		fullWindow += "\n\n  " + helpText + "\n  " + fullHelpText + "\n"
	} else {
		fullWindow += "\n\n  " + helpText + "\n\n"
	}

	return fullWindow
}

func InitialModel() SumPointsModel {
	return SumPointsModel{
		eqComp:         equationcomponent.NewEqModel(),
		p1:             pointcomponent.NewPointModel(),
		p2:             pointcomponent.NewPointModel(),
		rightComponent: rightview.InitialModel(),
		checkInput:     minicheckcomponent.NewCheckModel("P1 es igual a P2?"),
		focusedInput:   eqView,
	}
}

func StartProgramEquation() {
	p := tea.NewProgram(InitialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (m *SumPointsModel) nextInput() {
	m.focusedInput = (m.focusedInput + 1) % 5
}

// prevInput focuses the previous input field
func (m *SumPointsModel) prevInput() {
	m.focusedInput--
	// Wrap around
	if m.focusedInput < 0 {
		m.focusedInput = 4
	}
}

func (m *SumPointsModel) FocusActualInput() {
	m.eqComp.Blur()
	m.p1.Blur()
	m.checkInput.Blur()
	m.p2.Blur()
	switch m.focusedInput {
	case eqView:
		m.eqComp.Focus()
	case p1View:
		m.p1.Focus()
	case checkView:
		m.checkInput.Focus()
	case p2View:
		m.p2.Focus()
	}
}
