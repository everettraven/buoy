package paneler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/everettraven/buoy/pkg/charm/models"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/everettraven/buoy/pkg/types"
	"github.com/treilik/bubbleboxer"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ Paneler = &Table{}

type Table struct {
	Client client.Client
}

func (t *Table) Node(panel types.Panel, bxr *bubbleboxer.Boxer) (bubbleboxer.Node, error) {
	tab := types.Table{}
	err := json.Unmarshal(panel.Blob, &tab)
	if err != nil {
		return bubbleboxer.Node{}, fmt.Errorf("unmarshalling panel to table type: %s", err)
	}
	mw, err := modelWrapperForTablePanel(t.Client, tab)
	if err != nil {
		return bubbleboxer.Node{}, fmt.Errorf("getting table widget: %s", err)
	}
	return nodeForModelWrapper(tab.Name, mw, bxr)
}

func modelWrapperForTablePanel(cli client.Client, tablePanel types.Table) (*models.ModelWrapper, error) {
	panelItems := &unstructured.UnstructuredList{}
	panelItems.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   tablePanel.Group,
		Version: tablePanel.Version,
		Kind:    tablePanel.Kind,
	})
	err := cli.List(context.Background(), panelItems)
	if err != nil {
		return nil, fmt.Errorf("fetching items for panel %q: %s", tablePanel.Name, err)
	}

	columns := []table.Column{}
	for _, column := range tablePanel.Columns {
		columns = append(columns, table.Column{Title: column.Header, Width: column.Width})
	}

	rows := []table.Row{}
	for _, item := range panelItems.Items {
		row := []string{}
		for _, column := range tablePanel.Columns {
			row = append(row, getDotNotationValue(item.Object, column.Path).(string))
		}
		rows = append(rows, row)
	}

	height := 5

	if len(rows) < height {
		height = len(rows)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	tw := &models.ModelWrapper{
		Model:   t,
		UpdateF: models.TableUpdateFunc,
		HeightF: models.TableHeightFunc,
	}
	tw.SetStyle(styles.FocusedModelStyle)
	return tw, nil
}
