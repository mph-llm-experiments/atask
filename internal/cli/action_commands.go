package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mph-llm-experiments/acore"
	"github.com/mph-llm-experiments/atask/internal/config"
	"github.com/mph-llm-experiments/atask/internal/denote"
	"github.com/mph-llm-experiments/atask/internal/task"
)

// ActionCommand creates the action command with all subcommands
func ActionCommand(cfg *config.Config) *Command {
	cmd := &Command{
		Name:        "action",
		Usage:       "atask action <command> [options]",
		Description: "Manage the action queue",
	}

	cmd.Subcommands = []*Command{
		actionNewCommand(cfg),
		actionListCommand(cfg),
		actionShowCommand(cfg),
		actionUpdateCommand(cfg),
		actionApproveCommand(cfg),
		actionRejectCommand(cfg),
	}

	return cmd
}

// lookupAction finds an action by integer index_id or ULID string.
func lookupAction(dir string, identifier string) (*denote.Action, error) {
	if num, err := strconv.Atoi(identifier); err == nil {
		return task.FindActionByID(dir, num)
	}
	return task.FindActionByEntityID(dir, identifier)
}

// fieldFlag collects repeatable --field key=value flags
type fieldFlag struct {
	values map[string]string
}

func (f *fieldFlag) String() string { return "" }

func (f *fieldFlag) Set(val string) error {
	parts := strings.SplitN(val, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("field must be key=value, got: %s", val)
	}
	f.values[parts[0]] = parts[1]
	return nil
}

func actionNewCommand(cfg *config.Config) *Command {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	actionType := fs.String("action-type", "", "Action type (e.g. task_create, calendar_reschedule, or any plugin type)")
	proposedBy := fs.String("proposed-by", "cli", "Agent identifier")
	body := fs.String("body", "", "Reasoning/context for the action")
	fields := &fieldFlag{values: make(map[string]string)}
	fs.Var(fields, "field", "key=value field (repeatable)")

	return &Command{
		Name:        "new",
		Usage:       "atask action new <title> [options]",
		Description: "Create a proposed action",
		Flags:       fs,
		Run: func(cmd *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: atask action new <title> --action-type <type> [--field key=value ...]")
			}

			title := args[0]
			if *actionType == "" {
				return fmt.Errorf("--action-type is required")
			}

			bodyText := *body

			action, err := task.CreateAction(cfg.NotesDirectory, title, *actionType, *proposedBy, bodyText, fields.values)
			if err != nil {
				return err
			}

			if globalFlags.JSON {
				return printActionJSON(action)
			}

			if !globalFlags.Quiet {
				fmt.Printf("Created action #%d: %s\n", action.IndexID, action.Title)
			}
			return nil
		},
	}
}

func actionListCommand(cfg *config.Config) *Command {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	showAll := fs.Bool("all", false, "Show all actions including archived")

	return &Command{
		Name:        "list",
		Usage:       "atask action list [options]",
		Description: "List pending actions",
		Flags:       fs,
		Run: func(cmd *Command, args []string) error {
			scanner := denote.NewScanner(cfg.NotesDirectory)
			actions, err := scanner.FindActions()
			if err != nil {
				return err
			}

			if *showAll {
				archived, err := scanner.FindArchivedActions()
				if err != nil {
					return err
				}
				actions = append(actions, archived...)
			}

			// Filter to pending only unless --all
			if !*showAll {
				var pending []*denote.Action
				for _, a := range actions {
					if a.Status == denote.ActionPending {
						pending = append(pending, a)
					}
				}
				actions = pending
			}

			if globalFlags.JSON {
				return printActionsJSON(actions)
			}

			if len(actions) == 0 {
				if !globalFlags.Quiet {
					fmt.Println("No pending actions")
				}
				return nil
			}

			if !globalFlags.Quiet {
				fmt.Println("# Pending Actions")
			}
			for _, a := range actions {
				age := formatAge(a.ProposedAt)
				statusColor := color.New(color.FgYellow)
				if a.Status == denote.ActionExecuted {
					statusColor = color.New(color.FgGreen)
				} else if a.Status == denote.ActionFailed || a.Status == denote.ActionRejected {
					statusColor = color.New(color.FgRed)
				}

				if globalFlags.NoColor {
					fmt.Printf("  %d  %-16s %-40s (%s, %s)\n",
						a.IndexID, a.ActionType, a.Title, a.ProposedBy, age)
				} else {
					fmt.Printf("  %d  %-16s %-40s (%s, %s)\n",
						a.IndexID,
						statusColor.Sprint(a.ActionType),
						a.Title,
						a.ProposedBy,
						age)
				}
			}
			return nil
		},
	}
}

