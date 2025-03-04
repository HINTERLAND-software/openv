/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	onepassword "github.com/hinterland-software/openv/internal/1password"
	"github.com/hinterland-software/openv/internal/cli"
	"github.com/hinterland-software/openv/internal/logging"
	"github.com/hinterland-software/openv/internal/profile"
	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import [flags]",
	Short: "Import environment variables into 1Password",
	Long: `Import environment variables from a specified .env file into 1Password. 
The variables are stored securely with metadata and can be synchronized with different profiles.
Example usage:
  openv import --url github.com/org/repo --env staging --file .env.staging`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		opServiceAuthToken, err = cli.GetToken(cmd)
		if err != nil {
			return fmt.Errorf("❌ failed to get token: %w", err)
		}

		url, err := cmd.Flags().GetString("url")
		if err != nil {
			return fmt.Errorf("❌ failed to get url flag: %w", err)
		}
		baseURL := onepassword.GetBaseName(url)
		name, err := onepassword.GetName("", url)
		if err != nil {
			return err
		}
		env, err := cmd.Flags().GetString("env")
		if err != nil {
			return fmt.Errorf("❌ failed to get env flag: %w", err)
		}
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			return fmt.Errorf("❌ failed to get file flag: %w", err)
		}
		vaultTitle, _ := cmd.Flags().GetString("vault")

		logging.Logger.Debug("starting import",
			"url", url,
			"name", name,
			"env", env,
			"file", file,
			"vault", vaultTitle)

		if vaultTitle == "" {
			vaultTitle = onepassword.DefaultVault
			logging.Logger.Debug("using default vault", "vault", vaultTitle)
		}
		syncProfiles, _ := cmd.Flags().GetStringSlice("sync-profiles")

		logging.Logger.Debug("connecting to 1Password")
		service, err := onepassword.NewService(cmd.Context(), opServiceAuthToken)
		if err != nil {
			return fmt.Errorf("❌ failed to create 1Password service: %w", err)
		}

		// Validate profiles exist
		profileSvc := profile.NewService()
		err = profileSvc.ValidateProfiles(syncProfiles)
		if err != nil {
			return err
		}

		logging.Logger.Debug("looking up vault", "vault", vaultTitle)
		vault, err := service.GetVault(vaultTitle)
		if err != nil {
			return fmt.Errorf("❌ failed to get vault: %w", err)
		}

		opts := onepassword.ImportOptions{
			Name:         name,
			Env:          env,
			FilePath:     file,
			URL:          baseURL,
			VaultID:      vault.ID,
			SyncProfiles: syncProfiles,
		}

		logging.Logger.Debug("importing environment variables",
			"name", name,
			"env", env,
			"sync-profiles", syncProfiles)

		item, err := service.Import(opts)
		if err != nil {
			return fmt.Errorf("❌ failed to import environment variables: %w", err)
		}

		logging.Logger.Info("successfully imported environment variables",
			"name", name,
			"item_id", item.ID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().String("env", "", "Environment (e.g., production, staging)")
	importCmd.Flags().String("file", "", "Path to the environment file to import")
	importCmd.Flags().String("url", "", "Service URL")
	importCmd.Flags().String("vault", onepassword.DefaultVault, "1Password vault to use")
	importCmd.Flags().StringSlice("sync-profiles", []string{}, "Sync profiles to use")

	cobra.CheckErr(importCmd.MarkFlagRequired("env"))
	cobra.CheckErr(importCmd.MarkFlagRequired("file"))
	cobra.CheckErr(importCmd.MarkFlagRequired("url"))
}
