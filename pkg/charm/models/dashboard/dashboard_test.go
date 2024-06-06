package dashboard

import (
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/models/panels/item"
	"github.com/everettraven/buoy/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestDashboardUpdate(t *testing.T) {
	panels := []tea.Model{
		item.New(types.Item{
			PanelBase: types.PanelBase{
				Name: "test",
			},
		}, viewport.New(10, 10), item.Styles{}),
	}

	d := New(DefaultDashboardKeys, DashboardStyleOptions{}, panels...)

	t.Log("WindowSizeUpdate")
	d.Update(tea.WindowSizeMsg{Width: 50, Height: 50})
	assert.Equal(t, 50, d.width)

	t.Log("toggle detailed help")
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ctrl+h")})
	assert.True(t, d.help.ShowAll)

	t.Log("quit the program")
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	assert.Equal(t, cmd(), tea.Quit())
}
