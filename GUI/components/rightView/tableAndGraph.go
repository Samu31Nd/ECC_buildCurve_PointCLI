package rightview

// TODO: QUITENLE LA FUNCION INIT Y USENLO COMO COMPONENTE CON SUS DEFAULTS COMO LOS DEMAS COMPONENTES

import (
	"fmt"
	mathcurve "puntosCurvaEliptica/MathCurve"
	"strings"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/linechart"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	tableView = iota
	graphView
)

var (
	Tabs = [2]string{
		"Points",
		"Graph View",
	}
)

var (
	baseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
)

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()

	graphStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	axisStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)

type Model struct {
	activeTab       int
	Focused         bool
	Width, Height   int
	pointsTableComp table.Model
	// for graphView
	notGraph       bool
	P              int
	selectedPoints [2]mathcurve.Point
	finalPoint     mathcurve.Point
	graphComp      linechart.Model
}

type PointMsg struct {
	point mathcurve.Point
	index int
}

type CustomSizeMsg struct {
	Width, Height int
}

type NewValuesTableMsg struct {
	points []mathcurve.Point
}

//func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, 2)

	reDrawCanvas := false

	switch msg := msg.(type) {
	case PointMsg:
		reDrawCanvas = true
		m.finalPoint = mathcurve.Point{X: -1}
		//Update
		if msg.index == 2 {
			m.finalPoint = msg.point
		} else {
			m.selectedPoints[msg.index] = msg.point
		}
	case NewValuesTableMsg:
		rows := make([]table.Row, len(msg.points))
		for i, pt := range msg.points {
			rows[i] = table.Row{
				fmt.Sprintf("%d", i),    // Index
				fmt.Sprintf("%d", pt.X), // X
				fmt.Sprintf("%d", pt.Y), // Y
			}
		}
		m.pointsTableComp.SetRows(rows)
	case tea.WindowSizeMsg:
		m.Width = msg.Width / 2
		m.graphComp.Resize(msg.Width-4, msg.Height-10)
		m.Height = 2 * msg.Height / 3
	case CustomSizeMsg:
		m.graphComp.Resize(msg.Width-10, msg.Height-10)
		m.Width = msg.Width
		m.Height = msg.Height
	case tea.KeyMsg:
		if !m.Focused {
			break
		} // ignore input
		switch msg.String() {
		case "e":
			m.Focused = false
		case "g":
			m.Focused = true
			reDrawCanvas = true
			m.pointsTableComp.Blur()
			m.activeTab = graphView

		case "p":
			m.Focused = true
			m.pointsTableComp.Focus()
			m.activeTab = tableView
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			// handle view in graph
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.pointsTableComp.SelectedRow()[1]),
			)
		}
	}

	if reDrawCanvas {
		m.notGraph = false
		m.graphComp = linechart.New(
			m.Width-6, m.Height-6,
			0, float64(m.P-1),
			0, float64(m.P-1),
			linechart.WithXYSteps(1, 1),
			linechart.WithStyles(axisStyle, labelStyle, graphStyle),
		)
		m.graphComp.DrawXYAxisAndLabel()

		//dibujar los 2 puntos
		for _, point := range m.selectedPoints {
			ptCanva := canvas.Float64Point{
				X: float64(point.X),
				Y: float64(point.Y),
			}
			m.graphComp.DrawRune(ptCanva, '●')
		}
		if m.finalPoint.X != -1 {
			ptCanva := canvas.Float64Point{
				X: float64(m.finalPoint.X),
				Y: float64(m.finalPoint.Y),
			}
			m.graphComp.DrawRune(ptCanva, '◎')
		}
	}

	m.pointsTableComp, cmds[0] = m.pointsTableComp.Update(msg)
	m.graphComp, cmds[1] = m.graphComp.Update(msg)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	viewport := strings.Builder{}

	var renderedTabs []string

	for i, t := range Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(Tabs)-1, i == m.activeTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "└"
		} else if isLast && !isActive {
			border.BottomRight = "┴"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	row += lipgloss.NewStyle().Foreground(highlightColor).Render(strings.Repeat("─", m.Width+1-lipgloss.Width(row)) + "╮")
	viewport.WriteString(row)
	viewport.WriteString("\n")

	switch m.activeTab {
	case tableView:
		viewport.WriteString(windowStyle.Width(m.Width).Height(m.Height).Render(baseStyle.Render(m.pointsTableComp.View())))
	case graphView:
		if m.notGraph {
			viewport.WriteString(windowStyle.Width(m.Width).Height(m.Height).Render("Add some points dude"))
		} else {
			viewport.WriteString(windowStyle.Width(m.Width).Height(m.Height).Render(m.graphComp.View()))
		}
	}

	return viewport.String()
}

func InitialModel() Model {
	columns := []table.Column{
		{Title: "n", Width: 5},
		{Title: "X", Width: 8},
		{Title: "Y", Width: 8},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(8),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(s)

	return Model{
		activeTab:       tableView,
		Focused:         false, //TODO: Turn this to false
		pointsTableComp: t,
		Width:           45,
		Height:          30,
	}
}

func TestInitialModel(allPoints []mathcurve.Point, selectedPoints [2]mathcurve.Point, finalPoint mathcurve.Point, p int) Model {
	columns := []table.Column{
		{Title: "n", Width: 5},
		{Title: "X", Width: 8},
		{Title: "Y", Width: 8},
	}

	rows := make([]table.Row, len(allPoints))
	for i, pt := range allPoints {
		rows[i] = table.Row{
			fmt.Sprintf("%d", i),    // Index
			fmt.Sprintf("%d", pt.X), // X
			fmt.Sprintf("%d", pt.Y), // Y
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(8),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(s)

	return Model{
		activeTab:       tableView,
		Focused:         true, //TODO: Turn this to false
		pointsTableComp: t,
		Width:           45,
		Height:          30,
		selectedPoints:  selectedPoints,
		finalPoint:      finalPoint,
		P:               p,
	}
}

// func TrySeeTable() {
// 	points := []mathcurve.Point{
// 		{X: 0, Y: 4},
// 		{X: 1, Y: 4},
// 		{X: 2, Y: 4},
// 	}

// 	pointsGraph := [2]mathcurve.Point{
// 		{X: 0, Y: 4},
// 		{X: 1, Y: 4},
// 	}

// 	finalPoint := mathcurve.Point{
// 		X: 5, Y: 5,
// 	}
// 	p := tea.NewProgram(TestInitialModel(points, pointsGraph, finalPoint, 11))

// 	if _, err := p.Run(); err != nil {
// 		log.Fatal(err)
// 	}
// }
