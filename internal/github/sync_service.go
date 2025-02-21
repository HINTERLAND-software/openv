package github

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-github/v69/github"
	"github.com/hinterland-software/openv/internal"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/oauth2"
)

// GitHubService handles syncing environment variables to GitHub
type GitHubService struct {
	client *github.Client
	ctx    context.Context
	token  string
}

// NewGitHubService creates a new GitHubService
func NewGitHubService(token string) *GitHubService {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	return &GitHubService{client: client, ctx: ctx, token: token}
}

// SyncToOrgSecret syncs environment variables to a GitHub organization secret
func (s *GitHubService) SyncToOrgSecret(org string, envVars map[string]string, withCleanup bool) error {
	publicKeyResponse, _, err := s.client.Actions.GetOrgPublicKey(s.ctx, org)
	if err != nil {
		return fmt.Errorf("failed to get organization public key: %w", err)
	}

	openvKey := internal.OPENV_KEYS_ORG_SECRETS
	if withCleanup {
		s.cleanupOrgSecrets(org, openvKey, envVars)
	}

	for key, value := range envVars {
		if key == internal.OPENV_KEYS {
			err = s.SyncToOrgVariable(org, map[string]string{openvKey: value}, false)
			if err != nil {
				return err
			}
			continue
		}

		encrypted, err := encodeWithPublicKey(value, *publicKeyResponse.Key)
		if err != nil {
			return err
		}

		_, err = s.client.Actions.CreateOrUpdateOrgSecret(s.ctx, org, &github.EncryptedSecret{
			Name:           key,
			EncryptedValue: encrypted,
			KeyID:          *publicKeyResponse.KeyID,
		})
		if err != nil {
			return fmt.Errorf("failed to sync organization secret %s: %w", key, err)
		}
	}
	return nil
}

func (s *GitHubService) cleanupOrgSecrets(org string, key string, envVars map[string]string) {
	existing, _, _ := s.client.Actions.GetOrgVariable(s.ctx, org, key)
	for _, removed := range getRemovedEnvVars(existing, envVars) {
		s.client.Actions.DeleteOrgSecret(s.ctx, org, removed)
	}
}

// SyncToRepoSecret syncs environment variables to a GitHub repository secret
func (s *GitHubService) SyncToRepoSecret(owner, repo string, envVars map[string]string, withCleanup bool) error {
	publicKeyResponse, _, err := s.client.Actions.GetRepoPublicKey(s.ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to get repository public key: %w", err)
	}
	openvKey := internal.OPENV_KEYS_REPO_SECRETS
	if withCleanup {
		s.cleanupRepoSecrets(owner, repo, openvKey, envVars)
	}

	for key, value := range envVars {
		if key == internal.OPENV_KEYS {
			err = s.SyncToRepoVariable(owner, repo, map[string]string{openvKey: value}, false)
			if err != nil {
				return err
			}
			continue
		}

		encrypted, err := encodeWithPublicKey(value, *publicKeyResponse.Key)
		if err != nil {
			return err
		}

		_, err = s.client.Actions.CreateOrUpdateRepoSecret(s.ctx, owner, repo, &github.EncryptedSecret{
			Name:           key,
			EncryptedValue: encrypted,
			KeyID:          *publicKeyResponse.KeyID,
		})
		if err != nil {
			return fmt.Errorf("failed to sync repository secret %s: %w", key, err)
		}
	}
	return nil
}

func (s *GitHubService) cleanupRepoSecrets(owner, repo, key string, envVars map[string]string) {
	existing, _, _ := s.client.Actions.GetRepoVariable(s.ctx, owner, repo, key)
	for _, removed := range getRemovedEnvVars(existing, envVars) {
		s.client.Actions.DeleteRepoSecret(s.ctx, owner, repo, removed)
	}
}

