package panels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/alecthomas/chroma/quick"
	tbl "github.com/calyptia/go-bubble-table"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/everettraven/buoy/pkg/charm/styles"
	buoytypes "github.com/everettraven/buoy/pkg/types"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/yaml"
)

type TableKeyMap struct {
	ViewModeToggle key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k TableKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k TableKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ViewModeToggle},
	}
}

var DefaultTableKeys = TableKeyMap{
	ViewModeToggle: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "toggle viewing contents of selected resource"),
	),
}

const modeView = "view"
const modeTable = "table"

type RowInfo struct {
	Row        tbl.Row
	Identifier *types.NamespacedName
	// Is this necessary? Can the index change on different iterations?
	Index int
}

// Table is a tea.Model implementation
// that represents a table panel
type Table struct {
	tableModel tbl.Model
	lister     cache.GenericLister
	scope      meta.RESTScopeName
	viewport   viewport.Model
	mode       string
	mutex      *sync.Mutex
	rows       map[types.UID]*RowInfo
	columns    []buoytypes.Column
	err        error
	tempRows   []tbl.Row
	keys       TableKeyMap
	theme      styles.Theme
	table      *buoytypes.Table
}

func NewTable(keys TableKeyMap, table *buoytypes.Table, theme styles.Theme) *Table {
	tblColumns := []string{}
	width := 0
	for _, column := range table.Columns {
		tblColumns = append(tblColumns, column.Header)
		width += column.Width
	}

	tab := tbl.New(tblColumns, 100, 10)
	tab.Styles.SelectedRow = theme.TableSelectedRowStyle()

	return &Table{
		tableModel: tab,
		viewport:   viewport.New(0, 0),
		mode:       modeTable,
		mutex:      &sync.Mutex{},
		rows:       map[types.UID]*RowInfo{},
		columns:    table.Columns,
		keys:       keys,
		theme:      theme,
		table:      table,
	}
}

func (m *Table) Init() tea.Cmd {
	return nil
}

func (m *Table) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tableModel.SetSize(msg.Width, msg.Height/2)
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height / 2
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultTableKeys.ViewModeToggle):
			switch m.mode {
			case modeTable:
				m.mode = modeView
				vpContent, err := m.FetchContentForIndex(m.tableModel.Cursor())
				if err != nil {
					m.viewport.SetContent(err.Error())
				} else {
					m.viewport.SetContent(vpContent)
				}
			case modeView:
				m.mode = modeTable
				m.viewport.SetContent("")
			}
		}
	}

	if len(m.tempRows) > 0 {
		m.tableModel.SetRows(m.tempRows)
		m.tempRows = []tbl.Row{}
	}

	switch m.mode {
	case modeTable:
		m.tableModel, cmd = m.tableModel.Update(msg)
	case modeView:
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

func (m *Table) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	switch m.mode {
	case modeTable:
		return m.tableModel.View()
	case modeView:
		return m.viewport.View()
	default:
		return "?"
	}
}

func (m *Table) AddOrUpdate(u *unstructured.Unstructured) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	uid := u.GetUID()
	row := tbl.SimpleRow{}
	for _, column := range m.Columns() {
		val, err := getDotNotationValue(u.Object, column.Path)
		if err != nil {
			m.SetError(err)
			break
		}

		row = append(row, fmt.Sprint(val))
	}

	m.rows[uid] = &RowInfo{
		Row:        row,
		Identifier: &types.NamespacedName{Namespace: u.GetNamespace(), Name: u.GetName()},
	}
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
	indice := 0
	for _, rowInfo := range m.rows {
		rows = append(rows, rowInfo.Row)
		rowInfo.Index = indice
		indice++
	}
	m.tempRows = rows
}

func (m *Table) Columns() []buoytypes.Column {
	return m.columns
}

func (m *Table) Name() string {
	return m.table.Name
}

func (m *Table) TableDefinition() *buoytypes.Table {
	return m.table
}

func (m *Table) SetLister(lister cache.GenericLister) {
	m.lister = lister
}

func (m *Table) SetScope(scope meta.RESTScopeName) {
	m.scope = scope
}

func (m *Table) SetError(err error) {
	m.err = err
}

func (m *Table) FetchContentForIndex(index int) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var rowInfo *RowInfo
	for _, row := range m.rows {
		if row.Index == index {
			rowInfo = row
			break
		}
	}

	name := rowInfo.Identifier.String()
	if m.scope == meta.RESTScopeNameRoot {
		name = rowInfo.Identifier.Name
	}

	obj, err := m.lister.Get(name)
	if err != nil {
		return "", fmt.Errorf("fetching definition for %q: %w", name, err)
	}

	itemJSON, err := obj.(*unstructured.Unstructured).MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("error marshalling item %q: %w", name, err)
	}

	itemYAML, err := yaml.JSONToYAML(itemJSON)
	if err != nil {
		return "", fmt.Errorf("converting JSON to YAML for item %q: %w", name, err)
	}

	theme := m.theme.SyntaxHighlightDarkTheme
	if !lipgloss.HasDarkBackground() {
		theme = m.theme.SyntaxHighlightLightTheme
	}
	rw := &bytes.Buffer{}
	err = quick.Highlight(rw, string(itemYAML), "yaml", "terminal16m", theme)
	if err != nil {
		return "", fmt.Errorf("highlighting YAML for item %q: %w", name, err)
	}
	highlighted, err := io.ReadAll(rw)
	if err != nil {
		return "", fmt.Errorf("reading highlighted YAML for item %q: %w", name, err)
	}
	return string(highlighted), nil
}

func (m *Table) Help() help.KeyMap {
	return m.keys
}

func getDotNotationValue(item map[string]interface{}, dotPath string) (interface{}, error) {
	jsonBytes, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("error marshalling item to json: %w", err)
	}
	res := gjson.Get(string(jsonBytes), dotPath)
	if !res.Exists() {
		return "n/a", nil
	}
	return res.Value(), nil
}
