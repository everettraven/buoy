package paneler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/models"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/everettraven/buoy/pkg/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var _ Paneler = &Item{}

type Item struct {
	Client client.Client
}

func (t *Item) Model(panel types.Panel) (tea.Model, error) {
	item := types.Item{}
	err := json.Unmarshal(panel.Blob, &item)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling panel to table type: %s", err)
	}
	return modelWrapperForItemPanel(t.Client, item)
}

func modelWrapperForItemPanel(cli client.Client, itemPanel types.Item) (*models.Panel, error) {
	item := &unstructured.Unstructured{}
	item.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   itemPanel.Group,
		Version: itemPanel.Version,
		Kind:    itemPanel.Kind,
	})

	err := cli.Get(context.Background(), itemPanel.Key, item)
	if err != nil {
		return nil, fmt.Errorf("getting item %q", itemPanel.Key.String())
	}

	vp := viewport.New(100, 8)
	itemJSON, err := item.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshalling item %q", itemPanel.Key.String())
	}

	itemYAML, err := yaml.JSONToYAML(itemJSON)
	if err != nil {
		return nil, fmt.Errorf("converting JSON to YAML for item %q", itemPanel.Key.String())
	}
	vp.SetContent(string(itemYAML))

	vpw := &models.Panel{
		Model:   vp,
		UpdateF: models.ViewportUpdateFunc,
		Name:    itemPanel.Name,
	}
	vpw.SetStyle(styles.ModelStyle)

	return vpw, nil
}
