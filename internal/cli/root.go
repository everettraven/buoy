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
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/everettraven/buoy/pkg/factories/datastream"
	"github.com/everettraven/buoy/pkg/factories/panel"
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
		themePath, err := cmd.Flags().GetString("theme")
		if err != nil {
			return fmt.Errorf("getting theme flag: %w", err)
		}
		return run(args[0], themePath)
	},
}

func init() {
	rootCommand.AddCommand(versionCommand)
	rootCommand.Flags().String("theme", styles.DefaultThemePath, "path to theme file")
}

type ErrorSetter interface {
	SetError(err error)
}

func run(path string, themePath string) error {
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

	theme, err := styles.LoadTheme(themePath)
	if err != nil {
		log.Fatalf("loading theme: %s", err)
	}

	p := panel.NewPanelFactory(theme)

	cfg := config.GetConfigOrDie()
	df, err := datastream.NewDatastreamFactory(cfg)
	if err != nil {
		log.Fatalf("configuring datastream factory: %s", err)
	}

	panelModels := []tea.Model{}
	for _, panel := range dash.Panels {
		mod, err := p.ModelForPanel(panel)
		if err != nil {
			log.Fatalf("getting model for panel %q: %s", panel.Name, err)
		}
		panelModels = append(panelModels, mod)
	}

	for _, panel := range panelModels {
		dataStream, err := df.DatastreamForModel(panel)
		if err != nil {
			if errSetter, ok := panel.(ErrorSetter); ok {
				errSetter.SetError(err)
			} else {
				log.Fatalf("getting datastream for model: %s", err)
			}
		}
		go dataStream.Run(make(<-chan struct{}))
	}

	m := models.NewDashboard(models.DefaultDashboardKeys, theme, panelModels...)
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
