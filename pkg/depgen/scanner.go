package depgen

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Scanner scans source files for imports
type Scanner struct{}

// NewScanner creates a new scanner
func NewScanner() *Scanner {
	return &Scanner{}
}

// ScanFiles scans multiple files for imports
func (s *Scanner) ScanFiles(files []string, language string) ([]ImportInfo, error) {
	var allImports []ImportInfo
	seen := make(map[string]bool)

	for _, file := range files {
		imports, err := s.scanFile(file, language)
		if err != nil {
			return nil, err
		}

		// Deduplicate
		for _, imp := range imports {
			if !seen[imp.Package] {
				seen[imp.Package] = true
				allImports = append(allImports, imp)
			}
		}
	}

	return allImports, nil
}

func (s *Scanner) scanFile(filename, language string) ([]ImportInfo, error) {
	ext := filepath.Ext(filename)

	// Determine scanner based on language or extension
	switch language {
	case "python":
		return s.scanPython(filename)
	case "javascript":
		return s.scanJavaScript(filename)
	case "go":
		return s.scanGo(filename)
	case "rust":
		return s.scanRust(filename)
	default:
		// Fallback to extension-based detection
		switch ext {
		case ".py":
			return s.scanPython(filename)
		case ".js", ".ts", ".jsx", ".tsx", ".mjs", ".cjs":
			return s.scanJavaScript(filename)
		case ".go":
			return s.scanGo(filename)
		case ".rs":
			return s.scanRust(filename)
		}
	}

	return []ImportInfo{}, nil
}

func (s *Scanner) scanPython(filename string) ([]ImportInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var imports []ImportInfo
	scanner := bufio.NewScanner(file)
	lineNum := 0

	stdlib := pythonStdlib()

	importRegex := regexp.MustCompile(`^\s*import\s+([a-zA-Z0-9_\.]+)`)
	fromImportRegex := regexp.MustCompile(`^\s*from\s+([a-zA-Z0-9_\.]+)\s+import`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		if matches := importRegex.FindStringSubmatch(line); matches != nil {
			pkg := strings.Split(matches[1], ".")[0]
			if !stdlib[pkg] && !strings.HasPrefix(matches[1], ".") {
				imports = append(imports, ImportInfo{
					Package:    pkg,
					SourceFile: filename,
					LineNumber: lineNum,
				})
			}
		}

		if matches := fromImportRegex.FindStringSubmatch(line); matches != nil {
			pkg := strings.Split(matches[1], ".")[0]
			if !stdlib[pkg] && !strings.HasPrefix(matches[1], ".") {
				imports = append(imports, ImportInfo{
					Package:    pkg,
					SourceFile: filename,
					LineNumber: lineNum,
				})
			}
		}
	}

	return imports, scanner.Err()
}

func (s *Scanner) scanJavaScript(filename string) ([]ImportInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var imports []ImportInfo
	scanner := bufio.NewScanner(file)
	lineNum := 0

	builtins := nodeBuiltins()

	importRegex := regexp.MustCompile(`^\s*import\s+.*?from\s+['"]([^'"]+)['"]`)
	requireRegex := regexp.MustCompile(`require\(['"]([^'"]+)['"]\)`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if matches := importRegex.FindStringSubmatch(line); matches != nil {
			pkg := matches[1]
			if isExternalJSPackage(pkg, builtins) {
				imports = append(imports, ImportInfo{
					Package:    pkg,
					SourceFile: filename,
					LineNumber: lineNum,
				})
			}
		}

		if matches := requireRegex.FindStringSubmatch(line); matches != nil {
			pkg := matches[1]
			if isExternalJSPackage(pkg, builtins) {
				imports = append(imports, ImportInfo{
					Package:    pkg,
					SourceFile: filename,
					LineNumber: lineNum,
				})
			}
		}
	}

	return imports, scanner.Err()
}

func isExternalJSPackage(pkg string, builtins map[string]bool) bool {
	if strings.HasPrefix(pkg, ".") || strings.HasPrefix(pkg, "/") {
		return false
	}
	// Strip node: prefix (e.g. "node:fs" → "fs")
	name := strings.TrimPrefix(pkg, "node:")
	return !builtins[name]
}

func (s *Scanner) scanGo(filename string) ([]ImportInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var imports []ImportInfo
	scanner := bufio.NewScanner(file)
	lineNum := 0
	inImportBlock := false

	importRegex := regexp.MustCompile(`^\s*"([^"]+)"`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "import (") {
			inImportBlock = true
			continue
		}
		if inImportBlock && trimmed == ")" {
			inImportBlock = false
			continue
		}

		if strings.HasPrefix(trimmed, "import \"") {
			if matches := importRegex.FindStringSubmatch(trimmed); matches != nil {
				pkg := matches[1]
				if strings.Contains(pkg, ".") { // External package
					imports = append(imports, ImportInfo{
						Package:    pkg,
						SourceFile: filename,
						LineNumber: lineNum,
					})
				}
			}
		}

		if inImportBlock {
			if matches := importRegex.FindStringSubmatch(line); matches != nil {
				pkg := matches[1]
				if strings.Contains(pkg, ".") {
					imports = append(imports, ImportInfo{
						Package:    pkg,
						SourceFile: filename,
						LineNumber: lineNum,
					})
				}
			}
		}
	}

	return imports, scanner.Err()
}

func (s *Scanner) scanRust(filename string) ([]ImportInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var imports []ImportInfo
	scanner := bufio.NewScanner(file)
	lineNum := 0

	useRegex := regexp.MustCompile(`^\s*use\s+([a-zA-Z0-9_:]+)`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if matches := useRegex.FindStringSubmatch(line); matches != nil {
			module := matches[1]
			if !strings.HasPrefix(module, "std::") && !strings.HasPrefix(module, "crate::") {
				pkg := strings.Split(module, "::")[0]
				imports = append(imports, ImportInfo{
					Package:    pkg,
					SourceFile: filename,
					LineNumber: lineNum,
				})
			}
		}
	}

	return imports, scanner.Err()
}

func pythonStdlib() map[string]bool {
	return map[string]bool{
		"os": true, "sys": true, "re": true, "json": true, "time": true,
		"datetime": true, "collections": true, "itertools": true, "functools": true,
		"pathlib": true, "typing": true, "abc": true, "io": true, "math": true,
		"random": true, "string": true, "subprocess": true, "threading": true,
		"multiprocessing": true, "logging": true, "argparse": true, "configparser": true,
		"unittest": true, "sqlite3": true, "csv": true, "xml": true, "html": true,
		"urllib": true, "http": true, "email": true, "base64": true, "hashlib": true,
	}
}

func nodeBuiltins() map[string]bool {
	return map[string]bool{
		"assert": true, "buffer": true, "child_process": true, "cluster": true,
		"console": true, "constants": true, "crypto": true, "dgram": true,
		"dns": true, "domain": true, "events": true, "fs": true,
		"http": true, "https": true, "module": true, "net": true,
		"os": true, "path": true, "perf_hooks": true, "process": true,
		"punycode": true, "querystring": true, "readline": true, "repl": true,
		"stream": true, "string_decoder": true, "timers": true, "tls": true,
		"tty": true, "url": true, "util": true, "v8": true,
		"vm": true, "worker_threads": true, "zlib": true,
	}
}
