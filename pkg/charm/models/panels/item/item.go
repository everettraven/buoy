package item

import (
	"bytes"
	"io"
	"sync"

	"github.com/alecthomas/chroma/quick"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/everettraven/buoy/pkg/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apimachtypes "k8s.io/apimachinery/pkg/types"
)

type Styles struct {
	SyntaxHighlightDark  string
	SyntaxHighlightLight string
}

// Model is a tea.Model implementation
// that represents an item panel
type Model struct {
	viewport viewport.Model
	mutex    *sync.Mutex
	item     types.Item
	theme    Styles
	err      error
}

func New(item types.Item, viewport viewport.Model, theme Styles) *Model {
	return &Model{
		viewport: viewport,
		mutex:    &sync.Mutex{},
		item:     item,
		theme:    theme,
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height / 2
	}
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	return m.viewport.View()
}

func (m *Model) SetContent(content string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	// by default set the content as the plain string passed in
	m.viewport.SetContent(content)

	// attempt to perform syntax highlighting
	theme := m.theme.SyntaxHighlightDark
	if !lipgloss.HasDarkBackground() {
		theme = m.theme.SyntaxHighlightLight
	}
	rw := &bytes.Buffer{}
	err := quick.Highlight(rw, content, "yaml", "terminal16m", theme)
	if err != nil {
		return
	}
	highlighted, err := io.ReadAll(rw)
	if err != nil {
		return
	}
	m.viewport.SetContent(string(highlighted))
}

func (m *Model) Name() string {
	return m.item.Name
}

func (m *Model) Key() apimachtypes.NamespacedName {
	return m.item.Key
}

func (m *Model) GVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   m.item.Group,
		Version: m.item.Version,
		Kind:    m.item.Kind,
	}
}

func (m *Model) Theme() Styles {
	return m.theme
}

func (m *Model) SetError(err error) {
	m.err = err
}
