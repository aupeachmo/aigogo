package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/aupeachmo/aigogo/pkg/lockfile"
	"github.com/aupeachmo/aigogo/pkg/manifest"
	"github.com/aupeachmo/aigogo/pkg/store"
)

func execCmd() *Command {
	return &Command{
		Name:        "exec",
		Description: "Execute an agent's script",
		Run: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: aigg exec <agent_name> [args...]\n\n" +
					"Run an agent's entrypoint script.\n" +
					"The agent must be in aigogo.lock and have a \"scripts\" field in its manifest.")
			}

			agentName := args[0]
			agentArgs := args[1:]

			return runExec(agentName, agentArgs)
		},
	}
}

func runExec(agentName string, args []string) error {
	// 1. Find lock file and resolve the package
	_, lock, err := lockfile.FindLockFile()
	if err != nil {
		return fmt.Errorf("failed to find aigogo.lock: %w\n"+
			"Run 'aigg add <source>' to add the agent first", err)
	}

	pkg, exists := lock.Get(agentName)
	if !exists {
		return fmt.Errorf("agent %q not found in aigogo.lock\n"+
			"To add it, run: aigg add <registry>/%s:<tag> && aigg install", agentName, agentName)
	}

	// 2. Get from content-addressable store
	cas, err := store.NewStore()
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	hash := pkg.GetIntegrityHash()
	if !cas.Has(hash) {
		return fmt.Errorf("agent %q not installed (not in store)\n"+
			"Run: aigg install", agentName)
	}

	storedPkg, err := cas.Get(hash)
	if err != nil {
		return fmt.Errorf("failed to get agent from store: %w", err)
	}

	// 3. Load manifest to get scripts and language info
	m, err := manifest.Load(storedPkg.Manifest)
	if err != nil {
		return fmt.Errorf("failed to load agent manifest: %w", err)
	}

	if len(m.Scripts) == 0 {
		return fmt.Errorf("agent %q has no scripts defined in its manifest\n"+
			"The package author must add a \"scripts\" field to aigogo.json", agentName)
	}

	// Find the script to run: prefer matching agent name, fallback to first
	scriptFile, ok := m.Scripts[agentName]
	if !ok {
		// Try the first (or only) script
		if len(m.Scripts) == 1 {
			for _, v := range m.Scripts {
				scriptFile = v
			}
		} else {
			var available []string
			for k := range m.Scripts {
				available = append(available, k)
			}
			return fmt.Errorf("no script named %q found. Available scripts: %s",
				agentName, strings.Join(available, ", "))
		}
	}

	scriptPath := filepath.Join(storedPkg.FilesDir, scriptFile)
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script file %q not found in package", scriptFile)
	}

	// 4. Find and validate the interpreter
	interpreter, err := findInterpreter(m.Language)
	if err != nil {
		return err
	}

	// 5. Set up dependencies environment if needed
	envDir, err := setupExecEnv(hash, m, interpreter)
	if err != nil {
		return fmt.Errorf("failed to set up execution environment: %w", err)
	}

	// 6. Build and execute the command
	return executeScript(interpreter, m.Language.Name, scriptPath, storedPkg.FilesDir, envDir, args)
}

// findInterpreter locates and validates the language interpreter
func findInterpreter(lang manifest.Language) (string, error) {
	switch lang.Name {
	case "python":
		return findPythonInterpreter(lang.Version)
	case "javascript", "typescript":
		return findNodeInterpreter(lang.Version)
	default:
		return "", fmt.Errorf("exec is not supported for language %q (supported: python, javascript)", lang.Name)
	}
}

// findPythonInterpreter searches for a suitable Python interpreter
func findPythonInterpreter(versionConstraint string) (string, error) {
	candidates := []string{}

	// 1. Check active virtual environment
	if venv := os.Getenv("VIRTUAL_ENV"); venv != "" {
		candidates = append(candidates, filepath.Join(venv, "bin", "python"))
	}

	// 2. Check local project venvs
	cwd, _ := os.Getwd()
	if cwd != "" {
		for _, dir := range []string{"venv", ".venv"} {
			candidates = append(candidates, filepath.Join(cwd, dir, "bin", "python"))
		}
	}

	// 3. System interpreters
	candidates = append(candidates, "python3", "python")

	for _, candidate := range candidates {
		path, err := exec.LookPath(candidate)
		if err != nil {
			continue
		}

		// Validate version if constraint is specified
		if versionConstraint != "" {
			version, err := getPythonVersion(path)
			if err != nil {
				continue
			}
			if !checkVersionConstraint(version, versionConstraint) {
				continue // Try next candidate
			}
		}

		return path, nil
	}

	if versionConstraint != "" {
		return "", fmt.Errorf("no Python interpreter found satisfying %s\n"+
			"Install a compatible Python version or activate a virtual environment", versionConstraint)
	}
	return "", fmt.Errorf("python not found on PATH\nInstall Python 3 to run Python agents")
}

