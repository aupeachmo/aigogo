# Shell Completion for aigg

This document explains how to set up and use shell tab completion for the `aigg` CLI.

## Overview

The `aigg completion` command generates shell completion scripts that enable:
- Auto-completion of commands (e.g., `init`, `build`, `push`)
- Auto-completion of subcommands (e.g., `add file`, `rm dep`)
- Auto-completion of flags (e.g., `--force`, `--output`)
- Smart completion of package names from your local cache
- File path completion for relevant commands

## Installation

### Bash

#### System-wide Installation (Recommended)

```bash
# Requires sudo
aigg completion bash | sudo tee /etc/bash_completion.d/aigg > /dev/null

# Reload shell
exec bash
```

#### User-specific Installation

```bash
# Create directory if needed
mkdir -p ~/.local/share/bash-completion/completions

# Install completion
aigg completion bash > ~/.local/share/bash-completion/completions/aigg

# Reload shell
exec bash
```

#### Session-only (For Testing)

```bash
# Load completion temporarily
source <(aigg completion bash)

# Or add to ~/.bashrc for automatic loading
echo 'source <(aigg completion bash)' >> ~/.bashrc
```

### Zsh

#### User Installation (Recommended)

```bash
# Create completions directory
mkdir -p ~/.zsh/completions

# Generate completion script
aigg completion zsh > ~/.zsh/completions/_aigg

# Add to ~/.zshrc if not already present
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
echo 'autoload -Uz compinit && compinit' >> ~/.zshrc

# Reload shell
exec zsh
```

#### Using oh-my-zsh

```bash
# Create custom plugins directory if needed
mkdir -p ~/.oh-my-zsh/custom/plugins/aigg

# Generate completion
aigg completion zsh > ~/.oh-my-zsh/custom/plugins/aigg/_aigg

# Add 'aigg' to plugins in ~/.zshrc
# plugins=(git ... aigg)

# Reload
exec zsh
```

### Fish

```bash
# Create completions directory if needed
mkdir -p ~/.config/fish/completions

# Generate completion script
aigg completion fish > ~/.config/fish/completions/aigg.fish

# Reload (or restart shell)
source ~/.config/fish/config.fish
```

## What Gets Completed

### Commands

When you type `aigg <TAB>`, you'll see:
```
init       add        install    rm         validate   scan
build      push       pull       login      logout     list
show-deps  remove     remove-all delete     search
version    completion
```

### Subcommands

#### For `aigg add <TAB>`:
- `file` - Add files to package
- `dep` - Add runtime dependency
- `dev` - Add development dependency

#### For `aigg rm <TAB>`:
- `file` - Remove files from package
- `dep` - Remove runtime dependency
- `dev` - Remove development dependency

#### For `aigg completion <TAB>`:
- `bash`
- `zsh`
- `fish`

### Cached Package Names

Commands that work with local packages will auto-complete from your cache:

```bash
aigg remove <TAB>           # Shows all cached packages
aigg push --from <TAB>      # Shows all cached packages
```

Example:
```
aigg remove <TAB>
# Shows: api-utils:1.0.0  helpers:2.0.0  test-utils:1.0.0
```

### Flags

Different commands have different flag completions:

```bash
aigg build <TAB>            # --force, --no-validate
aigg push <TAB>             # --from
aigg delete <TAB>           # --all
```

### File Paths

File-related commands complete file paths:

```bash
aigg add file <TAB>         # Completes file paths
aigg rm file <TAB>          # Completes file paths
```

## Usage Examples

### Complete Commands
```bash
$ aigg <TAB>
init       add        install    rm         validate   scan
build      push       pull       login      logout     ...
```

### Complete Subcommands
```bash
$ aigg add <TAB>
file  dep  dev

$ aigg add d<TAB>
dep  dev
```

### Complete Package Names
```bash
$ aigg remove <TAB>
api-utils:1.0.0  test-utils:1.0.0  helpers:2.0.0

$ aigg remove api<TAB>
api-utils:1.0.0
```

### Complete Flags
```bash
$ aigg build --<TAB>
--force  --no-validate
```

### Complete Flags with Values
```bash
$ aigg push --from <TAB>
api-utils:1.0.0  test-utils:1.0.0  helpers:2.0.0
```

## Troubleshooting

### Completion Not Working in Bash

1. Verify bash-completion is installed:
   ```bash
   # Ubuntu/Debian
   sudo apt install bash-completion
   
   # macOS
   brew install bash-completion
   ```

2. Ensure your `~/.bashrc` sources completion:
   ```bash
   # Add to ~/.bashrc if missing
   if [ -f /etc/bash_completion ]; then
       . /etc/bash_completion
   fi
   ```

3. Reload your shell:
   ```bash
   exec bash
   ```

### Completion Not Working in Zsh

1. Ensure `compinit` is called in your `~/.zshrc`:
   ```zsh
   autoload -Uz compinit && compinit
   ```

2. Check if completion directory is in fpath:
   ```zsh
   echo $fpath
   ```

3. Rebuild completion cache:
   ```zsh
   rm -f ~/.zcompdump
   compinit
   ```

### Completion Not Working in Fish

1. Check if the completion file exists:
   ```bash
   ls ~/.config/fish/completions/aigg.fish
   ```

2. Reload completions:
   ```fish
   source ~/.config/fish/completions/aigg.fish
   ```

### Package Names Not Completing

This can happen if:
- You have no cached packages (run `aigg list` to check)
- The `aigg` binary is not in your PATH
- Shell completion is parsing `aigg list` output incorrectly

To test:
```bash
# Should show your cached packages
aigg list | grep -E '^[🔨📦]' | awk '{print $2}'
```

## Updating Completion

If you update `aigg` or add new commands, regenerate the completion script:

```bash
# Bash
aigg completion bash | sudo tee /etc/bash_completion.d/aigg > /dev/null
exec bash

# Zsh
aigg completion zsh > ~/.zsh/completions/_aigg
rm -f ~/.zcompdump
exec zsh

# Fish
aigg completion fish > ~/.config/fish/completions/aigg.fish
source ~/.config/fish/config.fish
```

## Technical Details

### How It Works

The completion scripts use shell-specific completion APIs:
- **Bash**: Uses `complete -F` and `_init_completion`
- **Zsh**: Uses `#compdef` and `_describe`
- **Fish**: Uses `complete -c` with conditions

### Cached Package Lookup

When completing package names, the scripts:
1. Check if `~/.aigogo/cache` directory exists
2. Run `aigg list` to get all packages
3. Parse the output using grep and awk
4. Filter results based on what you've typed so far

This happens dynamically each time you press TAB, so newly built or removed packages are immediately reflected in completions.

### Performance

Completion scripts are optimized for speed:
- Package names are only queried when needed
- Cached results are used within a single completion session
- File system checks are minimal

On most systems, completion feels instant even with many cached packages.

