package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aupeachmo/aigogo/pkg/auth"
)

type Pusher struct {
	client *http.Client
}

func NewPusher() *Pusher {
	return &Pusher{
		client: &http.Client{},
	}
}

// Push uploads an image to a registry using Docker Registry HTTP API V2
func (p *Pusher) Push(imageRef string) error {
	// Parse image reference
	registry, repository, tag, err := parseImageRef(imageRef)
	if err != nil {
		return err
	}

	// Get the image from local cache
	cache, err := getCacheDir()
	if err != nil {
		return err
	}

	imagePath := filepath.Join(cache, "images", sanitizeImageRef(imageRef))
	layerPath := filepath.Join(imagePath, "layer.tar")

	layerData, err := os.ReadFile(layerPath)
	if err != nil {
		return fmt.Errorf("image not found locally, build it first: %w", err)
	}

	// Get auth token (with repository scope for Docker Hub)
	authManager := auth.NewManager()
	token, err := authManager.GetToken(registry, repository)
	if err != nil {
		return fmt.Errorf("authentication required, run 'aigogo login %s': %w", registry, err)
	}

	// Upload config blob first (required by Docker Registry API)
	// The config is a minimal JSON object: {}
	configData := []byte("{}")
	configDigest, err := p.uploadBlob(registry, repository, configData, token)
	if err != nil {
		return fmt.Errorf("failed to upload config blob: %w", err)
	}

	// Upload layer blob
	layerDigest, err := p.uploadBlob(registry, repository, layerData, token)
	if err != nil {
		return fmt.Errorf("failed to upload layer blob: %w", err)
	}

	// Create and upload manifest (references both config and layer blobs)
	manifest := createManifest(configDigest, layerDigest, int64(len(layerData)))
	if err := p.uploadManifest(registry, repository, tag, manifest, token); err != nil {
		return fmt.Errorf("failed to upload manifest: %w", err)
	}

	return nil
}

func (p *Pusher) uploadBlob(registry, repository string, data []byte, token string) (string, error) {
	// Get actual API endpoint (Docker Hub uses registry-1.docker.io)
	apiEndpoint := getRegistryAPIEndpoint(registry)

	// Initiate blob upload
	url := fmt.Sprintf("https://%s/v2/%s/blobs/uploads/", apiEndpoint, repository)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	setAuthHeader(req, registry, token)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to initiate upload: %s - %s", resp.Status, string(body))
	}

	// Get upload URL
	uploadURL := resp.Header.Get("Location")
	if uploadURL == "" {
		return "", fmt.Errorf("no upload URL in response")
	}

	// If relative URL, make it absolute
	if !strings.HasPrefix(uploadURL, "http") {
		uploadURL = fmt.Sprintf("https://%s%s", apiEndpoint, uploadURL)
	}

	// Calculate digest
	digest := calculateDigest(data)

	// Upload the blob
	uploadURL = fmt.Sprintf("%s&digest=%s", uploadURL, digest)
	req, err = http.NewRequest("PUT", uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	setAuthHeader(req, registry, token)

	resp, err = p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to upload blob: %s - %s", resp.Status, string(body))
	}

	return digest, nil
}

func (p *Pusher) uploadManifest(registry, repository, tag string, manifest interface{}, token string) error {
	manifestData, err := json.Marshal(manifest)
	if err != nil {
		return err
	}

	// Get actual API endpoint (Docker Hub uses registry-1.docker.io)
	apiEndpoint := getRegistryAPIEndpoint(registry)

	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", apiEndpoint, repository, tag)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(manifestData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
	setAuthHeader(req, registry, token)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to upload manifest: %s - %s", resp.Status, string(body))
	}

	return nil
}

func createManifest(configDigest, layerDigest string, layerSize int64) map[string]interface{} {
	// Create a minimal Docker manifest v2
	// Both config and layer blobs must be uploaded before creating the manifest
	return map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.docker.distribution.manifest.v2+json",
		"config": map[string]interface{}{
			"mediaType": "application/vnd.docker.container.image.v1+json",
			"size":      2, // {} is 2 bytes
			"digest":    configDigest,
		},
		"layers": []map[string]interface{}{
			{
				"mediaType": "application/vnd.docker.image.rootfs.diff.tar",
				"size":      layerSize,
				"digest":    layerDigest,
			},
		},
	}
}
