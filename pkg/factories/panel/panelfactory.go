package panel

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/models/panels/item"
	"github.com/everettraven/buoy/pkg/charm/models/panels/logs"
	"github.com/everettraven/buoy/pkg/charm/models/panels/table"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/everettraven/buoy/pkg/types"
)

type PanelFactory interface {
	ModelForPanel(types.Panel) (tea.Model, error)
}

type paneler struct {
	panelerRegistry map[string]PanelFactory
}

var _ PanelFactory = &paneler{}

func (p *paneler) ModelForPanel(panel types.Panel) (tea.Model, error) {
	if p, ok := p.panelerRegistry[panel.Type]; ok {
		return p.ModelForPanel(panel)
	}
	return nil, fmt.Errorf("panel %q has unknown panel type: %q", panel.Name, panel.Type)
}

func NewPanelFactory(theme styles.Theme) PanelFactory {
	return &paneler{
		panelerRegistry: map[string]PanelFactory{
			types.PanelTypeTable: &Table{theme: table.Styles{
				SelectedRow:          theme.TableSelectedRowStyle(),
				SyntaxHighlightDark:  theme.SyntaxHighlightDarkTheme,
				SyntaxHighlightLight: theme.SyntaxHighlightLightTheme,
			}},
			types.PanelTypeItem: &Item{theme: item.Styles{
				SyntaxHighlightDark:  theme.SyntaxHighlightDarkTheme,
				SyntaxHighlightLight: theme.SyntaxHighlightLightTheme,
			}},
			types.PanelTypeLogs: &Log{theme: logs.Styles{
				SearchPrompt:              "> ",
				SearchPlaceholder:         "query",
				SearchModeStyle:           theme.LogSearchModeStyle(),
				SearchMatchHighlightStyle: theme.LogSearchHighlightStyle(),
			}},
		},
	}
}