// findNodeInterpreter searches for a suitable Node.js interpreter
func findNodeInterpreter(versionConstraint string) (string, error) {
	path, err := exec.LookPath("node")
	if err != nil {
		return "", fmt.Errorf("node not found on PATH\nInstall Node.js to run JavaScript agents")
	}

	if versionConstraint != "" {
		version, err := getNodeVersion(path)
		if err != nil {
			return path, nil // Can't check, proceed anyway
		}
		if !checkVersionConstraint(version, versionConstraint) {
			return "", fmt.Errorf("node %s does not satisfy %s\nInstall a compatible Node.js version",
				version, versionConstraint)
		}
	}

	return path, nil
}

// getPythonVersion runs python --version and returns the version string
func getPythonVersion(pythonPath string) (string, error) {
	out, err := exec.Command(pythonPath, "--version").Output()
	if err != nil {
		return "", err
	}
	// Output: "Python 3.11.5"
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) >= 2 {
		return parts[1], nil
	}
	return "", fmt.Errorf("unexpected python version output: %s", string(out))
}

// getNodeVersion runs node --version and returns the version string
func getNodeVersion(nodePath string) (string, error) {
	out, err := exec.Command(nodePath, "--version").Output()
	if err != nil {
		return "", err
	}
	// Output: "v20.10.0"
	version := strings.TrimSpace(string(out))
	version = strings.TrimPrefix(version, "v")
	return version, nil
}

// checkVersionConstraint checks if a version satisfies a constraint like ">=3.8,<4.0" or ">=18"
func checkVersionConstraint(version, constraint string) bool {
	parts := parseVersion(version)
	if parts == nil {
		return true // Can't parse, allow it
	}

	// Split constraints by comma
	constraints := strings.Split(constraint, ",")
	for _, c := range constraints {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}

		if !evaluateConstraint(parts, c) {
			return false
		}
	}

	return true
}

// parseVersion parses "3.11.5" into [3, 11, 5]
func parseVersion(v string) []int {
	v = strings.TrimPrefix(v, "v")
	segments := strings.Split(v, ".")
	result := make([]int, 0, len(segments))
	for _, s := range segments {
		// Strip any pre-release suffix (e.g., "5rc1")
		re := regexp.MustCompile(`^(\d+)`)
		match := re.FindString(s)
		if match == "" {
			break
		}
		n, err := strconv.Atoi(match)
		if err != nil {
			break
		}
		result = append(result, n)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// evaluateConstraint evaluates a single constraint like ">=3.8"
func evaluateConstraint(version []int, constraint string) bool {
	var op string
	var target string

	if strings.HasPrefix(constraint, ">=") {
		op = ">="
		target = strings.TrimPrefix(constraint, ">=")
	} else if strings.HasPrefix(constraint, "<=") {
		op = "<="
		target = strings.TrimPrefix(constraint, "<=")
	} else if strings.HasPrefix(constraint, "!=") {
		op = "!="
		target = strings.TrimPrefix(constraint, "!=")
	} else if strings.HasPrefix(constraint, ">") {
		op = ">"
		target = strings.TrimPrefix(constraint, ">")
	} else if strings.HasPrefix(constraint, "<") {
		op = "<"
		target = strings.TrimPrefix(constraint, "<")
	} else if strings.HasPrefix(constraint, "==") {
		op = "=="
		target = strings.TrimPrefix(constraint, "==")
	} else {
		return true // Unknown operator, allow
	}

	targetParts := parseVersion(target)
	if targetParts == nil {
		return true
	}

	cmp := compareVersions(version, targetParts)

	switch op {
	case ">=":
		return cmp >= 0
	case "<=":
		return cmp <= 0
	case ">":
		return cmp > 0
	case "<":
		return cmp < 0
	case "==":
		return cmp == 0
	case "!=":
		return cmp != 0
	}

	return true
}

// compareVersions compares two version arrays, returns -1, 0, or 1
func compareVersions(a, b []int) int {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	for i := 0; i < maxLen; i++ {
		av := 0
		if i < len(a) {
			av = a[i]
		}
		bv := 0
		if i < len(b) {
			bv = b[i]
		}

		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}

// envsDir returns the path to ~/.aigogo/envs/
func envsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".aigogo", "envs"), nil
}

// envPath returns the path for a specific package's env directory
func envPath(hash string) (string, error) {
	base, err := envsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, hash), nil
}

// setupExecEnv creates the execution environment (venv/node_modules) if needed
func setupExecEnv(hash string, m *manifest.Manifest, interpreter string) (string, error) {
	dir, err := envPath(hash)
	if err != nil {
		return "", err
	}

	// Check if env already exists
	if _, err := os.Stat(dir); err == nil {
		return dir, nil
	}

	// No dependencies? No env needed
	if m.Dependencies == nil || len(m.Dependencies.Runtime) == 0 {
		return "", nil
	}

	fmt.Printf("Installing dependencies for %s (first run)...\n", m.Name)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create env directory: %w", err)
	}

	switch m.Language.Name {
	case "python":
		if err := setupPythonEnv(dir, m, interpreter); err != nil {
			// Clean up on failure
			_ = os.RemoveAll(dir)
			return "", err
		}
	case "javascript", "typescript":
		if err := setupNodeEnv(dir, m); err != nil {
			_ = os.RemoveAll(dir)
			return "", err
		}
	}

	return dir, nil
}

