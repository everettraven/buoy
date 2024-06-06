package panel

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/models/panels/item"
	"github.com/everettraven/buoy/pkg/types"
)

var _ PanelFactory = &Item{}

type Item struct {
	theme item.Styles
}

func (t *Item) ModelForPanel(panel types.Panel) (tea.Model, error) {
	item := types.Item{}
	err := json.Unmarshal(panel.Blob, &item)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling panel to item type: %s", err)
	}
	iw := t.modelWrapperForItemPanel(item)
	return iw, nil
}

func (t *Item) modelWrapperForItemPanel(itemPanel types.Item) *item.Model {
	vp := viewport.New(100, 20)
	return item.New(itemPanel, vp, t.theme)
}
