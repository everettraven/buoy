package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "unknown"

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "installed version of buoy",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("buoy", version)
	},
}
