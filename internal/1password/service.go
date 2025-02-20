package onepassword

import (
	"context"
	"errors"
	"fmt"
	"strings"

	op "github.com/1password/onepassword-sdk-go"
	"github.com/hinterland-software/openv/internal/version"
)

// DefaultVault is the default vault name
const DefaultVault = "service-account"

// ImportOptions represents the options for importing environment variables
type ImportOptions struct {
	Name         string
	Env          string
	FilePath     string
	URL          string
	VaultID      string
	SyncProfiles []string
}

// GetEnvironmentOptions represents the options for getting environment variables
type GetEnvironmentOptions struct {
	URL     string
	Env     string
	VaultID string
}

// EnvironmentResult contains environment variables and metadata
type EnvironmentResult struct {
	Variables map[string]string
	ItemID    string
}

// Service handles 1Password operations including environment variable management
type Service struct {
	client *op.Client
	ctx    context.Context
}

// NewService creates a new 1Password service
func NewService(ctx context.Context, token string) (*Service, error) {
	client, err := op.NewClient(ctx, op.WithServiceAccountToken(token), op.WithIntegrationInfo("openv", version.Version))
	if err != nil {
		return nil, fmt.Errorf("failed to create 1Password client: %w", err)
	}
	return &Service{
		client: client,
		ctx:    ctx,
	}, nil
}

// Import imports environment variables into 1Password
func (s *Service) Import(opts ImportOptions) (*op.Item, error) {
	// Read and parse env file
	envVars, err := parseEnvFile(opts.FilePath)
	if err != nil {
		return nil, err
	}

	// Create item template
	item := createItemTemplate(opts, envVars)

	// Check for existing item
	existingItem, err := s.FindExistingItem(opts)
	if err != nil {
		return nil, err
	}

	if existingItem != nil {
		// Update existing item with new template values
		*existingItem = op.Item{
			ID:       existingItem.ID,      // Preserve the original ID
			VaultID:  existingItem.VaultID, // Preserve the vault
			Version:  existingItem.Version, // Preserve the version
			Title:    item.Title,
			Category: item.Category,
			Tags:     item.Tags,
			Sections: item.Sections,
			Fields:   item.Fields,
			Notes:    *item.Notes,
			Websites: item.Websites,
		}
		return s.UpdateItem(*existingItem)
	}

	return s.CreateItem(item)
}

// GetVault retrieves a vault by title
func (s *Service) GetVault(title string) (*op.VaultOverview, error) {
	vaults, err := s.client.Vaults.ListAll(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list vaults: %w", err)
	}

	for {
		vault, err := vaults.Next()
		if errors.Is(err, op.ErrorIteratorDone) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to iterate vault: %w", err)
		} else if vault.Title == title {
			return vault, nil
		}
	}

	return nil, fmt.Errorf("vault %s not found", title)
}

// CreateItem creates a new item in 1Password
func (s *Service) CreateItem(item op.ItemCreateParams) (*op.Item, error) {
	createdItem, err := s.client.Items.Create(s.ctx, item)
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}
	return &createdItem, nil
}

// UpdateItem updates an existing item in 1Password
func (s *Service) UpdateItem(item op.Item) (*op.Item, error) {
	updatedItem, err := s.client.Items.Put(s.ctx, item)
	if err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}
	return &updatedItem, nil
}

// FindExistingItem looks for an existing item with the same name and environment
func (s *Service) FindExistingItem(opts ImportOptions) (*op.Item, error) {
	items, err := s.client.Items.ListAll(s.ctx, opts.VaultID)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	envTag := fmt.Sprintf("env:%s", opts.Env)

	for {
		itemOverview, err := items.Next()
		if errors.Is(err, op.ErrorIteratorDone) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to iterate item: %w", err)
		} else if strings.ToLower(itemOverview.Title) != opts.Name {
			continue
		}

		item, err := s.client.Items.Get(s.ctx, itemOverview.VaultID, itemOverview.ID)
		if err != nil {
			if err.Error() == "item is not in an active state" {
				continue
			}
			return nil, fmt.Errorf("failed to get item: %w", err)
		}
		for _, tag := range item.Tags {
			if tag == envTag {
				return &item, nil
			}
		}
	}

	return nil, nil
}

// GetEnvironment retrieves environment variables for a given URL and environment.
// Returns an EnvironmentResult containing the variables and item ID.
// Returns an error if no matching environment is found or if there's an API error.
func (s *Service) GetEnvironment(opts GetEnvironmentOptions) (*EnvironmentResult, error) {
	// Find the item by URL and environment
	items, err := s.client.Items.ListAll(s.ctx, opts.VaultID)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	envTag := fmt.Sprintf("env:%s", opts.Env)
	envVars := make(map[string]string)

	for {
		itemOverview, err := items.Next()
		if errors.Is(err, op.ErrorIteratorDone) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to iterate items: %w", err)
		}

		item, err := s.client.Items.Get(s.ctx, itemOverview.VaultID, itemOverview.ID)
		if err != nil {
			if err.Error() == "item is not in an active state" {
				continue
			}
			return nil, fmt.Errorf("failed to get item: %w", err)
		}

		// Check if this is the right item
		hasEnvTag := false
		hasURLField := false
		correctURL := false

		for _, tag := range item.Tags {
			if tag == envTag {
				hasEnvTag = true
				break
			}
		}

		for _, field := range item.Fields {
			if field.ID == "url" {
				hasURLField = true
				if field.Value == opts.URL {
					correctURL = true
				}
				break
			}
		}

		if !hasEnvTag || !hasURLField || !correctURL {
			continue
		}

		// Found the right item, extract environment variables
		for _, field := range item.Fields {
			if field.SectionID != nil && *field.SectionID == "variables" {
				envVars[field.Title] = field.Value
			}
		}
		return &EnvironmentResult{
			Variables: envVars,
			ItemID:    item.ID,
		}, nil
	}

	return nil, fmt.Errorf("no environment variables found for %s (%s)", opts.URL, opts.Env)
}

func (s *Service) ResolveToken(ctx context.Context, token string) (string, error) {
	resolvedToken, err := s.client.Secrets.Resolve(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to resolve token: %w", err)
	}
	return resolvedToken, nil
}
