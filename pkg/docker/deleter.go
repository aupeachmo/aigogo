package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/aupeach/aigogo/pkg/auth"
)

// Deleter handles deleting images from registries
type Deleter struct {
	client *http.Client
}

// NewDeleter creates a new deleter
func NewDeleter() *Deleter {
	return &Deleter{
		client: &http.Client{},
	}
}

// Delete removes an image from a registry
func (d *Deleter) Delete(imageRef string) error {
	registry, repository, tag, err := parseImageRef(imageRef)
	if err != nil {
		return err
	}

	// Get authentication token
	authManager := auth.NewManager()
	token, err := authManager.GetToken(registry, repository)
	if err != nil {
		return fmt.Errorf("not logged in to %s: %w\nRun 'aigogo login %s' first", registry, err, registry)
	}

	// First, get the manifest digest
	// We need the digest to delete (can't delete by tag directly)
	apiEndpoint := getRegistryAPIEndpoint(registry)
	manifestURL := fmt.Sprintf("https://%s/v2/%s/manifests/%s", apiEndpoint, repository, tag)

	req, err := http.NewRequest("HEAD", manifestURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	setAuthHeader(req, registry, token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Set("Accept", "application/vnd.oci.image.manifest.v1+json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get manifest: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 401 {
		return fmt.Errorf("authentication required, run 'aigogo login %s'", registry)
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("image not found: %s", imageRef)
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get manifest: %s (status %d)", string(body), resp.StatusCode)
	}

	// Get the digest from the response header
	digest := resp.Header.Get("Docker-Content-Digest")
	if digest == "" {
		return fmt.Errorf("registry did not return manifest digest (may not support deletion)")
	}

	// Now delete using the digest
	deleteURL := fmt.Sprintf("https://%s/v2/%s/manifests/%s", apiEndpoint, repository, digest)

	deleteReq, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	setAuthHeader(deleteReq, registry, token)

	deleteResp, err := d.client.Do(deleteReq)
	if err != nil {
		return fmt.Errorf("failed to delete manifest: %w", err)
	}
	defer func() { _ = deleteResp.Body.Close() }()

	if deleteResp.StatusCode == 401 {
		return fmt.Errorf("authentication failed (insufficient permissions)")
	}

	if deleteResp.StatusCode == 404 {
		return fmt.Errorf("manifest not found (may have been already deleted)")
	}

	if deleteResp.StatusCode == 405 {
		return fmt.Errorf("registry does not support deletion (check registry configuration)")
	}

	if deleteResp.StatusCode != 202 && deleteResp.StatusCode != 200 {
		body, _ := io.ReadAll(deleteResp.Body)

		// Try to parse error response
		var errResp struct {
			Errors []struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"errors"`
		}

		if json.Unmarshal(body, &errResp) == nil && len(errResp.Errors) > 0 {
			return fmt.Errorf("registry error: %s - %s (status %d)",
				errResp.Errors[0].Code, errResp.Errors[0].Message, deleteResp.StatusCode)
		}

		return fmt.Errorf("failed to delete: %s (status %d)", string(body), deleteResp.StatusCode)
	}

	return nil
}

// listTags lists all tags in a repository
func (d *Deleter) listTags(registry, repository string) ([]string, error) {
	authManager := auth.NewManager()
	token, err := authManager.GetToken(registry, repository)
	if err != nil {
		return nil, fmt.Errorf("not logged in to %s: %w", registry, err)
	}

	// Docker Registry API: GET /v2/<name>/tags/list
	apiEndpoint := getRegistryAPIEndpoint(registry)
	url := fmt.Sprintf("https://%s/v2/%s/tags/list", apiEndpoint, repository)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	setAuthHeader(req, registry, token)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("authentication required, run 'aigogo login %s'", registry)
	}

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("repository not found: %s/%s", registry, repository)
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list tags: %s (status %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Tags, nil
}

// DeleteAll deletes all tags in a repository
func (d *Deleter) DeleteAll(registry, repository string) error {
	// List all tags
	tags, err := d.listTags(registry, repository)
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	if len(tags) == 0 {
		return fmt.Errorf("no tags found in repository %s/%s", registry, repository)
	}

	fmt.Printf("Found %d tag(s):\n", len(tags))
	for _, tag := range tags {
		fmt.Printf("  - %s\n", tag)
	}
	fmt.Println()

	// Delete each tag
	failed := []string{}
	succeeded := 0

	for _, tag := range tags {
		fullRef := fmt.Sprintf("%s/%s:%s", registry, repository, tag)
		fmt.Printf("Deleting %s... ", tag)

		err := d.Delete(fullRef)
		if err != nil {
			fmt.Printf("✗ Failed: %v\n", err)
			failed = append(failed, tag)
		} else {
			fmt.Printf("✓ Deleted\n")
			succeeded++
		}
	}

	fmt.Println()

	if len(failed) > 0 {
		fmt.Printf("⚠️  Warning: %d tag(s) failed to delete: %v\n", len(failed), failed)
		fmt.Printf("Successfully deleted %d out of %d tags\n", succeeded, len(tags))
		return fmt.Errorf("failed to delete %d tag(s)", len(failed))
	}

	fmt.Printf("✓ Successfully deleted all %d tag(s) from %s/%s\n", succeeded, registry, repository)
	return nil
}
