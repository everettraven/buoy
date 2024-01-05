package panel

import (
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/models/panels"
	"github.com/everettraven/buoy/pkg/charm/styles"
	buoytypes "github.com/everettraven/buoy/pkg/types"
)

var _ PanelFactory = &Table{}

type Table struct {
	theme styles.Theme
}

func (t *Table) ModelForPanel(panel buoytypes.Panel) (tea.Model, error) {
	tab := &buoytypes.Table{}
	err := json.Unmarshal(panel.Blob, tab)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling panel to table type: %s", err)
	}
	table := panels.NewTable(panels.DefaultTableKeys, tab, t.theme)
	return table, nil
}
