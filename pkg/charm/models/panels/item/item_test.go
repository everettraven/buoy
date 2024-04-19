package item

import (
	"errors"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestItemUpdate(t *testing.T) {
	item := New(types.Item{}, viewport.New(10, 10), Styles{})
	item.Update(tea.WindowSizeMsg{Width: 50, Height: 50})
	assert.Equal(t, 50, item.viewport.Width)
	assert.Equal(t, 25, item.viewport.Height)
}

func TestItemViewWithError(t *testing.T) {
	item := New(types.Item{}, viewport.New(10, 10), Styles{})
	err := errors.New("some error")
	item.SetError(err)
	assert.Equal(t, err.Error(), item.View())
}

func TestViewWithContent(t *testing.T) {
	item := New(types.Item{}, viewport.New(50, 50), Styles{})
	item.SetContent("some content")
	assert.Contains(t, item.View(), "some content")
}
