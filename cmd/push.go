package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/hinterland-software/openv/internal"
	onepassword "github.com/hinterland-software/openv/internal/1password"
	"github.com/hinterland-software/openv/internal/cli"
	"github.com/hinterland-software/openv/internal/github"
	"github.com/hinterland-software/openv/internal/logging"
	"github.com/hinterland-software/openv/internal/profile"
	"github.com/hinterland-software/openv/internal/version"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push environment variables to a file or sync profile",
	Long: `Push environment variables stored in 1Password to a local .env file or sync them to a specified service using a sync profile.
Example usage:
  openv push --url github.com/org/repo --env production --file .env
  openv push --url github.com/org/repo --env production --profile my-github-profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName, _ := cmd.Flags().GetString("profile")
		url, _ := cmd.Flags().GetString("url")
		env, _ := cmd.Flags().GetString("env")
		file, _ := cmd.Flags().GetString("file")
		force, _ := cmd.Flags().GetBool("force")
		vaultTitle, _ := cmd.Flags().GetString("vault")

		logging.Logger.Debug("push command parameters",
			"profile", profileName,
			"url", url,
			"env", env,
			"file", file,
			"force", force,
			"vault", vaultTitle)

		var err error
		opServiceAuthToken, err = cli.GetToken(cmd)
		if err != nil {
			return fmt.Errorf("❌ failed to get token: %w", err)
		}
		logging.Logger.Debug("successfully obtained auth token")

		op, vaultID, err := initializeOnePasswordService(cmd, vaultTitle)
		if err != nil {
			return err
		}
		logging.Logger.Debug("1Password service initialized", "vault_id", vaultID)

		var configProfile *profile.Profile
		if profileName != "" {
			svc := profile.NewService()
			configProfile, err = svc.GetProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to get profile: %w", err)
			}
			logging.Logger.Debug("profile found", "profile", configProfile.Name)
		}

		envVars, err := retrieveEnvironmentVariables(op, configProfile, vaultID, url, env)
		if err != nil {
			return err
		}

		switch {
		case file != "":
			return exportToFile(envVars, url, env, file)
		case profileName != "":
			return syncToProfile(envVars, url, configProfile, force)
		default:
			return fmt.Errorf("no file or profile specified")
		}
	},
}

func initializeOnePasswordService(cmd *cobra.Command, vaultTitle string) (*onepassword.Service, string, error) {
	if vaultTitle == "" {
		vaultTitle = onepassword.DefaultVault
		logging.Logger.Debug("using default vault", "vault", vaultTitle)
	}

	logging.Logger.Debug("connecting to 1Password")
	service, err := onepassword.NewService(cmd.Context(), opServiceAuthToken)
	if err != nil {
		return nil, "", fmt.Errorf("❌ failed to create 1Password service: %w", err)
	}

	logging.Logger.Debug("looking up vault", "vault", vaultTitle)
	vault, err := service.GetVault(vaultTitle)
	if err != nil {
		return nil, "", fmt.Errorf("❌ failed to get vault: %w", err)
	}

	return service, vault.ID, nil
}

func retrieveEnvironmentVariables(service *onepassword.Service, configProfile *profile.Profile, vaultID, url, env string) (*onepassword.EnvironmentResult, error) {
	logging.Logger.Debug("retrieving environment variables from 1Password", "vault", vaultID, "env", env)
	envVars, err := service.GetEnvironment(onepassword.GetEnvironmentOptions{
		URL:     onepassword.GetBaseName(url),
		Env:     env,
		VaultID: vaultID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve environment variables: %w", err)
	}

	if configProfile != nil && slices.Contains(configProfile.Flags, profile.FlagPrefixWithEnv) {
		envVars.Variables = prefixKeys(envVars, env)
	}

	if err := addOpenvKey(envVars); err != nil {
		return nil, err
	}

	return envVars, nil
}

func addOpenvKey(envVars *onepassword.EnvironmentResult) error {
	keys := make([]string, 0, len(envVars.Variables))
	for key := range envVars.Variables {
		keys = append(keys, key)
	}
	keysJSON, err := json.Marshal(keys)
	if err != nil {
		return fmt.Errorf("failed to marshal keys: %w", err)
	}
	envVars.Variables[internal.OPENV_KEYS] = string(keysJSON)
	return nil
}

func prefixKeys(envVars *onepassword.EnvironmentResult, env string) map[string]string {
	variables := make(map[string]string)
	logging.Logger.Debug("prefixing environment variables with environment", "env", env)
	for k, v := range envVars.Variables {
		prefixed := fmt.Sprintf("%s_%s", strings.ToUpper(env), k)
		variables[prefixed] = v
	}
	return variables
}

func syncToProfile(envVars *onepassword.EnvironmentResult, url string, configProfile *profile.Profile, force bool) error {
	if !force {
		confirmPrompt := promptui.Prompt{
			Label:     fmt.Sprintf("Are you sure you want to sync to %s using profile %s", url, configProfile.Name),
			IsConfirm: true,
		}
		_, err := confirmPrompt.Run()
		if err != nil {
			return fmt.Errorf("operation cancelled")
		}
	}

	var err error
	switch {
	case slices.Contains(profile.ProfileSyncsGithub, configProfile.Sync):
		var owner, repo string
		owner, repo, err = github.DeriveLocationFromURL(url)
		if err != nil {
			return err
		}
		githubService := github.NewGitHubService(configProfile.Token)

		switch configProfile.Sync {
		case profile.GithubEnvironmentSecret:
			err = githubService.SyncToRepoEnvironmentSecret(owner, repo, envVars.Env, envVars.Variables, true)
		case profile.GithubEnvironmentVariable:
			err = githubService.SyncToRepoEnvironmentVariable(owner, repo, envVars.Env, envVars.Variables, true)
		case profile.GithubRepoSecret:
			err = githubService.SyncToRepoSecret(owner, repo, envVars.Variables, true)
		case profile.GithubRepoVariable:
			err = githubService.SyncToRepoVariable(owner, repo, envVars.Variables, true)
		case profile.GithubOrgSecret:
			err = githubService.SyncToOrgSecret(owner, envVars.Variables, true)
		case profile.GithubOrgVariable:
			err = githubService.SyncToOrgVariable(owner, envVars.Variables, true)
		default:
			err = fmt.Errorf("sync not implemented for %s", configProfile.Sync)
		}
	default:
		// TODO: Implement missing syncs
		err = fmt.Errorf("sync not implemented for %s", configProfile.Sync)
	}

	if err != nil {
		return fmt.Errorf("failed to sync environment variables: %w", err)
	}

	logging.Logger.Info("Environment variables synced successfully")
	return nil
}

func exportToFile(envVars *onepassword.EnvironmentResult, url, env string, file string) error {
	logging.Logger.Debug("starting export",
		"url", url,
		"env", env,
		"file", file)

	var keys []string
	for k := range envVars.Variables {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var content strings.Builder
	content.WriteString(fmt.Sprintf("# Environment variables for %s (%s)\n", url, env))
	content.WriteString(fmt.Sprintf("# Generated by openv %s\n\n", version.Info()))
	for _, key := range keys {
		value := envVars.Variables[key]
		if strings.Contains(value, " ") || strings.Contains(value, "#") {
			value = fmt.Sprintf(`"%s"`, value)
		}
		content.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}

	logging.Logger.Debug("writing environment variables to file", "file", file, "count", len(envVars.Variables))
	if err := os.WriteFile(file, []byte(content.String()), 0600); err != nil {
		return fmt.Errorf("❌ failed to write environment file: %w", err)
	}

	logging.Logger.Info("successfully exported environment variables to file", "file", file)
	return nil
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().String("profile", "", "Sync profile name")
	pushCmd.Flags().String("url", "", "Service URL")
	pushCmd.Flags().String("env", "", "Environment (e.g., production, staging)")
	pushCmd.Flags().StringP("file", "f", ".env", "Path to the output environment file")
	pushCmd.Flags().BoolP("force", "y", false, "Do not prompt for confirmation")
	pushCmd.Flags().String("vault", onepassword.DefaultVault, "1Password vault to use")

	cobra.CheckErr(pushCmd.MarkFlagRequired("url"))
	cobra.CheckErr(pushCmd.MarkFlagRequired("env"))
}
