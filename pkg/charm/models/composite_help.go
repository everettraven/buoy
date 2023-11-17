package models

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

type CompositeHelpKeyMap struct {
	helps []help.KeyMap
}

func NewCompositeHelpKeyMap(helps ...help.KeyMap) *CompositeHelpKeyMap {
	return &CompositeHelpKeyMap{helps: helps}
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (ch CompositeHelpKeyMap) ShortHelp() []key.Binding {
	bindings := []key.Binding{}
	for _, h := range ch.helps {
		bindings = append(bindings, h.ShortHelp()...)
	}
	return bindings
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (ch CompositeHelpKeyMap) FullHelp() [][]key.Binding {
	bindings := [][]key.Binding{}
	for _, h := range ch.helps {
		bindings = append(bindings, h.FullHelp()...)
	}
	return bindings
}
