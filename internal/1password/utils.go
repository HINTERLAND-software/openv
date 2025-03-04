package onepassword

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	op "github.com/1password/onepassword-sdk-go"
	"github.com/hinterland-software/openv/internal/logging"
)

var (
	metadataSection = op.ItemSection{
		ID:    "metadata",
		Title: "Metadata",
	}

	variablesSection = op.ItemSection{
		ID:    "variables",
		Title: "Environment Variables",
	}

	// Create sections
	sections = []op.ItemSection{
		metadataSection,
		variablesSection,
	}
)

func parseEnvFile(filePath string) (map[string]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read env file: %w", err)
	}

	vars := make(map[string]string)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		vars[key] = value
	}

	return vars, nil
}

func createItemTemplate(opts ImportOptions, envVars map[string]string) op.ItemCreateParams {
	isoDate := time.Now().UTC().Format(time.RFC3339)

	tags := []string{"autogenerated", ".env", fmt.Sprintf("env:%s", opts.Env)}

	// Create fields
	fields := []op.ItemField{
		{
			ID:        "env",
			FieldType: op.ItemFieldTypeText,
			Title:     "Environment",
			Value:     opts.Env,
			SectionID: &metadataSection.ID,
		},
		{
			ID:        "url",
			FieldType: op.ItemFieldTypeText,
			Title:     "URL",
			Value:     opts.URL,
			SectionID: &metadataSection.ID,
		},
	}

	// Add sync profiles if provided
	if len(opts.SyncProfiles) > 0 {
		fields = append(fields, op.ItemField{
			ID:        "sync_profiles",
			FieldType: op.ItemFieldTypeText,
			Title:     "Sync Profiles",
			Value:     strings.Join(opts.SyncProfiles, ","),
			SectionID: &metadataSection.ID,
		})
	}

	// Add environment variables
	envFields := []op.ItemField{}
	for key, value := range envVars {
		envFields = append(envFields, op.ItemField{
			ID:        key,
			FieldType: op.ItemFieldTypeConcealed,
			Title:     key,
			Value:     value,
			SectionID: &variablesSection.ID,
		})
	}

	sort.Slice(envFields, func(i, j int) bool {
		return envFields[i].Title < envFields[j].Title
	})

	// Add JSON and .env representations
	return op.ItemCreateParams{
		Title:    opts.Name,
		Category: op.ItemCategorySecureNote,
		Tags:     tags,
		Sections: sections,
		Fields:   append(fields, envFields...),
		VaultID:  opts.VaultID,
		Notes:    &[]string{fmt.Sprintf("%s - %s\nDate: %s\n", opts.URL, opts.Env, isoDate)}[0],
	}
}

func GetName(name string, url string) (string, error) {
	// If name is not provided, try to derive it from GitHub URL
	if name == "" {
		if url != "" {
			derivedName := deriveNameFromGithubURL(url)
			if derivedName != "" {
				name = derivedName
				logging.Logger.Debug("using derived name from GitHub URL", "name", name)
			} else {
				return "", fmt.Errorf("❌ name is required when URL is not a GitHub repository")
			}
		} else {
			return "", fmt.Errorf("❌ name is required when URL is not provided")
		}
	}
	return getItemName(name), nil
}

func deriveNameFromGithubURL(url string) string {
	if url == "" {
		return ""
	}
	// Remove trailing slash if present
	url = strings.TrimSuffix(url, "/")
	// Split by / and get the last part
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	return ""
}

func getItemName(name string) string {
	return strings.ToLower(fmt.Sprintf(".env.%s", name))
}

func GetBaseName(url string) string {
	if url == "" {
		return ""
	}
	parts := strings.Split(url, "//")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return url
}
