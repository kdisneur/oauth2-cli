package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kdisneur/oauth2-cli/internal"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "displays the current command line version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(internal.GetVersionInfo())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
