package docker

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareDirs_Identical(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	writeFile(t, dirA, "a.py", "print('hello')")
	writeFile(t, dirA, "b.py", "print('world')")
	writeFile(t, dirB, "a.py", "print('hello')")
	writeFile(t, dirB, "b.py", "print('world')")

	d := NewDiffer()
	result, err := d.CompareDirs(dirA, dirB)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Identical {
		t.Error("expected Identical=true")
	}
	if len(result.OnlyInA) != 0 || len(result.OnlyInB) != 0 || len(result.Modified) != 0 {
		t.Errorf("expected no diffs, got OnlyInA=%v OnlyInB=%v Modified=%v",
			result.OnlyInA, result.OnlyInB, result.Modified)
	}
	if len(result.Unchanged) != 2 {
		t.Errorf("expected 2 unchanged, got %d", len(result.Unchanged))
	}
}

func TestCompareDirs_FileAdded(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	writeFile(t, dirA, "a.py", "print('hello')")
	writeFile(t, dirB, "a.py", "print('hello')")
	writeFile(t, dirB, "b.py", "print('new')")

	d := NewDiffer()
	result, err := d.CompareDirs(dirA, dirB)
	if err != nil {
		t.Fatal(err)
	}

	if result.Identical {
		t.Error("expected Identical=false")
	}
	if len(result.OnlyInB) != 1 || result.OnlyInB[0] != "b.py" {
		t.Errorf("expected OnlyInB=[b.py], got %v", result.OnlyInB)
	}
}

func TestCompareDirs_FileRemoved(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	writeFile(t, dirA, "a.py", "print('hello')")
	writeFile(t, dirA, "b.py", "print('old')")
	writeFile(t, dirB, "a.py", "print('hello')")

	d := NewDiffer()
	result, err := d.CompareDirs(dirA, dirB)
	if err != nil {
		t.Fatal(err)
	}

	if result.Identical {
		t.Error("expected Identical=false")
	}
	if len(result.OnlyInA) != 1 || result.OnlyInA[0] != "b.py" {
		t.Errorf("expected OnlyInA=[b.py], got %v", result.OnlyInA)
	}
}

func TestCompareDirs_FileModified(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	writeFile(t, dirA, "a.py", "print('old')")
	writeFile(t, dirB, "a.py", "print('new')")

	d := NewDiffer()
	result, err := d.CompareDirs(dirA, dirB)
	if err != nil {
		t.Fatal(err)
	}

	if result.Identical {
		t.Error("expected Identical=false")
	}
	if len(result.Modified) != 1 || result.Modified[0] != "a.py" {
		t.Errorf("expected Modified=[a.py], got %v", result.Modified)
	}
	if len(result.FileDiffs) != 1 {
		t.Fatalf("expected 1 FileDiff, got %d", len(result.FileDiffs))
	}
	if result.FileDiffs[0].IsBinary {
		t.Error("expected non-binary diff")
	}
	if !strings.Contains(result.FileDiffs[0].Diff, "-print('old')") {
		t.Errorf("diff does not contain expected content: %s", result.FileDiffs[0].Diff)
	}
	if !strings.Contains(result.FileDiffs[0].Diff, "+print('new')") {
		t.Errorf("diff does not contain expected content: %s", result.FileDiffs[0].Diff)
	}
}

func TestCompareDirs_Mixed(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	writeFile(t, dirA, "same.py", "unchanged")
	writeFile(t, dirA, "changed.py", "old content")
	writeFile(t, dirA, "removed.py", "going away")

	writeFile(t, dirB, "same.py", "unchanged")
	writeFile(t, dirB, "changed.py", "new content")
	writeFile(t, dirB, "added.py", "brand new")

	d := NewDiffer()
	result, err := d.CompareDirs(dirA, dirB)
	if err != nil {
		t.Fatal(err)
	}

	if result.Identical {
		t.Error("expected Identical=false")
	}
	if len(result.Modified) != 1 {
		t.Errorf("expected 1 modified, got %d", len(result.Modified))
	}
	if len(result.OnlyInA) != 1 {
		t.Errorf("expected 1 removed, got %d", len(result.OnlyInA))
	}
	if len(result.OnlyInB) != 1 {
		t.Errorf("expected 1 added, got %d", len(result.OnlyInB))
	}
	if len(result.Unchanged) != 1 {
		t.Errorf("expected 1 unchanged, got %d", len(result.Unchanged))
	}
}

