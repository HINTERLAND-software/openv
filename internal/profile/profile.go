package profile

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/spf13/viper"
)

type (
	SyncType string
	FlagType string
)

type Types interface {
	SyncType | FlagType
}

const (
	GithubEnvironmentSecret   SyncType = "github-environment-secret"
	GithubEnvironmentVariable SyncType = "github-environment-variable"
	GithubRepoSecret          SyncType = "github-repo-secret"
	GithubRepoVariable        SyncType = "github-repo-variable"
	GithubOrgSecret           SyncType = "github-org-secret"
	GithubOrgVariable         SyncType = "github-org-variable"

	NetlifyDeployContext                    SyncType = "netlify-deploy-context"
	NetlifyDeployContextScopeBuild          SyncType = "netlify-deploy-context-scope-build"
	NetlifyDeployContextScopeFunctions      SyncType = "netlify-deploy-context-scope-functions"
	NetlifyDeployContextScopeRuntime        SyncType = "netlify-deploy-context-scope-runtime"
	NetlifyDeployContextScopePostProcessing SyncType = "netlify-deploy-context-scope-post-processing"

	VercelEnvironmentCustom      SyncType = "vercel-environment-custom"
	VercelEnvironmentProduction  SyncType = "vercel-environment-production"
	VercelEnvironmentPreview     SyncType = "vercel-environment-preview"
	VercelEnvironmentDevelopment SyncType = "vercel-environment-development"

	DenoDeploy SyncType = "deno-deploy"

	ShopifyHydrogenEnvironment SyncType = "shopify-hydrogen-environment"

	FlagPrefixWithEnv FlagType = "prefix-with-env"
)

var (
	ProfileSyncsGithub = []SyncType{
		GithubEnvironmentSecret,
		GithubEnvironmentVariable,
		GithubRepoSecret,
		GithubRepoVariable,
		GithubOrgSecret,
		GithubOrgVariable,
	}

	ProfileSyncsNetlify = []SyncType{
		NetlifyDeployContext,
		NetlifyDeployContextScopeBuild,
		NetlifyDeployContextScopeFunctions,
		NetlifyDeployContextScopeRuntime,
		NetlifyDeployContextScopePostProcessing,
	}

	ProfileSyncsVercel = []SyncType{
		VercelEnvironmentCustom,
		VercelEnvironmentProduction,
		VercelEnvironmentPreview,
		VercelEnvironmentDevelopment,
	}

	ProfileSyncsDeno = []SyncType{
		DenoDeploy,
	}

	ProfileSyncsShopify = []SyncType{
		ShopifyHydrogenEnvironment,
	}

	Flags = []FlagType{
		FlagPrefixWithEnv,
	}
)

var ActiveProfileSyncs = append([]SyncType{}, ProfileSyncsGithub...)

func TypesToStrings[T Types](types []T) []string {
	strings := []string{}
	for _, t := range types {
		strings = append(strings, string(t))
	}
	return strings
}

// Profile represents a sync profile configuration
type Profile struct {
	Name  string     `json:"name" mapstructure:"name"`
	Sync  SyncType   `json:"sync" mapstructure:"sync"`
	Token string     `json:"token" mapstructure:"token"`
	URL   string     `json:"url" mapstructure:"url"`
	Flags []FlagType `json:"flags" mapstructure:"flags"`
}

// Service manages sync profiles
type Service struct {
	Profiles []Profile
}

// NewService creates a new profile service
func NewService() *Service {
	var profiles []Profile
	if err := viper.UnmarshalKey("sync_profiles", &profiles); err != nil {
		return &Service{Profiles: []Profile{}}
	}
	return &Service{Profiles: profiles}
}

// Save persists profiles to config
func (s *Service) Save() error {
	viper.Set("sync_profiles", s.Profiles)
	return viper.WriteConfig()
}

// AddProfile adds a new sync profile
func (s *Service) AddProfile(name, url string, syncType SyncType, token string, flags []FlagType) error {
	// Validate name uniqueness
	for _, p := range s.Profiles {
		if p.Name == name {
			return fmt.Errorf("profile with name %s already exists", name)
		}
	}

	// Validate sync type
	if err := ValidateType(syncType, ActiveProfileSyncs); err != nil {
		return err
	}

	// Validate flags
	if err := ValidateFlags(flags); err != nil {
		return err
	}

	s.Profiles = append(s.Profiles, Profile{
		Name:  name,
		Sync:  syncType,
		Token: token,
		URL:   url,
		Flags: flags,
	})
	return s.Save()
}

// GetProfile retrieves a profile by name
func (s *Service) GetProfile(name string) (*Profile, error) {
	for _, p := range s.Profiles {
		if p.Name == name {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("profile %s not found", name)
}

// ListProfiles returns all available profiles
func (s *Service) ListProfiles() []Profile {
	return s.Profiles
}

// RemoveProfile removes a profile by name
func (s *Service) RemoveProfile(name string) error {
	for i, p := range s.Profiles {
		if p.Name == name {
			s.Profiles = append(s.Profiles[:i], s.Profiles[i+1:]...)
			return s.Save()
		}
	}
	return fmt.Errorf("profile %s not found", name)
}

// UpdateProfile updates an existing profile
func (s *Service) UpdateProfile(name, url string, syncType SyncType, token string, flags []FlagType) error {
	// Validate sync type
	if err := ValidateType(syncType, ActiveProfileSyncs); err != nil {
		return err
	}

	// Validate flags
	if err := ValidateFlags(flags); err != nil {
		return err
	}

	for i, p := range s.Profiles {
		if p.Name == name {
			s.Profiles[i].Sync = syncType
			s.Profiles[i].Token = token
			s.Profiles[i].URL = url
			s.Profiles[i].Flags = flags
			return s.Save()
		}
	}
	return fmt.Errorf("profile %s not found", name)
}

func (s *Service) ValidateProfiles(profiles []string) error {
	for _, profile := range profiles {
		if _, err := s.GetProfile(profile); err != nil {
			return fmt.Errorf("invalid sync profile %s: %w", profile, err)
		}
	}
	return nil
}

func ValidateType[T Types](t T, validTypes []T) error {
	if !slices.Contains(validTypes, t) {
		return fmt.Errorf("%s is not a valid %s", t, reflect.TypeOf(t).String())
	}
	return nil
}

func ValidateFlags(flags []FlagType) error {
	for _, flag := range flags {
		if !slices.Contains(Flags, flag) {
			return fmt.Errorf("%s is not a valid flag", flag)
		}
	}
	return nil
}
