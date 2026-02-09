package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Manager struct {
	configPath string
}

type AuthConfig struct {
	Auths map[string]AuthEntry `json:"auths"`
}

type AuthEntry struct {
	Auth string `json:"auth"` // base64 encoded username:password
}

func NewManager() *Manager {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".aigogo", "auth.json")
	return &Manager{configPath: configPath}
}

// Login stores credentials for a registry
func (m *Manager) Login(registry, username, password string) error {
	config, err := m.loadConfig()
	if err != nil {
		config = &AuthConfig{Auths: make(map[string]AuthEntry)}
	}

	// Encode credentials
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

	config.Auths[registry] = AuthEntry{Auth: auth}

	return m.saveConfig(config)
}

// Logout removes credentials for a registry
func (m *Manager) Logout(registry string) error {
	config, err := m.loadConfig()
	if err != nil {
		return err
	}

	delete(config.Auths, registry)

	return m.saveConfig(config)
}

// GetToken retrieves an auth token for a registry
// For Docker Hub, exchanges credentials for OAuth2 token with repository scope
// For other registries, returns base64 encoded username:password
// repository is optional but required for Docker Hub to get proper scopes
func (m *Manager) GetToken(registry, repository string) (string, error) {
	config, err := m.loadConfig()
	if err != nil {
		return "", err
	}

	entry, exists := config.Auths[registry]
	if !exists {
		return "", fmt.Errorf("not logged in to %s", registry)
	}

	// For Docker Hub, exchange credentials for OAuth2 token
	if registry == "docker.io" {
		return m.getDockerHubToken(entry.Auth, repository)
	}

	// For other registries, return base64 encoded credentials
	// Decode to verify it's valid
	_, err = base64.StdEncoding.DecodeString(entry.Auth)
	if err != nil {
		return "", fmt.Errorf("invalid auth token")
	}

	return entry.Auth, nil
}

// getDockerHubToken exchanges Docker Hub credentials for an OAuth2 token
func (m *Manager) getDockerHubToken(base64Auth, repository string) (string, error) {
	// Decode credentials
	decoded, err := base64.StdEncoding.DecodeString(base64Auth)
	if err != nil {
		return "", fmt.Errorf("invalid auth token")
	}

	parts := splitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid auth token format")
	}

	username, password := parts[0], parts[1]

	// Docker Hub OAuth2 token endpoint
	// Use specific repository scope if provided, otherwise use wildcard
	scope := "repository:*:push,pull"
	if repository != "" {
		// Extract username from repository if it's in format username/repo
		repoParts := strings.Split(repository, "/")
		if len(repoParts) >= 2 {
			// Use the repository as-is for scope
			scope = fmt.Sprintf("repository:%s:push,pull", repository)
		}
	}

	authURL := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=%s", scope)

	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create auth request: %w", err)
	}

	// Use Basic Auth
	req.SetBasicAuth(username, password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate with Docker Hub: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("docker hub authentication failed: %s - %s", resp.Status, string(body))
	}

	var tokenResponse struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResponse.Token == "" {
		return "", fmt.Errorf("no token in response")
	}

	return tokenResponse.Token, nil
}

// GetCredentials returns the username and password for a registry
func (m *Manager) GetCredentials(registry string) (username, password string, err error) {
	token, err := m.GetToken(registry, "")
	if err != nil {
		return "", "", err
	}

	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", "", fmt.Errorf("invalid auth token")
	}

	parts := splitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid auth token format")
	}

	return parts[0], parts[1], nil
}

func (m *Manager) loadConfig() (*AuthConfig, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &AuthConfig{Auths: make(map[string]AuthEntry)}, nil
		}
		return nil, err
	}

	var config AuthConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.Auths == nil {
		config.Auths = make(map[string]AuthEntry)
	}

	return &config, nil
}

func (m *Manager) saveConfig(config *AuthConfig) error {
	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with restricted permissions
	if err := os.WriteFile(m.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func splitN(s, sep string, n int) []string {
	result := []string{}
	for i := 0; i < n-1; i++ {
		idx := -1
		for j, c := range s {
			if string(c) == sep {
				idx = j
				break
			}
		}
		if idx == -1 {
			result = append(result, s)
			return result
		}
		result = append(result, s[:idx])
		s = s[idx+1:]
	}
	result = append(result, s)
	return result
}
