package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/models"
	"github.com/everettraven/buoy/pkg/paneler"
	"github.com/everettraven/buoy/pkg/types"
	"github.com/spf13/cobra"
	"github.com/treilik/bubbleboxer"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var rootCommand = &cobra.Command{
	Use:   "buoy [file.json]",
	Short: "declarative kubernetes dashboard in the terminal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(args[0])
	},
}

func run(file string) error {
	rawJson, err := os.ReadFile("test.json")
	if err != nil {
		log.Fatalf("reading test.json: %s", err)
	}

	dash := &types.Dashboard{}
	err = json.Unmarshal(rawJson, dash)
	if err != nil {
		log.Fatalf("unmarshalling dashboard: %s", err)
	}

	cfg := config.GetConfigOrDie()
	p, err := paneler.NewDefaultPaneler(cfg)
	if err != nil {
		log.Fatalf("configuring paneler: %s", err)
	}

	boxer := &bubbleboxer.Boxer{}
	leafs := []bubbleboxer.Node{}

	for _, panel := range dash.Panels {
		leaf, err := p.Node(panel, boxer)
		if err != nil {
			log.Fatalf("getting leaf node for panel %q: %s", panel.Name, err)
		}
		leafs = append(leafs, leaf)
	}

	boxer.LayoutTree = bubbleboxer.Node{
		VerticalStacked: true,
		SizeFunc: func(node bubbleboxer.Node, widthOrHeight int) []int {
			sizes := []int{}
			for _, children := range node.Children {
				sizes = append(sizes, children.SizeFunc(children, children.GetHeight())...)
			}
			return sizes
		},
		Children: leafs,
	}

	// for now only render the first table
	m := &models.Dashboard{Tui: boxer, Panels: dash.Panels}
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	return nil
}

func Execute() {
	if err := rootCommand.Execute(); err != nil {
		log.Fatal(err)
	}
}
