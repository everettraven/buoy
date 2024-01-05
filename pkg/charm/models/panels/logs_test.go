package panels

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/stretchr/testify/assert"
)

func TestEnterSearchMode(t *testing.T) {
	logs := NewLogs(DefaultLogsKeys, nil, styles.Theme{})
	logs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	assert.Equal(t, logs.mode, modeSearching)
}

func TestExecuteSearch(t *testing.T) {
	logs := NewLogs(DefaultLogsKeys, nil, styles.Theme{})
	logs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	logs.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, logs.mode, modeSearched)
}

func TestExitSearchMode(t *testing.T) {
	logs := NewLogs(DefaultLogsKeys, nil, styles.Theme{})
	logs.mode = modeSearching
	logs.searchbar.Focus()
	logs.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Equal(t, logs.mode, modeLogs)
	assert.False(t, logs.searchbar.Focused())

	logs.mode = modeSearched
	logs.searchbar.Focus()
	logs.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Equal(t, logs.mode, modeLogs)
	assert.False(t, logs.searchbar.Focused())
}

func TestSearchModeToggle(t *testing.T) {
	logs := NewLogs(DefaultLogsKeys, nil, styles.Theme{})
	logs.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	logs.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	assert.True(t, logs.strictSearch)
	logs.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	assert.False(t, logs.strictSearch)
}

func TestSearchLogs(t *testing.T) {
	t.Log("strict search")
	logs := NewLogs(DefaultLogsKeys, nil, styles.Theme{})
	logs.strictSearch = true
	logs.content = "some log line\nlog line with a search term\n"
	logs.viewport.Width = 50
	logs.searchbar.SetValue("search")
	match := logs.searchLogs()
	assert.Equal(t, "log line with a search term\n", match)

	t.Log("fuzzy search")
	logs.searchbar.SetValue("sll")
	logs.strictSearch = false
	match = logs.searchLogs()
	assert.Equal(t, "some log line\n", match)
}

func TestLogsWindowSizeUpdate(t *testing.T) {
	logs := NewLogs(DefaultLogsKeys, nil, styles.Theme{})
	logs.Update(tea.WindowSizeMsg{Width: 100, Height: 100})
	assert.Equal(t, logs.viewport.Width, 100)
	assert.Equal(t, logs.viewport.Height, 50)
}

func TestLogsAddContent(t *testing.T) {
	logs := NewLogs(DefaultLogsKeys, nil, styles.Theme{})
	logs.AddContent("some log line\n")
	assert.Equal(t, "\nsome log line\n", logs.content)
	assert.True(t, logs.contentUpdated)

	logs.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.False(t, logs.contentUpdated)
}

func TestLogsViewWithError(t *testing.T) {
	logs := NewLogs(DefaultLogsKeys, nil, styles.Theme{})
	err := errors.New("some error")
	logs.SetError(err)
	assert.Equal(t, err, logs.err)
	assert.Equal(t, err.Error(), logs.View())
}
