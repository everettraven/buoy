package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/models"
	"github.com/everettraven/buoy/pkg/paneler"
	"github.com/everettraven/buoy/pkg/types"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/yaml"
)

var rootCommand = &cobra.Command{
	Use:   "buoy [config]",
	Short: "declarative kubernetes dashboard in the terminal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(args[0])
	},
}

func init() {
	rootCommand.AddCommand(versionCommand)
}

func run(path string) error {
	var raw []byte
	var ext string
	u, err := url.ParseRequestURI(path)
	if err != nil {
		ext = filepath.Ext(path)
		raw, err = os.ReadFile(path)
		if err != nil {
			log.Fatalf("reading local config: %s", err)
		}
	} else {
		ext = filepath.Ext(u.Path)
		resp, err := http.Get(u.String())
		if err != nil {
			log.Fatalf("fetching remote config: %s", err)
		}
		defer resp.Body.Close()
		raw, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("reading remote config: %s", err)
		}
	}

	dash := &types.Dashboard{}
	if ext == ".yaml" {
		err = yaml.Unmarshal(raw, dash)
	} else {
		err = json.Unmarshal(raw, dash)
	}
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

	m := models.NewDashboard(models.DefaultDashboardKeys, panelModels...)
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
