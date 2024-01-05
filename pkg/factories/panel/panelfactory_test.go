package panel

import (
	"testing"

	"github.com/everettraven/buoy/pkg/charm/models/panels"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/everettraven/buoy/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestUnknownPanelType(t *testing.T) {
	panelFactory := NewPanelFactory(styles.Theme{})
	_, err := panelFactory.ModelForPanel(types.Panel{
		PanelBase: types.PanelBase{
			Name: "test",
			Type: "unknown",
		},
	})
	assert.Error(t, err)
}

func TestTablePanel(t *testing.T) {
	panelJSON := `{
		"name": "Deployments",
		"group": "apps",
		"version": "v1",
		"kind": "Deployment",
		"type": "table",
		"columns": [
			{
				"header": "Namespace",
				"path": "metadata.namespace"
			},
			{
				"header": "Name",
				"path": "metadata.name"
			},
			{
				"header": "Replicas",
				"path": "status.replicas"
			}
		]
	}`

	panel := &types.Panel{}
	err := panel.UnmarshalJSON([]byte(panelJSON))
	assert.NoError(t, err)

	panelFactory := NewPanelFactory(styles.Theme{})
	tbl, err := panelFactory.ModelForPanel(*panel)
	assert.NoError(t, err)
	assert.NotNil(t, tbl)
	assert.IsType(t, &panels.Table{}, tbl)
}

func TestItemPanel(t *testing.T) {
	panelJSON := `{
		"name": "Kube API Server",
		"group": "",
		"version": "v1",
		"kind": "Pod",
		"type": "item",
		"key": {
			"namespace": "kube-system",
			"name": "kube-apiserver-kind-control-plane"
		} 
	}`

	panel := &types.Panel{}
	err := panel.UnmarshalJSON([]byte(panelJSON))
	assert.NoError(t, err)

	panelFactory := NewPanelFactory(styles.Theme{})
	item, err := panelFactory.ModelForPanel(*panel)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.IsType(t, &panels.Item{}, item)
}

func TestLogPanel(t *testing.T) {
	panelJSON := `{
		"name": "Kube API Server Logs",
		"group": "",
		"version": "v1",
		"kind": "Pod",
		"type": "logs",
		"key": {
			"namespace": "kube-system",
			"name": "kube-apiserver-kind-control-plane"
		} 
	}`

	panel := &types.Panel{}
	err := panel.UnmarshalJSON([]byte(panelJSON))
	assert.NoError(t, err)

	panelFactory := NewPanelFactory(styles.Theme{})
	log, err := panelFactory.ModelForPanel(*panel)
	assert.NoError(t, err)
	assert.NotNil(t, log)
	assert.IsType(t, &panels.Logs{}, log)
}
