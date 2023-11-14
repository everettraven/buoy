package panels

import (
	"sync"

	tbl "github.com/calyptia/go-bubble-table"
	tea "github.com/charmbracelet/bubbletea"
	buoytypes "github.com/everettraven/buoy/pkg/types"
	"k8s.io/apimachinery/pkg/types"
)

// Table is a tea.Model implementation
// that represents a table panel
type Table struct {
	table   tbl.Model
	name    string
	mutex   *sync.Mutex
	rows    map[types.UID]tbl.Row
	columns []buoytypes.Column
	err     error
}

func NewTable(name string, table tbl.Model, columns []buoytypes.Column) *Table {
	return &Table{
		table:   table,
		name:    name,
		mutex:   &sync.Mutex{},
		rows:    map[types.UID]tbl.Row{},
		columns: columns,
	}
}

func (m *Table) Init() tea.Cmd { return nil }

func (m *Table) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.table.SetSize(msg.Width, msg.Height/2)
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *Table) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	return m.table.View()
}

func (m *Table) AddOrUpdateRow(uid types.UID, row tbl.Row) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.rows[uid] = row
	m.updateRows()
}

func (m *Table) DeleteRow(uid types.UID) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.rows, uid)
	m.updateRows()
}

func (m *Table) updateRows() {
	rows := []tbl.Row{}
	for _, row := range m.rows {
		rows = append(rows, row)
	}
	m.table.SetRows(rows)
}

func (m *Table) Columns() []buoytypes.Column {
	return m.columns
}

func (m *Table) Name() string {
	return m.name
}

func (m *Table) SetError(err error) {
	m.err = err
}
