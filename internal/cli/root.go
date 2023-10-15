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

	panelModels := []tea.Model{}
	for _, panel := range dash.Panels {
		mod, err := p.Model(panel)
		if err != nil {
			log.Fatalf("getting model for panel %q: %s", panel.Name, err)
		}
		panelModels = append(panelModels, mod)
	}

	m := &models.Dashboard{Panels: panelModels}
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
