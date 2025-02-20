package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of openv",
	Long:  `All software has versions. This is openv's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("openv version %s\n", viper.GetString("version"))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
