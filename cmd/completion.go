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
				return fmt.Errorf("usage: aigogo completion <bash|zsh|fish>\n\nExamples:\n  # Bash\n  aigogo completion bash | sudo tee /etc/bash_completion.d/aigogo && sudo chmod 755 /etc/bash_completion.d/aigogo\n  # or add to ~/.bashrc:\n  source <(aigogo completion bash)\n\n  # Zsh\n  aigogo completion zsh > ~/.zsh/completions/_aigogo\n  # or add to ~/.zshrc:\n  source <(aigogo completion zsh)\n\n  # Fish\n  aigogo completion fish > ~/.config/fish/completions/aigogo.fish")
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

const bashCompletion = `# aigogo bash completion script

_aigogo_completions() {
    local cur prev words cword
    _init_completion || return

    # Main commands
    local commands="init add install uninstall rm validate scan build push pull login logout list show-deps remove remove-all delete search version completion"

    # Subcommands for add/rm
    local add_subcommands="file dep dev"
    local rm_subcommands="file dep dev"

    # Flags
    local build_flags="--force --no-validate"
    local push_flags="--from"
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
        cached_images=$(aigogo list 2>/dev/null | grep -E '^[🔨📦]' | awk '{print $2}' || echo "")
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

complete -F _aigogo_completions aigogo
`

const zshCompletion = `#compdef aigogo

_aigogo() {
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
        cached_images=(${(f)"$(aigogo list 2>/dev/null | grep -E '^[🔨📦]' | awk '{print $2}')"})
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

_aigogo "$@"
`

const fishCompletion = `# aigogo fish completion script

# Main commands
complete -c aigogo -f
complete -c aigogo -n "__fish_use_subcommand" -a "init" -d "Initialize a new aigogo package"
complete -c aigogo -n "__fish_use_subcommand" -a "add" -d "Add packages, files or dependencies"
complete -c aigogo -n "__fish_use_subcommand" -a "install" -d "Install packages from aigogo.lock"
complete -c aigogo -n "__fish_use_subcommand" -a "uninstall" -d "Remove installed packages and import configuration"
complete -c aigogo -n "__fish_use_subcommand" -a "rm" -d "Remove files or dependencies"
complete -c aigogo -n "__fish_use_subcommand" -a "validate" -d "Validate the manifest"
complete -c aigogo -n "__fish_use_subcommand" -a "scan" -d "Scan for dependencies"
complete -c aigogo -n "__fish_use_subcommand" -a "build" -d "Build a package locally"
complete -c aigogo -n "__fish_use_subcommand" -a "push" -d "Push a package to registry"
complete -c aigogo -n "__fish_use_subcommand" -a "pull" -d "Pull a package from registry"
complete -c aigogo -n "__fish_use_subcommand" -a "login" -d "Login to a registry"
complete -c aigogo -n "__fish_use_subcommand" -a "logout" -d "Logout from a registry"
complete -c aigogo -n "__fish_use_subcommand" -a "list" -d "List cached packages"
complete -c aigogo -n "__fish_use_subcommand" -a "show-deps" -d "Show dependencies in various formats"
complete -c aigogo -n "__fish_use_subcommand" -a "remove" -d "Remove a cached package"
complete -c aigogo -n "__fish_use_subcommand" -a "remove-all" -d "Remove all cached packages"
complete -c aigogo -n "__fish_use_subcommand" -a "delete" -d "Delete a package from registry"
complete -c aigogo -n "__fish_use_subcommand" -a "search" -d "Search for packages"
complete -c aigogo -n "__fish_use_subcommand" -a "version" -d "Show version information"
complete -c aigogo -n "__fish_use_subcommand" -a "completion" -d "Generate completion scripts"

# add subcommands
complete -c aigogo -n "__fish_seen_subcommand_from add; and not __fish_seen_subcommand_from file dep dev" -a "file" -d "Add files to include list"
complete -c aigogo -n "__fish_seen_subcommand_from add; and not __fish_seen_subcommand_from file dep dev" -a "dep" -d "Add runtime dependency"
complete -c aigogo -n "__fish_seen_subcommand_from add; and not __fish_seen_subcommand_from file dep dev" -a "dev" -d "Add development dependency"

# rm subcommands
complete -c aigogo -n "__fish_seen_subcommand_from rm; and not __fish_seen_subcommand_from file dep dev" -a "file" -d "Remove files from include list"
complete -c aigogo -n "__fish_seen_subcommand_from rm; and not __fish_seen_subcommand_from file dep dev" -a "dep" -d "Remove runtime dependency"
complete -c aigogo -n "__fish_seen_subcommand_from rm; and not __fish_seen_subcommand_from file dep dev" -a "dev" -d "Remove development dependency"

# completion shells
complete -c aigogo -n "__fish_seen_subcommand_from completion" -a "bash zsh fish"

# Cached images for remove, build, push
function __aigogo_cached_images
    if test -d "$HOME/.aigogo/cache"
        aigogo list 2>/dev/null | grep -E '^[🔨📦]' | awk '{print $2}'
    end
end

complete -c aigogo -n "__fish_seen_subcommand_from remove" -a "(__aigogo_cached_images)" -d "Cached package"
complete -c aigogo -n "__fish_seen_subcommand_from build" -a "(__aigogo_cached_images)" -d "Package reference"
complete -c aigogo -n "__fish_seen_subcommand_from push" -a "(__aigogo_cached_images)" -d "Package reference"

# Flags
complete -c aigogo -n "__fish_seen_subcommand_from build" -l "force" -d "Force rebuild"
complete -c aigogo -n "__fish_seen_subcommand_from build" -l "no-validate" -d "Skip validation"
complete -c aigogo -n "__fish_seen_subcommand_from push" -l "from" -d "Push from local build"
complete -c aigogo -n "__fish_seen_subcommand_from delete" -l "all" -d "Delete all tags"
complete -c aigogo -n "__fish_seen_subcommand_from remove-all" -l "force" -d "Skip confirmation"
complete -c aigogo -n "__fish_seen_subcommand_from add; and __fish_seen_subcommand_from file" -l "force" -d "Add files even if ignored"
complete -c aigogo -n "__fish_seen_subcommand_from add; and __fish_seen_subcommand_from dep" -l "from-pyproject" -d "Import from pyproject.toml"
complete -c aigogo -n "__fish_seen_subcommand_from add; and __fish_seen_subcommand_from dev" -l "from-pyproject" -d "Import from pyproject.toml"
complete -c aigogo -n "__fish_seen_subcommand_from show-deps" -l "format" -d "Output format" -a "text pyproject pep621 poetry requirements pip npm package-json yarn"
complete -c aigogo -n "__fish_seen_subcommand_from login" -s "u" -l "username" -d "Username"
complete -c aigogo -n "__fish_seen_subcommand_from login" -s "p" -d "Read password from stdin (prevents password in shell history)"
complete -c aigogo -n "__fish_seen_subcommand_from login" -l "dockerhub" -d "Use Docker Hub (docker.io) as registry"

# Complete --from with cached images
complete -c aigogo -n "__fish_seen_subcommand_from push; and __fish_seen_argument -l from" -a "(__aigogo_cached_images)" -d "Local build"
`
