.PHONY: build install clean test qa fmt vet lint help build-all build-linux build-linux-arm build-darwin build-darwin-arm build-windows deps

# Binary name
BINARY=aigg

# Output directory
BIN_DIR=bin

# Version from git (fallback to 0.0.1 if no tags)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.0.1")

# Build flags for stripped binaries with version injection
LDFLAGS=-ldflags="-s -w -X main.Version=$(VERSION)"
BUILD_FLAGS=-trimpath $(LDFLAGS)

# Default target
.DEFAULT_GOAL := build

# Build the binary (stripped)
build:
	@echo "Building $(BINARY)..."
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY) .
	@strip $(BIN_DIR)/$(BINARY) 2>/dev/null || true
	@echo "Build complete: $(BIN_DIR)/$(BINARY)"
	@ls -lh $(BIN_DIR)/$(BINARY) | awk '{print "Size: " $$5}'
	@file $(BIN_DIR)/$(BINARY)

# Install to /usr/local/bin (with shell completion)
install: build
	@echo "Installing $(BINARY) to /usr/local/bin..."
	@sudo cp $(BIN_DIR)/$(BINARY) /usr/local/bin/
	@sudo chmod 755 /usr/local/bin/$(BINARY)
	@echo "✓ Binary installed"
	@echo ""
	@echo "Installing shell completion..."
	@if echo "$$SHELL" | grep -q bash; then \
		echo "  Detected bash shell"; \
		sudo $(BIN_DIR)/$(BINARY) completion bash > /tmp/aigg-completion-bash; \
		sudo mv /tmp/aigg-completion-bash /etc/bash_completion.d/aigg; \
		sudo chmod 644 /etc/bash_completion.d/aigg; \
		echo "  ✓ Bash completion installed to /etc/bash_completion.d/aigg"; \
		echo "  ✓ Permissions set to 644 (world readable)"; \
		echo ""; \
		echo "  Reload your shell or run: source /etc/bash_completion.d/aigg"; \
	elif echo "$$SHELL" | grep -q zsh; then \
		echo "  Detected zsh shell"; \
		mkdir -p ~/.zsh/completions; \
		$(BIN_DIR)/$(BINARY) completion zsh > ~/.zsh/completions/_aigg; \
		echo "  ✓ Zsh completion installed to ~/.zsh/completions/_aigg"; \
		if ! grep -q "fpath=(~/.zsh/completions" ~/.zshrc 2>/dev/null; then \
			echo "" >> ~/.zshrc; \
			echo "# aigg completion" >> ~/.zshrc; \
			echo "fpath=(~/.zsh/completions \$$fpath)" >> ~/.zshrc; \
			echo "autoload -Uz compinit && compinit" >> ~/.zshrc; \
			echo "  ✓ Updated ~/.zshrc"; \
		fi; \
		echo ""; \
		echo "  Reload your shell or run: exec zsh"; \
	else \
		echo "  Shell not detected or not supported (bash/zsh)"; \
		echo "  Run manually: aigg completion bash|zsh"; \
	fi
	@echo ""
	@echo "Installation complete. Run '$(BINARY) version' to verify."

# Install to ~/bin (no sudo required, with shell completion)
install-user: build
	@echo "Installing $(BINARY) to ~/bin..."
	@mkdir -p ~/bin
	@cp $(BIN_DIR)/$(BINARY) ~/bin/
	@echo "✓ Binary installed"
	@echo ""
	@echo "Installing shell completion..."
	@if echo "$$SHELL" | grep -q bash; then \
		echo "  Detected bash shell"; \
		mkdir -p ~/.local/share/bash-completion/completions; \
		$(BIN_DIR)/$(BINARY) completion bash > ~/.local/share/bash-completion/completions/aigg; \
		chmod 644 ~/.local/share/bash-completion/completions/aigg; \
		echo "  ✓ Bash completion installed to ~/.local/share/bash-completion/completions/aigg"; \
		echo ""; \
		echo "  Reload your shell or run: source ~/.local/share/bash-completion/completions/aigg"; \
	elif echo "$$SHELL" | grep -q zsh; then \
		echo "  Detected zsh shell"; \
		mkdir -p ~/.zsh/completions; \
		$(BIN_DIR)/$(BINARY) completion zsh > ~/.zsh/completions/_aigg; \
		echo "  ✓ Zsh completion installed to ~/.zsh/completions/_aigg"; \
		if ! grep -q "fpath=(~/.zsh/completions" ~/.zshrc 2>/dev/null; then \
			echo "" >> ~/.zshrc; \
			echo "# aigg completion" >> ~/.zshrc; \
			echo "fpath=(~/.zsh/completions \$$fpath)" >> ~/.zshrc; \
			echo "autoload -Uz compinit && compinit" >> ~/.zshrc; \
			echo "  ✓ Updated ~/.zshrc"; \
		fi; \
		echo ""; \
		echo "  Reload your shell or run: exec zsh"; \
	else \
		echo "  Shell not detected or not supported (bash/zsh)"; \
		echo "  Run manually: aigg completion bash|zsh"; \
	fi
	@echo ""
	@echo "Installation complete. Make sure ~/bin is in your PATH."
	@echo "Add this to your ~/.bashrc or ~/.zshrc if needed:"
	@echo '  export PATH="$$HOME/bin:$$PATH"'

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -f $(BINARY) $(BINARY)-* $(BINARY).exe aigg-before
	@rm -rf dist/
	@go clean
	@echo "Clean complete."

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run QA integration tests (local-only, no registry)
qa: build
	@echo "Running QA integration tests..."
	@AIGOGO=$(CURDIR)/$(BIN_DIR)/$(BINARY) ./qa/run.sh --local

