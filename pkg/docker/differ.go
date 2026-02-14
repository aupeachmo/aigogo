package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aupeachmo/aigogo/pkg/auth"
)

// DiffResult holds the result of comparing two package versions
type DiffResult struct {
	Identical bool
	OnlyInA   []string   // files only in left side
	OnlyInB   []string   // files only in right side
	Modified  []string   // files changed
	Unchanged []string   // files identical
	FileDiffs []FileDiff // unified diffs for modified files
}

// FileDiff holds the unified diff for a single modified file
type FileDiff struct {
	Path     string
	Diff     string // unified diff text
	IsBinary bool
}

// Differ compares package versions
type Differ struct {
	client *http.Client
}

// NewDiffer creates a new Differ
func NewDiffer() *Differ {
	return &Differ{
		client: &http.Client{},
	}
}

// CompareDirs walks both directories, computes file sets, and generates unified diffs for modified files
func (d *Differ) CompareDirs(dirA, dirB string) (*DiffResult, error) {
	filesA, err := collectFiles(dirA)
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory A: %w", err)
	}
	filesB, err := collectFiles(dirB)
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory B: %w", err)
	}

	setA := make(map[string]bool, len(filesA))
	for _, f := range filesA {
		setA[f] = true
	}
	setB := make(map[string]bool, len(filesB))
	for _, f := range filesB {
		setB[f] = true
	}

	result := &DiffResult{}

	// Files only in A
	for _, f := range filesA {
		if !setB[f] {
			result.OnlyInA = append(result.OnlyInA, f)
		}
	}

	// Files only in B
	for _, f := range filesB {
		if !setA[f] {
			result.OnlyInB = append(result.OnlyInB, f)
		}
	}

	// Files in both — check content
	for _, f := range filesA {
		if !setB[f] {
			continue
		}
		contentA, errA := os.ReadFile(filepath.Join(dirA, f))
		if errA != nil {
			return nil, fmt.Errorf("failed to read %s from A: %w", f, errA)
		}
		contentB, errB := os.ReadFile(filepath.Join(dirB, f))
		if errB != nil {
			return nil, fmt.Errorf("failed to read %s from B: %w", f, errB)
		}

		if bytes.Equal(contentA, contentB) {
			result.Unchanged = append(result.Unchanged, f)
		} else {
			result.Modified = append(result.Modified, f)
			fd := FileDiff{Path: f}
			if isBinary(contentA) || isBinary(contentB) {
				fd.IsBinary = true
				fd.Diff = fmt.Sprintf("Binary files a/%s and b/%s differ", f, f)
			} else {
				fd.Diff = computeUnifiedDiff("a/"+f, "b/"+f, contentA, contentB)
			}
			result.FileDiffs = append(result.FileDiffs, fd)
		}
	}

	sort.Strings(result.OnlyInA)
	sort.Strings(result.OnlyInB)
	sort.Strings(result.Modified)
	sort.Strings(result.Unchanged)

	result.Identical = len(result.OnlyInA) == 0 && len(result.OnlyInB) == 0 && len(result.Modified) == 0

	return result, nil
}

// CompareWithRemote compares a local ref with a remote ref.
// It first tries a cheap digest comparison. If that shows they differ (or can't be determined),
// it falls back to content-level comparison.
func (d *Differ) CompareWithRemote(localRef, remoteRef string) (*DiffResult, error) {
	// Try cheap check first
	same, err := d.CheckSameAsRemote(remoteRef)
	if err == nil && same {
		return &DiffResult{Identical: true}, nil
	}

	// Fall back to content-level comparison
	localDir, err := d.ExtractToTemp(localRef)
	if err != nil {
		return nil, fmt.Errorf("failed to extract local ref: %w", err)
	}
	defer func() { _ = os.RemoveAll(localDir) }()

	remoteDir, err := d.extractRemoteToTemp(remoteRef)
	if err != nil {
		return nil, fmt.Errorf("failed to extract remote ref: %w", err)
	}
	defer func() { _ = os.RemoveAll(remoteDir) }()

	return d.CompareDirs(localDir, remoteDir)
}

// CheckSameAsRemote does a cheap sameness check by comparing tar layer digests.
// The local tar is looked up under remoteRef because BuildImageFromPath stores
// the layer.tar keyed by the registry ref (the push destination), not the local build ref.
// Returns (true, nil) if identical, (false, nil) if different, or (false, err) on error.
func (d *Differ) CheckSameAsRemote(remoteRef string) (bool, error) {
	localDigest, err := d.GetLocalTarDigest(remoteRef)
	if err != nil {
		return false, fmt.Errorf("failed to get local tar digest: %w", err)
	}

	remoteDigest, err := d.FetchRemoteLayerDigest(remoteRef)
	if err != nil {
		return false, fmt.Errorf("failed to fetch remote layer digest: %w", err)
	}

	return localDigest == remoteDigest, nil
}

