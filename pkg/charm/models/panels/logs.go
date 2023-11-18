package panels

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/sahilm/fuzzy"
)

type LogsKeyMap struct {
	Search       key.Binding
	SubmitSearch key.Binding
	QuitSearch   key.Binding
	ToggleStrict key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k LogsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k LogsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Search, k.SubmitSearch, k.QuitSearch, k.ToggleStrict},
	}
}

var DefaultLogsKeys = LogsKeyMap{
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "open a prompt to search logs"),
	),
	SubmitSearch: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit search prompt"),
	),
	QuitSearch: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "exit search mode"),
	),
	ToggleStrict: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "toggle strict search mode"),
	),
}

const modeLogs = "logs"
const modeSearching = "searching"
const modeSearched = "searched"

// Logs is a tea.Model implementation
// that represents an item panel
type Logs struct {
	viewport       viewport.Model
	searchbar      textinput.Model
	name           string
	mutex          *sync.Mutex
	content        string
	contentUpdated bool
	mode           string
	keys           LogsKeyMap
	strictSearch   bool
}

func NewLogs(keys LogsKeyMap, name string) *Logs {
	searchbar := textinput.New()
	searchbar.Prompt = "> "
	searchbar.Placeholder = "search term"
	return &Logs{
		viewport:  viewport.New(10, 10),
		searchbar: searchbar,
		name:      name,
		mutex:     &sync.Mutex{},
		content:   "",
		mode:      modeLogs,
		keys:      keys,
	}
}

func (m *Logs) Init() tea.Cmd { return nil }

func (m *Logs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height / 2
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Search):
			m.mode = modeSearching
			if !m.searchbar.Focused() {
				m.searchbar.Focus()
			}
			m.searchbar.SetValue("")
			return m, nil
		case key.Matches(msg, m.keys.QuitSearch):
			m.mode = modeLogs
			if m.searchbar.Focused() {
				m.searchbar.Blur()
			}
			m.viewport.SetContent(wrapLogs(m.content, m.viewport.Width))
			m.contentUpdated = false
		case key.Matches(msg, m.keys.SubmitSearch):
			if m.mode == modeSearching {
				m.mode = modeSearched
				if m.searchbar.Focused() {
					m.searchbar.Blur()
				}
			}
		case key.Matches(msg, m.keys.ToggleStrict):
			m.strictSearch = !m.strictSearch
		}
	}

	if m.contentUpdated && m.mode == modeLogs {
		m.viewport.SetContent(wrapLogs(m.content, m.viewport.Width))
		m.contentUpdated = false
	}

	if m.mode == modeSearching {
		m.searchbar, cmd = m.searchbar.Update(msg)
		return m, cmd
	}

	if m.mode == modeSearched {
		m.viewport.SetContent(searchLogs(m.content, m.searchbar.Value(), m.viewport.Width, m.strictSearch))
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Logs) View() string {
	searchMode := "fuzzy"
	if m.strictSearch {
		searchMode = "strict"
	}
	searchModeOutput := styles.LogSearchModeStyle().Render(fmt.Sprintf("search mode: %s", searchMode))

	if m.mode == modeSearching {
		return lipgloss.JoinVertical(lipgloss.Top,
			m.searchbar.View(),
			searchModeOutput,
		)
	}

	if m.mode == modeSearched {
		return lipgloss.JoinVertical(lipgloss.Top,
			m.viewport.View(),
			searchModeOutput,
		)
	}

	return m.viewport.View()
}

func (m *Logs) Help() help.KeyMap {
	return m.keys
}

func (m *Logs) AddContent(content string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.content = strings.Join([]string{m.content, content}, "\n")
	m.contentUpdated = true
}

func (m *Logs) Name() string {
	return m.name
}

// searchLogs searches the logs for the given term
// and returns a string with the matching log lines
// and the matched term highlighted. Uses fuzzy search
// if strict is false. Wraps logs to the given width if wrap > 0.
func searchLogs(logs, term string, wrap int, strict bool) string {
	splitLogs := strings.Split(logs, "\n")
	if strict {
		return strictMatchLogs(term, splitLogs, wrap)
	}
	return fuzzyMatchLogs(term, splitLogs, wrap)
}

func strictMatchLogs(searchTerm string, logLines []string, wrap int) string {
	var results strings.Builder
	for _, log := range logLines {
		if wrap > 0 {
			log = wrapLogs(log, wrap)
		}
		if strings.Contains(log, searchTerm) {
			highlighted := strings.Replace(
				log,
				searchTerm,
				styles.LogSearchHighlightStyle().Render(searchTerm), -1,
			)
			results.WriteString(highlighted + "\n")
		}
	}
	return results.String()
}

func fuzzyMatchLogs(searchTerm string, logLines []string, wrap int) string {
	var matches []fuzzy.Match
	if wrap > 0 {
		wrappedLogs := []string{}
		for _, log := range logLines {
			wrappedLogs = append(wrappedLogs, wrapLogs(log, wrap))
		}
		matches = fuzzy.Find(searchTerm, wrappedLogs)
	} else {
		matches = fuzzy.Find(searchTerm, logLines)
	}

	var results strings.Builder
	for _, match := range matches {
		for i := 0; i < len(match.Str); i++ {
			if matched(i, match.MatchedIndexes) {
				results.WriteString(styles.LogSearchHighlightStyle().Render(string(match.Str[i])))
			} else {
				results.WriteString(string(match.Str[i]))
			}
		}
		results.WriteString("\n")
	}

	return results.String()
}

func matched(index int, matches []int) bool {
	for _, i := range matches {
		if index == i {
			return true
		}
	}
	return false
}

func wrapLogs(logs string, maxWidth int) string {
	splitLogs := strings.Split(logs, "\n")
	var logsBuilder strings.Builder
	for _, log := range splitLogs {
		if len(log) > maxWidth {
			segs := (len(log) / maxWidth)
			for seg := 0; seg < segs; seg++ {
				logsBuilder.WriteString(log[:maxWidth])
				logsBuilder.WriteString("\n")
				log = log[maxWidth:]
			}
			//write any leftovers
			logsBuilder.WriteString(log)
		} else {
			logsBuilder.WriteString(log)
		}
		logsBuilder.WriteString("\n")
	}
	return logsBuilder.String()
}