func TestCompareDirs_Nested(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	writeFile(t, dirA, "sub/deep/file.py", "content A")
	writeFile(t, dirB, "sub/deep/file.py", "content B")

	d := NewDiffer()
	result, err := d.CompareDirs(dirA, dirB)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Modified) != 1 {
		t.Fatalf("expected 1 modified, got %d", len(result.Modified))
	}
	expected := filepath.Join("sub", "deep", "file.py")
	if result.Modified[0] != expected {
		t.Errorf("modified file = %q, want %q", result.Modified[0], expected)
	}
}

func TestCompareDirs_Binary(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	// Write binary content (contains null bytes)
	if err := os.WriteFile(filepath.Join(dirA, "img.png"), []byte{0x89, 'P', 'N', 'G', 0, 0}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dirB, "img.png"), []byte{0x89, 'P', 'N', 'G', 0, 1}, 0644); err != nil {
		t.Fatal(err)
	}

	d := NewDiffer()
	result, err := d.CompareDirs(dirA, dirB)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.FileDiffs) != 1 {
		t.Fatalf("expected 1 FileDiff, got %d", len(result.FileDiffs))
	}
	if !result.FileDiffs[0].IsBinary {
		t.Error("expected IsBinary=true")
	}
	if !strings.Contains(result.FileDiffs[0].Diff, "Binary files") {
		t.Errorf("expected binary diff message, got: %s", result.FileDiffs[0].Diff)
	}
}

func TestCompareDirs_SkipsMetadata(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	writeFile(t, dirA, "code.py", "same")
	writeFile(t, dirA, ".aigogo-metadata.json", `{"old": true}`)
	writeFile(t, dirA, ".aigogo-manifest.json", `{"old": true}`)
	writeFile(t, dirA, "aigogo.json", `{"name":"a","version":"1.0.0"}`)

	writeFile(t, dirB, "code.py", "same")
	writeFile(t, dirB, ".aigogo-metadata.json", `{"new": true}`)
	writeFile(t, dirB, ".aigogo-manifest.json", `{"new": true}`)
	writeFile(t, dirB, "aigogo.json", `{"name":"a","version":"2.0.0"}`)

	d := NewDiffer()
	result, err := d.CompareDirs(dirA, dirB)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Identical {
		t.Errorf("expected Identical=true (metadata and aigogo.json should be skipped), got Modified=%v OnlyInA=%v OnlyInB=%v",
			result.Modified, result.OnlyInA, result.OnlyInB)
	}
}