// SyncToRepoEnvironment syncs environment variables to a GitHub repository environment
func (s *GitHubService) SyncToRepoEnvironmentSecret(owner, repoName, env string, envVars map[string]string, withCleanup bool) error {
	repo, _, err := s.client.Repositories.Get(s.ctx, owner, repoName)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	repoID := int(repo.GetID())

	publicKeyResponse, _, err := s.client.Actions.GetEnvPublicKey(s.ctx, repoID, env)
	if err != nil {
		return fmt.Errorf("failed to get environment public key: %w", err)
	}

	if publicKeyResponse == nil {
		return fmt.Errorf("public key response is nil")
	}

	openvKey := internal.OPENV_KEYS_REPO_ENVIRONMENT_SECRETS
	if withCleanup {
		s.cleanupRepoEnvironmentSecrets(owner, repoName, env, openvKey, repoID, envVars)
	}

	for key, value := range envVars {
		if key == internal.OPENV_KEYS {
			err = s.SyncToRepoEnvironmentVariable(owner, repoName, env, map[string]string{openvKey: value}, false)
			if err != nil {
				return err
			}
			continue
		}

		encrypted, err := encodeWithPublicKey(value, *publicKeyResponse.Key)
		if err != nil {
			return fmt.Errorf("failed to encrypt value: %w", err)
		}

		_, err = s.client.Actions.CreateOrUpdateEnvSecret(s.ctx, repoID, env, &github.EncryptedSecret{
			Name:           key,
			EncryptedValue: encrypted,
			KeyID:          *publicKeyResponse.KeyID,
		})
		if err != nil {
			return fmt.Errorf("failed to sync repository environment secret %s: %w", key, err)
		}
	}
	return nil
}

func (s *GitHubService) cleanupRepoEnvironmentSecrets(owner, repoName, env, key string, repoID int, envVars map[string]string) {
	existing, _, _ := s.client.Actions.GetEnvVariable(s.ctx, owner, repoName, env, key)
	for _, removed := range getRemovedEnvVars(existing, envVars) {
		s.client.Actions.DeleteEnvSecret(s.ctx, repoID, env, removed)
	}
}

// SyncToOrgVariable syncs environment variables to a GitHub organization variable
func (s *GitHubService) SyncToOrgVariable(org string, envVars map[string]string, withCleanup bool) error {
	openVKey := internal.OPENV_KEYS_ORG_VARIABLES
	if withCleanup {
		s.cleanupOrgVariables(org, openVKey, envVars)
	}

	for key, value := range envVars {
		if key == internal.OPENV_KEYS {
			key = openVKey
		}

		if existing, _, err := s.client.Actions.GetOrgVariable(s.ctx, org, key); err == nil && existing != nil {
			if existing.Value != value {
				existing.Value = value
				_, err := s.client.Actions.UpdateOrgVariable(s.ctx, org, existing)
				if err != nil {
					return fmt.Errorf("failed to update organization variable %s: %w", key, err)
				}
			}
		} else {
			_, err := s.client.Actions.CreateOrgVariable(s.ctx, org, &github.ActionsVariable{
				Name:  key,
				Value: value,
			})
			if err != nil {
				return fmt.Errorf("failed to sync organization variable %s: %w", key, err)
			}
		}
	}
	return nil
}

func (s *GitHubService) cleanupOrgVariables(org string, key string, envVars map[string]string) {
	existing, _, _ := s.client.Actions.GetOrgVariable(s.ctx, org, key)
	for _, removed := range getRemovedEnvVars(existing, envVars) {
		s.client.Actions.DeleteOrgVariable(s.ctx, org, removed)
	}
}

// SyncToRepoVariable syncs environment variables to a GitHub repository variable
func (s *GitHubService) SyncToRepoVariable(owner, repo string, envVars map[string]string, withCleanup bool) error {
	openvKey := internal.OPENV_KEYS_REPO_VARIABLES
	if withCleanup {
		s.cleanupRepoVariables(owner, repo, openvKey, envVars)
	}

	for key, value := range envVars {
		if key == internal.OPENV_KEYS {
			key = openvKey
		}

		if existing, _, err := s.client.Actions.GetRepoVariable(s.ctx, owner, repo, key); err == nil && existing != nil {
			if existing.Value != value {
				existing.Value = value
				_, err := s.client.Actions.UpdateRepoVariable(s.ctx, owner, repo, existing)
				if err != nil {
					return fmt.Errorf("failed to update repository variable %s: %w", key, err)
				}
			}
		} else {
			_, err := s.client.Actions.CreateRepoVariable(s.ctx, owner, repo, &github.ActionsVariable{
				Name:  key,
				Value: value,
			})
			if err != nil {
				return fmt.Errorf("failed to sync repository variable %s: %w", key, err)
			}
		}
	}
	return nil
}

