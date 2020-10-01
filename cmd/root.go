package cmd

import (
	"fmt"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// RootFlags stores all the global flags
type RootFlags struct {
	ConfigFolder string
}

var rootFlags = RootFlags{
	ConfigFolder: "~/.config/oauth2",
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "oauth2",
	Short: "Handle OAuth 2 flows",
	Long: `Handle OAuth 2 flows:

- Authorization Code flow
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&rootFlags.ConfigFolder, "config", "c", rootFlags.ConfigFolder, "path to the folder where caching and conf will be stored")
}

func initConfig() {
	if rootFlags.ConfigFolder == "" {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		rootFlags.ConfigFolder = path.Join(home, ".config", "oauth2")
	}

	folder, err := homedir.Expand(rootFlags.ConfigFolder)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	rootFlags.ConfigFolder = folder
}
