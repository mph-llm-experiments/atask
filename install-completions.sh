#!/bin/sh
# Install shell completions for atask

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
COMPLETIONS_DIR="$SCRIPT_DIR/completions"

# Detect shell if not specified
if [ -z "$1" ]; then
    # Try to detect from parent shell
    PARENT_SHELL=$(ps -p $PPID -o comm= 2>/dev/null | sed 's/^-//')
    case "$PARENT_SHELL" in
        *bash*)
            SHELL_TYPE="bash"
            ;;
        *zsh*)
            SHELL_TYPE="zsh"
            ;;
        *)
            # Fallback to $SHELL variable
            case "$SHELL" in
                */bash)
                    SHELL_TYPE="bash"
                    ;;
                */zsh)
                    SHELL_TYPE="zsh"
                    ;;
                *)
                    echo "Could not detect shell type. Please specify 'bash' or 'zsh' as argument."
                    echo "Usage: $0 [bash|zsh]"
                    exit 1
                    ;;
            esac
            ;;
    esac
    echo "Detected shell: $SHELL_TYPE"
else
    SHELL_TYPE="$1"
fi

install_bash_completions() {
    echo "Installing bash completions..."
    
    # Common bash completion directories
    local completion_dirs=(
        "/usr/local/etc/bash_completion.d"
        "/etc/bash_completion.d"
        "$HOME/.local/share/bash-completion/completions"
        "$HOME/.bash_completion.d"
    )
    
    local installed=false
    
    # Try to find a suitable directory
    for dir in "${completion_dirs[@]}"; do
        if [ -d "$dir" ] || mkdir -p "$dir" 2>/dev/null; then
            if cp "$COMPLETIONS_DIR/atask.bash" "$dir/atask" 2>/dev/null; then
                echo "Installed to $dir/atask"
                installed=true
                break
            fi
        fi
    done
    
    if [ "$installed" = false ]; then
        # Fallback: add to user's bashrc
        local bashrc="$HOME/.bashrc"
        local completion_file="$COMPLETIONS_DIR/atask.bash"
        
        echo "Could not install to system directories. Adding to $bashrc"
        echo "" >> "$bashrc"
        echo "# atask completion" >> "$bashrc"
        echo "[ -f $completion_file ] && source $completion_file" >> "$bashrc"
        echo "Added completion source to $bashrc"
    fi
    
    echo ""
    echo "Bash completion installed! Restart your shell or run:"
    echo "  source ~/.bashrc"
}

install_zsh_completions() {
    echo "Installing zsh completions..."
    
    # Common zsh completion directories
    local completion_dirs=(
        "/usr/local/share/zsh/site-functions"
        "/usr/share/zsh/site-functions"
        "$HOME/.zsh/completions"
        "$HOME/.config/zsh/completions"
    )
    
    # Add custom fpath directories if Oh My Zsh is installed
    if [ -d "$HOME/.oh-my-zsh" ]; then
        completion_dirs+=("$HOME/.oh-my-zsh/custom/plugins/atask")
        completion_dirs+=("$HOME/.oh-my-zsh/completions")
    fi
    
    local installed=false
    
    # Try to find a suitable directory
    for dir in "${completion_dirs[@]}"; do
        if [ -d "$dir" ] || mkdir -p "$dir" 2>/dev/null; then
            if cp "$COMPLETIONS_DIR/_atask" "$dir/_atask" 2>/dev/null; then
                echo "Installed to $dir/_atask"
                installed=true
                break
            fi
        fi
    done
    
    if [ "$installed" = false ]; then
        # Fallback: create directory and update fpath
        local zsh_completions="$HOME/.zsh/completions"
        mkdir -p "$zsh_completions"
        cp "$COMPLETIONS_DIR/_atask" "$zsh_completions/_atask"
        
        local zshrc="$HOME/.zshrc"
        echo "" >> "$zshrc"
        echo "# atask completion" >> "$zshrc"
        echo "fpath=($zsh_completions \$fpath)" >> "$zshrc"
        echo "autoload -Uz compinit && compinit" >> "$zshrc"
        
        echo "Installed to $zsh_completions/_atask"
        echo "Updated fpath in $zshrc"
    fi
    
    echo ""
    echo "Zsh completion installed! Restart your shell or run:"
    echo "  source ~/.zshrc"
    echo "  rm -f ~/.zcompdump && compinit"
}

case "$SHELL_TYPE" in
    bash)
        install_bash_completions
        ;;
    zsh)
        install_zsh_completions
        ;;
    *)
        echo "Unknown shell type: $SHELL_TYPE"
        echo "Usage: $0 [bash|zsh]"
        exit 1
        ;;
esac

echo ""
echo "Installation complete!"
echo ""
echo "The completions provide:"
echo "  - Command and subcommand suggestions"
echo "  - Flag and option completions"
echo "  - Dynamic task ID completion"
echo "  - Priority, status, and area value suggestions"
echo "  - Smart context-aware completions"