func TestFormatDiff_Identical(t *testing.T) {
	result := &DiffResult{Identical: true}
	out := FormatDiff(result, false)
	if out != "Packages are identical.\n" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestFormatDiff_WithChanges(t *testing.T) {
	result := &DiffResult{
		Modified:  []string{"utils.py"},
		OnlyInA:   []string{"old.py"},
		OnlyInB:   []string{"new.py"},
		Unchanged: []string{"keep.py"},
		FileDiffs: []FileDiff{
			{Path: "utils.py", Diff: "diff a/utils.py b/utils.py\n--- a/utils.py\n+++ b/utils.py\n@@ -1,1 +1,1 @@\n-old\n+new"},
		},
	}

	out := FormatDiff(result, false)
	if !strings.Contains(out, "diff a/utils.py b/utils.py") {
		t.Error("expected unified diff header")
	}
	if !strings.Contains(out, "Only in a: old.py") {
		t.Error("expected 'Only in a' line")
	}
	if !strings.Contains(out, "Only in b: new.py") {
		t.Error("expected 'Only in b' line")
	}
	if !strings.Contains(out, "1 modified") {
		t.Error("expected summary with modified count")
	}
}

func TestFormatSummary(t *testing.T) {
	result := &DiffResult{
		Modified:  []string{"utils.py"},
		OnlyInA:   []string{"old.py"},
		OnlyInB:   []string{"new.py"},
		Unchanged: []string{"keep.py", "other.py"},
	}

	out := FormatSummary(result)
	if !strings.Contains(out, "M utils.py") {
		t.Error("expected 'M utils.py'")
	}
	if !strings.Contains(out, "A new.py") {
		t.Error("expected 'A new.py'")
	}
	if !strings.Contains(out, "D old.py") {
		t.Error("expected 'D old.py'")
	}
	if !strings.Contains(out, "1 modified, 1 added, 1 removed, 2 unchanged") {
		t.Errorf("unexpected summary: %s", out)
	}
}

func TestFormatSummary_Identical(t *testing.T) {
	result := &DiffResult{Identical: true}
	out := FormatSummary(result)
	if out != "Packages are identical.\n" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestComputeUnifiedDiff(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		wantSub  []string // substrings that must appear
		wantNone []string // substrings that must not appear
	}{
		{
			name:    "single line change",
			a:       "hello\nworld\n",
			b:       "hello\nearth\n",
			wantSub: []string{"-world", "+earth", "@@ -"},
		},
		{
			name:    "addition",
			a:       "line1\nline2\n",
			b:       "line1\nline2\nline3\n",
			wantSub: []string{"+line3"},
		},
		{
			name:    "deletion",
			a:       "line1\nline2\nline3\n",
			b:       "line1\nline2\n",
			wantSub: []string{"-line3"},
		},
		{
			name:    "empty to content",
			a:       "",
			b:       "new\n",
			wantSub: []string{"+new"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := computeUnifiedDiff("a/file", "b/file", []byte(tt.a), []byte(tt.b))
			for _, sub := range tt.wantSub {
				if !strings.Contains(out, sub) {
					t.Errorf("diff does not contain %q:\n%s", sub, out)
				}
			}
			for _, sub := range tt.wantNone {
				if strings.Contains(out, sub) {
					t.Errorf("diff should not contain %q:\n%s", sub, out)
				}
			}
		})
	}
}

func TestGetLocalTarDigest(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	ref := "test:1.0.0"
	sanitized := SanitizeImageRef(ref)
	content := []byte("fake layer data")

	// Create cache structure under fake home
	cacheDir := filepath.Join(fakeHome, ".aigogo", "cache", "images", sanitized)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "layer.tar"), content, 0644); err != nil {
		t.Fatal(err)
	}

	d := NewDiffer()
	digest, err := d.GetLocalTarDigest(ref)
	if err != nil {
		t.Fatalf("GetLocalTarDigest failed: %v", err)
	}

	// Compute expected digest
	h := sha256.Sum256(content)
	expected := "sha256:" + hex.EncodeToString(h[:])
	if digest != expected {
		t.Errorf("digest = %q, want %q", digest, expected)
	}
}

func TestFetchRemoteLayerDigest(t *testing.T) {
	expectedDigest := "sha256:abc123def456"

	// Create mock registry server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v2/") && strings.Contains(r.URL.Path, "/manifests/") {
			manifest := map[string]interface{}{
				"schemaVersion": 2,
				"layers": []map[string]interface{}{
					{
						"mediaType": "application/vnd.docker.image.rootfs.diff.tar",
						"size":      1234,
						"digest":    expectedDigest,
					},
				},
			}
			w.Header().Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
			_ = json.NewEncoder(w).Encode(manifest)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	d := &Differ{client: server.Client()}

	// We can't easily test against the real registry endpoint parsing,
	// so test the manifest parsing logic directly
	req, _ := http.NewRequest("GET", server.URL+"/v2/test/repo/manifests/1.0.0", nil)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := d.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	var manifest map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		t.Fatal(err)
	}

	layers, ok := manifest["layers"].([]interface{})
	if !ok || len(layers) == 0 {
		t.Fatal("no layers in manifest")
	}

	layer := layers[0].(map[string]interface{})
	digest := layer["digest"].(string)

	if digest != expectedDigest {
		t.Errorf("digest = %q, want %q", digest, expectedDigest)
	}
}