func actionShowCommand(cfg *config.Config) *Command {
	return &Command{
		Name:        "show",
		Usage:       "atask action show <id>",
		Description: "Show action details",
		Run: func(cmd *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: atask action show <id>")
			}

			action, err := lookupAction(cfg.NotesDirectory, args[0])
			if err != nil {
				return err
			}

			if globalFlags.JSON {
				type jsonAction struct {
					*denote.Action
					Content string `json:"content,omitempty"`
				}
				ja := jsonAction{Action: action, Content: action.Content}
				data, err := json.MarshalIndent(ja, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Printf("# %s (#%d)\n\n", action.Title, action.IndexID)
			fmt.Printf("  Action Type: %s\n", action.ActionType)
			fmt.Printf("  Status:      %s\n", action.Status)
			fmt.Printf("  Proposed By: %s\n", action.ProposedBy)
			fmt.Printf("  Proposed At: %s\n", action.ProposedAt)
			fmt.Println()

			if len(action.Fields) > 0 {
				fmt.Println("  Fields:")
				for k, v := range action.Fields {
					fmt.Printf("    %s: %s\n", k, v)
				}
				fmt.Println()
			}

			if action.Content != "" {
				fmt.Println("  Reasoning:")
				fmt.Printf("  %s\n", action.Content)
			}

			return nil
		},
	}
}

func actionUpdateCommand(cfg *config.Config) *Command {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	title := fs.String("title", "", "Update action title")
	actionType := fs.String("action-type", "", "Update action type")
	fields := &fieldFlag{values: make(map[string]string)}
	fs.Var(fields, "field", "key=value field (repeatable)")

	return &Command{
		Name:        "update",
		Usage:       "atask action update <id> [options]",
		Description: "Modify action fields before approval",
		Flags:       fs,
		Run: func(cmd *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: atask action update <id> [--field key=value ...]")
			}

			action, err := lookupAction(cfg.NotesDirectory, args[0])
			if err != nil {
				return err
			}

			if action.Status != denote.ActionPending {
				return fmt.Errorf("cannot update action with status: %s", action.Status)
			}

			changed := false

			if *title != "" {
				action.Title = *title
				changed = true
			}

			if *actionType != "" {
				action.ActionType = *actionType
				changed = true
			}

			for k, v := range fields.values {
				action.Fields[k] = v
				changed = true
			}

			if !changed {
				return fmt.Errorf("no changes specified")
			}

			action.Modified = acore.Now()
			if err := acore.UpdateFrontmatter(acore.NewLocalStore(filepath.Dir(action.FilePath)), filepath.Base(action.FilePath), action); err != nil {
				return fmt.Errorf("failed to update action: %w", err)
			}

			// Re-read to return updated state
			action, err = denote.ParseActionFile(action.FilePath)
			if err != nil {
				return err
			}

			if globalFlags.JSON {
				return printActionJSON(action)
			}

			if !globalFlags.Quiet {
				fmt.Printf("Updated action #%d\n", action.IndexID)
			}
			return nil
		},
	}
}

