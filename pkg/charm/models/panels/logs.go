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
	"github.com/muesli/reflow/wrap"
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
	theme          *styles.Theme
}

func NewLogs(keys LogsKeyMap, name string, theme *styles.Theme) *Logs {
	searchbar := textinput.New()
	searchbar.Prompt = "> "
	searchbar.Placeholder = "search term"
	vp := viewport.New(10, 10)
	return &Logs{
		viewport:  vp,
		searchbar: searchbar,
		name:      name,
		mutex:     &sync.Mutex{},
		content:   "",
		mode:      modeLogs,
		keys:      keys,
		theme:     theme,
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
		m.viewport.SetContent(m.searchLogs())
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Logs) View() string {
	searchMode := "fuzzy"
	if m.strictSearch {
		searchMode = "strict"
	}
	searchModeOutput := m.theme.LogSearchModeStyle().Render(fmt.Sprintf("search mode: %s", searchMode))

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

// searchLogs searches the logs for the term in the searchbar
// and returns a string with the matching log lines
// and the matched term highlighted. Uses fuzzy search
// if strict search is not enabled. Wraps logs to the width of the viewport.
func (m *Logs) searchLogs() string {
	term := m.searchbar.Value()
	wrap := m.viewport.Width
	strict := m.strictSearch
	splitLogs := strings.Split(m.content, "\n")
	if strict {
		return strictMatchLogs(term, splitLogs, m.viewport.Width, m.theme.LogSearchHighlightStyle())
	}
	return fuzzyMatchLogs(term, splitLogs, wrap, m.theme.LogSearchHighlightStyle())
}

func strictMatchLogs(searchTerm string, logLines []string, wrap int, style lipgloss.Style) string {
	var results strings.Builder
	for _, log := range logLines {
		if wrap > 0 {
			log = wrapLogs(log, wrap)
		}
		if strings.Contains(log, searchTerm) {
			highlighted := strings.Replace(
				log,
				searchTerm,
				style.Render(searchTerm), -1,
			)
			results.WriteString(highlighted + "\n")
		}
	}
	return results.String()
}

func fuzzyMatchLogs(searchTerm string, logLines []string, wrap int, style lipgloss.Style) string {
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
				results.WriteString(style.Render(string(match.Str[i])))
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
	return wrap.String(logs, maxWidth)
}
