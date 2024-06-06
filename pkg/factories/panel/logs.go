package panel

import (
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/models/panels/logs"
	"github.com/everettraven/buoy/pkg/types"
)

var _ PanelFactory = &Log{}

type Log struct {
	theme logs.Styles
}

func (t *Log) ModelForPanel(panel types.Panel) (tea.Model, error) {
	log := &types.Logs{}
	err := json.Unmarshal(panel.Blob, log)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling panel to table type: %s", err)
	}
	logPanel := logs.New(logs.DefaultKeys, log, t.theme)
	return logPanel, nil
}
