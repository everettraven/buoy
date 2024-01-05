package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/stretchr/testify/assert"
)

func TestTabberUpdate(t *testing.T) {
	tabber := NewTabber(DefaultTabberKeys, styles.Theme{}, Tab{Name: "test", Model: nil}, Tab{Name: "test2", Model: nil})

	t.Log("navigate to next tab")
	tabber.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 1, tabber.selected)

	t.Log("navigate to previous tab")
	tabber.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("shift+tab")})
	assert.Equal(t, 0, tabber.selected)

	t.Log("navigate to previous tab (out of bounds -> last tab)")
	tabber.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("shift+tab")})
	assert.Equal(t, 1, tabber.selected)

	t.Log("navigate to next tab (out of bounds -> first tab)")
	tabber.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 0, tabber.selected)
}
