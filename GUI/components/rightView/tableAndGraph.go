package rightview

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
	exit            bool
	Width, Height   int
	pointsTableComp table.Model
	// for graphView
	graphIsReady   bool
	P              int
	selectedPoints [2]mathcurve.Point
	finalPoint     mathcurve.Point
	graphComp      linechart.Model
}

type PointMsg struct {
	Point        mathcurve.Point
	IsP1SameAsP2 bool
	Index        int
}

type CustomSizeMsg struct {
	Width, Height int
}

type NewValuesTableMsg struct {
	Points []mathcurve.Point
}

//func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, 1)

	switch msg := msg.(type) {
	case PointMsg:
		switch msg.Index {
		case 0:
			m.selectedPoints[0] = msg.Point
			if msg.IsP1SameAsP2 {
				m.selectedPoints[1] = msg.Point
			}
		case 1:
			if !msg.IsP1SameAsP2 {
				m.selectedPoints[1] = msg.Point
			}
		case 2:
			m.finalPoint = msg.Point
		}
		if m.graphIsReady {
			m.graphComp.Resize(m.Width-4, m.Height)
			m.graphComp.Clear()
			m.graphComp.DrawXYAxisAndLabel()
			for _, p := range m.selectedPoints {
				if p.X == -1 {
					continue
				}
				pointToDraw := canvas.Float64Point{X: float64(p.X), Y: float64(p.Y)}
				m.graphComp.DrawRune(pointToDraw, 'X')
			}

			if m.finalPoint.X != -1 {
				pointToDraw := canvas.Float64Point{X: float64(m.finalPoint.X), Y: float64(m.finalPoint.Y)}
				m.graphComp.DrawRune(pointToDraw, 'O')

			}
		}
	case NewValuesTableMsg:
		rows := make([]table.Row, len(msg.Points))
		for i, pt := range msg.Points {
			rows[i] = table.Row{
				fmt.Sprintf("%d", i),      // Index
				fmt.Sprintf("{ %d", pt.X), // X
				",  ",                     // X
				fmt.Sprintf("%d }", pt.Y), // Y
			}
		}
		m.pointsTableComp.SetRows(rows)
		//at this point surely m.P is defined, so lets create a new Graph
		//CREATE AND REDRAW ENTIRE GRAPH WITHOUT POINTS
		{
			m.graphIsReady = true
			m.graphComp = linechart.New(
				0, 0,
				0, float64(m.P),
				0, float64(m.P),
				linechart.WithXYSteps(1, 1),
				linechart.WithStyles(axisStyle, labelStyle, graphStyle),
			)
			m.graphComp.Resize(m.Width-4, m.Height)
			m.graphComp.Clear()
			m.graphComp.DrawXYAxisAndLabel()
		}
	case CustomSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height / 2
		//REDRAW GRAPH
		{
			if m.graphIsReady {
				m.graphComp.Resize(m.Width-4, m.Height)
				m.graphComp.Clear()
				m.graphComp.DrawXYAxisAndLabel()
				for _, p := range m.selectedPoints {
					if p.X == -1 {
						continue
					}
					pointToDraw := canvas.Float64Point{X: float64(p.X), Y: float64(p.Y)}
					m.graphComp.DrawRune(pointToDraw, 'X')
				}

				if m.finalPoint.X != -1 {
					pointToDraw := canvas.Float64Point{X: float64(m.finalPoint.X), Y: float64(m.finalPoint.Y)}
					m.graphComp.DrawRune(pointToDraw, 'O')

				}
			}
		}
	case tea.KeyMsg:
		if !m.Focused {
			return m, nil
		}
		switch msg.Type {
		case tea.KeyCtrlC:
			m.exit = true
			m.Focused = false
			return m, tea.Quit
		}
		switch msg.String() {
		case "e":
			m.Focused = false
			m.pointsTableComp.Blur()
		case "g":
			m.Focused = true
			m.pointsTableComp.Blur()
			m.activeTab = graphView
		case "p":
			m.Focused = true
			m.pointsTableComp.Focus()
			m.activeTab = tableView
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	m.pointsTableComp, cmds[0] = m.pointsTableComp.Update(msg)
	if m.graphIsReady {
		var cmd tea.Cmd
		m.graphComp, cmd = m.graphComp.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func drawTabs(width, activeTab int) string {
	var renderedTabs []string

	for i, t := range Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(Tabs)-1, i == activeTab
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
	row += lipgloss.NewStyle().Foreground(highlightColor).Render(strings.Repeat("─", width+1-lipgloss.Width(row)) + "╮")

	return row
}

func (m Model) View() string {
	if m.exit {
		return ""
	}
	viewport := strings.Builder{}
	viewport.WriteString(drawTabs(m.Width, m.activeTab) + "\n")

	switch m.activeTab {
	case tableView:
		//viewport.WriteString(windowStyle.Width(m.Width).Height(m.Height).Render(baseStyle.Render(strings.Repeat("a", 100))))
		viewport.WriteString(windowStyle.Width(m.Width).Height(m.Height).Render(baseStyle.Render(m.pointsTableComp.View())))
	case graphView:
		if m.graphIsReady {
			viewport.WriteString(windowStyle.Width(m.Width).Height(m.Height).Render(m.graphComp.View()))
		} else {
			viewport.WriteString(windowStyle.Width(m.Width).Height(m.Height).Render("Insert P value to see the graph"))
		}
	}

	return viewport.String()
}

func InitialModel() Model {
	columns := []table.Column{
		{Title: "n", Width: 5},
		{Title: "X", Width: 8},
		{Title: "", Width: 3},
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
		Focused:         false,
		pointsTableComp: t,
		Width:           45,
		Height:          20,
		finalPoint:      mathcurve.Point{X: -1, Y: -1},
		selectedPoints: [2]mathcurve.Point{
			{X: -1, Y: -1},
			{X: -1, Y: -1},
		},
		graphIsReady: false,
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

func (m *Model) Focus(input string) {
	switch input {
	case "g":
		m.graphComp.Focus()
	case "p":
		m.pointsTableComp.Focus()
	}
	m.Focused = true
}

func (m *Model) Blur() {
	switch m.activeTab {
	case graphView:
		m.graphComp.Focus()
	case tableView:
		m.pointsTableComp.Focus()
	}

	m.Focused = false
}

// func TrySeeTable() {

// 	p := tea.NewProgram(InitialModel())

// 	if _, err := p.Run(); err != nil {
// 		log.Fatal(err)
// 	}
// }