func (s *GitHubService) cleanupRepoVariables(owner, repo string, key string, envVars map[string]string) {
	existing, _, _ := s.client.Actions.GetRepoVariable(s.ctx, owner, repo, key)
	for _, removed := range getRemovedEnvVars(existing, envVars) {
		s.client.Actions.DeleteRepoVariable(s.ctx, owner, repo, removed)
	}
}

// SyncToRepoEnvironmentVariable syncs environment variables to a GitHub repository environment variable
func (s *GitHubService) SyncToRepoEnvironmentVariable(owner, repoName, env string, envVars map[string]string, withCleanup bool) error {
	openvKey := internal.OPENV_KEYS_REPO_ENVIRONMENT_VARIABLES
	if withCleanup {
		s.cleanupRepoEnvironmentVariables(owner, repoName, env, openvKey, envVars)
	}

	for key, value := range envVars {
		if key == internal.OPENV_KEYS {
			key = openvKey
		}

		if existing, _, err := s.client.Actions.GetEnvVariable(s.ctx, owner, repoName, env, key); err == nil && existing != nil {
			if existing.Value != value {
				existing.Value = value
				_, err := s.client.Actions.UpdateEnvVariable(s.ctx, owner, repoName, env, existing)
				if err != nil {
					return fmt.Errorf("failed to update variable %s: %w", key, err)
				}
			}
		} else {
			_, err := s.client.Actions.CreateEnvVariable(s.ctx, owner, repoName, env, &github.ActionsVariable{
				Name:  key,
				Value: value,
			})
			if err != nil {
				return fmt.Errorf("failed to sync repository environment variable %s: %w", key, err)
			}
		}
	}
	return nil
}

func (s *GitHubService) cleanupRepoEnvironmentVariables(owner, repoName, env string, key string, envVars map[string]string) {
	existing, _, _ := s.client.Actions.GetEnvVariable(s.ctx, owner, repoName, env, key)
	for _, removed := range getRemovedEnvVars(existing, envVars) {
		s.client.Actions.DeleteEnvVariable(s.ctx, owner, repoName, env, removed)
	}
}

// DeriveLocationFromURL derives the GitHub location (org/repo) from a URL
func DeriveLocationFromURL(url string) (string, string, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub URL: %s", url)
	}
	return parts[len(parts)-2], parts[len(parts)-1], nil
}

// getRemovedEnvVars returns a list of variable names that exist in the stored list
// (existing.Value as JSON array of strings) but are not present as keys in the new variables map
func getRemovedEnvVars(existing *github.ActionsVariable, newVarMap map[string]string) []string {
	existingVarNames := []string{}
	if existing != nil {
		// existing.Value contains a JSON array of variable names
		json.Unmarshal([]byte(existing.Value), &existingVarNames)
	}

	removedVarNames := []string{}
	for _, varName := range existingVarNames {
		if _, exists := newVarMap[varName]; !exists {
			removedVarNames = append(removedVarNames, varName)
		}
	}
	return removedVarNames
}

func encodeWithPublicKey(text string, publicKey string) (string, error) {
	// Decode the public key from base64
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return "", err
	}

	// Decode the public key
	var publicKeyDecoded [32]byte
	copy(publicKeyDecoded[:], publicKeyBytes)

	// Encrypt the secret value
	encrypted, err := box.SealAnonymous(nil, []byte(text), (*[32]byte)(publicKeyBytes), rand.Reader)

	if err != nil {
		return "", err
	}
	// Encode the encrypted value in base64
	encryptedBase64 := base64.StdEncoding.EncodeToString(encrypted)

	return encryptedBase64, nil
}
