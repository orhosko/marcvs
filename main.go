package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"test.bubble.tea/ankiconnect"

	"github.com/joho/godotenv"
)

/*
This example assumes an existing understanding of commands and messages. If you
haven't already read our tutorials on the basics of Bubble Tea and working
with commands, we recommend reading those first.

Find them at:
https://github.com/charmbracelet/bubbletea/tree/master/tutorials/commands
https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics
*/

// sessionState is used to track which model is focused
type sessionState uint

const (
	defaultTime              = time.Minute
	timerView   sessionState = iota
	spinnerView
	listView
	inputView
)

var (
	// Available spinners
	spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}

	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)

	listModelStyle = lipgloss.NewStyle().
			Width(60).
			Height(20).
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.NormalBorder())
	focusedListModelStyle = lipgloss.NewStyle().
				Width(60).
				Height(20).
				Align(lipgloss.Center, lipgloss.Center).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("69"))

	modelStyle = lipgloss.NewStyle().
			Width(30).
			Height(5).
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.NormalBorder())
	focusedModelStyle = lipgloss.NewStyle().
				Width(30).
				Height(5).
				Align(lipgloss.Center, lipgloss.Center).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("69"))
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	inputModelStyle = lipgloss.NewStyle().
			Width(60).
			Height(3).
			Align(lipgloss.Left, lipgloss.Center).
			BorderStyle(lipgloss.NormalBorder()).
			PaddingLeft(2)
	focusedInputModelStyle = lipgloss.NewStyle().
				Width(60).
				Height(3).
				Align(lipgloss.Left, lipgloss.Center).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("69")).
				PaddingLeft(2)
)

type mainModel struct {
	state     sessionState
	timer     timer.Model
	spinner   spinner.Model
	list      list.Model
	textInput textinput.Model
	index     int
}

func newModel(timeout time.Duration) mainModel {
	m := mainModel{state: timerView}
	m.timer = timer.New(timeout)
	m.spinner = spinner.New()
	m.list = list.New(items, itemDelegate{}, 20, 14)
	m.list.Title = "List of words"
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowHelp(false)
	m.textInput = textinput.New()
	m.textInput.Placeholder = "Enter a word..."
	m.textInput.Focus()
	m.textInput.CharLimit = 156
	m.textInput.Width = 50

	return m
}

