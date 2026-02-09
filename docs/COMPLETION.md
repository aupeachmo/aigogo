# Shell Completion for aigogo

This document explains how to set up and use shell tab completion for the `aigogo` CLI.

## Overview

The `aigogo completion` command generates shell completion scripts that enable:
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
aigogo completion bash | sudo tee /etc/bash_completion.d/aigogo > /dev/null

# Reload shell
exec bash
```

#### User-specific Installation

```bash
# Create directory if needed
mkdir -p ~/.local/share/bash-completion/completions

# Install completion
aigogo completion bash > ~/.local/share/bash-completion/completions/aigogo

# Reload shell
exec bash
```

#### Session-only (For Testing)

```bash
# Load completion temporarily
source <(aigogo completion bash)

# Or add to ~/.bashrc for automatic loading
echo 'source <(aigogo completion bash)' >> ~/.bashrc
```

### Zsh

#### User Installation (Recommended)

```bash
# Create completions directory
mkdir -p ~/.zsh/completions

# Generate completion script
aigogo completion zsh > ~/.zsh/completions/_aigogo

# Add to ~/.zshrc if not already present
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
echo 'autoload -Uz compinit && compinit' >> ~/.zshrc

# Reload shell
exec zsh
```

#### Using oh-my-zsh

```bash
# Create custom plugins directory if needed
mkdir -p ~/.oh-my-zsh/custom/plugins/aigogo

# Generate completion
aigogo completion zsh > ~/.oh-my-zsh/custom/plugins/aigogo/_aigogo

# Add 'aigogo' to plugins in ~/.zshrc
# plugins=(git ... aigogo)

# Reload
exec zsh
```

### Fish

```bash
# Create completions directory if needed
mkdir -p ~/.config/fish/completions

# Generate completion script
aigogo completion fish > ~/.config/fish/completions/aigogo.fish

# Reload (or restart shell)
source ~/.config/fish/config.fish
```

## What Gets Completed

### Commands

When you type `aigogo <TAB>`, you'll see:
```
init       add        install    rm         validate   scan
build      push       pull       login      logout     list
show-deps  remove     remove-all delete     search
version    completion
```

### Subcommands

#### For `aigogo add <TAB>`:
- `file` - Add files to package
- `dep` - Add runtime dependency
- `dev` - Add development dependency

#### For `aigogo rm <TAB>`:
- `file` - Remove files from package
- `dep` - Remove runtime dependency
- `dev` - Remove development dependency

#### For `aigogo completion <TAB>`:
- `bash`
- `zsh`
- `fish`

### Cached Package Names

Commands that work with local packages will auto-complete from your cache:

```bash
aigogo remove <TAB>           # Shows all cached packages
aigogo push --from <TAB>      # Shows all cached packages
```

Example:
```
aigogo remove <TAB>
# Shows: api-utils:1.0.0  helpers:2.0.0  test-utils:1.0.0
```

### Flags

Different commands have different flag completions:

```bash
aigogo build <TAB>            # --force, --no-validate
aigogo push <TAB>             # --from
aigogo delete <TAB>           # --all
```

### File Paths

File-related commands complete file paths:

```bash
aigogo add file <TAB>         # Completes file paths
aigogo rm file <TAB>          # Completes file paths
```

## Usage Examples

### Complete Commands
```bash
$ aigogo <TAB>
init       add        install    rm         validate   scan
build      push       pull       login      logout     ...
```

### Complete Subcommands
```bash
$ aigogo add <TAB>
file  dep  dev

$ aigogo add d<TAB>
dep  dev
```

### Complete Package Names
```bash
$ aigogo remove <TAB>
api-utils:1.0.0  test-utils:1.0.0  helpers:2.0.0

$ aigogo remove api<TAB>
api-utils:1.0.0
```

### Complete Flags
```bash
$ aigogo build --<TAB>
--force  --no-validate
```

### Complete Flags with Values
```bash
$ aigogo push --from <TAB>
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
   ls ~/.config/fish/completions/aigogo.fish
   ```

2. Reload completions:
   ```fish
   source ~/.config/fish/completions/aigogo.fish
   ```

### Package Names Not Completing

This can happen if:
- You have no cached packages (run `aigogo list` to check)
- The `aigogo` binary is not in your PATH
- Shell completion is parsing `aigogo list` output incorrectly

To test:
```bash
# Should show your cached packages
aigogo list | grep -E '^[🔨📦]' | awk '{print $2}'
```

## Updating Completion

If you update `aigogo` or add new commands, regenerate the completion script:

```bash
# Bash
aigogo completion bash | sudo tee /etc/bash_completion.d/aigogo > /dev/null
exec bash

# Zsh
aigogo completion zsh > ~/.zsh/completions/_aigogo
rm -f ~/.zcompdump
exec zsh

# Fish
aigogo completion fish > ~/.config/fish/completions/aigogo.fish
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
2. Run `aigogo list` to get all packages
3. Parse the output using grep and awk
4. Filter results based on what you've typed so far

This happens dynamically each time you press TAB, so newly built or removed packages are immediately reflected in completions.

### Performance

Completion scripts are optimized for speed:
- Package names are only queried when needed
- Cached results are used within a single completion session
- File system checks are minimal

On most systems, completion feels instant even with many cached packages.