// FetchRemoteLayerDigest gets the layer digest from a remote registry manifest
func (d *Differ) FetchRemoteLayerDigest(registryRef string) (string, error) {
	registry, repository, tag, err := parseImageRef(registryRef)
	if err != nil {
		return "", err
	}

	apiEndpoint := getRegistryAPIEndpoint(registry)
	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", apiEndpoint, repository, tag)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	// Try with auth first
	authManager := auth.NewManager()
	token, err := authManager.GetToken(registry, repository)
	if err != nil {
		token = ""
	}
	setAuthHeader(req, registry, token)

	resp, err := d.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("remote image not found: %s", registryRef)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get manifest: %s - %s", resp.Status, string(body))
	}

	var manifest map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return "", fmt.Errorf("failed to decode manifest: %w", err)
	}

	layers, ok := manifest["layers"].([]interface{})
	if !ok || len(layers) == 0 {
		return "", fmt.Errorf("no layers found in manifest")
	}

	layer, ok := layers[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid layer format in manifest")
	}

	digest, ok := layer["digest"].(string)
	if !ok {
		return "", fmt.Errorf("no digest in layer")
	}

	return digest, nil
}

// GetLocalTarDigest reads the local layer.tar and returns its digest.
// The ref is typically a registry ref, since BuildImageFromPath stores layer.tar
// under the registry ref (push destination).
func (d *Differ) GetLocalTarDigest(ref string) (string, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return "", err
	}

	sanitized := sanitizeImageRef(ref)

	// Check images/ subdirectory (registry push cache)
	tarPath := filepath.Join(cacheDir, "images", sanitized, "layer.tar")
	if _, err := os.Stat(tarPath); err != nil {
		return "", fmt.Errorf("no layer.tar for %s: %w", ref, err)
	}

	data, err := os.ReadFile(tarPath)
	if err != nil {
		return "", fmt.Errorf("failed to read layer.tar: %w", err)
	}

	return CalculateDigest(data), nil
}

// ExtractToTemp extracts a package (local build or pulled image) to a temp directory.
// Returns the path to the temp directory containing the package files.
func (d *Differ) ExtractToTemp(ref string) (string, error) {
	cachePath := GetCachePath(ref)
	if cachePath == "" {
		return "", fmt.Errorf("package not found in cache: %s", ref)
	}

	tmpDir, err := os.MkdirTemp("", "aigogo-diff-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Check if it's a local build (has .aigogo-metadata.json) or pulled image (has layer.tar)
	if _, err := os.Stat(filepath.Join(cachePath, ".aigogo-metadata.json")); err == nil {
		// Local build — copy files directly (skip metadata)
		if err := copyDirFiltered(cachePath, tmpDir); err != nil {
			_ = os.RemoveAll(tmpDir)
			return "", fmt.Errorf("failed to copy local build: %w", err)
		}
		return tmpDir, nil
	}

	// Pulled image — extract from tar
	extractor := NewExtractor()
	if _, err := extractor.Extract(ref, tmpDir, true); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to extract: %w", err)
	}

	return tmpDir, nil
}