func (m mainModel) Init() tea.Cmd {
	godotenv.Load()
	// start the timer and spinner on program start
	return tea.Batch(m.timer.Init(), m.spinner.Tick, textinput.Blink)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c": // "q" un da çalışmasını sağlayabiliriz
			return m, tea.Quit
		case "tab":
			m.state = (m.state + 1) % 4
			if m.state == 0 {
				m.state = 4
			}
		case "n":
			if m.state == timerView {
				m.timer = timer.New(defaultTime)
				cmds = append(cmds, m.timer.Init())
			} else {
				m.Next()
				m.resetSpinner()
				cmds = append(cmds, m.spinner.Tick)
			}
		}
		switch m.state {
		// update whichever model is focused
		case spinnerView:
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		case listView:
			m.list, cmd = m.list.Update(msg)
		case inputView:
			if msg.String() == "enter" {

				s := m.textInput.Value()
				s = strings.Trim(s, " ")
				s = strings.TrimLeft(s, " ")

				if len(strings.TrimSpace(s)) != 0 {
					items = append([]list.Item{item{s, ""}}, items...)
					m.list.SetItems(items)
					m.textInput.Reset()
				}

				c := &http.Client{}

				key := os.Getenv("collinsKey")

				req, _ := http.NewRequest("GET", "https://api.collinsdictionary.com/api/v1/dictionaries/english/search/first/?q="+s, nil)

				req.Header.Set("Accept", "application/json")
				req.Header.Set("accessKey", key)

				res, err := c.Do(req)

				if err != nil {
					fmt.Println(err)
				}

				defer res.Body.Close()

				body, err := io.ReadAll(res.Body)

				items = append([]list.Item{item{string(body), ""}}, items...)

				note := ankiconnect.Note{
					DeckName:  "test1",
					ModelName: "Basic",
					Fields: ankiconnect.Fields_{
						Front: s,
						Back:  "back",
					},
				}

				newcmd := ankiconnect.Invoke("addNote", map[string]interface{}{"note": note})

				cmds = append(cmds, newcmd)

			} else {
				m.textInput, cmd = m.textInput.Update(msg)
			}
		default:
			m.timer, cmd = m.timer.Update(msg)
			cmds = append(cmds, cmd)
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case timer.TickMsg:
		m.timer, cmd = m.timer.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// type item struct {
// 	title, desc string
// }

type item struct {
	word string
	etm  string
}

// func (i item) Title() string       { return i.title }
// func (i item) Description() string { return i.desc }
// func (i item) FilterValue() string { return i.title }
func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", len(items)-index, i.word)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

//	var items = []list.Item{
//		item{title: "Raspberry Pi’s", desc: "I have ’em all over my house"},
//		item{title: "Nutella", desc: "It's good on toast"},
//		item{title: "Bitter melon", desc: "It cools you down"},
//		item{title: "Nice socks", desc: "And by that I mean socks without holes"},
//		item{title: "Eight hours of sleep", desc: "I had this once"},
//		item{title: "Cats", desc: "Usually"},
//		item{title: "Plantasia, the album", desc: "My plants love it too"},
//		item{title: "Pour over coffee", desc: "It takes forever to make though"},
//		item{title: "VR", desc: "Virtual reality...what is there to say?"},
//		item{title: "Noguchi Lamps", desc: "Such pleasing organic forms"},
//		item{title: "Linux", desc: "Pretty much the best OS"},
//		item{title: "Business school", desc: "Just kidding"},
//		item{title: "Pottery", desc: "Wet clay is a great feeling"},
//		item{title: "Shampoo", desc: "Nothing like clean hair"},
//		item{title: "Table tennis", desc: "It’s surprisingly exhausting"},
//		item{title: "Milk crates", desc: "Great for packing in your extra stuff"},
//		item{title: "Afternoon tea", desc: "Especially the tea sandwich part"},
//		item{title: "Stickers", desc: "The thicker the vinyl the better"},
//		item{title: "20° Weather", desc: "Celsius, not Fahrenheit"},
//		item{title: "Warm light", desc: "Like around 2700 Kelvin"},
//		item{title: "The vernal equinox", desc: "The autumnal equinox is pretty good too"},
//		item{title: "Gaffer’s tape", desc: "Basically sticky fabric"},
//		item{title: "Terrycloth", desc: "In other words, towel fabric"},
//	}
var items = []list.Item{
	item{"Ramen", ""},
	item{"Tomato Soup", ""},
	item{"Hamburgers", ""},
	// item("Cheeseburgers"),
	// item("Currywurst"),
	// item("Okonomiyaki"),
	// item("Pasta"),
	// item("Fillet Mignon"),
	// item("Caviar"),
	// item("Just Wine"),
}

func (m mainModel) View() string {
	var s string

	// s += fmt.Sprintf("%v", m.state) + "\n"

	model := m.currentFocusedModel()
	if m.state == timerView {

		s += inputModelStyle.Render(m.textInput.View())
		s += "\n"

		s += listModelStyle.Render(m.list.View())
		s += "\n"
		s += lipgloss.JoinHorizontal(lipgloss.Top, focusedModelStyle.Render(fmt.Sprintf("%4s", m.timer.View())), modelStyle.Render(m.spinner.View()))
	}
	if m.state == spinnerView {

		s += inputModelStyle.Render(m.textInput.View())
		s += "\n"

		s += listModelStyle.Render(m.list.View())
		s += "\n"
		s += lipgloss.JoinHorizontal(lipgloss.Top, modelStyle.Render(fmt.Sprintf("%4s", m.timer.View())), focusedModelStyle.Render(m.spinner.View()))
	}
	if m.state == listView {

		s += inputModelStyle.Render(m.textInput.View())
		s += "\n"

		s += focusedListModelStyle.Render(m.list.View())
		s += "\n"
		s += lipgloss.JoinHorizontal(lipgloss.Top, modelStyle.Render(fmt.Sprintf("%4s", m.timer.View())), modelStyle.Render(m.spinner.View()))
		// s += lipgloss.JoinVertical(lipgloss.Left, lipgloss.JoinHorizontal(lipgloss.Top, modelStyle.Render(fmt.Sprintf("%4s", m.timer.View())), focusedModelStyle.Render(m.spinner.View()), modelStyle.Render(m.list.View())))
	}
	if m.state == inputView {

		s += focusedInputModelStyle.Render(m.textInput.View())
		s += "\n"

		s += listModelStyle.Render(m.list.View())
		s += "\n"
		s += lipgloss.JoinHorizontal(lipgloss.Top, modelStyle.Render(fmt.Sprintf("%4s", m.timer.View())), modelStyle.Render(m.spinner.View()))
	}
	s += helpStyle.Render(fmt.Sprintf("\ntab: focus next • n: new %s • q: exit\n", model))
	return s
}

func (m mainModel) currentFocusedModel() string {
	if m.state == timerView {
		return "timer"
	}
	if m.state == spinnerView {
		return "spinner"
	}
	if m.state == listView {
		return "list"
	}
	return "wtf"
}

func (m *mainModel) Next() {
	if m.index == len(spinners)-1 {
		m.index = 0
	} else {
		m.index++
	}
}

func (m *mainModel) resetSpinner() {
	m.spinner = spinner.New()
	m.spinner.Style = spinnerStyle
	m.spinner.Spinner = spinners[m.index]
}

func main() {
	p := tea.NewProgram(newModel(defaultTime), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
