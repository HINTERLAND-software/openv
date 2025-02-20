package cmd

import (
	"fmt"
	"slices"

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
		setToken(cmd)

		svc := profile.NewService()
		name, _ := cmd.Flags().GetString("name")
		typeStr, _ := cmd.Flags().GetString("type")
		token, _ := cmd.Flags().GetString("token")
		url, _ := cmd.Flags().GetString("url")
		service, _ := cmd.Flags().GetString("service")

		// Interactive mode if flags not provided
		if name == "" {
			prompt := promptui.Prompt{
				Label: "Profile Name",
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
				Items: profile.ProfileTypes,
			}
			_, result, err := prompt.Run()
			if err != nil {
				return err
			}
			typeStr = result
		}

		if token == "" {
			prompt := promptui.Prompt{
				Label:       "Service Token",
				HideEntered: true,
			}
			var err error
			token, err = prompt.Run()
			if err != nil {
				return err
			}
		}

		if url == "" {
			prompt := promptui.Prompt{
				Label: "Service URL",
			}
			var err error
			url, err = prompt.Run()
			if err != nil {
				return err
			}
		}

		if service == "" {
			prompt := promptui.Select{
				Label: "Service",
				Items: profile.ValidServices,
			}
			_, result, err := prompt.Run()
			if err != nil {
				return err
			}
			service = result
		}

		return svc.AddProfile(name, url, profile.ServiceType(service), profile.SyncType(typeStr), token)
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
			fmt.Println("No profiles configured")
			return nil
		}

		for _, p := range profiles {
			fmt.Printf("- %s (%s)\n", p.Name, p.Type)
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
	Long:  `Update the details of an existing sync profile, such as its type, token, URL, and service type. The profile is identified by its name.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		typeStr, _ := cmd.Flags().GetString("type")
		token, _ := cmd.Flags().GetString("token")
		url, _ := cmd.Flags().GetString("url")
		service, _ := cmd.Flags().GetString("service")

		svc := profile.NewService()
		existingProfile, err := svc.GetProfile(name)
		if err != nil {
			return err
		}

		// Interactive mode if flags not provided
		if typeStr == "" {
			prompt := promptui.Select{
				Label:     "Sync Type",
				Items:     profile.ProfileTypes,
				CursorPos: slices.Index(profile.ProfileTypes, existingProfile.Type),
			}
			_, result, err := prompt.Run()
			if err != nil {
				return err
			}
			typeStr = result
		}

		if token == "" {
			prompt := promptui.Prompt{
				Label:       "Service Token",
				HideEntered: true,
				Default:     existingProfile.Token,
			}
			var err error
			token, err = prompt.Run()
			if err != nil {
				return err
			}
		}

		if url == "" {
			prompt := promptui.Prompt{
				Label:   "Service URL",
				Default: existingProfile.URL,
			}
			var err error
			url, err = prompt.Run()
			if err != nil {
				return err
			}
		}

		if service == "" {
			prompt := promptui.Select{
				Label:     "Service",
				Items:     profile.ValidServices,
				CursorPos: slices.Index(profile.ValidServices, existingProfile.Service),
			}
			_, result, err := prompt.Run()
			if err != nil {
				return err
			}
			service = result
		}

		return svc.UpdateProfile(name, url, profile.ServiceType(service), profile.SyncType(typeStr), token)
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileAddCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileRemoveCmd)
	profileCmd.AddCommand(profileUpdateCmd)

	profileAddCmd.Flags().String("name", "", "Profile name")
	profileAddCmd.Flags().String("type", "", "Sync type (environment or prefixed)")
	profileAddCmd.Flags().String("token", "", "Service token")
	profileAddCmd.Flags().String("url", "", "Service URL")
	profileAddCmd.Flags().String("service", "", "Service type (github, netlify, vercel, deno-deploy, shopify)")

	profileUpdateCmd.Flags().String("type", "", "Sync type (environment or prefixed)")
	profileUpdateCmd.Flags().String("token", "", "Service token")
	profileUpdateCmd.Flags().String("url", "", "Service URL")
	profileUpdateCmd.Flags().String("service", "", "Service type (github, netlify, vercel, deno-deploy, shopify)")
}
