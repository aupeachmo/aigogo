package cmd

import (
	"fmt"
)

func completionCmd() *Command {
	return &Command{
		Name:        "completion",
		Description: "Generate shell completion scripts",
		Run: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: aigg completion <bash|zsh|fish>\n\nExamples:\n  # Bash\n  aigg completion bash | sudo tee /etc/bash_completion.d/aigg && sudo chmod 755 /etc/bash_completion.d/aigg\n  # or add to ~/.bashrc:\n  source <(aigg completion bash)\n\n  # Zsh\n  aigg completion zsh > ~/.zsh/completions/_aigg\n  # or add to ~/.zshrc:\n  source <(aigg completion zsh)\n\n  # Fish\n  aigg completion fish > ~/.config/fish/completions/aigg.fish")
			}

			shell := args[0]

			switch shell {
			case "bash":
				fmt.Print(bashCompletion)
			case "zsh":
				fmt.Print(zshCompletion)
			case "fish":
				fmt.Print(fishCompletion)
			default:
				return fmt.Errorf("unsupported shell: %s\nSupported shells: bash, zsh, fish", shell)
			}

			return nil
		},
	}
}

const bashCompletion = `# aigg bash completion script

_aigg_completions() {
    local cur prev words cword
    _init_completion || return

    # Main commands
    local commands="init add install uninstall rm validate scan build diff push pull login logout list show-deps remove remove-all delete search version completion"

    # Subcommands for add/rm
    local add_subcommands="file dep dev"
    local rm_subcommands="file dep dev"

    # Flags
    local build_flags="--force --no-validate"
    local push_flags="--from --dry-run"
    local diff_flags="--remote --summary"
    local delete_flags="--all"
    local add_file_flags="--force"
    local add_dep_flags="--from-pyproject"
    local add_dev_flags="--from-pyproject"
    local show_deps_flags="--format"
    local show_deps_formats="text pyproject pep621 poetry requirements pip npm package-json yarn"
    local login_flags="-u -p --dockerhub"

    # Get cached images for completion
    local cached_images=""
    if [ -d "$HOME/.aigogo/cache" ]; then
        cached_images=$(aigg list 2>/dev/null | grep -E '^[🔨📦]' | awk '{print $2}' || echo "")
    fi

    case $cword in
        1)
            # Complete main commands
            COMPREPLY=($(compgen -W "$commands" -- "$cur"))
            ;;
        2)
            # Complete subcommands or arguments based on previous command
            case $prev in
                add)
                    COMPREPLY=($(compgen -W "$add_subcommands" -- "$cur"))
                    ;;
                rm)
                    COMPREPLY=($(compgen -W "$rm_subcommands" -- "$cur"))
                    ;;
                completion)
                    COMPREPLY=($(compgen -W "bash zsh fish" -- "$cur"))
                    ;;
                remove)
                    # Complete with cached image names
                    COMPREPLY=($(compgen -W "$cached_images" -- "$cur"))
                    ;;
                remove-all)
                    # Complete with flags only
                    if [[ $cur == -* ]]; then
                        COMPREPLY=($(compgen -W "--force" -- "$cur"))
                    fi
                    ;;
                show-deps)
                    # Complete with files/directories or --format flag
                    if [[ $cur == -* ]]; then
                        COMPREPLY=($(compgen -W "$show_deps_flags" -- "$cur"))
                    else
                        COMPREPLY=($(compgen -f -- "$cur"))
                    fi
                    ;;
                build|push)
                    # Complete with cached images for reference
                    COMPREPLY=($(compgen -W "$cached_images" -- "$cur"))
                    ;;
                diff)
                    if [[ $cur == -* ]]; then
                        COMPREPLY=($(compgen -W "$diff_flags" -- "$cur"))
                    else
                        COMPREPLY=($(compgen -W "$cached_images" -- "$cur"))
                    fi
                    ;;
                *)
                    ;;
            esac
            ;;
        *)
            # Complete flags and additional arguments
            case ${words[1]} in
                add)
                    case ${words[2]} in
                        file)
                            # Complete file paths and --force flag
                            if [[ $cur == -* ]]; then
                                COMPREPLY=($(compgen -W "$add_file_flags" -- "$cur"))
                            else
                                COMPREPLY=($(compgen -f -- "$cur"))
                            fi
                            ;;
                        dep)
                            # Complete --from-pyproject flag
                            if [[ $cur == -* ]]; then
                                COMPREPLY=($(compgen -W "$add_dep_flags" -- "$cur"))
                            fi
                            ;;
                        dev)
                            # Complete --from-pyproject flag
                            if [[ $cur == -* ]]; then
                                COMPREPLY=($(compgen -W "$add_dev_flags" -- "$cur"))
                            fi
                            ;;
                        *)
                            ;;
                    esac
                    ;;
                rm)
                    case ${words[2]} in
                        file)
                            # Complete file paths
                            COMPREPLY=($(compgen -f -- "$cur"))
                            ;;
                        *)
                            ;;
                    esac
                    ;;
                build)
                    if [[ $cur == -* ]]; then
                        COMPREPLY=($(compgen -W "$build_flags" -- "$cur"))
                    fi
                    ;;
                push)
                    if [[ $cur == -* ]]; then
                        COMPREPLY=($(compgen -W "$push_flags" -- "$cur"))
                    elif [[ $prev == "--from" ]]; then
                        COMPREPLY=($(compgen -W "$cached_images" -- "$cur"))
                    fi
                    ;;
                diff)
                    if [[ $cur == -* ]]; then
                        COMPREPLY=($(compgen -W "$diff_flags" -- "$cur"))
                    else
                        COMPREPLY=($(compgen -W "$cached_images" -- "$cur"))
                    fi
                    ;;
                delete)
                    if [[ $cur == -* ]]; then
                        COMPREPLY=($(compgen -W "$delete_flags" -- "$cur"))
                    fi
                    ;;
                remove-all)
                    if [[ $cur == -* ]]; then
                        COMPREPLY=($(compgen -W "--force" -- "$cur"))
                    fi
                    ;;
                show-deps)
                    if [[ $cur == -* ]]; then
                        COMPREPLY=($(compgen -W "$show_deps_flags" -- "$cur"))
                    elif [[ $prev == "--format" ]]; then
                        COMPREPLY=($(compgen -W "$show_deps_formats" -- "$cur"))
                    else
                        COMPREPLY=($(compgen -f -- "$cur"))
                    fi
                    ;;
                login)
                    if [[ $cur == -* ]]; then
                        COMPREPLY=($(compgen -W "$login_flags" -- "$cur"))
                    fi
                    ;;
                *)
                    ;;
            esac
            ;;
    esac
}

complete -F _aigg_completions aigg
`

