package ankiconnect

import (
	"bytes"
	"encoding/json"
	// "fmt"
	"io"
	"net/http"
	// "os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	// choices  []string         // items on the to-do list
	// cursor   int              // which to-do list item our cursor is pointing at
	// selected map[int]struct{} // which to-do items are selected

	respText string
	err      error
}

type AnkiConnectAction struct {
	Action  string                 `json:"action"`
	Version int                    `json:"version"`
	Params  map[string]interface{} `json:"params"`
}

const ankiConnectUrl = "http://localhost:8765"

func Invoke(action string, params map[string]interface{}) tea.Cmd {
	return func() tea.Msg {

		jsonLoad, err := json.Marshal(AnkiConnectAction{Action: action, Params: params, Version: 6})

		// invalid JSON
		if err != nil {
			return errMsg{err}
		}

		c := &http.Client{Timeout: 10 * time.Second}
		res, err := c.Post(ankiConnectUrl, "application/json", bytes.NewBuffer(jsonLoad))

		// network error
		if err != nil {
			return errMsg{err}
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)

		// unable to read body
		if err != nil {
			return errMsg{err}
		}

		// var resp1 map[string]interface{}
		// err = json.Unmarshal(body, &resp1)
		//
		// // unable to parse JSON
		// if err != nil {
		// 	return errMsg{err}
		// }
		//
		// if _, ok := resp1["error"]; !ok {
		// 	return errMsg{fmt.Errorf("response is missing required error field")}
		// }
		//
		// if _, ok := resp1["result"]; !ok {
		// 	return errMsg{fmt.Errorf("response is missing required result field")}
		// }
		//
		// if resp1["error"] != nil {
		// 	return errMsg{fmt.Errorf("error from AnkiConnect: %s", resp1["error"].(string))}
		// }

		switch action {
		case "createDeck", "addNote":
			var resp map[string]int
			err = json.Unmarshal(body, &resp)

			if err != nil {
				return errMsg{err}
			}

			// println(resp["result"])

		case "deckNames":
			var resp map[string][]string
			err = json.Unmarshal(body, &resp)

			if err != nil {
				return errMsg{err}
			}
			// println(resp["result"][0])
		}

		return statusMsg("ok")
	}
}

type (
	Note struct {
		DeckName  string   `json:"deckName"`
		ModelName string   `json:"modelName"`
		Fields    Fields_  `json:"fields"`
		Options   Opts     `json:"options"`
		Tags      []string `json:"tags"`
		Audio     []Media  `json:"audio"`
		Video     []Media  `json:"video"`
		Picture   []Media  `json:"picture"`
	}

	Media struct {
		Url      string   `json:"url"`
		Filename string   `json:"filename"`
		SkipHash string   `json:"skipHash"`
		Fields   []string `json:"fields"`
	}

	DuplicateScopeOptions_ struct {
		DeckName       string `json:"deckName"`
		CheckChildren  bool   `json:"checkChildren"`
		CheckAllModels bool   `json:"checkAllModels"`
	}

	Opts struct {
		AllowDuplicate        bool                   `json:"allowDuplicate"`
		DuplicateScope        string                 `json:"duplicateScope"`
		DuplicateScopeOptions DuplicateScopeOptions_ `json:"duplicateScopeOptions"`
	}

	Fields_ struct {
		Front string `json:"Front"`
		Back  string `json:"Back"`
	}
)

var note1 = Note{
	DeckName:  "test1",
	ModelName: "Basic",
	Fields:    Fields_{Front: "front content", Back: "back content"},
	Options: Opts{
		AllowDuplicate: false,
		DuplicateScope: "deck",
		DuplicateScopeOptions: DuplicateScopeOptions_{
			DeckName:       "Default",
			CheckChildren:  false,
			CheckAllModels: false,
		},
	},
	Tags: []string{"yomichan"},
	Audio: []Media{{
		Url:      "https://assets.languagepod101.com/dictionary/japanese/audiomp3.php?kanji=猫&kana=ねこ",
		Filename: "yomichan_ねこ_猫.mp3",
		SkipHash: "7e2c2f954ef6051373ba916f000168dc",
		Fields:   []string{"Front"},
	},
	},

	Video: []Media{
		{
			Url:      "https://cdn.videvo.net/videvo_files/video/free/2015-06/small_watermarked/Contador_Glam_preview.mp4",
			Filename: "countdown.mp4",
			SkipHash: "4117e8aab0d37534d9c8eac362388bbe",
			Fields:   []string{"Back"},
		},
	},

	Picture: []Media{
		{
			Url:      "https://upload.wikimedia.org/wikipedia/commons/thumb/c/c7/A_black_cat_named_Tilly.jpg/220px-A_black_cat_named_Tilly.jpg",
			Filename: "black_cat.jpg",
			SkipHash: "8d6e4646dfae812bf39651b59d7429ce",
			Fields:   []string{"Back"},
		},
	},
}

