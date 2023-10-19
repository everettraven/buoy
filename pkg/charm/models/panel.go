package models

import (
	"strings"

	tbl "github.com/calyptia/go-bubble-table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/everettraven/buoy/pkg/charm/styles"
)

type Viewable interface {
	View() string
}

type Initable interface {
	Init() tea.Cmd
}

type UpdateFunc func(interface{}, tea.Msg) (interface{}, tea.Cmd)

type HeightFunc func(interface{}) int

var ViewportUpdateFunc UpdateFunc = func(model interface{}, msg tea.Msg) (interface{}, tea.Cmd) {
	vp, ok := model.(viewport.Model)
	if !ok {
		return model, tea.Println("model not of type viewport.Model!")
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		vp.Height = msg.Height
		vp.Width = msg.Width
	}
	vp, cmd = vp.Update(msg)
	return vp, cmd
}

var TableUpdateFunc UpdateFunc = func(model interface{}, msg tea.Msg) (interface{}, tea.Cmd) {
	tab, ok := model.(tbl.Model)
	if !ok {
		return model, tea.Println("model not of type table.Model!")
	}
	var cmd tea.Cmd
	tab, cmd = tab.Update(msg)
	return tab, cmd
}

// Model is a helper to wrap model types used by
// buoy into a tea.Model interface implementation
type Panel struct {
	Model   interface{}
	UpdateF UpdateFunc
	HeightF HeightFunc
	Name    string
	style   lipgloss.Style
}

func (m *Panel) Init() tea.Cmd {
	if init, ok := m.Model.(Initable); ok {
		return init.Init()
	}
	return nil
}

func (m *Panel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Model, cmd = m.UpdateF(m.Model, msg)
	return m, cmd
}

func (m *Panel) View() string {
	if view, ok := m.Model.(Viewable); ok {
		var out strings.Builder
		out.WriteString(styles.TitleStyle.Render(m.Name))
		out.WriteString("\n")
		out.WriteString(m.style.Render(view.View()))
		return out.String()
	}

	return "model not a Viewable"
}

func (m *Panel) SetStyle(style lipgloss.Style) {
	m.style = style
}