const zshCompletion = `#compdef aigg

_aigg() {
    local -a commands
    commands=(
        'init:Initialize a new aigogo package'
        'add:Add packages, files or dependencies'
        'install:Install packages from aigogo.lock'
        'uninstall:Remove installed packages and import configuration'
        'rm:Remove files or dependencies'
        'validate:Validate the manifest'
        'scan:Scan for dependencies'
        'build:Build a package locally'
        'diff:Compare package versions'
        'push:Push a package to registry'
        'pull:Pull a package from registry'
        'login:Login to a registry'
        'logout:Logout from a registry'
        'list:List cached packages'
        'show-deps:Show dependencies in various formats'
        'remove:Remove a cached package'
        'remove-all:Remove all cached packages'
        'delete:Delete a package from registry'
        'search:Search for packages'
        'version:Show version information'
        'completion:Generate completion scripts'
    )

    local -a add_subcommands
    add_subcommands=(
        'file:Add files to include list'
        'dep:Add runtime dependency'
        'dev:Add development dependency'
    )

    local -a rm_subcommands
    rm_subcommands=(
        'file:Remove files from include list'
        'dep:Remove runtime dependency'
        'dev:Remove development dependency'
    )

    local -a shells
    shells=('bash' 'zsh' 'fish')

    # Get cached images
    local -a cached_images
    if [[ -d "$HOME/.aigogo/cache" ]]; then
        cached_images=(${(f)"$(aigg list 2>/dev/null | grep -E '^[🔨📦]' | awk '{print $2}')"})
    fi

    case $state in
        command)
            _describe 'command' commands
            ;;
        *)
            case $words[2] in
                add)
                    if [[ $words[3] == "" ]] || [[ ${#words[@]} -eq 3 ]]; then
                        _describe 'subcommand' add_subcommands
                    elif [[ $words[3] == "file" ]]; then
                        if [[ $words[$CURRENT] == -* ]]; then
                            _arguments '--force[Add files even if ignored]'
                        else
                            _files
                        fi
                    elif [[ $words[3] == "dep" ]] || [[ $words[3] == "dev" ]]; then
                        if [[ $words[$CURRENT] == -* ]]; then
                            _arguments '--from-pyproject[Import from pyproject.toml]'
                        fi
                    fi
                    ;;
                rm)
                    if [[ $words[3] == "" ]] || [[ ${#words[@]} -eq 3 ]]; then
                        _describe 'subcommand' rm_subcommands
                    elif [[ $words[3] == "file" ]]; then
                        _files
                    fi
                    ;;
                completion)
                    _values 'shell' $shells
                    ;;
                remove)
                    _values 'cached images' $cached_images
                    ;;
                build|push)
                    _values 'image reference' $cached_images
                    ;;
                diff)
                    if [[ $words[$CURRENT] == -* ]]; then
                        _arguments '--remote[Compare against remote registry]' '--summary[Show compact summary only]'
                    else
                        _values 'image reference' $cached_images
                    fi
                    ;;
                show-deps)
                    if [[ $words[$CURRENT] == -* ]]; then
                        _arguments '--format[Output format]:format:(text pyproject pep621 poetry requirements pip npm package-json yarn)'
                    else
                        _files
                    fi
                    ;;
                login)
                    if [[ $words[$CURRENT] == -* ]]; then
                        _arguments '-u[Username]' '-p[Read password from stdin]' '--dockerhub[Use Docker Hub]'
                    fi
                    ;;
                *)
                    ;;
            esac
            ;;
    esac
}

_aigg "$@"
`

