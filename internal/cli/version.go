package cli

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/charmbracelet/lipgloss"
	figure "github.com/common-nighthawk/go-figure"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/spf13/cobra"
)

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "installed version of buoy",
	Run: func(cmd *cobra.Command, args []string) {
		var out strings.Builder
		fig := figure.NewFigure("buoy", "rounded", true)
		header := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true).
			BorderForeground(styles.DefaultColor).
			Render(fig.String())

		out.WriteString(header + "\n\n")
		settingNameStyle := lipgloss.NewStyle().Bold(true)
		if bi, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range bi.Settings {
				name := settingNameStyle.Render(setting.Key)
				out.WriteString(fmt.Sprintf("%s: %s\n", name, setting.Value))
			}
		} else {
			out.WriteString("unable to read build info")
		}
		fmt.Print(out.String())
	},
}
