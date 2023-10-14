package types

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/types"
)

// TODO: Think about how to link panels to allow for
// configurable layouts using the tree concept of bubble boxer

const (
	PanelTypeTable = "table"
	PanelTypeItem  = "item"
	PanelTypeLogs  = "logs"
)

type PanelBase struct {
	Name    string `json:"name"`
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
	Type    string `json:"type"`
}

type Panel struct {
	PanelBase
	Blob json.RawMessage `json:"blob"`
}

func (p *Panel) UnmarshalJSON(data []byte) error {
	pb := PanelBase{}
	if err := json.Unmarshal(data, &pb); err != nil {
		return err
	}
	p.PanelBase = pb
	p.Blob = data
	return nil
}

type Table struct {
	PanelBase
	Columns []Column `json:"columns"`
}

type Column struct {
	Header string `json:"header"`
	Width  int    `json:"width"`
	Path   string `json:"path"`
}

type Item struct {
	PanelBase
	Key types.NamespacedName
}

type Logs struct {
	PanelBase
	Key types.NamespacedName
}
