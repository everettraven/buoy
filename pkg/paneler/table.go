package paneler

import (
	"context"
	"encoding/json"
	"fmt"

	tbl "github.com/calyptia/go-bubble-table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/models"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/everettraven/buoy/pkg/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ Paneler = &Table{}

type Table struct {
	Client client.Client
}

func (t *Table) Model(panel types.Panel) (tea.Model, error) {
	tab := types.Table{}
	err := json.Unmarshal(panel.Blob, &tab)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling panel to table type: %s", err)
	}
	return modelWrapperForTablePanel(t.Client, tab)
}

func modelWrapperForTablePanel(cli client.Client, tablePanel types.Table) (*models.Panel, error) {
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

	columns := []string{}
	width := 0
	for _, column := range tablePanel.Columns {
		columns = append(columns, column.Header)
		width += column.Width
	}

	rows := []tbl.Row{}
	for _, item := range panelItems.Items {
		row := tbl.SimpleRow{}
		for _, column := range tablePanel.Columns {
			val, err := getDotNotationValue(item.Object, column.Path)
			if err != nil {
				return nil, err
			}
			switch val := val.(type) {
			case string:
				row = append(row, val)
			case map[string]interface{}:
				data, err := json.Marshal(val)
				if err != nil {
					return nil, fmt.Errorf("marshalling object data to string: %w", err)
				}
				row = append(row, string(data))
			default:
				row = append(row, fmt.Sprint(val))
			}
		}
		rows = append(rows, row)
	}

	height := 6

	if len(rows) < height {
		height = len(rows) + 1
	}

	t := tbl.New(columns, 100, height)
	t.SetRows(rows)

	tw := &models.Panel{
		Model:   t,
		UpdateF: models.TableUpdateFunc,
		Name:    tablePanel.Name,
	}
	tw.SetStyle(styles.ModelStyle)
	return tw, nil
}
