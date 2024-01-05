package panels

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/styles"
	buoytypes "github.com/everettraven/buoy/pkg/types"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

func TestTableUpdate(t *testing.T) {
	t.Log("WindowSizeUpdate")
	table := NewTable(DefaultTableKeys, &buoytypes.Table{}, styles.Theme{})
	table.Update(tea.WindowSizeMsg{Width: 50, Height: 50})
	assert.Equal(t, 50, table.viewport.Width)
	assert.Equal(t, 25, table.viewport.Height)

	t.Log("toggle view mode on")
	table.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	assert.Equal(t, modeView, table.mode)

	t.Log("toggle view mode off")
	table.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	assert.Equal(t, modeTable, table.mode)
}

func TestAddOrUpdate(t *testing.T) {
	table := NewTable(DefaultTableKeys, &buoytypes.Table{
		Columns: []buoytypes.Column{
			{Header: "Name", Width: 10, Path: "metadata.name"},
		},
	}, styles.Theme{})

	t.Log("add a row")
	u := &unstructured.Unstructured{}
	u.SetName("test")
	u.SetNamespace("test-ns")
	u.SetUID(types.UID("test"))
	table.AddOrUpdate(u)
	assert.Len(t, table.rows, 1)
	assert.NotNil(t, table.rows[types.UID("test")].Row)
	assert.Equal(t, &types.NamespacedName{Namespace: "test-ns", Name: "test"}, table.rows[types.UID("test")].Identifier)

	t.Log("update a row")
	u.SetName("test2")
	table.AddOrUpdate(u)
	assert.Len(t, table.rows, 1)
	assert.NotNil(t, table.rows[types.UID("test")].Row)
	assert.Equal(t, &types.NamespacedName{Namespace: "test-ns", Name: "test2"}, table.rows[types.UID("test")].Identifier)
}

func TestDeleteRow(t *testing.T) {
	table := NewTable(DefaultTableKeys, &buoytypes.Table{
		Columns: []buoytypes.Column{
			{Header: "Name", Width: 10, Path: "metadata.name"},
		},
	}, styles.Theme{})

	t.Log("add a row")
	u := &unstructured.Unstructured{}
	u.SetName("test")
	u.SetNamespace("test-ns")
	u.SetUID(types.UID("test"))
	table.AddOrUpdate(u)
	assert.Len(t, table.rows, 1)

	t.Log("delete a row")
	table.DeleteRow(types.UID("test"))
	assert.Len(t, table.rows, 0)
}

func TestTableView(t *testing.T) {
	table := NewTable(DefaultTableKeys, &buoytypes.Table{}, styles.Theme{})

	t.Log("view with error state")
	err := errors.New("some error")
	table.SetError(err)
	assert.Equal(t, err.Error(), table.View())

	t.Log("view with table mode")
	table.SetError(nil)
	table.mode = modeTable
	assert.Equal(t, table.tableModel.View(), table.View())

	t.Log("view with view mode toggled on")
	table.mode = modeView
	table.viewport.SetContent("floof")
	assert.Equal(t, table.viewport.View(), table.View())
}

func TestGetDotNotationValueNonExistentValue(t *testing.T) {
	obj := map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": "baz",
		},
	}

	val, err := getDotNotationValue(obj, "foo.baz")
	assert.NoError(t, err)
	assert.Equal(t, "n/a", val)
}
