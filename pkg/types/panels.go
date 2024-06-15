package types

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/types"
)

const (
	PanelTypeTable = "table"
	PanelTypeItem  = "item"
	PanelTypeLogs  = "logs"
)

type PanelBase struct {
	Name    string `json:"name" yaml:"name"`
	Group   string `json:"group" yaml:"group"`
	Version string `json:"version" yaml:"version"`
	Kind    string `json:"kind" yaml:"kind"`
	Type    string `json:"type" yaml:"type"`
}

type Panel struct {
	PanelBase
	Blob json.RawMessage `json:"blob" yaml:"blob"`
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
	Columns       []Column          `json:"columns" yaml:"columns"`
	Namespace     string            `json:"namespace" yaml:"namespace"`
	LabelSelector map[string]string `json:"labelSelector" yaml:"labelSelector"`
	PageSize      int               `json:"pageSize" yaml:"pageSize"`
}

type Column struct {
	Header string `json:"header" yaml:"header"`
	Width  int    `json:"width" yaml:"width"`
	Path   string `json:"path" yaml:"path"`
}

type Item struct {
	PanelBase
	Key types.NamespacedName `json:"key" yaml:"key"`
}

type Logs struct {
	PanelBase
	Key       types.NamespacedName `json:"key" yaml:"key"`
	Container string               `json:"container" yaml:"container"`
}