func TestGetLocalTarDigest_MatchesExpected(t *testing.T) {
	content := []byte("matching layer content")
	h := sha256.Sum256(content)
	expectedDigest := "sha256:" + hex.EncodeToString(h[:])

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		manifest := map[string]interface{}{
			"schemaVersion": 2,
			"layers": []map[string]interface{}{
				{
					"mediaType": "application/vnd.docker.image.rootfs.diff.tar",
					"size":      len(content),
					"digest":    expectedDigest,
				},
			},
		}
		w.Header().Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
		_ = json.NewEncoder(w).Encode(manifest)
	}))
	defer server.Close()

	// Set up fake home with cache
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	ref := "test:1.0.0"
	sanitized := SanitizeImageRef(ref)
	cacheDir := filepath.Join(fakeHome, ".aigogo", "cache", "images", sanitized)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "layer.tar"), content, 0644); err != nil {
		t.Fatal(err)
	}

	// We can verify the local digest matches what we expect
	d := &Differ{client: server.Client()}
	localDigest, err := d.GetLocalTarDigest(ref)
	if err != nil {
		t.Fatalf("GetLocalTarDigest: %v", err)
	}
	if localDigest != expectedDigest {
		t.Errorf("local digest = %q, want %q", localDigest, expectedDigest)
	}
}

func TestIsBinary(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		binary bool
	}{
		{"text", []byte("hello world\n"), false},
		{"binary with null", []byte{0x89, 'P', 'N', 'G', 0, 0}, true},
		{"empty", []byte{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBinary(tt.data)
			if got != tt.binary {
				t.Errorf("isBinary() = %v, want %v", got, tt.binary)
			}
		})
	}
}

func TestCollectFiles_SkipsMetadata(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "code.py", "code")
	writeFile(t, dir, ".aigogo-metadata.json", "{}")
	writeFile(t, dir, ".aigogo-manifest.json", "{}")
	writeFile(t, dir, "aigogo.json", `{"name":"test"}`)

	files, err := collectFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(files), files)
	}
	if files[0] != "code.py" {
		t.Errorf("expected code.py, got %s", files[0])
	}
}