func actionApproveCommand(cfg *config.Config) *Command {
	return &Command{
		Name:        "approve",
		Usage:       "atask action approve <id>",
		Description: "Approve and execute the action",
		Run: func(cmd *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: atask action approve <id>")
			}

			action, err := lookupAction(cfg.NotesDirectory, args[0])
			if err != nil {
				return err
			}

			if action.Status != denote.ActionPending {
				return fmt.Errorf("cannot approve action with status: %s", action.Status)
			}

			// Execute the action directly — stay pending on failure so user can fix and retry
			result, execErr := executeAction(action)

			if execErr != nil {
				if globalFlags.JSON {
					errResult := map[string]interface{}{
						"status": "failed",
						"error":  execErr.Error(),
					}
					data, _ := json.MarshalIndent(errResult, "", "  ")
					fmt.Println(string(data))
				} else if !globalFlags.Quiet {
					fmt.Fprintf(os.Stderr, "Action failed: %s\n", execErr.Error())
				}
				return execErr
			}

			// Mark as executed and archive
			action.Status = denote.ActionExecuted
			action.Modified = acore.Now()
			if err := acore.UpdateFrontmatter(acore.NewLocalStore(filepath.Dir(action.FilePath)), filepath.Base(action.FilePath), action); err != nil {
				return fmt.Errorf("failed to update action status: %w", err)
			}

			if err := task.ArchiveAction(cfg.NotesDirectory, action); err != nil {
				return fmt.Errorf("failed to archive action: %w", err)
			}

			if globalFlags.JSON {
				resultMap := map[string]interface{}{
					"status": "executed",
					"result": string(result),
				}
				data, _ := json.MarshalIndent(resultMap, "", "  ")
				fmt.Println(string(data))
			} else if !globalFlags.Quiet {
				fmt.Printf("Action #%d executed successfully\n", action.IndexID)
			}

			return nil
		},
	}
}

func actionRejectCommand(cfg *config.Config) *Command {
	return &Command{
		Name:        "reject",
		Usage:       "atask action reject <id>",
		Description: "Reject and archive the action",
		Run: func(cmd *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: atask action reject <id>")
			}

			action, err := lookupAction(cfg.NotesDirectory, args[0])
			if err != nil {
				return err
			}

			if action.Status != denote.ActionPending {
				return fmt.Errorf("cannot reject action with status: %s", action.Status)
			}

			action.Status = denote.ActionRejected
			action.Modified = acore.Now()
			if err := acore.UpdateFrontmatter(acore.NewLocalStore(filepath.Dir(action.FilePath)), filepath.Base(action.FilePath), action); err != nil {
				return fmt.Errorf("failed to update action status: %w", err)
			}

			if err := task.ArchiveAction(cfg.NotesDirectory, action); err != nil {
				return fmt.Errorf("failed to archive action: %w", err)
			}

			if globalFlags.JSON {
				resultMap := map[string]interface{}{
					"status": "rejected",
				}
				data, _ := json.MarshalIndent(resultMap, "", "  ")
				fmt.Println(string(data))
			} else if !globalFlags.Quiet {
				fmt.Printf("Action #%d rejected\n", action.IndexID)
			}
			return nil
		},
	}
}

// executePlugin runs an external plugin script with JSON on stdin.
func executePlugin(pluginPath string, action *denote.Action) ([]byte, error) {
	input := map[string]interface{}{
		"action_type": action.ActionType,
		"title":       action.Title,
		"fields":      action.Fields,
	}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plugin input: %w", err)
	}

	cmd := exec.Command(pluginPath)
	cmd.Stdin = bytes.NewReader(inputJSON)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("plugin failed: %s\nStderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// pluginDir returns the path to the acore plugins directory (~/.config/acore/plugins).
func pluginDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "acore", "plugins")
}

