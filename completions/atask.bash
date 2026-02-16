#!/bin/bash
# Bash completion for atask

_denote_tasks_completions() {
    local cur prev words cword
    _get_comp_words_by_ref -n : cur prev words cword

    local prog="${words[0]}"
    
    # Helper function to get task IDs
    _get_task_ids() {
        "$prog" completion task-ids 2>/dev/null
    }
    
    # Helper function to get project IDs/names
    _get_project_ids() {
        "$prog" completion project-ids 2>/dev/null | cut -d: -f1
    }
    
    # Helper function to get project names for display
    _get_project_names() {
        "$prog" completion project-ids 2>/dev/null
    }
    
    # Helper function to get areas
    _get_areas() {
        "$prog" completion areas 2>/dev/null
    }
    
    # Helper function to get tags
    _get_tags() {
        "$prog" completion tags 2>/dev/null
    }

    # Global flags available everywhere
    local global_flags="--config --dir --json --no-color --quiet -q --area --tui -t --help --version"

    # Main command - check if it's the first word after the program name
    if [[ $cword -eq 1 ]]; then
        # Task commands (implicit) + other commands
        COMPREPLY=($(compgen -W "new list update done log edit delete project completion $global_flags" -- "$cur"))
        return
    fi

    # Check what command we're completing
    local cmd=""
    local subcmd=""
    
    for ((i=1; i<cword; i++)); do
        case "${words[i]}" in
            # Skip global flags
            --config|--dir|--area|-t|--tui|-q|--quiet|--json|--no-color)
                # Skip the flag and its argument if needed
                if [[ "${words[i]}" =~ ^--(config|dir|area)$ ]]; then
                    ((i++))
                fi
                ;;
            # Commands
            new|list|update|done|log|edit|delete|project|completion)
                if [[ -z "$cmd" ]]; then
                    cmd="${words[i]}"
                else
                    subcmd="${words[i]}"
                fi
                ;;
        esac
    done

    # Handle completion based on command
    case "$cmd" in
        # Task commands (implicit)
        new)
            case "$prev" in
                -p|--priority)
                    COMPREPLY=($(compgen -W "p1 p2 p3" -- "$cur"))
                    ;;
                --due)
                    COMPREPLY=($(compgen -W "today tomorrow monday tuesday wednesday thursday friday saturday sunday" -- "$cur"))
                    ;;
                --area)
                    local areas=$(_get_areas)
                    COMPREPLY=($(compgen -W "$areas" -- "$cur"))
                    ;;
                --project)
                    local projects=$(_get_project_ids)
                    COMPREPLY=($(compgen -W "$projects" -- "$cur"))
                    ;;
                --estimate)
                    COMPREPLY=($(compgen -W "1 2 3 5 8 13 21" -- "$cur"))
                    ;;
                *)
                    COMPREPLY=($(compgen -W "-p --priority --due --area --project --estimate --tags $global_flags" -- "$cur"))
                    ;;
            esac
            ;;
            
        list)
            case "$prev" in
                --status)
                    COMPREPLY=($(compgen -W "open done paused delegated dropped" -- "$cur"))
                    ;;
                -p|--priority)
                    COMPREPLY=($(compgen -W "p1 p2 p3" -- "$cur"))
                    ;;
                --area)
                    local areas=$(_get_areas)
                    COMPREPLY=($(compgen -W "$areas" -- "$cur"))
                    ;;
                --project)
                    local projects=$(_get_project_ids)
                    COMPREPLY=($(compgen -W "$projects" -- "$cur"))
                    ;;
                -s|--sort)
                    COMPREPLY=($(compgen -W "modified priority due created" -- "$cur"))
                    ;;
                *)
                    COMPREPLY=($(compgen -W "-a --all --area --status -p --priority --project --overdue --soon -s --sort -r --reverse $global_flags" -- "$cur"))
                    ;;
            esac
            ;;
            
        update)
            case "$prev" in
                -p|--priority)
                    COMPREPLY=($(compgen -W "p1 p2 p3" -- "$cur"))
                    ;;
                --due)
                    COMPREPLY=($(compgen -W "today tomorrow monday tuesday wednesday thursday friday saturday sunday" -- "$cur"))
                    ;;
                --area)
                    local areas=$(_get_areas)
                    COMPREPLY=($(compgen -W "$areas" -- "$cur"))
                    ;;
                --project)
                    local projects=$(_get_project_ids)
                    COMPREPLY=($(compgen -W "$projects" -- "$cur"))
                    ;;
                --status)
                    COMPREPLY=($(compgen -W "open done paused delegated dropped" -- "$cur"))
                    ;;
                --estimate)
                    COMPREPLY=($(compgen -W "1 2 3 5 8 13 21" -- "$cur"))
                    ;;
                *)
                    # Check if we've already got flags, then suggest task IDs
                    local has_flags=false
                    for word in "${words[@]:2}"; do
                        if [[ "$word" =~ ^- ]]; then
                            has_flags=true
                            break
                        fi
                    done
                    
                    if [[ "$has_flags" == true ]] || [[ "$cur" != -* ]]; then
                        local tasks=$(_get_task_ids)
                        COMPREPLY=($(compgen -W "$tasks" -- "$cur"))
                    else
                        COMPREPLY=($(compgen -W "-p --priority --due --area --project --status --estimate $global_flags" -- "$cur"))
                    fi
                    ;;
            esac
            ;;
            
        done|delete)
            # Always suggest task IDs
            local tasks=$(_get_task_ids)
            COMPREPLY=($(compgen -W "$tasks $global_flags" -- "$cur"))
            ;;
            
        log)
            # First argument should be task ID
            if [[ $cword -eq 2 ]] || [[ "$prev" == "log" ]]; then
                local tasks=$(_get_task_ids)
                COMPREPLY=($(compgen -W "$tasks" -- "$cur"))
            fi
            # After task ID, no completion (free text log message)
            ;;
            
        edit)
            # Task ID
            local tasks=$(_get_task_ids)
            COMPREPLY=($(compgen -W "$tasks $global_flags" -- "$cur"))
            ;;
            
        # Project command
        project)
            if [[ -z "$subcmd" ]]; then
                COMPREPLY=($(compgen -W "new list update tasks $global_flags" -- "$cur"))
                return
            fi
            
            case "$subcmd" in
                new)
                    case "$prev" in
                        -p|--priority)
                            COMPREPLY=($(compgen -W "p1 p2 p3" -- "$cur"))
                            ;;
                        --due|--start)
                            COMPREPLY=($(compgen -W "today tomorrow monday tuesday wednesday thursday friday saturday sunday" -- "$cur"))
                            ;;
                        --area)
                            local areas=$(_get_areas)
                            COMPREPLY=($(compgen -W "$areas" -- "$cur"))
                            ;;
                        *)
                            COMPREPLY=($(compgen -W "-p --priority --due --area --start --tags $global_flags" -- "$cur"))
                            ;;
                    esac
                    ;;
                    
                list)
                    case "$prev" in
                        --status)
                            COMPREPLY=($(compgen -W "active completed paused cancelled" -- "$cur"))
                            ;;
                        -p|--priority)
                            COMPREPLY=($(compgen -W "p1 p2 p3" -- "$cur"))
                            ;;
                        --area)
                            local areas=$(_get_areas)
                            COMPREPLY=($(compgen -W "$areas" -- "$cur"))
                            ;;
                        -s|--sort)
                            COMPREPLY=($(compgen -W "modified priority due created title" -- "$cur"))
                            ;;
                        *)
                            COMPREPLY=($(compgen -W "-a --all --area --status -p --priority -s --sort -r --reverse $global_flags" -- "$cur"))
                            ;;
                    esac
                    ;;
                    
                update)
                    case "$prev" in
                        -p|--priority)
                            COMPREPLY=($(compgen -W "p1 p2 p3" -- "$cur"))
                            ;;
                        --due)
                            COMPREPLY=($(compgen -W "today tomorrow monday tuesday wednesday thursday friday saturday sunday" -- "$cur"))
                            ;;
                        --area)
                            local areas=$(_get_areas)
                            COMPREPLY=($(compgen -W "$areas" -- "$cur"))
                            ;;
                        --status)
                            COMPREPLY=($(compgen -W "active completed paused cancelled" -- "$cur"))
                            ;;
                        *)
                            # Check if we need project IDs
                            local has_flags=false
                            for word in "${words[@]:3}"; do
                                if [[ "$word" =~ ^- ]]; then
                                    has_flags=true
                                    break
                                fi
                            done
                            
                            if [[ "$has_flags" == true ]] || [[ "$cur" != -* ]]; then
                                local projects=$(_get_project_ids)
                                COMPREPLY=($(compgen -W "$projects" -- "$cur"))
                            else
                                COMPREPLY=($(compgen -W "-p --priority --due --area --status $global_flags" -- "$cur"))
                            fi
                            ;;
                    esac
                    ;;
                    
                tasks)
                    # Project ID
                    local projects=$(_get_project_ids)
                    COMPREPLY=($(compgen -W "$projects $global_flags" -- "$cur"))
                    ;;
            esac
            ;;
            
        # Completion command
        completion)
            COMPREPLY=($(compgen -W "task-ids project-ids areas tags" -- "$cur"))
            ;;
    esac
}

complete -F _denote_tasks_completions atask