package cli

import (
	"fmt"
	"slices"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func GetToken(cmd *cobra.Command) (string, error) {
	// First check if token is set via flag
	token := cmd.Flag("op-token").Value.String()

	// If not set via flag, check config
	if token == "" {
		token = viper.GetString("op-token")
	}

	if token == "" {
		prompt := promptui.Prompt{
			Label:       "1Password Service Account Token",
			HideEntered: true,
		}

		return prompt.Run()
	}

	return token, nil
}

type SelectItem struct {
	ID       string
	Selected bool
}

func TypesToItems(slice, selected []string) []*SelectItem {
	items := []*SelectItem{}
	for _, item := range slice {
		items = append(items, &SelectItem{
			ID:       item,
			Selected: slices.Contains(selected, item),
		})
	}
	return items
}

// SelectItems() prompts user to select one or more items in the given slice
func SelectItems(selectedPos int, allItems []*SelectItem) ([]*SelectItem, error) {
	// Always prepend a "Done" item to the slice if it doesn't
	// already exist.
	const doneID = "Done"
	if len(allItems) > 0 && allItems[len(allItems)-1].ID != doneID {
		var items = []*SelectItem{}
		allItems = append(items, allItems...)
		allItems = append(allItems, &SelectItem{ID: doneID})
	}

	// Define promptui template
	templates := &promptui.SelectTemplates{
		Label: `{{if .Selected}}
									✔
							{{end}} {{ .ID }} - label`,
		Active:   "→ {{if .Selected}}✔ {{end}}{{ .ID | cyan }}",
		Inactive: "{{if .Selected}}✔ {{end}}{{ .ID | cyan }}",
	}

	prompt := promptui.Select{
		Label:     "Item",
		Items:     allItems,
		Templates: templates,
		Size:      5,
		// Start the cursor at the currently selected index
		CursorPos:    selectedPos,
		HideSelected: true,
	}

	selectionIdx, _, err := prompt.Run()
	if err != nil {
		return nil, fmt.Errorf("prompt failed: %w", err)
	}

	chosenItem := allItems[selectionIdx]

	if chosenItem.ID != doneID {
		// If the user selected something other than "Done",
		// toggle selection on this item and run the function again.
		chosenItem.Selected = !chosenItem.Selected
		return SelectItems(selectionIdx, allItems)
	}

	// If the user selected the "Done" item, return
	// all selected items.
	var selectedItems []*SelectItem
	for _, i := range allItems {
		if i.Selected {
			selectedItems = append(selectedItems, i)
		}
	}
	return selectedItems, nil
}