// setupPythonEnv creates a venv and installs dependencies
func setupPythonEnv(dir string, m *manifest.Manifest, interpreter string) error {
	venvPath := filepath.Join(dir, ".venv")

	// Try uv first, then fall back to python3 -m venv
	uvPath, uvErr := exec.LookPath("uv")
	if uvErr == nil {
		// Use uv to create venv
		cmd := exec.Command(uvPath, "venv", venvPath, "--python", interpreter)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			// Fall through to python -m venv
			uvPath = ""
		}
	}

	if uvPath == "" {
		// Fall back to python3 -m venv
		cmd := exec.Command(interpreter, "-m", "venv", venvPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create virtual environment: %w\n"+
				"Install the python3-venv package or install uv", err)
		}
	}

	// Build requirements list from dependencies
	var reqs []string
	for _, dep := range m.Dependencies.Runtime {
		reqs = append(reqs, dep.Package+dep.Version)
	}

	if len(reqs) == 0 {
		return nil
	}

	// Write requirements to a temp file
	reqFile := filepath.Join(dir, "requirements.txt")
	if err := os.WriteFile(reqFile, []byte(strings.Join(reqs, "\n")+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write requirements: %w", err)
	}

	// Install dependencies
	if uvPath != "" {
		cmd := exec.Command(uvPath, "pip", "install", "--python", filepath.Join(venvPath, "bin", "python"), "-r", reqFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("uv pip install failed: %w", err)
		}
	} else {
		// Use the venv's own pip
		venvPip := filepath.Join(venvPath, "bin", "pip")
		if _, err := os.Stat(venvPip); os.IsNotExist(err) {
			// Try pip3
			venvPip = filepath.Join(venvPath, "bin", "pip3")
		}
		cmd := exec.Command(venvPip, "install", "-r", reqFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pip install failed: %w", err)
		}
	}

	fmt.Println("Dependencies installed successfully")
	return nil
}

// setupNodeEnv creates a node_modules directory and installs dependencies
func setupNodeEnv(dir string, m *manifest.Manifest) error {
	// Generate a package.json with the dependencies
	deps := make(map[string]string)
	for _, dep := range m.Dependencies.Runtime {
		deps[dep.Package] = dep.Version
	}

	packageJSON := map[string]interface{}{
		"name":         m.Name + "-exec-env",
		"version":      "1.0.0",
		"private":      true,
		"dependencies": deps,
	}

	data, err := json.MarshalIndent(packageJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to generate package.json: %w", err)
	}

	pkgPath := filepath.Join(dir, "package.json")
	if err := os.WriteFile(pkgPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write package.json: %w", err)
	}

	// Run npm install
	npmPath, err := exec.LookPath("npm")
	if err != nil {
		return fmt.Errorf("npm not found on PATH\nInstall Node.js to install JavaScript dependencies")
	}

	cmd := exec.Command(npmPath, "install", "--prefix", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm install failed: %w", err)
	}

	fmt.Println("Dependencies installed successfully")
	return nil
}

// executeScript runs the script with the appropriate interpreter and environment
func executeScript(interpreter, language, scriptPath, filesDir, envDir string, args []string) error {
	var cmdArgs []string
	env := os.Environ()

	switch language {
	case "python":
		if envDir != "" {
			// Use the venv's Python interpreter which has access to installed packages
			venvPython := filepath.Join(envDir, ".venv", "bin", "python")
			if _, err := os.Stat(venvPython); err == nil {
				interpreter = venvPython
			}
		}

		// Add the files directory to PYTHONPATH so imports work
		pythonPath := filesDir
		if existing := os.Getenv("PYTHONPATH"); existing != "" {
			pythonPath = filesDir + ":" + existing
		}
		env = setEnv(env, "PYTHONPATH", pythonPath)

		cmdArgs = append([]string{scriptPath}, args...)

	case "javascript", "typescript":
		// Set NODE_PATH to include both the files dir and any env node_modules
		nodePath := filesDir
		if envDir != "" {
			nodeModules := filepath.Join(envDir, "node_modules")
			nodePath = filesDir + ":" + nodeModules
		}
		if existing := os.Getenv("NODE_PATH"); existing != "" {
			nodePath = nodePath + ":" + existing
		}
		env = setEnv(env, "NODE_PATH", nodePath)

		cmdArgs = append([]string{scriptPath}, args...)
	}

	// Use syscall.Exec to replace the process (like exec in bash)
	// This ensures signals, exit codes, and stdio are properly forwarded
	binary, err := exec.LookPath(interpreter)
	if err != nil {
		return fmt.Errorf("interpreter not found: %s", interpreter)
	}

	execArgs := append([]string{binary}, cmdArgs...)
	return syscall.Exec(binary, execArgs, env)
}

// setEnv sets an environment variable in the env slice
func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}