var note2 = Note{
	DeckName:  "test1",
	ModelName: "Basic",
	Fields:    Fields_{Front: "front content", Back: "back content"},
}

type statusMsg string

type errMsg struct{ err error }

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }

// func initialModel() model {
// 	return model{
// 		// Our to-do list is a grocery list
// 		choices: []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},
//
// 		// A map which indicates which choices are selected. We're using
// 		// the  map like a mathematical set. The keys refer to the indexes
// 		// of the `choices` slice, above.
// 		selected: make(map[int]struct{}),
// 	}
// }

// func (m model) Init() tea.Cmd {
// 	// Just return `nil`, which means "no I/O right now, please."
// 	return invoke("addNote", map[string]interface{}{"note": note1})
// 	// return invoke("createDeck", map[string]string{"deck": "test1"})
// 	// return invoke("deckNames", map[string]string{})
// }

// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {
//
// 	case statusMsg:
// 		// The server returned a status message. Save it to our model. Also
// 		// tell the Bubble Tea runtime we want to exit because we have nothing
// 		// else to do. We'll still be able to render a final view with our
// 		// status message.
// 		m.respText = string(msg)
// 		return m, tea.Quit
//
// 	case errMsg:
// 		// There was an error. Note it in the model. And tell the runtime
// 		// we're done and want to quit.
// 		m.err = msg
// 		return m, tea.Quit
//
// 	// Is it a key press?
// 	case tea.KeyMsg:
//
// 		// Cool, what was the actual key pressed?
// 		switch msg.String() {
//
// 		// These keys should exit the program.
// 		case "ctrl+c", "q":
// 			return m, tea.Quit
//
// 		// The "up" and "k" keys move the cursor up
// 		case "up", "k":
// 			if m.cursor > 0 {
// 				m.cursor--
// 			}
//
// 		// The "down" and "j" keys move the cursor down
// 		case "down", "j":
// 			if m.cursor < len(m.choices)-1 {
// 				m.cursor++
// 			}
//
// 		// The "enter" key and the spacebar (a literal space) toggle
// 		// the selected state for the item that the cursor is pointing at.
// 		case "enter", " ":
// 			_, ok := m.selected[m.cursor]
// 			if ok {
// 				delete(m.selected, m.cursor)
// 			} else {
// 				m.selected[m.cursor] = struct{}{}
// 			}
// 		}
// 	}
//
// 	// Return the updated model to the Bubble Tea runtime for processing.
// 	// Note that we're not returning a command.
// 	return m, nil
// }
//
// func (m model) View() string {
// 	// The header
// 	s := "What should we buy at the market?\n\n"
//
// 	// https://biletinial.com/tr-tr/muzik/star-wars-a-new-hope-in-concert#Iterate over our choices
// 	for i, choice := range m.choices {
//
// 		// Is the cursor pointing at this choice?
// 		cursor := " " // no cursor
// 		if m.cursor == i {
// 			cursor = ">" // cursor!
// 		}
//
// 		// Is this choice selected?
// 		checked := " " // not selected
// 		if _, ok := m.selected[i]; ok {
// 			checked = "x" // selected!
// 		}
//
// 		// Render the row
// 		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
// 	}
//
// 	// The footer
// 	s += "\nPress q to quit.\n"
//
// 	// If there's an error, print it out and don't do anything else.
// 	if m.err != nil {
// 		return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
// 	}
//
// 	// When the server responds with a status, add it to the current line.
// 	if m.respText != "" {
// 		s += fmt.Sprintf("%s!", m.respText)
// 	}
//
// 	// Send off whatever we came up with above for rendering.
// 	return "\n" + s + "\n\n"
// }
//
// func main() {
// 	p := tea.NewProgram(initialModel())
// 	if _, err := p.Run(); err != nil {
// 		fmt.Printf("Alas, there's been an error: %v", err)
// 		os.Exit(1)
// 	}
// }
