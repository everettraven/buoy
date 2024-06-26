package table

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/alecthomas/chroma/quick"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	buoytypes "github.com/everettraven/buoy/pkg/types"
	tbl "github.com/evertras/bubble-table/table"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

const (
	modeView  = "view"
	modeTable = "table"
	// TODO: These default sizes should probably
	// be configurable
	defaultPageSize    = 5
	defaultColumnWidth = 20
)

type KeyMap struct {
	ViewModeToggle key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.ViewModeToggle,
		},
	}
}

var DefaultKeys = KeyMap{
	ViewModeToggle: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "toggle viewing contents of selected resource"),
	),
}

type RowInfo struct {
	Row        tbl.Row
	Identifier *types.NamespacedName
	Index      int
}

type Styles struct {
	SelectedRow          lipgloss.Style
	TextAlignment        lipgloss.Style
	SyntaxHighlightDark  string
	SyntaxHighlightLight string
}

type ViewActionFunc func(row *RowInfo) (string, error)

// Model is a tea.Model implementation
// that represents a table panel
type Model struct {
	tableModel tbl.Model
	viewport   viewport.Model
	mode       string
	mutex      *sync.Mutex
	rows       map[types.UID]*RowInfo
	columns    []buoytypes.Column
	err        error
	tempRows   []tbl.Row
	keys       KeyMap
	table      *buoytypes.Table
	styles     Styles
	viewAction ViewActionFunc
}

func New(keys KeyMap, table *buoytypes.Table, styles Styles) *Model {
	tblColumns := []tbl.Column{}
	width := 0
	for _, column := range table.Columns {
		if column.Width > 0 {
			tblColumns = append(tblColumns, tbl.NewColumn(column.Header, column.Header, column.Width))
			width += column.Width
		} else {
			tblColumns = append(tblColumns, tbl.NewFlexColumn(column.Header, column.Header, 1))
			width += defaultColumnWidth
		}
	}

	pageSize := table.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	tab := tbl.New(tblColumns).
		SelectableRows(false).
		HighlightStyle(styles.SelectedRow).
		WithBaseStyle(styles.TextAlignment).
		WithPageSize(pageSize).
		WithHorizontalFreezeColumnCount(1).
		WithMultiline(true).
		WithTargetWidth(width).
		BorderRounded()

	return &Model{
		tableModel: tab,
		viewport:   viewport.New(0, 0),
		mode:       modeTable,
		mutex:      &sync.Mutex{},
		rows:       map[types.UID]*RowInfo{},
		columns:    table.Columns,
		keys:       keys,
		table:      table,
		styles:     styles,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.tableModel = m.tableModel.WithMaxTotalWidth(msg.Width)
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height / 2
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeys.ViewModeToggle):
			switch m.mode {
			case modeTable:
				m.mode = modeView
				m.tableModel = m.tableModel.Focused(false)
				row := m.FetchRowForIndex(m.tableModel.GetHighlightedRowIndex())
				vpContent, err := m.viewAction(row)
				if err != nil {
					m.viewport.SetContent(err.Error())
				} else {
					m.viewport.SetContent(highlight(vpContent, m.styles))
				}
			case modeView:
				m.mode = modeTable
				m.viewport.SetContent("")
				m.tableModel = m.tableModel.Focused(true)
			}
		}
	}

	if len(m.tempRows) > 0 {
		m.tableModel = m.tableModel.WithRows(m.tempRows)
		m.tempRows = []tbl.Row{}
	}

	switch m.mode {
	case modeTable:
		m.tableModel = m.tableModel.Focused(true)
		m.tableModel, cmd = m.tableModel.Update(msg)
	case modeView:
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

func (m *Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	switch m.mode {
	case modeTable:
		return m.tableModel.View()
	case modeView:
		return m.viewport.View()
	default:
		return fmt.Sprintf("unknown table state. table.mode=%q", m.mode)
	}
}

func (m *Model) AddOrUpdate(u *unstructured.Unstructured) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	uid := u.GetUID()
	rowData := tbl.RowData{}
	for _, column := range m.Columns() {
		val, err := getDotNotationValue(u.Object, column.Path)
		if err != nil {
			m.SetError(err)
			break
		}
		rowData[column.Header] = val
	}
	row := tbl.NewRow(rowData)
	row = row.WithStyle(m.styles.TextAlignment)

	m.rows[uid] = &RowInfo{
		Row:        row,
		Identifier: &types.NamespacedName{Namespace: u.GetNamespace(), Name: u.GetName()},
	}
	m.updateRows()
}

func (m *Model) DeleteRow(uid types.UID) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.rows, uid)
	m.updateRows()
}

func (m *Model) updateRows() {
	rows := []tbl.Row{}
	indice := 0
	for _, rowInfo := range m.rows {
		rows = append(rows, rowInfo.Row)
		rowInfo.Index = indice
		indice++
	}
	m.tempRows = rows
}

func (m *Model) Columns() []buoytypes.Column {
	return m.columns
}

func (m *Model) Name() string {
	return m.table.Name
}

func (m *Model) GVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   m.table.Group,
		Version: m.table.Version,
		Kind:    m.table.Kind,
	}
}

func (m *Model) Namespace() string {
	return m.table.Namespace
}

func (m *Model) LabelSelector() labels.Set {
	return m.table.LabelSelector
}

func (m *Model) SetError(err error) {
	m.err = err
}

func (m *Model) FetchRowForIndex(index int) *RowInfo {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var rowInfo *RowInfo
	for _, row := range m.rows {
		if row.Index == index {
			rowInfo = row
			break
		}
	}

	return rowInfo
}

func (m *Model) SetViewActionFunc(vaf ViewActionFunc) {
	m.viewAction = vaf
}

func (m *Model) Help() help.KeyMap {
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

func highlight(s string, styles Styles) string {
	// attempt to perform syntax highlighting
	theme := styles.SyntaxHighlightDark
	if !lipgloss.HasDarkBackground() {
		theme = styles.SyntaxHighlightLight
	}
	rw := &bytes.Buffer{}
	err := quick.Highlight(rw, s, "yaml", "terminal16m", theme)
	if err != nil {
		return s
	}
	highlighted, err := io.ReadAll(rw)
	if err != nil {
		return s
	}
	return string(highlighted)
}
