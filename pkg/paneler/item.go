package paneler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/everettraven/buoy/pkg/charm/models"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/everettraven/buoy/pkg/types"
	"github.com/treilik/bubbleboxer"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var _ Paneler = &Item{}

type Item struct {
	Client client.Client
}

func (t *Item) Node(panel types.Panel, bxr *bubbleboxer.Boxer) (bubbleboxer.Node, error) {
	item := types.Item{}
	err := json.Unmarshal(panel.Blob, &item)
	if err != nil {
		return bubbleboxer.Node{}, fmt.Errorf("unmarshalling panel to table type: %s", err)
	}
	mw, err := modelWrapperForItemPanel(t.Client, item)
	if err != nil {
		return bubbleboxer.Node{}, fmt.Errorf("getting table widget: %s", err)
	}
	return nodeForModelWrapper(item.Name, mw, bxr)
}

func modelWrapperForItemPanel(cli client.Client, itemPanel types.Item) (*models.ModelWrapper, error) {
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

	vpw := &models.ModelWrapper{
		Model:   vp,
		UpdateF: models.ViewportUpdateFunc,
		HeightF: models.ViewportHeightFunc,
	}
	vpw.SetStyle(styles.FocusedModelStyle)

	return vpw, nil
}