// executeAction maps action_type + fields to a CLI command and runs it.
func executeAction(action *denote.Action) ([]byte, error) {
	// Try plugin first
	if dir := pluginDir(); dir != "" {
		pluginPath := filepath.Join(dir, action.ActionType)
		if info, err := os.Stat(pluginPath); err == nil && !info.IsDir() {
			return executePlugin(pluginPath, action)
		}
	}

	var bin string
	var args []string

	switch action.ActionType {
	case denote.ActionTypeTaskCreate:
		bin = "atask"
		title := action.Fields["title"]
		if title == "" {
			title = action.Title
		}
		args = []string{"new", title}
		addFieldFlag(action.Fields, &args, "priority", "--priority")
		addFieldFlag(action.Fields, &args, "due", "--due")
		addFieldFlag(action.Fields, &args, "area", "--area")
		addFieldFlag(action.Fields, &args, "project", "--project")
		addFieldFlag(action.Fields, &args, "tags", "--tags")
		addFieldFlag(action.Fields, &args, "estimate", "--estimate")
		addFieldFlag(action.Fields, &args, "recur", "--recur")
		// add_person handled as post-creation step below

	case denote.ActionTypeTaskUpdate:
		bin = "atask"
		targetID := action.Fields["target_id"]
		if targetID == "" {
			return nil, fmt.Errorf("task_update requires target_id field")
		}
		args = []string{"update"}
		addFieldFlag(action.Fields, &args, "title", "--title")
		addFieldFlag(action.Fields, &args, "status", "--status")
		addFieldFlag(action.Fields, &args, "priority", "--priority")
		addFieldFlag(action.Fields, &args, "due", "--due")
		addFieldFlag(action.Fields, &args, "area", "--area")
		addFieldFlag(action.Fields, &args, "project", "--project")
		addFieldFlag(action.Fields, &args, "plan_for", "--plan-for")
		addFieldFlag(action.Fields, &args, "add_person", "--add-person")
		args = append(args, targetID)

	case denote.ActionTypeIdeaCreate:
		bin = "anote"
		title := action.Fields["title"]
		if title == "" {
			title = action.Title
		}
		args = []string{"new", title}
		addFieldFlag(action.Fields, &args, "kind", "--kind")
		addFieldFlag(action.Fields, &args, "tags", "--tags")

	case denote.ActionTypeIdeaUpdate:
		bin = "anote"
		targetID := action.Fields["target_id"]
		if targetID == "" {
			return nil, fmt.Errorf("idea_update requires target_id field")
		}
		args = []string{"update", targetID}
		addFieldFlag(action.Fields, &args, "title", "--title")
		addFieldFlag(action.Fields, &args, "state", "--state")
		addFieldFlag(action.Fields, &args, "kind", "--kind")
		addFieldFlag(action.Fields, &args, "maturity", "--maturity")

	case denote.ActionTypePeopleUpdate:
		bin = "apeople"
		targetID := action.Fields["target_id"]
		if targetID == "" {
			return nil, fmt.Errorf("people_update requires target_id field")
		}
		args = []string{"update", targetID}
		addFieldFlag(action.Fields, &args, "state", "-state")
		addFieldFlag(action.Fields, &args, "plan_for", "-plan-for")

	case denote.ActionTypePeopleLog:
		bin = "apeople"
		targetID := action.Fields["target_id"]
		if targetID == "" {
			return nil, fmt.Errorf("people_log requires target_id field")
		}
		note := action.Fields["note"]
		if note == "" {
			return nil, fmt.Errorf("people_log requires note field")
		}
		args = []string{"log", targetID, note}
		addFieldFlag(action.Fields, &args, "interaction", "-interaction")

	default:
		return nil, fmt.Errorf("unknown action type: %s (no plugin found at %s)", action.ActionType, filepath.Join(pluginDir(), action.ActionType))
	}

	args = append(args, "--json", "--quiet")
	c := exec.Command(bin, args...)
	output, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command failed: %s\nOutput: %s", err, string(output))
	}

	// For task_create: if add_person is set, run a follow-up update to link people
	// (atask new doesn't support --add-person, only atask update does)
	if action.ActionType == denote.ActionTypeTaskCreate {
		if addPerson, ok := action.Fields["add_person"]; ok && addPerson != "" {
			// Parse index_id from the created task's JSON output
			var created struct {
				IndexID int `json:"index_id"`
			}
			if err := json.Unmarshal(output, &created); err == nil && created.IndexID > 0 {
				updateArgs := []string{"update"}
				addFieldFlag(action.Fields, &updateArgs, "add_person", "--add-person")
				updateArgs = append(updateArgs, fmt.Sprintf("%d", created.IndexID), "--json", "--quiet")
				uc := exec.Command("atask", updateArgs...)
				if updateOut, updateErr := uc.CombinedOutput(); updateErr != nil {
					// Non-fatal: task was created but linking failed
					return output, fmt.Errorf("task created but linking people failed: %s\nOutput: %s", updateErr, string(updateOut))
				}
			}
		}
	}

	return output, nil
}

