package panels

import (
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
)

type LogsKeyMap struct {
	Search       key.Binding
	SubmitSearch key.Binding
	QuitSearch   key.Binding
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
		{k.Search},
	}
}

var DefaultLogsKeys = LogsKeyMap{
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "open a prompt to enter a term to fuzzy search logs"),
	),
	SubmitSearch: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit search prompt"),
	),
	QuitSearch: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "exit search mode"),
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
}

func NewLogs(name string, viewport viewport.Model) *Logs {
	searchbar := textinput.New()
	searchbar.Prompt = "> "
	searchbar.Placeholder = "search term"
	return &Logs{
		viewport:  viewport,
		searchbar: searchbar,
		name:      name,
		mutex:     &sync.Mutex{},
		content:   "",
		mode:      modeLogs,
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
		case key.Matches(msg, DefaultLogsKeys.Search):
			m.mode = modeSearching
			if !m.searchbar.Focused() {
				m.searchbar.Focus()
			}
			m.searchbar.SetValue("")
			return m, nil
		case key.Matches(msg, DefaultLogsKeys.QuitSearch):
			m.mode = modeLogs
			if m.searchbar.Focused() {
				m.searchbar.Blur()
			}
			m.viewport.SetContent(wrapLogs(m.content, m.viewport.Width))
			m.contentUpdated = false
		case key.Matches(msg, DefaultLogsKeys.SubmitSearch):
			if m.mode == modeSearching {
				m.mode = modeSearched
				if m.searchbar.Focused() {
					m.searchbar.Blur()
				}
			}
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
	if m.mode == modeSearching {
		return m.searchbar.View()
	}
	return m.viewport.View()
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

func (m *Logs) searchLogs() string {
	searchTerm := m.searchbar.Value()
	splitLogs := strings.Split(m.content, "\n")
	matches := fuzzy.Find(searchTerm, splitLogs)
	matchedLogs := []string{}
	for _, match := range matches {
		matchedLog := splitLogs[match.Index]
		// TODO: highlight matched term
		matchedLogs = append(matchedLogs, matchedLog)
	}
	return wrapLogs(strings.Join(matchedLogs, "\n"), m.viewport.Width)
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