# Format code
fmt:
	@echo "Formatting..."
	@gofmt -w .
	@echo "Format complete."

# Vet code
vet:
	@echo "Vetting..."
	@go vet ./...
	@echo "Vet complete."

# Lint (requires golangci-lint installed)
lint:
	@echo "Linting..."
	@golangci-lint run --timeout=5m
	@echo "Lint complete."

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated."

# Build for Linux AMD64
build-linux:
	@echo "Building for Linux AMD64..."
	@mkdir -p $(BIN_DIR)
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY)-linux-amd64 .
	@strip $(BIN_DIR)/$(BINARY)-linux-amd64 2>/dev/null || true
	@echo "Build complete: $(BIN_DIR)/$(BINARY)-linux-amd64"
	@ls -lh $(BIN_DIR)/$(BINARY)-linux-amd64 | awk '{print "Size: " $$5}'

# Build for Linux ARM64
build-linux-arm:
	@echo "Building for Linux ARM64..."
	@mkdir -p $(BIN_DIR)
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY)-linux-arm64 .
	@echo "Build complete: $(BIN_DIR)/$(BINARY)-linux-arm64"
	@ls -lh $(BIN_DIR)/$(BINARY)-linux-arm64 | awk '{print "Size: " $$5}'

# Build for macOS Intel
build-darwin:
	@echo "Building for macOS Intel..."
	@mkdir -p $(BIN_DIR)
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY)-darwin-amd64 .
	@echo "Build complete: $(BIN_DIR)/$(BINARY)-darwin-amd64"
	@ls -lh $(BIN_DIR)/$(BINARY)-darwin-amd64 | awk '{print "Size: " $$5}'

# Build for macOS ARM (Apple Silicon)
build-darwin-arm:
	@echo "Building for macOS ARM (Apple Silicon)..."
	@mkdir -p $(BIN_DIR)
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY)-darwin-arm64 .
	@echo "Build complete: $(BIN_DIR)/$(BINARY)-darwin-arm64"
	@ls -lh $(BIN_DIR)/$(BINARY)-darwin-arm64 | awk '{print "Size: " $$5}'

# Build for Windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BIN_DIR)
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY)-windows-amd64.exe .
	@echo "Build complete: $(BIN_DIR)/$(BINARY)-windows-amd64.exe"
	@ls -lh $(BIN_DIR)/$(BINARY)-windows-amd64.exe | awk '{print "Size: " $$5}'

# Build for all platforms
build-all: clean
	@echo "Building for all platforms..."
	@echo ""
	@$(MAKE) build-linux
	@echo ""
	@$(MAKE) build-linux-arm
	@echo ""
	@$(MAKE) build-darwin
	@echo ""
	@$(MAKE) build-darwin-arm
	@echo ""
	@$(MAKE) build-windows
	@echo ""
	@echo "✅ All builds complete:"
	@ls -lh $(BIN_DIR)/$(BINARY)-* 2>/dev/null | awk '{printf "  %-40s %s\n", $$9, $$5}'

# Show help
help:
	@echo "aigg v2.0 - Code Snippet Package Manager"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build              Build stripped binary for current platform"
	@echo "  build-linux        Build for Linux AMD64 (stripped)"
	@echo "  build-linux-arm    Build for Linux ARM64 (stripped)"
	@echo "  build-darwin       Build for macOS Intel (stripped)"
	@echo "  build-darwin-arm   Build for macOS ARM/Apple Silicon (stripped)"
	@echo "  build-windows      Build for Windows AMD64 (stripped)"
	@echo "  build-all          Build for all platforms"
	@echo "  install            Install to /usr/local/bin + shell completion (requires sudo)"
	@echo "  install-user       Install to ~/bin + shell completion (no sudo)"
	@echo "  clean              Remove build artifacts"
	@echo "  test               Run unit tests"
	@echo "  qa                 Run QA integration tests (builds first)"
	@echo "  fmt                Format code with gofmt"
	@echo "  vet                Run go vet"
	@echo "  lint               Run golangci-lint (must be installed)"
	@echo "  deps               Download and tidy dependencies"
	@echo "  help               Show this help message"
	@echo ""
	@echo "Build flags:"
	@echo "  - Binaries are built with -ldflags=\"-s -w\" (stripped)"
	@echo "  - CGO disabled for static binaries"
	@echo "  - Linux binaries additionally stripped with strip command"
	@echo "  - Typical size: 8-9 MB (stripped) vs 12-13 MB (unstripped)"
