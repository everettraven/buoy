package panels

import (
	"sync"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Item is a tea.Model implementation
// that represents an item panel
type Item struct {
	viewport viewport.Model
	name     string
	mutex    *sync.Mutex
}

func NewItem(name string, viewport viewport.Model) *Item {
	return &Item{
		viewport: viewport,
		name:     name,
		mutex:    &sync.Mutex{},
	}
}

func (m *Item) Init() tea.Cmd { return nil }

func (m *Item) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height / 2
	}
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Item) View() string {
	return m.viewport.View()
}

func (m *Item) SetContent(content string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.viewport.SetContent(content)
}

func (m *Item) Name() string {
	return m.name
}