const fishCompletion = `# aigg fish completion script

# Main commands
complete -c aigg -f
complete -c aigg -n "__fish_use_subcommand" -a "init" -d "Initialize a new aigogo package"
complete -c aigg -n "__fish_use_subcommand" -a "add" -d "Add packages, files or dependencies"
complete -c aigg -n "__fish_use_subcommand" -a "install" -d "Install packages from aigogo.lock"
complete -c aigg -n "__fish_use_subcommand" -a "uninstall" -d "Remove installed packages and import configuration"
complete -c aigg -n "__fish_use_subcommand" -a "rm" -d "Remove files or dependencies"
complete -c aigg -n "__fish_use_subcommand" -a "validate" -d "Validate the manifest"
complete -c aigg -n "__fish_use_subcommand" -a "scan" -d "Scan for dependencies"
complete -c aigg -n "__fish_use_subcommand" -a "build" -d "Build a package locally"
complete -c aigg -n "__fish_use_subcommand" -a "diff" -d "Compare package versions"
complete -c aigg -n "__fish_use_subcommand" -a "push" -d "Push a package to registry"
complete -c aigg -n "__fish_use_subcommand" -a "pull" -d "Pull a package from registry"
complete -c aigg -n "__fish_use_subcommand" -a "login" -d "Login to a registry"
complete -c aigg -n "__fish_use_subcommand" -a "logout" -d "Logout from a registry"
complete -c aigg -n "__fish_use_subcommand" -a "list" -d "List cached packages"
complete -c aigg -n "__fish_use_subcommand" -a "show-deps" -d "Show dependencies in various formats"
complete -c aigg -n "__fish_use_subcommand" -a "remove" -d "Remove a cached package"
complete -c aigg -n "__fish_use_subcommand" -a "remove-all" -d "Remove all cached packages"
complete -c aigg -n "__fish_use_subcommand" -a "delete" -d "Delete a package from registry"
complete -c aigg -n "__fish_use_subcommand" -a "search" -d "Search for packages"
complete -c aigg -n "__fish_use_subcommand" -a "version" -d "Show version information"
complete -c aigg -n "__fish_use_subcommand" -a "completion" -d "Generate completion scripts"

# add subcommands
complete -c aigg -n "__fish_seen_subcommand_from add; and not __fish_seen_subcommand_from file dep dev" -a "file" -d "Add files to include list"
complete -c aigg -n "__fish_seen_subcommand_from add; and not __fish_seen_subcommand_from file dep dev" -a "dep" -d "Add runtime dependency"
complete -c aigg -n "__fish_seen_subcommand_from add; and not __fish_seen_subcommand_from file dep dev" -a "dev" -d "Add development dependency"

# rm subcommands
complete -c aigg -n "__fish_seen_subcommand_from rm; and not __fish_seen_subcommand_from file dep dev" -a "file" -d "Remove files from include list"
complete -c aigg -n "__fish_seen_subcommand_from rm; and not __fish_seen_subcommand_from file dep dev" -a "dep" -d "Remove runtime dependency"
complete -c aigg -n "__fish_seen_subcommand_from rm; and not __fish_seen_subcommand_from file dep dev" -a "dev" -d "Remove development dependency"

# completion shells
complete -c aigg -n "__fish_seen_subcommand_from completion" -a "bash zsh fish"

# Cached images for remove, build, push
function __aigg_cached_images
    if test -d "$HOME/.aigogo/cache"
        aigg list 2>/dev/null | grep -E '^[🔨📦]' | awk '{print $2}'
    end
end

complete -c aigg -n "__fish_seen_subcommand_from remove" -a "(__aigg_cached_images)" -d "Cached package"
complete -c aigg -n "__fish_seen_subcommand_from build" -a "(__aigg_cached_images)" -d "Package reference"
complete -c aigg -n "__fish_seen_subcommand_from diff" -a "(__aigg_cached_images)" -d "Package reference"
complete -c aigg -n "__fish_seen_subcommand_from push" -a "(__aigg_cached_images)" -d "Package reference"

# Flags
complete -c aigg -n "__fish_seen_subcommand_from build" -l "force" -d "Force rebuild"
complete -c aigg -n "__fish_seen_subcommand_from build" -l "no-validate" -d "Skip validation"
complete -c aigg -n "__fish_seen_subcommand_from diff" -l "remote" -d "Compare against remote registry"
complete -c aigg -n "__fish_seen_subcommand_from diff" -l "summary" -d "Show compact summary only"
complete -c aigg -n "__fish_seen_subcommand_from push" -l "from" -d "Push from local build"
complete -c aigg -n "__fish_seen_subcommand_from push" -l "dry-run" -d "Check if push is needed without pushing"
complete -c aigg -n "__fish_seen_subcommand_from delete" -l "all" -d "Delete all tags"
complete -c aigg -n "__fish_seen_subcommand_from remove-all" -l "force" -d "Skip confirmation"
complete -c aigg -n "__fish_seen_subcommand_from add; and __fish_seen_subcommand_from file" -l "force" -d "Add files even if ignored"
complete -c aigg -n "__fish_seen_subcommand_from add; and __fish_seen_subcommand_from dep" -l "from-pyproject" -d "Import from pyproject.toml"
complete -c aigg -n "__fish_seen_subcommand_from add; and __fish_seen_subcommand_from dev" -l "from-pyproject" -d "Import from pyproject.toml"
complete -c aigg -n "__fish_seen_subcommand_from show-deps" -l "format" -d "Output format" -a "text pyproject pep621 poetry requirements pip npm package-json yarn"
complete -c aigg -n "__fish_seen_subcommand_from login" -s "u" -l "username" -d "Username"
complete -c aigg -n "__fish_seen_subcommand_from login" -s "p" -d "Read password from stdin (prevents password in shell history)"
complete -c aigg -n "__fish_seen_subcommand_from login" -l "dockerhub" -d "Use Docker Hub (docker.io) as registry"

# Complete --from with cached images
complete -c aigg -n "__fish_seen_subcommand_from push; and __fish_seen_argument -l from" -a "(__aigg_cached_images)" -d "Local build"
`