// extractRemoteToTemp pulls a remote image and extracts it to a temp directory.
func (d *Differ) extractRemoteToTemp(remoteRef string) (string, error) {
	puller := NewPuller()
	if err := puller.Pull(remoteRef); err != nil {
		return "", fmt.Errorf("failed to pull remote: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "aigogo-diff-remote-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	extractor := NewExtractor()
	if _, err := extractor.Extract(remoteRef, tmpDir, true); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to extract remote: %w", err)
	}

	return tmpDir, nil
}

// FormatDiff produces git-style unified diff output
func FormatDiff(result *DiffResult, summaryOnly bool) string {
	if result.Identical {
		return "Packages are identical.\n"
	}

	if summaryOnly {
		return FormatSummary(result)
	}

	var buf strings.Builder

	// Unified diffs for modified files
	for _, fd := range result.FileDiffs {
		buf.WriteString(fd.Diff)
		buf.WriteString("\n")
	}

	// Files only in A (deleted)
	for _, f := range result.OnlyInA {
		fmt.Fprintf(&buf, "Only in a: %s\n", f)
	}

	// Files only in B (added)
	for _, f := range result.OnlyInB {
		fmt.Fprintf(&buf, "Only in b: %s\n", f)
	}

	// Summary line
	buf.WriteString("\n")
	buf.WriteString(summaryLine(result))
	buf.WriteString("\n")

	return buf.String()
}

// FormatSummary produces compact M/A/D listing with counts
func FormatSummary(result *DiffResult) string {
	if result.Identical {
		return "Packages are identical.\n"
	}

	var buf strings.Builder

	for _, f := range result.Modified {
		fmt.Fprintf(&buf, "M %s\n", f)
	}
	for _, f := range result.OnlyInB {
		fmt.Fprintf(&buf, "A %s\n", f)
	}
	for _, f := range result.OnlyInA {
		fmt.Fprintf(&buf, "D %s\n", f)
	}

	buf.WriteString(summaryLine(result))
	buf.WriteString("\n")

	return buf.String()
}

func summaryLine(result *DiffResult) string {
	parts := []string{}
	if n := len(result.Modified); n > 0 {
		parts = append(parts, fmt.Sprintf("%d modified", n))
	}
	if n := len(result.OnlyInB); n > 0 {
		parts = append(parts, fmt.Sprintf("%d added", n))
	}
	if n := len(result.OnlyInA); n > 0 {
		parts = append(parts, fmt.Sprintf("%d removed", n))
	}
	if n := len(result.Unchanged); n > 0 {
		parts = append(parts, fmt.Sprintf("%d unchanged", n))
	}
	return strings.Join(parts, ", ")
}

// collectFiles walks a directory and returns sorted relative file paths,
// skipping metadata files
func collectFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		// Skip aigogo internal metadata and manifest files
		base := filepath.Base(rel)
		if base == ".aigogo-metadata.json" || base == ".aigogo-manifest.json" || base == "aigogo.json" {
			return nil
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// isBinary detects if content is binary by checking for null bytes in the first 8KB
func isBinary(data []byte) bool {
	check := data
	if len(check) > 8192 {
		check = check[:8192]
	}
	return bytes.ContainsRune(check, 0)
}

// computeUnifiedDiff generates a unified diff with 3 lines of context
func computeUnifiedDiff(nameA, nameB string, a, b []byte) string {
	linesA := splitLines(string(a))
	linesB := splitLines(string(b))

	// Compute LCS using O(NM) DP approach
	edits := diffLines(linesA, linesB)

	// Generate unified diff hunks with 3 lines of context
	hunks := buildHunks(edits, 3)

	if len(hunks) == 0 {
		return ""
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "diff %s %s\n", nameA, nameB)
	fmt.Fprintf(&buf, "--- %s\n", nameA)
	fmt.Fprintf(&buf, "+++ %s\n", nameB)

	for _, h := range hunks {
		fmt.Fprintf(&buf, "@@ -%d,%d +%d,%d @@\n", h.startA+1, h.countA, h.startB+1, h.countB)
		for _, line := range h.lines {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

// editOp represents a line-level edit operation
type editOp struct {
	kind byte   // ' ' = keep, '+' = insert, '-' = delete
	line string // the line content
	idxA int    // index in A (-1 if insert)
	idxB int    // index in B (-1 if delete)
}

// hunk represents a unified diff hunk
type hunk struct {
	startA int
	countA int
	startB int
	countB int
	lines  []string
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	// Remove trailing empty string from final newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// diffLines computes a minimal edit script between two line slices
func diffLines(a, b []string) []editOp {
	n, m := len(a), len(b)

	// Build LCS table
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// Backtrack to build edit ops
	var ops []editOp
	i, j := n, m
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && a[i-1] == b[j-1] {
			ops = append(ops, editOp{kind: ' ', line: a[i-1], idxA: i - 1, idxB: j - 1})
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			ops = append(ops, editOp{kind: '+', line: b[j-1], idxA: -1, idxB: j - 1})
			j--
		} else {
			ops = append(ops, editOp{kind: '-', line: a[i-1], idxA: i - 1, idxB: -1})
			i--
		}
	}

	// Reverse to get forward order
	for l, r := 0, len(ops)-1; l < r; l, r = l+1, r-1 {
		ops[l], ops[r] = ops[r], ops[l]
	}

	return ops
}

// buildHunks groups edit ops into hunks with the given context size
func buildHunks(ops []editOp, context int) []hunk {
	if len(ops) == 0 {
		return nil
	}

	// Find indices of changed lines
	var changeIndices []int
	for i, op := range ops {
		if op.kind != ' ' {
			changeIndices = append(changeIndices, i)
		}
	}

	if len(changeIndices) == 0 {
		return nil
	}

	// Group changes that are within 2*context of each other
	var hunks []hunk
	groupStart := changeIndices[0]
	groupEnd := changeIndices[0]

	flush := func() {
		// Expand with context
		start := groupStart - context
		if start < 0 {
			start = 0
		}
		end := groupEnd + context
		if end >= len(ops) {
			end = len(ops) - 1
		}

		h := hunk{}

		// Calculate start positions in A and B
		aPos := 0
		bPos := 0
		for i := 0; i < start; i++ {
			switch ops[i].kind {
			case ' ':
				aPos++
				bPos++
			case '-':
				aPos++
			case '+':
				bPos++
			}
		}
		h.startA = aPos
		h.startB = bPos

		// Build hunk lines and count
		for i := start; i <= end; i++ {
			op := ops[i]
			prefix := string(op.kind)
			h.lines = append(h.lines, prefix+op.line)
			switch op.kind {
			case ' ':
				h.countA++
				h.countB++
			case '-':
				h.countA++
			case '+':
				h.countB++
			}
		}

		hunks = append(hunks, h)
	}

	for i := 1; i < len(changeIndices); i++ {
		if changeIndices[i]-groupEnd > 2*context {
			flush()
			groupStart = changeIndices[i]
		}
		groupEnd = changeIndices[i]
	}
	flush()

	return hunks
}

// copyDirFiltered copies directory contents, skipping metadata files
func copyDirFiltered(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		// Skip metadata and manifest
		base := filepath.Base(rel)
		if base == ".aigogo-metadata.json" || base == ".aigogo-manifest.json" || base == "aigogo.json" {
			return nil
		}

		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