func addFieldFlag(fields map[string]string, args *[]string, fieldName, flagName string) {
	if v, ok := fields[fieldName]; ok && v != "" {
		// Support comma-separated values for repeatable flags (e.g. add_person)
		if strings.Contains(v, ",") && strings.HasPrefix(flagName, "--add-") {
			for _, part := range strings.Split(v, ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					*args = append(*args, flagName, part)
				}
			}
		} else {
			*args = append(*args, flagName, v)
		}
	}
}

func appendToBody(filepath string, text string) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return
	}
	content = append(content, []byte(text)...)
	os.WriteFile(filepath, content, 0644)
}

func formatAge(proposedAt string) string {
	if proposedAt == "" {
		return "unknown"
	}

	t, err := time.Parse(time.RFC3339, proposedAt)
	if err != nil {
		return proposedAt
	}

	diff := time.Since(t)
	hours := int(diff.Hours())
	if hours < 1 {
		mins := int(diff.Minutes())
		if mins < 1 {
			return "just now"
		}
		return fmt.Sprintf("%dm ago", mins)
	}
	if hours < 24 {
		return fmt.Sprintf("%dh ago", hours)
	}
	days := hours / 24
	return fmt.Sprintf("%dd ago", days)
}

func printActionJSON(action *denote.Action) error {
	type jsonAction struct {
		ID         string            `json:"id"`
		IndexID    int               `json:"index_id"`
		Title      string            `json:"title"`
		Type       string            `json:"type"`
		ActionType string            `json:"action_type"`
		Status     string            `json:"status"`
		ProposedAt string            `json:"proposed_at"`
		ProposedBy string            `json:"proposed_by"`
		Fields     map[string]string `json:"fields"`
		Content    string            `json:"content,omitempty"`
		Created    string            `json:"created,omitempty"`
		Modified   string            `json:"modified,omitempty"`
	}

	ja := jsonAction{
		ID:         action.ID,
		IndexID:    action.IndexID,
		Title:      action.Title,
		Type:       action.Type,
		ActionType: action.ActionType,
		Status:     action.Status,
		ProposedAt: action.ProposedAt,
		ProposedBy: action.ProposedBy,
		Fields:     action.Fields,
		Content:    action.Content,
		Created:    action.Created,
		Modified:   action.Modified,
	}

	data, err := json.MarshalIndent(ja, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func printActionsJSON(actions []*denote.Action) error {
	type jsonAction struct {
		ID         string            `json:"id"`
		IndexID    int               `json:"index_id"`
		Title      string            `json:"title"`
		Type       string            `json:"type"`
		ActionType string            `json:"action_type"`
		Status     string            `json:"status"`
		ProposedAt string            `json:"proposed_at"`
		ProposedBy string            `json:"proposed_by"`
		Fields     map[string]string `json:"fields"`
		Content    string            `json:"content,omitempty"`
	}

	var items []jsonAction
	for _, a := range actions {
		items = append(items, jsonAction{
			ID:         a.ID,
			IndexID:    a.IndexID,
			Title:      a.Title,
			Type:       a.Type,
			ActionType: a.ActionType,
			Status:     a.Status,
			ProposedAt: a.ProposedAt,
			ProposedBy: a.ProposedBy,
			Fields:     a.Fields,
			Content:    a.Content,
		})
	}

	if items == nil {
		items = []jsonAction{}
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
