package models

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/everettraven/buoy/pkg/charm/styles"
)

type Helper interface {
	Help() help.KeyMap
}

type Tab struct {
	Name  string
	Model tea.Model
}

type Tabber struct {
	tabs     []Tab
	selected int
	keyMap   TabberKeyMap
	width    int
}

func NewTabber(keyMap TabberKeyMap, tabs ...Tab) *Tabber {
	return &Tabber{
		tabs:   tabs,
		keyMap: keyMap,
	}
}

func (t *Tabber) Init() tea.Cmd {
	return nil
}

func (t *Tabber) Update(msg tea.Msg) (*Tabber, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, t.keyMap.TabRight):
			t.selected++
			if t.selected > len(t.tabs)-1 {
				t.selected = 0
			}
			return t, tea.ClearScreen
		case key.Matches(msg, t.keyMap.TabLeft):
			t.selected--
			if t.selected < 0 {
				t.selected = len(t.tabs) - 1
			}
			return t, tea.ClearScreen
		}
	case tea.WindowSizeMsg:
		t.width = msg.Width
		var cmd tea.Cmd
		for i := range t.tabs {
			var tempCmd tea.Cmd
			t.tabs[i].Model, tempCmd = t.tabs[i].Model.Update(msg)
			cmd = tea.Batch(cmd, tempCmd)
		}
		return t, cmd
	}

	var cmd tea.Cmd
	t.tabs[t.selected].Model, cmd = t.tabs[t.selected].Model.Update(msg)
	return t, cmd
}

func (t *Tabber) View() string {
	tabRightArrow := styles.TabGap().Render("  ▶  ")
	tabLeftArrow := styles.TabGap().Render("  ◀  ")

	pager := &pager{
		tabRightArrow: tabRightArrow,
		tabLeftArrow:  tabLeftArrow,
		pages:         []page{},
	}
	pager.setPages(t.tabs, t.selected, t.width)

	tabBlock := pager.renderForSelectedTab(t.selected)
	// gap is a repeating of the spaces so that the bottom border continues the entire width
	// of the terminal. This allows it to look like a proper set of tabs
	gap := styles.TabGap().Render(strings.Repeat(" ", max(0, t.width-lipgloss.Width(tabBlock)-2)))
	tabsWithBorder := lipgloss.JoinHorizontal(lipgloss.Bottom, tabBlock, gap)
	content := styles.ContentStyle().Render(t.tabs[t.selected].Model.View())
	return lipgloss.JoinVertical(0, tabsWithBorder, content)
}

func (t *Tabber) Help() help.KeyMap {
	helps := []help.KeyMap{}
	if helper, ok := t.tabs[t.selected].Model.(Helper); ok {
		helps = append(helps, helper.Help())
	}

	return CompositeHelpKeyMap{
		helps: append(helps, t.keyMap),
	}
}

type TabberKeyMap struct {
	TabRight key.Binding
	TabLeft  key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k TabberKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k TabberKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.TabLeft, k.TabRight},
	}
}

var DefaultTabberKeys = TabberKeyMap{
	TabRight: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "change tabs to the right"),
	),
	TabLeft: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "change tabs to the left"),
	),
}

type page struct {
	tabs  []string
	start int
	end   int
}

type pager struct {
	pages         []page
	tabRightArrow string
	tabLeftArrow  string
}

func (p *pager) renderForSelectedTab(selected int) string {
	tabPage := page{}
	for _, page := range p.pages {
		if page.start <= selected && page.end >= selected {
			tabPage = page
		}
	}

	tabBlock := lipgloss.JoinHorizontal(lipgloss.Top, tabPage.tabs...)
	if len(p.pages) > 1 {
		tabBlock = lipgloss.JoinHorizontal(lipgloss.Bottom, p.tabLeftArrow, tabBlock, p.tabRightArrow)
	}

	return tabBlock
}

func (p *pager) setPages(tabs []Tab, selected int, width int) {
	tabPages := []page{}
	tempTab := ""
	tempPage := page{start: 0, tabs: []string{}}
	for i, tab := range tabs {
		renderedTab := styles.TabStyle().Render(tab.Name)
		if i == selected {
			renderedTab = styles.SelectedTabStyle().Render(tab.Name)
		}
		tempTab = lipgloss.JoinHorizontal(lipgloss.Top, tempTab, renderedTab)
		joined := lipgloss.JoinHorizontal(lipgloss.Bottom, p.tabLeftArrow, tempTab, p.tabRightArrow)
		if lipgloss.Width(joined) > width-5 {
			tempPage.end = i
			tabPages = append(tabPages, tempPage)
			tempPage = page{start: i, tabs: []string{}}
			tempTab = lipgloss.JoinHorizontal(lipgloss.Top, "", renderedTab)
		}

		tempPage.tabs = append(tempPage.tabs, renderedTab)
	}
	if tempTab != "" {
		tempPage.end = len(tabs) - 1
		tabPages = append(tabPages, tempPage)
	}
	p.pages = tabPages
}
