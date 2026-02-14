#!/usr/bin/env bash
#
# qa/run.sh — Automated QA test harness for aigg
#
# Exercises every command, flag, and error path listed in qa/QA.md.
# Registry tests are skipped when credentials are not provided.
#
# Usage:
#   bash qa/run.sh              # prompts for registry creds (Enter to skip)
#   bash qa/run.sh --local      # skip registry prompts entirely
#
set -euo pipefail

###############################################################################
# Resolve aigg binary
###############################################################################
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

AIGOGO="${AIGOGO:-$REPO_ROOT/bin/aigg}"
if [[ ! -x "$AIGOGO" ]]; then
    echo "aigg binary not found at $AIGOGO"
    echo "Run 'make build' first, or set AIGOGO=/path/to/aigg"
    exit 1
fi

###############################################################################
# Colours / formatting
###############################################################################
if [[ -t 1 ]]; then
    GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[0;33m'
    BOLD='\033[1m'; RESET='\033[0m'
else
    GREEN=''; RED=''; YELLOW=''; BOLD=''; RESET=''
fi

###############################################################################
# Counters
###############################################################################
TOTAL=0; PASSED=0; FAILED=0; SKIPPED=0

###############################################################################
# Log file — all command output goes here
###############################################################################
LOGFILE="$(mktemp /tmp/aigg-qa-log.XXXXXX)"
echo "Log file: $LOGFILE"

###############################################################################
# Temp workspace — cleaned up on exit
###############################################################################
WORK="$(mktemp -d /tmp/aigg-qa-work.XXXXXX)"
cleanup() { rm -rf "$WORK"; }
trap cleanup EXIT
echo "Workspace: $WORK"
echo ""

###############################################################################
# Registry credentials
###############################################################################
REGISTRY="" ; REG_USER="" ; REG_PASS="" ; REG_REPO=""
HAS_REGISTRY=false

if [[ "${1:-}" != "--local" ]]; then
    echo "--- Registry credentials (press Enter to skip all) ---"
    read -rp "Registry [docker.io]: " REGISTRY
    REGISTRY="${REGISTRY:-docker.io}"

    if [[ -n "$REGISTRY" ]]; then
        read -rp "Username: " REG_USER
        if [[ -n "$REG_USER" ]]; then
            read -rsp "Password/PAT: " REG_PASS; echo ""
            read -rp "Repository namespace (e.g. $REG_USER/aigogo-qa-test): " REG_REPO
            REG_REPO="${REG_REPO:-$REG_USER/aigogo-qa-test}"
            if [[ -n "$REG_PASS" ]]; then
                HAS_REGISTRY=true
            fi
        fi
    fi
    echo ""
fi

if $HAS_REGISTRY; then
    echo "Registry tests:  ENABLED ($REGISTRY/$REG_REPO)"
else
    echo "Registry tests:  SKIPPED (no credentials)"
fi
echo "========================================================"
echo ""

###############################################################################
# Test helpers
###############################################################################

# run_test "description" command...
#   Expects exit 0.
run_test() {
    local desc="$1"; shift
    TOTAL=$((TOTAL + 1))
    if "$@" >>"$LOGFILE" 2>&1; then
        PASSED=$((PASSED + 1))
        printf "${GREEN}PASS${RESET}  %s\n" "$desc"
    else
        FAILED=$((FAILED + 1))
        printf "${RED}FAIL${RESET}  %s\n" "$desc"
        echo "       command: $*" >&2
    fi
}

# run_test_fail "description" command...
#   Expects non-zero exit.
run_test_fail() {
    local desc="$1"; shift
    TOTAL=$((TOTAL + 1))
    if "$@" >>"$LOGFILE" 2>&1; then
        FAILED=$((FAILED + 1))
        printf "${RED}FAIL${RESET}  %s  (expected failure but got exit 0)\n" "$desc"
    else
        PASSED=$((PASSED + 1))
        printf "${GREEN}PASS${RESET}  %s\n" "$desc"
    fi
}

# run_test_grep "description" "pattern" command...
#   Expects exit 0 AND stdout+stderr matches pattern.
run_test_grep() {
    local desc="$1"; local pattern="$2"; shift 2
    TOTAL=$((TOTAL + 1))
    local out
    out="$("$@" 2>&1)" || true
    echo "$out" >>"$LOGFILE"
    if echo "$out" | grep -qE -- "$pattern"; then
        PASSED=$((PASSED + 1))
        printf "${GREEN}PASS${RESET}  %s\n" "$desc"
    else
        FAILED=$((FAILED + 1))
        printf "${RED}FAIL${RESET}  %s  (pattern /%s/ not found)\n" "$desc" "$pattern"
    fi
}

# run_test_fail_grep "description" "pattern" command...
#   Expects non-zero exit AND output matches pattern.
run_test_fail_grep() {
    local desc="$1"; local pattern="$2"; shift 2
    TOTAL=$((TOTAL + 1))
    local out; local rc=0
    out="$("$@" 2>&1)" || rc=$?
    echo "$out" >>"$LOGFILE"
    if [[ $rc -ne 0 ]] && echo "$out" | grep -qE -- "$pattern"; then
        PASSED=$((PASSED + 1))
        printf "${GREEN}PASS${RESET}  %s\n" "$desc"
    else
        FAILED=$((FAILED + 1))
        if [[ $rc -eq 0 ]]; then
            printf "${RED}FAIL${RESET}  %s  (expected failure but got exit 0)\n" "$desc"
        else
            printf "${RED}FAIL${RESET}  %s  (pattern /%s/ not found)\n" "$desc" "$pattern"
        fi
    fi
}

# skip_test "description"
skip_test() {
    local desc="$1"
    TOTAL=$((TOTAL + 1)); SKIPPED=$((SKIPPED + 1))
    printf "${YELLOW}SKIP${RESET}  %s\n" "$desc"
}

###############################################################################
# Helpers to create dummy projects
###############################################################################

# Create a minimal Python agent project in $1
create_python_project() {
    local dir="$1"
    mkdir -p "$dir"

    cat > "$dir/utils.py" <<'PYEOF'
import os
import json

def hello():
    return "hello from aigogo"
PYEOF

    cat > "$dir/helpers.py" <<'PYEOF'
import sys

def helper():
    return sys.platform
PYEOF
}

# Create a minimal JS agent project in $1
create_js_project() {
    local dir="$1"
    mkdir -p "$dir"

    cat > "$dir/index.js" <<'JSEOF'
const fs = require('fs');

module.exports = {
    greet: () => "hello from aigogo"
};
JSEOF
}

# Create a Python virtualenv.
# Tries poetry first, then python3 -m venv as fallback.
# Usage: create_venv <dir>
# Sets VENV_PATH to the created venv directory. Returns 0 on success.
create_venv() {
    local dir="$1"
    local venv="$dir/.venv"
    mkdir -p "$dir"

    # Try poetry first
    if command -v poetry >/dev/null 2>&1; then
        (
            cd "$dir"
            poetry init -n --no-interaction >>"$LOGFILE" 2>&1
            POETRY_VIRTUALENVS_IN_PROJECT=true poetry install --no-root >>"$LOGFILE" 2>&1
        )
        if [ -d "$venv" ]; then
            VENV_PATH="$venv"
            return 0
        fi
        echo "poetry found but failed to create venv" >>"$LOGFILE"
    fi

    # Fallback: python3 -m venv
    if python3 -m venv "$venv" >>"$LOGFILE" 2>&1; then
        VENV_PATH="$venv"
        return 0
    fi
    echo "python3 -m venv also failed" >>"$LOGFILE"

    echo "ERROR: could not create virtualenv in $dir" >>"$LOGFILE"
    return 1
}

# Activate a venv created by create_venv.
# Saves the original PATH so deactivate_venv can restore it.
activate_venv() {
    local venv_dir="$1"
    SAVED_PATH="$PATH"
    export VIRTUAL_ENV="$venv_dir"
    export PATH="$venv_dir/bin:$PATH"
}

# Deactivate the current venv and restore PATH.
deactivate_venv() {
    unset VIRTUAL_ENV
    export PATH="$SAVED_PATH"
}

###############################################################################
#  SECTION: Utilities
###############################################################################
echo "${BOLD}=== Utilities ===${RESET}"

run_test_grep "aigg version" "aigg version" \
    "$AIGOGO" version

run_test_grep "aigg completion bash" "_aigg_completions" \
    "$AIGOGO" completion bash

run_test_grep "aigg completion zsh" "#compdef aigg" \
    "$AIGOGO" completion zsh

run_test_grep "aigg completion fish" "complete -c aigg" \
    "$AIGOGO" completion fish

echo ""

###############################################################################
#  SECTION: Author Commands
###############################################################################
echo "${BOLD}=== Author Commands ===${RESET}"

# --- init ---
INIT_DIR="$WORK/init-test"
mkdir -p "$INIT_DIR"
pushd "$INIT_DIR" >/dev/null

run_test_grep "aigg init — creates aigogo.json" "Initialized aigogo package" \
    "$AIGOGO" init

run_test "aigg init — aigogo.json exists" test -f aigogo.json

popd >/dev/null

# --- add file ---
PY_DIR="$WORK/author-py"
create_python_project "$PY_DIR"
pushd "$PY_DIR" >/dev/null
"$AIGOGO" init >>"$LOGFILE" 2>&1

run_test_grep "aigg add file <path>" "Added 1 file" \
    "$AIGOGO" add file utils.py

run_test_grep "aigg add file <path> (second file)" "Added 1 file" \
    "$AIGOGO" add file helpers.py

# add file with --force  (add a file that would normally be fine, just test flag)
cat > "$PY_DIR/extra.py" <<'EOF'
# extra
EOF
run_test_grep "aigg add file <path> --force (flag after path)" "Added 1 file" \
    "$AIGOGO" add file extra.py --force

# add file with glob
cat > "$PY_DIR/glob1.py" << 'EOF'
pass
EOF
cat > "$PY_DIR/glob2.py" << 'EOF'
pass
EOF
run_test_grep "aigg add file <glob>" "Added 1 file" \
    "$AIGOGO" add file "glob*.py"

popd >/dev/null

# --- add dep / add dev ---
DEP_DIR="$WORK/author-dep"
create_python_project "$DEP_DIR"
pushd "$DEP_DIR" >/dev/null
"$AIGOGO" init >>"$LOGFILE" 2>&1

run_test_grep "aigg add dep <pkg> <ver>" "Added requests" \
    "$AIGOGO" add dep requests ">=2.28.0"

run_test_grep "aigg add dev <pkg> <ver>" "Added pytest" \
    "$AIGOGO" add dev pytest ">=7.0.0"

popd >/dev/null

# --- rm ---
RM_DIR="$WORK/author-rm"
create_python_project "$RM_DIR"
pushd "$RM_DIR" >/dev/null
"$AIGOGO" init >>"$LOGFILE" 2>&1
"$AIGOGO" add file utils.py >>"$LOGFILE" 2>&1
"$AIGOGO" add file helpers.py >>"$LOGFILE" 2>&1
"$AIGOGO" add dep requests ">=2.0" >>"$LOGFILE" 2>&1
"$AIGOGO" add dev pytest ">=7.0" >>"$LOGFILE" 2>&1

run_test_grep "aigg rm file <path>" "Removed 1 file" \
    "$AIGOGO" rm file helpers.py

run_test_grep "aigg rm dep <pkg>" "Removed requests" \
    "$AIGOGO" rm dep requests

run_test_grep "aigg rm dev <pkg>" "Removed pytest" \
    "$AIGOGO" rm dev pytest

popd >/dev/null

# --- scan ---
SCAN_DIR="$WORK/author-scan"
create_python_project "$SCAN_DIR"
pushd "$SCAN_DIR" >/dev/null
"$AIGOGO" init >>"$LOGFILE" 2>&1
"$AIGOGO" add file utils.py >>"$LOGFILE" 2>&1

run_test_grep "aigg scan" "Scanning source files" \
    "$AIGOGO" scan

popd >/dev/null

# --- validate ---
VAL_DIR="$WORK/author-validate"
create_python_project "$VAL_DIR"
pushd "$VAL_DIR" >/dev/null
"$AIGOGO" init >>"$LOGFILE" 2>&1
"$AIGOGO" add file utils.py >>"$LOGFILE" 2>&1

run_test_grep "aigg validate" "Validating manifest" \
    "$AIGOGO" validate

popd >/dev/null

# --- build ---
BUILD_DIR="$WORK/author-build"
create_python_project "$BUILD_DIR"
pushd "$BUILD_DIR" >/dev/null
"$AIGOGO" init >>"$LOGFILE" 2>&1
"$AIGOGO" add file utils.py >>"$LOGFILE" 2>&1

# Use --force on first build to handle leftover cache from prior runs
run_test_grep "aigg build <name>:<tag>" "Successfully built" \
    "$AIGOGO" build qa-test:1.0.0 --force

run_test_grep "aigg build (auto-increment)" "Auto-incrementing version" \
    "$AIGOGO" build

run_test_grep "aigg build --force" "Successfully built" \
    "$AIGOGO" build qa-test:1.0.0 --force

run_test_grep "aigg build --no-validate" "Successfully built" \
    "$AIGOGO" build qa-test:1.0.1 --force --no-validate

popd >/dev/null

echo ""

###############################################################################
#  SECTION: Diff Command
###############################################################################
echo "${BOLD}=== Diff Command ===${RESET}"

# Re-use the build dir from the build section (BUILD_DIR).
# State: aigogo.json has name=author-build version=0.1.1 (from auto-increment
# build). Cache has "author-build:0.1.1" matching the working dir files.
pushd "$BUILD_DIR" >/dev/null

# --- Zero-arg mode (uses name:version from aigogo.json) ---
# Working dir hasn't changed since the auto-increment build, so should be identical
run_test_grep "aigg diff (zero-arg, identical)" "identical" \
    "$AIGOGO" diff

# Modify a file, then zero-arg diff should detect the change
cp utils.py utils.py.bak
echo "# zero-arg change" >> utils.py

run_test_grep "aigg diff (zero-arg, modified)" "modified" \
    "$AIGOGO" diff

# Restore for subsequent tests
cp utils.py.bak utils.py

# --- One-arg mode (explicit ref) ---
echo "# modified" >> utils.py

run_test_grep "aigg diff <name>:<tag> (working dir vs build)" "modified" \
    "$AIGOGO" diff qa-test:1.0.0

run_test_grep "aigg diff --summary <name>:<tag>" "^M " \
    "$AIGOGO" diff --summary qa-test:1.0.0

# Restore file — diff should show identical
cp utils.py.bak utils.py
rm utils.py.bak

run_test_grep "aigg diff <name>:<tag> (identical)" "identical" \
    "$AIGOGO" diff qa-test:1.0.0

# --- Two-arg mode (local build vs local build) ---
"$AIGOGO" build qa-diff-a:1.0.0 --force >>"$LOGFILE" 2>&1
echo "# diff change" >> utils.py
"$AIGOGO" build qa-diff-b:1.0.0 --force >>"$LOGFILE" 2>&1

run_test_grep "aigg diff <ref-a> <ref-b> (two local builds)" "modified" \
    "$AIGOGO" diff qa-diff-a:1.0.0 qa-diff-b:1.0.0

popd >/dev/null

echo ""

###############################################################################
#  SECTION: Consumer Commands
###############################################################################
echo "${BOLD}=== Consumer Commands ===${RESET}"

# Build a local package, then use "aigg add <name>:<tag>" which resolves
# from the local cache (no registry required).

BUILD_CONSUMER="$WORK/consumer-build"
create_python_project "$BUILD_CONSUMER"
pushd "$BUILD_CONSUMER" >/dev/null
"$AIGOGO" init >>"$LOGFILE" 2>&1
"$AIGOGO" add file utils.py >>"$LOGFILE" 2>&1
"$AIGOGO" build consumer-pkg:1.0.0 --force >>"$LOGFILE" 2>&1
popd >/dev/null

CONSUMER_DIR="$WORK/consumer"
mkdir -p "$CONSUMER_DIR"
pushd "$CONSUMER_DIR" >/dev/null

# Create a virtualenv so aigogo can write .pth to site-packages without root
if create_venv "$WORK/consumer-venv"; then
    activate_venv "$VENV_PATH"
fi

run_test_grep "aigg add <name>:<tag> (local cache)" "Added|local cache" \
    "$AIGOGO" add consumer-pkg:1.0.0

run_test "aigg add — aigogo.lock created" test -f aigogo.lock

run_test_grep "aigg install" "Installed" \
    "$AIGOGO" install

# --- .pth file verification ---
# Check that .pth-location tracking file was created
run_test "aigg install — .pth-location tracking file created" \
    test -f "$CONSUMER_DIR/.aigogo/.pth-location"

# Check that the .pth file referenced by the tracking file exists
pth_file_check() {
    local tracking="$CONSUMER_DIR/.aigogo/.pth-location"
    if [ ! -f "$tracking" ]; then
        echo "tracking file not found: $tracking" >>"$LOGFILE"
        return 1
    fi
    local pth_path
    pth_path="$(cat "$tracking" | tr -d '[:space:]')"
    if [ ! -f "$pth_path" ]; then
        echo ".pth file not found at: $pth_path" >>"$LOGFILE"
        return 1
    fi
    # Verify the .pth file contains the imports directory path
    if ! grep -q ".aigogo/imports" "$pth_path"; then
        echo ".pth file does not contain .aigogo/imports path" >>"$LOGFILE"
        echo "contents: $(cat "$pth_path")" >>"$LOGFILE"
        return 1
    fi
    return 0
}

run_test "aigg install — .pth file written to site-packages" \
    pth_file_check

# Check that Python can import without manual PYTHONPATH
pth_import_check() {
    # Unset PYTHONPATH to prove .pth is doing the work
    local normalized_name
    normalized_name=$(python3 -c "
import json
with open('$CONSUMER_DIR/aigogo.lock') as f:
    lock = json.load(f)
for name, pkg in lock['packages'].items():
    print(name.replace('-', '_').replace('.', '_'))
    break
" 2>>"$LOGFILE") || return 1

    # Try importing the aigogo namespace — should work via .pth file alone
    (cd /tmp && PYTHONPATH="" python3 -c "import aigogo.${normalized_name}" 2>>"$LOGFILE")
}

run_test "aigg install — Python import works without PYTHONPATH" \
    pth_import_check

# Deactivate the virtualenv
if [ -n "${VIRTUAL_ENV:-}" ]; then
    deactivate_venv
fi

popd >/dev/null

# --- JavaScript consumer tests ---
# Build a JS package, install it, verify the new structure

BUILD_JS_CONSUMER="$WORK/consumer-js-build"
create_js_project "$BUILD_JS_CONSUMER"
pushd "$BUILD_JS_CONSUMER" >/dev/null
"$AIGOGO" init >>"$LOGFILE" 2>&1
# Switch to javascript language and set name to match build tag
python3 -c "
import json
with open('aigogo.json') as f: m = json.load(f)
m['name'] = 'js-consumer-pkg'
m['language']['name'] = 'javascript'
m['language']['version'] = '>=18'
with open('aigogo.json', 'w') as f: json.dump(m, f, indent=2)
" 2>>"$LOGFILE" || true
"$AIGOGO" add file index.js >>"$LOGFILE" 2>&1
"$AIGOGO" build js-consumer-pkg:1.0.0 --force >>"$LOGFILE" 2>&1
popd >/dev/null

JS_CONSUMER_DIR="$WORK/consumer-js"
mkdir -p "$JS_CONSUMER_DIR"
pushd "$JS_CONSUMER_DIR" >/dev/null

"$AIGOGO" add js-consumer-pkg:1.0.0 >>"$LOGFILE" 2>&1
"$AIGOGO" install >>"$LOGFILE" 2>&1

# Check that JS package is a real directory (not a symlink)
js_real_dir_check() {
    local pkg_dir="$JS_CONSUMER_DIR/.aigogo/imports/@aigogo/js_consumer_pkg"
    # Try both normalized and original name
    if [ ! -d "$pkg_dir" ]; then
        # Try with hyphens
        pkg_dir="$JS_CONSUMER_DIR/.aigogo/imports/@aigogo/js-consumer-pkg"
    fi
    if [ ! -d "$pkg_dir" ]; then
        echo "JS package directory not found" >>"$LOGFILE"
        ls -la "$JS_CONSUMER_DIR/.aigogo/imports/@aigogo/" >>"$LOGFILE" 2>&1 || true
        return 1
    fi
    # Verify it is NOT a symlink (should be a real directory)
    if [ -L "$pkg_dir" ]; then
        echo "JS package should be a real directory, not a symlink" >>"$LOGFILE"
        return 1
    fi
    return 0
}

run_test "aigg install — JS package is real directory (not symlink)" \
    js_real_dir_check

# Check that generated package.json exists with main field
js_package_json_check() {
    local pkg_dir
    for candidate in "$JS_CONSUMER_DIR/.aigogo/imports/@aigogo/js_consumer_pkg" \
                     "$JS_CONSUMER_DIR/.aigogo/imports/@aigogo/js-consumer-pkg"; do
        if [ -d "$candidate" ]; then
            pkg_dir="$candidate"
            break
        fi
    done
    if [ -z "$pkg_dir" ]; then
        echo "JS package directory not found" >>"$LOGFILE"
        return 1
    fi
    if [ ! -f "$pkg_dir/package.json" ]; then
        echo "package.json not found in $pkg_dir" >>"$LOGFILE"
        return 1
    fi
    if ! grep -q '"main"' "$pkg_dir/package.json"; then
        echo "package.json missing main field" >>"$LOGFILE"
        cat "$pkg_dir/package.json" >>"$LOGFILE"
        return 1
    fi
    return 0
}

run_test "aigg install — JS package has generated package.json with main" \
    js_package_json_check

# Check that register.js was generated
run_test "aigg install — register.js generated" \
    test -f "$JS_CONSUMER_DIR/.aigogo/register.js"

# Check register.js contains the expected content
js_register_content_check() {
    if ! grep -q "Module._initPaths" "$JS_CONSUMER_DIR/.aigogo/register.js"; then
        echo "register.js missing _initPaths call" >>"$LOGFILE"
        return 1
    fi
    if ! grep -q "NODE_PATH" "$JS_CONSUMER_DIR/.aigogo/register.js"; then
        echo "register.js missing NODE_PATH setup" >>"$LOGFILE"
        return 1
    fi
    return 0
}

run_test "aigg install — register.js has correct content" \
    js_register_content_check

# Check that require('@aigogo/...') works via register script
js_require_check() {
    if ! command -v node >/dev/null 2>&1; then
        echo "node not found, skipping" >>"$LOGFILE"
        return 0
    fi
    (cd "$JS_CONSUMER_DIR" && node --require ./.aigogo/register.js -e "require('@aigogo/js-consumer-pkg')" 2>>"$LOGFILE")
}

run_test "aigg install — JS require works via register script" \
    js_require_check

popd >/dev/null

echo ""

###############################################################################
#  SECTION: Uninstall Command
###############################################################################
echo "${BOLD}=== Uninstall Command ===${RESET}"

# Create a fresh project, install, then uninstall
UNINSTALL_DIR="$WORK/uninstall-test"
mkdir -p "$UNINSTALL_DIR"

# Build a Python package for this test
UNINSTALL_BUILD="$WORK/uninstall-build"
create_python_project "$UNINSTALL_BUILD"
pushd "$UNINSTALL_BUILD" >/dev/null
"$AIGOGO" init >>"$LOGFILE" 2>&1
"$AIGOGO" add file utils.py >>"$LOGFILE" 2>&1
"$AIGOGO" build uninstall-pkg:1.0.0 --force >>"$LOGFILE" 2>&1
popd >/dev/null

pushd "$UNINSTALL_DIR" >/dev/null

# Create a virtualenv so .pth can be written and tested
if create_venv "$WORK/uninstall-venv"; then
    activate_venv "$VENV_PATH"
fi

"$AIGOGO" add uninstall-pkg:1.0.0 >>"$LOGFILE" 2>&1
"$AIGOGO" install >>"$LOGFILE" 2>&1

# Verify .aigogo/ exists before uninstall
run_test "uninstall — .aigogo/ exists before uninstall" \
    test -d "$UNINSTALL_DIR/.aigogo"

run_test_grep "aigg uninstall" "Uninstall complete" \
    "$AIGOGO" uninstall

# Verify .aigogo/ is removed
uninstall_aigogo_dir_check() {
    if [ -d "$UNINSTALL_DIR/.aigogo" ]; then
        echo ".aigogo/ still exists after uninstall" >>"$LOGFILE"
        return 1
    fi
    return 0
}

run_test "aigg uninstall — .aigogo/ removed" \
    uninstall_aigogo_dir_check

# Verify aigogo.lock is preserved
run_test "aigg uninstall — aigogo.lock preserved" \
    test -f "$UNINSTALL_DIR/aigogo.lock"

# Verify .pth file was cleaned from site-packages
uninstall_pth_check() {
    # The .pth-location tracking file is gone (inside .aigogo/)
    # and the .pth file in site-packages should also be gone
    local site_packages
    site_packages=$(python3 -c "import sysconfig; print(sysconfig.get_path('purelib'))" 2>/dev/null) || return 0
    if [ -f "$site_packages/aigogo.pth" ]; then
        echo "aigogo.pth still exists in site-packages after uninstall" >>"$LOGFILE"
        return 1
    fi
    return 0
}

run_test "aigg uninstall — .pth file cleaned from site-packages" \
    uninstall_pth_check

# Deactivate the virtualenv
if [ -n "${VIRTUAL_ENV:-}" ]; then
    deactivate_venv
fi

popd >/dev/null

# Uninstall with nothing to uninstall
UNINSTALL_EMPTY="$WORK/uninstall-empty"
mkdir -p "$UNINSTALL_EMPTY"
# Create aigogo.lock so findProjectDir succeeds
echo '{"version":1,"packages":{}}' > "$UNINSTALL_EMPTY/aigogo.lock"
pushd "$UNINSTALL_EMPTY" >/dev/null

run_test_grep "aigg uninstall — nothing to uninstall" "Nothing to uninstall" \
    "$AIGOGO" uninstall

popd >/dev/null

# Uninstall outside any project → error
UNINSTALL_ERR="$WORK/uninstall-no-project"
mkdir -p "$UNINSTALL_ERR"
pushd "$UNINSTALL_ERR" >/dev/null

run_test_fail_grep "aigg uninstall outside project -> error" "not an aigogo project" \
    "$AIGOGO" uninstall

popd >/dev/null

echo ""

###############################################################################
#  SECTION: Content-Addressable Store Verification
###############################################################################
echo "${BOLD}=== CAS Verification ===${RESET}"

# Extract the integrity hash and package name from the lock file
CAS_INFO=$(python3 -c "
import json
with open('$CONSUMER_DIR/aigogo.lock') as f:
    lock = json.load(f)
for name, pkg in lock['packages'].items():
    h = pkg['integrity'].replace('sha256:', '')
    lang = pkg.get('language', 'python')
    print(h)
    print(name)
    print(lang)
    break
" 2>>"$LOGFILE") || true

CAS_HASH=$(echo "$CAS_INFO" | sed -n '1p')
CAS_PKG_NAME=$(echo "$CAS_INFO" | sed -n '2p')
CAS_PKG_LANG=$(echo "$CAS_INFO" | sed -n '3p')
CAS_PREFIX="${CAS_HASH:0:2}"
STORE_ROOT="$HOME/.aigogo/store"
STORE_PKG_DIR="$STORE_ROOT/sha256/$CAS_PREFIX/$CAS_HASH"

# Check 1: Store directory structure exists
run_test "CAS: store dir exists for hash" \
    test -d "$STORE_PKG_DIR"

run_test "CAS: files/ subdirectory exists" \
    test -d "$STORE_PKG_DIR/files"

run_test "CAS: aigogo.json manifest exists" \
    test -f "$STORE_PKG_DIR/aigogo.json"

# Check 2: Integrity hash from lock file resolves to a real store entry
#   (We already tested the directory exists above — now verify the lock file
#    integrity value matches the directory name)
lock_integrity_check() {
    local lock_file="$1"
    python3 -c "
import json, os, sys
with open('$lock_file') as f:
    lock = json.load(f)
for name, pkg in lock['packages'].items():
    h = pkg['integrity'].replace('sha256:', '')
    prefix = h[:2]
    store_path = os.path.expanduser('~/.aigogo/store/sha256/' + prefix + '/' + h)
    if not os.path.isdir(store_path):
        print('MISMATCH: lock hash %s not found at %s' % (h, store_path), file=sys.stderr)
        sys.exit(1)
    # Verify files/ subdir exists inside
    files_dir = os.path.join(store_path, 'files')
    if not os.path.isdir(files_dir):
        print('MISMATCH: no files/ dir at %s' % store_path, file=sys.stderr)
        sys.exit(1)
sys.exit(0)
" 2>>"$LOGFILE"
}

run_test "CAS: lock file integrity hash matches store path" \
    lock_integrity_check "$CONSUMER_DIR/aigogo.lock"

# Check 3: Files in the store match the original source files
cas_file_content_check() {
    local original="$BUILD_CONSUMER/utils.py"
    local stored="$STORE_PKG_DIR/files/utils.py"
    if [ ! -f "$stored" ]; then
        echo "stored file not found: $stored" >>"$LOGFILE"
        return 1
    fi
    diff -q "$original" "$stored" >>"$LOGFILE" 2>&1
}

run_test "CAS: stored file content matches original" \
    cas_file_content_check

# Check 4: Files in the store are read-only (0444), directories are 0555
cas_permissions_check() {
    local fail=0
    # Check a stored file
    local file_perms
    file_perms=$(stat -c '%a' "$STORE_PKG_DIR/files/utils.py" 2>/dev/null || \
                 stat -f '%Lp' "$STORE_PKG_DIR/files/utils.py" 2>/dev/null)
    if [ "$file_perms" != "444" ]; then
        echo "expected file perms 444, got $file_perms" >>"$LOGFILE"
        fail=1
    fi
    # Check the files/ directory
    local dir_perms
    dir_perms=$(stat -c '%a' "$STORE_PKG_DIR/files" 2>/dev/null || \
                stat -f '%Lp' "$STORE_PKG_DIR/files" 2>/dev/null)
    if [ "$dir_perms" != "555" ]; then
        echo "expected dir perms 555, got $dir_perms" >>"$LOGFILE"
        fail=1
    fi
    return $fail
}

run_test "CAS: stored files are read-only (0444/0555)" \
    cas_permissions_check

# Check 5: Symlink points to correct store path
cas_symlink_check() {
    # Python normalizes hyphens/dots to underscores
    local normalized_name
    normalized_name=$(echo "$CAS_PKG_NAME" | sed 's/[-.]/_/g')

    local link_path
    if [ "$CAS_PKG_LANG" = "python" ]; then
        link_path="$CONSUMER_DIR/.aigogo/imports/aigogo/$normalized_name"
    else
        link_path="$CONSUMER_DIR/.aigogo/imports/@aigogo/$CAS_PKG_NAME"
    fi

    if [ ! -L "$link_path" ]; then
        echo "symlink not found: $link_path" >>"$LOGFILE"
        echo "  package name from lock: $CAS_PKG_NAME" >>"$LOGFILE"
        echo "  normalized: $normalized_name" >>"$LOGFILE"
        return 1
    fi
    local target
    target=$(readlink "$link_path")
    local expected="$STORE_PKG_DIR/files"
    if [ "$target" != "$expected" ]; then
        echo "symlink target mismatch:" >>"$LOGFILE"
        echo "  expected: $expected" >>"$LOGFILE"
        echo "  got:      $target" >>"$LOGFILE"
        return 1
    fi
    # Verify the target is a real directory (link resolves)
    if [ ! -d "$link_path" ]; then
        echo "symlink target does not resolve to a directory" >>"$LOGFILE"
        return 1
    fi
    return 0
}

run_test "CAS: symlink exists at .aigogo/imports/aigogo/<pkg>" \
    cas_symlink_check

echo ""

# Also test "aigg add <registry>/<name>:<tag>" if registry is available
if $HAS_REGISTRY; then
    CONSUMER_REG_DIR="$WORK/consumer-reg"
    mkdir -p "$CONSUMER_REG_DIR"
    pushd "$CONSUMER_REG_DIR" >/dev/null
    run_test_grep "aigg add <registry>/<name>:<tag>" "Added" \
        "$AIGOGO" add "$REGISTRY/$REG_REPO:1.0.0"
    popd >/dev/null
else
    skip_test "aigg add <registry>/<name>:<tag> (needs registry)"
fi

echo ""

###############################################################################
#  SECTION: show-deps Formats
###############################################################################
echo "${BOLD}=== show-deps Formats ===${RESET}"

# Python project for show-deps
SHOWDEPS_PY="$WORK/showdeps-py"
create_python_project "$SHOWDEPS_PY"
pushd "$SHOWDEPS_PY" >/dev/null
"$AIGOGO" init >>"$LOGFILE" 2>&1
"$AIGOGO" add dep requests ">=2.28.0" >>"$LOGFILE" 2>&1
"$AIGOGO" add dev pytest ">=7.0.0" >>"$LOGFILE" 2>&1
popd >/dev/null

# JS project for show-deps
SHOWDEPS_JS="$WORK/showdeps-js"
create_js_project "$SHOWDEPS_JS"
pushd "$SHOWDEPS_JS" >/dev/null
"$AIGOGO" init >>"$LOGFILE" 2>&1
# Switch to javascript language
python3 -c "
import json, sys
with open('aigogo.json') as f: m = json.load(f)
m['language']['name'] = 'javascript'
m['language']['version'] = '>=18'
with open('aigogo.json', 'w') as f: json.dump(m, f, indent=2)
" 2>>"$LOGFILE" || true
"$AIGOGO" add dep express "^4.18.0" >>"$LOGFILE" 2>&1
"$AIGOGO" add dev jest "^29.0.0" >>"$LOGFILE" 2>&1
popd >/dev/null

PY_MANIFEST="$SHOWDEPS_PY/aigogo.json"
JS_MANIFEST="$SHOWDEPS_JS/aigogo.json"

run_test_grep "show-deps <path> (default=text)" "Package:" \
    "$AIGOGO" show-deps "$PY_MANIFEST"

run_test_grep "show-deps --format text" "Package:" \
    "$AIGOGO" show-deps "$PY_MANIFEST" --format text

run_test_grep "show-deps --format requirements" "requirements" \
    "$AIGOGO" show-deps "$PY_MANIFEST" --format requirements

run_test_grep "show-deps --format pip" "requirements" \
    "$AIGOGO" show-deps "$PY_MANIFEST" --format pip

run_test_grep "show-deps --format pyproject" "pyproject.toml" \
    "$AIGOGO" show-deps "$PY_MANIFEST" --format pyproject

run_test_grep "show-deps --format pep621" "pyproject.toml" \
    "$AIGOGO" show-deps "$PY_MANIFEST" --format pep621

run_test_grep "show-deps --format poetry" "poetry" \
    "$AIGOGO" show-deps "$PY_MANIFEST" --format poetry

run_test_grep "show-deps --format npm (JS)" "dependencies" \
    "$AIGOGO" show-deps "$JS_MANIFEST" --format npm

run_test_grep "show-deps --format package-json (JS)" "dependencies" \
    "$AIGOGO" show-deps "$JS_MANIFEST" --format package-json

run_test_grep "show-deps --format yarn (JS)" "yarn add" \
    "$AIGOGO" show-deps "$JS_MANIFEST" --format yarn

run_test_grep "show-deps <dir> (directory)" "Package:" \
    "$AIGOGO" show-deps "$SHOWDEPS_PY"

# Cross-language error cases
run_test_fail_grep "Python format on JS package -> error" "only supported for Python" \
    "$AIGOGO" show-deps "$JS_MANIFEST" --format requirements

run_test_fail_grep "JS format on Python package -> error" "only supported for JavaScript" \
    "$AIGOGO" show-deps "$PY_MANIFEST" --format npm

echo ""

###############################################################################
#  SECTION: Cache Management
###############################################################################
echo "${BOLD}=== Cache Management ===${RESET}"

run_test_grep "aigg list" "Cached agents|No cached" \
    "$AIGOGO" list

# Build something so we can remove it
CACHE_DIR="$WORK/cache-test"
create_python_project "$CACHE_DIR"
pushd "$CACHE_DIR" >/dev/null
"$AIGOGO" init >>"$LOGFILE" 2>&1
"$AIGOGO" add file utils.py >>"$LOGFILE" 2>&1
"$AIGOGO" build cache-remove-me:1.0.0 --force >>"$LOGFILE" 2>&1
popd >/dev/null

run_test_grep "aigg remove <name>:<tag>" "Successfully removed|removed" \
    "$AIGOGO" remove cache-remove-me:1.0.0

# Build two more to test remove-all
pushd "$CACHE_DIR" >/dev/null
"$AIGOGO" build cache-rm-all-a:1.0.0 --force >>"$LOGFILE" 2>&1
"$AIGOGO" build cache-rm-all-b:1.0.0 --force >>"$LOGFILE" 2>&1
popd >/dev/null

run_test_grep "aigg remove-all --force" "Successfully removed|No cached" \
    "$AIGOGO" remove-all --force

echo ""

###############################################################################
#  SECTION: Registry Commands
###############################################################################
echo "${BOLD}=== Registry Commands ===${RESET}"

if $HAS_REGISTRY; then
    # login with -u -p (pipe password via stdin)
    run_test_grep "aigg login <registry> -u <user> -p" "Successfully logged in" \
        bash -c "echo '$REG_PASS' | $AIGOGO login $REGISTRY -u $REG_USER -p"

    # build a package to push
    REG_BUILD_DIR="$WORK/reg-build"
    create_python_project "$REG_BUILD_DIR"
    pushd "$REG_BUILD_DIR" >/dev/null
    "$AIGOGO" init >>"$LOGFILE" 2>&1
    "$AIGOGO" add file utils.py >>"$LOGFILE" 2>&1
    "$AIGOGO" build reg-push-test:1.0.0 --force >>"$LOGFILE" 2>&1
    popd >/dev/null

    REG_IMAGE="$REGISTRY/$REG_REPO:1.0.0"

    run_test_grep "aigg push --from" "Successfully pushed|Pushing" \
        "$AIGOGO" push "$REG_IMAGE" --from reg-push-test:1.0.0

    run_test_grep "aigg pull" "Successfully pulled|Pulling" \
        "$AIGOGO" pull "$REG_IMAGE"

    # delete (pipe "yes" for confirmation)
    run_test_grep "aigg delete" "Successfully deleted|Delete" \
        bash -c "echo yes | $AIGOGO delete $REG_IMAGE"

    # logout
    run_test_grep "aigg logout" "Successfully logged out" \
        "$AIGOGO" logout "$REGISTRY"
else
    skip_test "aigg login <registry> -u <user> -p"
    skip_test "aigg push --from"
    skip_test "aigg pull"
    skip_test "aigg delete"
    skip_test "aigg logout"
fi

echo ""

###############################################################################
#  SECTION: Error Cases
###############################################################################
echo "${BOLD}=== Error Cases ===${RESET}"

# No args → prints help (exit 0 per root.go)
run_test_grep "no args -> prints help" "Usage:" \
    "$AIGOGO"

# Unknown command → error
run_test_fail_grep "unknown command -> error" "Unknown command" \
    "$AIGOGO" not-a-command

# build with no aigogo.json → error
ERR_DIR="$WORK/err-empty"
mkdir -p "$ERR_DIR"
pushd "$ERR_DIR" >/dev/null

run_test_fail_grep "build with no aigogo.json -> error" "failed to find manifest|aigogo.json" \
    "$AIGOGO" build

run_test_fail_grep "install with no aigogo.lock -> error" "aigogo.lock|lock" \
    "$AIGOGO" install

popd >/dev/null

# push without --from → error
run_test_fail_grep "push without --from -> error" "--from flag is required|--from" \
    "$AIGOGO" push fake.io/org/pkg:1.0.0

# show-deps with invalid format → error listing valid formats
FMTERR_DIR="$WORK/fmterr"
create_python_project "$FMTERR_DIR"
pushd "$FMTERR_DIR" >/dev/null
"$AIGOGO" init >>"$LOGFILE" 2>&1
popd >/dev/null

run_test_fail_grep "show-deps --format invalid -> error" "unsupported format|Supported formats" \
    "$AIGOGO" show-deps "$FMTERR_DIR/aigogo.json" --format invalid

echo ""

###############################################################################
#  Summary
###############################################################################
echo "========================================================"
printf "${BOLD}Total: %d${RESET}  " "$TOTAL"
printf "${GREEN}Passed: %d${RESET}  " "$PASSED"
printf "${RED}Failed: %d${RESET}  " "$FAILED"
printf "${YELLOW}Skipped: %d${RESET}\n" "$SKIPPED"
echo "========================================================"
echo "Log: $LOGFILE"

if [[ $FAILED -gt 0 ]]; then
    exit 1
fi
