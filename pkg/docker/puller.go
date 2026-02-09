package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aupeach/aigogo/pkg/auth"
)

type Puller struct {
	client *http.Client
}

func NewPuller() *Puller {
	return &Puller{
		client: &http.Client{},
	}
}

// Pull downloads an image from a registry
func (p *Puller) Pull(imageRef string) error {
	// Parse image reference
	registry, repository, tag, err := parseImageRef(imageRef)
	if err != nil {
		return err
	}

	// Get auth token
	authManager := auth.NewManager()
	token, err := authManager.GetToken(registry, repository)
	if err != nil {
		// Try without auth for public registries
		token = ""
	}

	// Get manifest
	manifest, err := p.getManifest(registry, repository, tag, token)
	if err != nil {
		return fmt.Errorf("failed to get manifest: %w", err)
	}

	// Download layers
	layers, ok := manifest["layers"].([]interface{})
	if !ok || len(layers) == 0 {
		return fmt.Errorf("no layers found in manifest")
	}

	// For simplicity, we'll download the first layer (should be our snippet data)
	layer := layers[0].(map[string]interface{})
	digest := layer["digest"].(string)
	size := int64(layer["size"].(float64))

	layerData, err := p.downloadBlob(registry, repository, digest, token)

	if err != nil {
		return fmt.Errorf("failed to download layer: %w", err)
	}

	// Save to local cache
	cache, err := getCacheDir()
	if err != nil {
		return err
	}

	imagePath := filepath.Join(cache, "images", sanitizeImageRef(imageRef))
	if err := os.MkdirAll(imagePath, 0755); err != nil {
		return fmt.Errorf("failed to create image directory: %w", err)
	}

	// Save layer
	layerPath := filepath.Join(imagePath, "layer.tar")
	if err := os.WriteFile(layerPath, layerData, 0644); err != nil {
		return fmt.Errorf("failed to write layer: %w", err)
	}

	// Save metadata
	metadata := ImageMetadata{
		Ref:       imageRef,
		CreatedAt: time.Now(),
		Size:      size,
	}

	metadataData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	metadataPath := filepath.Join(imagePath, "metadata.json")
	if err := os.WriteFile(metadataPath, metadataData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

func (p *Puller) getManifest(registry, repository, tag, token string) (map[string]interface{}, error) {
	apiEndpoint := getRegistryAPIEndpoint(registry)
	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", apiEndpoint, repository, tag)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	setAuthHeader(req, registry, token)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get manifest: %s - %s", resp.Status, string(body))
	}

	var manifest map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}

	return manifest, nil
}

func (p *Puller) downloadBlob(registry, repository, digest, token string) ([]byte, error) {
	apiEndpoint := getRegistryAPIEndpoint(registry)
	url := fmt.Sprintf("https://%s/v2/%s/blobs/%s", apiEndpoint, repository, digest)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	setAuthHeader(req, registry, token)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to download blob: %s - %s", resp.Status, string(body))
	}

	return io.ReadAll(resp.Body)
}
