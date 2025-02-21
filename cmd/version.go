package cmd

import (
	"github.com/hinterland-software/openv/internal/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of openv",
	Long:  `All software has versions. This is openv's`,
	Run: func(cmd *cobra.Command, args []string) {
		logging.Logger.Info("openv version", "version", viper.GetString("version"))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
