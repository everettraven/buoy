package panels

import (
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/everettraven/buoy/pkg/charm/styles"
)

// Logs is a tea.Model implementation
// that represents an item panel
type Logs struct {
	viewport viewport.Model
	name     string
	style    lipgloss.Style
	mutex    *sync.Mutex
	content  string
}

func NewLogs(name string, viewport viewport.Model) *Logs {
	return &Logs{
		viewport: viewport,
		name:     name,
		style:    styles.ModelStyle,
		mutex:    &sync.Mutex{},
		content:  "",
	}
}

func (m *Logs) Init() tea.Cmd { return nil }

func (m *Logs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height / 2
	}
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Logs) View() string {
	return m.style.Render(m.viewport.View())
}

func (m *Logs) SetStyle(style lipgloss.Style) {
	m.style = style
}

func (m *Logs) AddContent(content string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.content = strings.Join([]string{m.content, content}, "\n")
	m.viewport.SetContent(m.content)
}

func (m *Logs) Name() string {
	return m.name
}