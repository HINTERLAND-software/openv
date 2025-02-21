package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hinterland-software/openv/internal/cli"
	"github.com/hinterland-software/openv/internal/logging"
	"github.com/hinterland-software/openv/internal/profile"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage sync profiles for different services",
	Long:  `Create, list, update, and remove sync profiles for various services like GitHub, Netlify, Vercel, etc. These profiles help in managing environment variables across different platforms.`,
}

var profileAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new sync profile",
	Long:  `Add a new sync profile to manage environment variables for a specific service. You can specify the profile name, type, token, URL, and service type.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		opServiceAuthToken, err = cli.GetToken(cmd)
		if err != nil {
			return fmt.Errorf("❌ failed to get token: %w", err)
		}

		svc := profile.NewService()
		name, _ := cmd.Flags().GetString("name")
		typeStr, _ := cmd.Flags().GetString("type")
		token, _ := cmd.Flags().GetString("token")
		url, _ := cmd.Flags().GetString("url")
		flagsArray, _ := cmd.Flags().GetStringArray("flags")

		// Interactive mode if flags not provided
		if name == "" {
			prompt := promptui.Prompt{
				Label:    "Profile Name",
				Validate: NotEmpty("name cannot be empty"),
			}
			var err error
			name, err = prompt.Run()
			if err != nil {
				return err
			}
		}

		if typeStr == "" {
			prompt := promptui.Select{
				Label: "Sync Type",
				Items: profile.ActiveProfileSyncs,
			}
			_, result, err := prompt.Run()
			if err != nil {
				return err
			}
			if result == "" {
				return fmt.Errorf("sync type cannot be empty")
			}
			typeStr = result
		}

		if token == "" {
			prompt := promptui.Prompt{
				Label:       "Service Token",
				HideEntered: true,
				Validate:    NotEmpty("token cannot be empty"),
			}
			var err error
			token, err = prompt.Run()
			if err != nil {
				return err
			}
		}

		if url == "" {
			prompt := promptui.Prompt{
				Label:    "Service URL",
				Validate: NotEmpty("URL cannot be empty"),
			}
			var err error
			url, err = prompt.Run()
			if err != nil {
				return err
			}
		}

		flags := []profile.FlagType{}
		if len(flagsArray) == 0 {
			selectedItems, err := cli.SelectItems(
				1,
				cli.TypesToItems(
					profile.TypesToStrings(profile.Flags),
					[]string{},
				),
			)
			if err != nil {
				return err
			}
			for _, item := range selectedItems {
				flags = append(flags, profile.FlagType(item.ID))
			}
		} else {
			for _, item := range flagsArray {
				flags = append(flags, profile.FlagType(item))
			}
		}

		return svc.AddProfile(name, url, profile.SyncType(typeStr), token, flags)
	},
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured sync profiles",
	Long:  `Display a list of all sync profiles that have been configured, showing their names and types.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := profile.NewService()
		profiles := svc.ListProfiles()

		if len(profiles) == 0 {
			logging.Logger.Info("No profiles configured")
			return nil
		}

		for _, p := range profiles {
			logging.Logger.Info("Profile", "name", p.Name, "type", p.Sync)
		}
		return nil
	},
}

var profileRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a sync profile by name",
	Long:  `Remove a specified sync profile from the configuration by providing its name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := profile.NewService()
		return svc.RemoveProfile(args[0])
	},
}

var profileUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update an existing sync profile",
	Long:  `Update the details of an existing sync profile, such as its type, token, URL. The profile is identified by its name.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		syncStr, _ := cmd.Flags().GetString("sync")
		token, _ := cmd.Flags().GetString("token")
		url, _ := cmd.Flags().GetString("url")
		flagsArray, _ := cmd.Flags().GetStringArray("flags")

		svc := profile.NewService()
		existingProfile, err := svc.GetProfile(name)
		if err != nil {
			return err
		}

		// Interactive mode if flags not provided
		if syncStr == "" {
			prompt := promptui.Select{
				Label:     "Sync Type",
				Items:     profile.ActiveProfileSyncs,
				CursorPos: slices.Index(profile.ActiveProfileSyncs, existingProfile.Sync),
			}
			_, result, err := prompt.Run()
			if err != nil {
				return err
			}
			if result == "" {
				return fmt.Errorf("sync type cannot be empty")
			}
			syncStr = result
		}

		if token == "" {
			prompt := promptui.Prompt{
				Label:       "Service Token",
				HideEntered: true,
				Default:     existingProfile.Token,
				Validate:    NotEmpty("token cannot be empty"),
			}
			var err error
			token, err = prompt.Run()
			if err != nil {
				return err
			}
		}

		if url == "" {
			prompt := promptui.Prompt{
				Label:    "Service URL",
				Default:  existingProfile.URL,
				Validate: NotEmpty("URL cannot be empty"),
			}
			var err error
			url, err = prompt.Run()
			if err != nil {
				return err
			}
		}

		flags := []profile.FlagType{}
		if len(flagsArray) == 0 {
			selectedPos := 0
			if len(existingProfile.Flags) > 0 {
				selectedPos = slices.Index(profile.Flags, existingProfile.Flags[0])
			}
			selectedItems, err := cli.SelectItems(
				selectedPos,
				cli.TypesToItems(
					profile.TypesToStrings(profile.Flags),
					profile.TypesToStrings(existingProfile.Flags),
				),
			)
			if err != nil {
				return err
			}
			for _, item := range selectedItems {
				flags = append(flags, profile.FlagType(item.ID))
			}
		} else {
			for _, item := range flagsArray {
				flags = append(flags, profile.FlagType(item))
			}
		}

		return svc.UpdateProfile(name, url, profile.SyncType(syncStr), token, flags)
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileAddCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileRemoveCmd)
	profileCmd.AddCommand(profileUpdateCmd)

	profileAddCmd.Flags().String("name", "", "Profile name")

	profileAddCmd.Flags().String("sync", "", fmt.Sprintf("Sync type (%s)", strings.Join(profile.TypesToStrings(profile.ActiveProfileSyncs), ", ")))
	profileAddCmd.Flags().String("token", "", "Service token")
	profileAddCmd.Flags().String("url", "", "Service URL")
	profileAddCmd.Flags().StringArray("flags", nil, fmt.Sprintf("Flags (%s)", strings.Join(profile.TypesToStrings(profile.Flags), ", ")))

	profileUpdateCmd.Flags().String("sync", "", fmt.Sprintf("Sync type (%s)", strings.Join(profile.TypesToStrings(profile.ActiveProfileSyncs), ", ")))
	profileUpdateCmd.Flags().String("token", "", "Service token")
	profileUpdateCmd.Flags().String("url", "", "Service URL")
	profileUpdateCmd.Flags().StringArray("flags", nil, fmt.Sprintf("Flags (%s)", strings.Join(profile.TypesToStrings(profile.Flags), ", ")))
}

func NotEmpty(errorMessage string) promptui.ValidateFunc {
	return promptui.ValidateFunc(func(input string) error {
		if input == "" {
			return fmt.Errorf("%s", errorMessage)
		}
		return nil
	})
}
