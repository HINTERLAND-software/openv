package cmd

import (
	"fmt"
	"os"

	"github.com/hinterland-software/openv/internal/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// genDocCmd represents the gen-doc command
var genDocCmd = &cobra.Command{
	Use:   "gen-doc",
	Short: "Generate CLI documentation in markdown format",
	Long:  `Generate comprehensive markdown documentation for all commands in the CLI, which can be used for reference or publishing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir, _ := cmd.Flags().GetString("output-dir")
		if outputDir == "" {
			outputDir = "./docs"
		}

		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		err := doc.GenMarkdownTree(rootCmd, outputDir)
		if err != nil {
			return fmt.Errorf("failed to generate documentation: %w", err)
		}

		logging.Logger.Info("Documentation generated", "dir", outputDir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(genDocCmd)
	genDocCmd.Flags().StringP("output-dir", "o", "./docs", "Directory to output the generated documentation")
}
