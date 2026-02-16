# Shell Completions

atask includes comprehensive shell completion support for both bash and zsh.

## Installation

### Quick Install

Run the included installation script:

```bash
./install-completions.sh
```

This will detect your shell and install completions to the appropriate location.

### Manual Installation

#### Bash

Copy the completion file to one of these locations:
- `/usr/local/etc/bash_completion.d/atask`
- `/etc/bash_completion.d/atask`
- `~/.local/share/bash-completion/completions/atask`

Or source it in your `.bashrc`:
```bash
source /path/to/atask/completions/atask.bash
```

#### Zsh

Copy the completion file to a directory in your `fpath`:
- `/usr/local/share/zsh/site-functions/_atask`
- `~/.zsh/completions/_atask`

Or add the completions directory to your `fpath` in `.zshrc`:
```bash
fpath=(~/path/to/atask/completions $fpath)
autoload -Uz compinit && compinit
```

## Features

### Command Completion
- Main commands: `task`, `project`, `note`
- Subcommands for each entity type
- Legacy command aliases

### Smart Argument Completion
- **Task IDs**: Dynamic completion of existing task IDs
- **Project IDs**: Shows both ID and project name
- **Areas**: Completes from existing areas in your notes
- **Tags**: Completes from existing tags

### Flag Completion
- All flags and options with descriptions
- Context-aware value suggestions:
  - Priority: `p1`, `p2`, `p3`
  - Status: `open`, `done`, `paused`, `delegated`, `dropped`
  - Sort options: `modified`, `priority`, `due`, `created`
  - Due dates: `today`, `tomorrow`, weekday names

### Examples

```bash
# Complete commands
atask <TAB>
# Shows: task project note --tui --help --version

# Complete task subcommands
atask task <TAB>
# Shows: new list update done edit delete log

# Complete task IDs for done command
atask task done <TAB>
# Shows: 1 2 3 5 8 13 (actual task IDs)

# Complete priority values
atask task new "Fix bug" -p <TAB>
# Shows: p1 p2 p3

# Complete areas
atask task list -area <TAB>
# Shows: work personal hobby (from your existing tasks)

# Complete due date shortcuts
atask task new "Meeting" -due <TAB>
# Shows: today tomorrow monday tuesday wednesday...
```

## Advanced Usage

### Range and List Completion

The completions understand the smart task argument format:
- Single IDs: `3`
- Ranges: `3-5`
- Lists: `3,5,7`
- Mixed: `3,5-7,10`

### Dynamic Data

The completion system uses a special `completion` command to get dynamic data:

```bash
# Get all task IDs
atask completion task-ids

# Get all project IDs with names
atask completion project-ids

# Get all areas
atask completion areas

# Get all tags
atask completion tags
```

This ensures completions always reflect your current data.

## Troubleshooting

### Completions Not Working

1. **Reload your shell**: `source ~/.bashrc` or `source ~/.zshrc`
2. **Clear zsh cache**: `rm -f ~/.zcompdump && compinit`
3. **Check installation**: Verify the completion file is in the right location
4. **Test manually**: Try running `atask completion task-ids` to ensure it works

### Performance

Completions are cached per shell session. If you add new tasks/projects/areas, you may need to start a new shell to see them in completions.