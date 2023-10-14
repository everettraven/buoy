package models

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

var ViewportHeightFunc HeightFunc = func(model interface{}) int {
	vp, ok := model.(viewport.Model)
	if !ok {
		return 0
	}

	return vp.Height + 1
}

var TableHeightFunc HeightFunc = func(model interface{}) int {
	tab, ok := model.(table.Model)
	if !ok {
		return 0
	}

	return tab.Height() + 1
}

var TableUpdateFunc UpdateFunc = func(model interface{}, msg tea.Msg) (interface{}, tea.Cmd) {
	tab, ok := model.(table.Model)
	if !ok {
		return model, tea.Println("model not of type table.Model!")
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if tab.Focused() {
				tab.Blur()
			} else {
				tab.Focus()
			}
		}
	}
	tab, cmd = tab.Update(msg)
	return tab, cmd
}

// ModelWrapper is a helper to wrap model types used by
// buoy into a tea.Model interface implementation
type ModelWrapper struct {
	Model   interface{}
	UpdateF UpdateFunc
	HeightF HeightFunc
	Name    string
	style   lipgloss.Style
}

func (m *ModelWrapper) Init() tea.Cmd {
	if init, ok := m.Model.(Initable); ok {
		return init.Init()
	}
	return nil
}

func (m *ModelWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Model, cmd = m.UpdateF(m.Model, msg)
	return m, cmd
}

func (m *ModelWrapper) View() string {
	if view, ok := m.Model.(Viewable); ok {
		return m.style.Render(view.View())
	}

	return "model not a Viewable"
}

func (m *ModelWrapper) Height() int {
	return m.HeightF(m.Model)
}

func (m *ModelWrapper) SetStyle(style lipgloss.Style) {
	m.style = style
}
