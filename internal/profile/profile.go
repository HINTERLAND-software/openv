package profile

import (
	"fmt"
	"slices"

	"github.com/spf13/viper"
)

type SyncType string
type ServiceType string

const (
	Environment SyncType    = "environment"
	Prefixed    SyncType    = "prefixed"
	GitHub      ServiceType = "github"
	Netlify     ServiceType = "netlify"
	Vercel      ServiceType = "vercel"
	DenoDeploy  ServiceType = "deno-deploy"
	Shopify     ServiceType = "shopify"
)

var ProfileTypes = []SyncType{Environment, Prefixed}
var ValidServices = []ServiceType{GitHub, Netlify, Vercel, DenoDeploy, Shopify}

// Profile represents a sync profile configuration
type Profile struct {
	Name    string      `json:"name" mapstructure:"name"`
	Type    SyncType    `json:"type" mapstructure:"type"`
	Token   string      `json:"token" mapstructure:"token"`
	URL     string      `json:"url" mapstructure:"url"`
	Service ServiceType `json:"service" mapstructure:"service"`
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
func (s *Service) AddProfile(name, url string, service ServiceType, syncType SyncType, token string) error {
	// Validate name uniqueness
	for _, p := range s.Profiles {
		if p.Name == name {
			return fmt.Errorf("profile with name %s already exists", name)
		}
	}

	// Validate sync type
	if !slices.Contains(ProfileTypes, syncType) {
		return fmt.Errorf("invalid sync type: %s", syncType)
	}

	// Validate service
	isValidService := false
	for _, s := range ValidServices {
		if service == s {
			isValidService = true
			break
		}
	}
	if !isValidService {
		return fmt.Errorf("invalid service: %s", service)
	}

	s.Profiles = append(s.Profiles, Profile{Name: name, Type: syncType, Token: token, URL: url, Service: service})
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
func (s *Service) UpdateProfile(name, url string, service ServiceType, syncType SyncType, token string) error {
	// Validate sync type
	if !slices.Contains(ProfileTypes, syncType) {
		return fmt.Errorf("invalid sync type: %s", syncType)
	}

	// Validate service
	isValidService := false
	for _, s := range ValidServices {
		if service == s {
			isValidService = true
			break
		}
	}
	if !isValidService {
		return fmt.Errorf("invalid service: %s", service)
	}

	for i, p := range s.Profiles {
		if p.Name == name {
			s.Profiles[i].Type = syncType
			s.Profiles[i].Token = token
			s.Profiles[i].URL = url
			s.Profiles[i].Service = service
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
