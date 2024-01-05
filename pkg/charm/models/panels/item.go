package panels

import (
	"sync"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/everettraven/buoy/pkg/types"
)

// Item is a tea.Model implementation
// that represents an item panel
type Item struct {
	viewport viewport.Model
	mutex    *sync.Mutex
	item     types.Item
	theme    styles.Theme
	err      error
}

func NewItem(item types.Item, viewport viewport.Model, theme styles.Theme) *Item {
	return &Item{
		viewport: viewport,
		mutex:    &sync.Mutex{},
		item:     item,
		theme:    theme,
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
	if m.err != nil {
		return m.err.Error()
	}
	return m.viewport.View()
}

func (m *Item) SetContent(content string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.viewport.SetContent(content)
}

func (m *Item) Name() string {
	return m.item.Name
}

func (m *Item) ItemDefinition() types.Item {
	return m.item
}

func (m *Item) Theme() styles.Theme {
	return m.theme
}

func (m *Item) SetError(err error) {
	m.err = err
}
