package cmd

import (
	"fmt"
	"os"
	"os/exec"

	onepassword "github.com/hinterland-software/openv/internal/1password"
	"github.com/hinterland-software/openv/internal/cli"
	"github.com/hinterland-software/openv/internal/logging"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [command]",
	Short: "Run a command with environment variables from 1Password",
	Long: `Execute a specified command with environment variables sourced from 1Password. 
This allows for secure and seamless integration of environment variables into your workflow.
Example usage:
  openv run --url github.com/org/repo --env production -- npm start`,
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
		env, err := cmd.Flags().GetString("env")
		if err != nil {
			return fmt.Errorf("❌ failed to get env flag: %w", err)
		}
		command, _ := cmd.Flags().GetString("command")
		vaultTitle, _ := cmd.Flags().GetString("vault")

		if command == "" && len(args) == 0 {
			return fmt.Errorf("❌ no command specified. Use -- <command> or --command '<command>'")
		}

		logging.Logger.Debug("starting command execution",
			"url", url,
			"env", env,
			"command", command,
			"args", args,
			"vault", vaultTitle)

		if vaultTitle == "" {
			vaultTitle = onepassword.DefaultVault
			logging.Logger.Debug("using default vault", "vault", vaultTitle)
		}

		logging.Logger.Debug("connecting to 1Password")
		service, err := onepassword.NewService(cmd.Context(), opServiceAuthToken)
		if err != nil {
			return fmt.Errorf("❌ failed to create 1Password service: %w", err)
		}

		logging.Logger.Debug("looking up vault", "vault", vaultTitle)
		vault, err := service.GetVault(vaultTitle)
		if err != nil {
			return fmt.Errorf("❌ failed to get vault: %w", err)
		}

		logging.Logger.Debug("fetching environment variables",
			"url", onepassword.GetBaseName(url),
			"env", env)
		envVars, err := service.GetEnvironment(onepassword.GetEnvironmentOptions{
			URL:     onepassword.GetBaseName(url),
			Env:     env,
			VaultID: vault.ID,
		})
		if err != nil {
			return fmt.Errorf("❌ failed to get environment variables: %w", err)
		}

		// Prepare command to run
		var cmdToRun *exec.Cmd
		if command != "" {
			logging.Logger.Debug("using shell command", "command", command)
			cmdToRun = exec.Command("sh", "-c", command)
		} else {
			logging.Logger.Debug("using direct command", "command", args[0], "args", args[1:])
			cmdToRun = exec.Command(args[0], args[1:]...)
		}

		// Set up environment
		cmdToRun.Env = os.Environ() // Start with current environment
		for key, value := range envVars.Variables {
			cmdToRun.Env = append(cmdToRun.Env, fmt.Sprintf("%s=%s", key, value))
		}

		logging.Logger.Info("running command",
			"env_vars_count", len(envVars.Variables))

		// Set up input/output
		cmdToRun.Stdin = os.Stdin
		cmdToRun.Stdout = os.Stdout
		cmdToRun.Stderr = os.Stderr

		logging.Logger.Info("running command", "env_vars_count", len(envVars.Variables))
		if err := cmdToRun.Run(); err != nil {
			return fmt.Errorf("❌ command failed: %w", err)
		}

		logging.Logger.Info("command completed successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().String("url", "", "Service URL")
	runCmd.Flags().String("env", "", "Environment (e.g., production, staging)")
	runCmd.Flags().String("vault", onepassword.DefaultVault, "1Password vault to use")
	runCmd.Flags().String("command", "", "Command to run (alternative to using --)")

	cobra.CheckErr(runCmd.MarkFlagRequired("url"))
	cobra.CheckErr(runCmd.MarkFlagRequired("env"))
}