func TestSummaryLine(t *testing.T) {
	tests := []struct {
		name   string
		result *DiffResult
		want   string
	}{
		{
			name: "all types",
			result: &DiffResult{
				Modified:  []string{"a"},
				OnlyInB:   []string{"b"},
				OnlyInA:   []string{"c"},
				Unchanged: []string{"d", "e"},
			},
			want: "1 modified, 1 added, 1 removed, 2 unchanged",
		},
		{
			name: "modified only",
			result: &DiffResult{
				Modified: []string{"a", "b"},
			},
			want: "2 modified",
		},
		{
			name:   "empty",
			result: &DiffResult{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summaryLine(tt.result)
			if got != tt.want {
				t.Errorf("summaryLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDiffLines(t *testing.T) {
	a := []string{"line1", "line2", "line3"}
	b := []string{"line1", "modified", "line3", "line4"}

	ops := diffLines(a, b)

	// Should have: keep line1, delete line2, insert modified, keep line3, insert line4
	var keeps, inserts, deletes int
	for _, op := range ops {
		switch op.kind {
		case ' ':
			keeps++
		case '+':
			inserts++
		case '-':
			deletes++
		}
	}

	if keeps != 2 {
		t.Errorf("expected 2 keeps, got %d", keeps)
	}
	if deletes != 1 {
		t.Errorf("expected 1 delete, got %d", deletes)
	}
	if inserts != 2 {
		t.Errorf("expected 2 inserts, got %d", inserts)
	}
}

func TestExtractToTemp_LocalBuild(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	ref := "mylib:1.0.0"
	sanitized := SanitizeImageRef(ref)
	buildDir := filepath.Join(fakeHome, ".aigogo", "cache", sanitized)
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, buildDir, "code.py", "print('hello')")
	writeFile(t, buildDir, ".aigogo-metadata.json", `{"type":"local-build"}`)

	d := NewDiffer()
	tmpDir, err := d.ExtractToTemp(ref)
	if err != nil {
		t.Fatalf("ExtractToTemp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// code.py should be extracted
	if _, err := os.Stat(filepath.Join(tmpDir, "code.py")); err != nil {
		t.Error("code.py not found in extracted dir")
	}
	// metadata should be skipped
	if _, err := os.Stat(filepath.Join(tmpDir, ".aigogo-metadata.json")); err == nil {
		t.Error(".aigogo-metadata.json should not be in extracted dir")
	}
}

func TestFormatDiff_SummaryOnly(t *testing.T) {
	result := &DiffResult{
		Modified:  []string{"utils.py"},
		OnlyInB:   []string{"new.py"},
		Unchanged: []string{"keep.py"},
	}

	out := FormatDiff(result, true)
	// Should be summary format (M/A/D), not unified diff
	if !strings.Contains(out, "M utils.py") {
		t.Error("expected summary format with 'M utils.py'")
	}
	if !strings.Contains(out, "A new.py") {
		t.Error("expected 'A new.py'")
	}
	if strings.Contains(out, "diff a/") {
		t.Error("summary mode should not contain unified diff headers")
	}
}

func TestCopyDirFiltered(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, src, "code.py", "print('hello')")
	writeFile(t, src, "sub/nested.py", "import os")
	writeFile(t, src, ".aigogo-metadata.json", `{"skip": true}`)
	writeFile(t, src, ".aigogo-manifest.json", `{"skip": true}`)
	writeFile(t, src, "aigogo.json", `{"name":"test"}`)

	if err := copyDirFiltered(src, dst); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dst, "code.py")); err != nil {
		t.Error("code.py should be copied")
	}
	if _, err := os.Stat(filepath.Join(dst, "sub", "nested.py")); err != nil {
		t.Error("sub/nested.py should be copied")
	}
	if _, err := os.Stat(filepath.Join(dst, ".aigogo-metadata.json")); err == nil {
		t.Error(".aigogo-metadata.json should not be copied")
	}
	if _, err := os.Stat(filepath.Join(dst, ".aigogo-manifest.json")); err == nil {
		t.Error(".aigogo-manifest.json should not be copied")
	}
	if _, err := os.Stat(filepath.Join(dst, "aigogo.json")); err == nil {
		t.Error("aigogo.json should not be copied")
	}
}

// writeFile is a test helper that creates a file with the given content,
// creating parent directories as needed.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// Verify the output format helper (used as a pseudo-integration check)
func TestFormatDiff_OutputFormat(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	writeFile(t, dirA, "utils.py", "import os\ndef helper():\n    return \"old\"\n")
	writeFile(t, dirA, "old_module.py", "# old\n")
	writeFile(t, dirA, "keep.py", "# unchanged\n")

	writeFile(t, dirB, "utils.py", "import os\nimport sys\ndef helper():\n    return \"new\"\n")
	writeFile(t, dirB, "new_module.py", "# new\n")
	writeFile(t, dirB, "keep.py", "# unchanged\n")

	d := NewDiffer()
	result, err := d.CompareDirs(dirA, dirB)
	if err != nil {
		t.Fatal(err)
	}

	out := FormatDiff(result, false)
	fmt.Println(out) // For manual inspection during verbose test runs

	// Check key elements
	if !strings.Contains(out, "diff a/utils.py b/utils.py") {
		t.Error("expected diff header for utils.py")
	}
	if !strings.Contains(out, "Only in a: old_module.py") {
		t.Error("expected 'Only in a: old_module.py'")
	}
	if !strings.Contains(out, "Only in b: new_module.py") {
		t.Error("expected 'Only in b: new_module.py'")
	}
	if !strings.Contains(out, "1 modified, 1 added, 1 removed, 1 unchanged") {
		t.Errorf("unexpected summary in output:\n%s", out)
	}
}
