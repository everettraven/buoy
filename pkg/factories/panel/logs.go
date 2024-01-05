package panel

import (
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/models/panels"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/everettraven/buoy/pkg/types"
)

var _ PanelFactory = &Log{}

type Log struct {
	theme styles.Theme
}

func (t *Log) ModelForPanel(panel types.Panel) (tea.Model, error) {
	log := &types.Logs{}
	err := json.Unmarshal(panel.Blob, log)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling panel to table type: %s", err)
	}
	logPanel := panels.NewLogs(panels.DefaultLogsKeys, log, t.theme)
	return logPanel, nil
}
