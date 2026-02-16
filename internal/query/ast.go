package query

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pdxmph/denote-tasks/internal/config"
	"github.com/pdxmph/denote-tasks/internal/denote"
)

// Node represents a node in the abstract syntax tree
type Node interface {
	Evaluate(task *denote.Task, cfg *config.Config) bool
	String() string
}

// ComparisonNode represents a field comparison (e.g., status:open, estimate>5)
type ComparisonNode struct {
	Field    string
	Operator string // ":", ">", "<", "=", "!="
	Value    string
}

func (n *ComparisonNode) String() string {
	return fmt.Sprintf("%s%s%s", n.Field, n.Operator, n.Value)
}

func (n *ComparisonNode) Evaluate(task *denote.Task, cfg *config.Config) bool {
	field := strings.ToLower(n.Field)
	value := strings.ToLower(n.Value)

	switch field {
	case "status":
		return compareString(strings.ToLower(task.TaskMetadata.Status), n.Operator, value)

	case "priority":
		return compareString(strings.ToLower(task.TaskMetadata.Priority), n.Operator, value)

	case "area":
		return compareString(strings.ToLower(task.TaskMetadata.Area), n.Operator, value)

	case "assignee":
		return compareString(strings.ToLower(task.TaskMetadata.Assignee), n.Operator, value)

	case "project_id":
		// Special values
		if value == "empty" {
			isEmpty := task.TaskMetadata.ProjectID == ""
			return n.Operator == ":" && isEmpty
		}
		if value == "set" {
			isSet := task.TaskMetadata.ProjectID != ""
			return n.Operator == ":" && isSet
		}
		return compareString(task.TaskMetadata.ProjectID, n.Operator, value)

	case "estimate":
		return compareInt(task.TaskMetadata.Estimate, n.Operator, value)

	case "index_id":
		return compareInt(task.TaskMetadata.IndexID, n.Operator, value)

	case "due", "due_date":
		// Special values
		switch value {
		case "empty":
			isEmpty := task.TaskMetadata.DueDate == ""
			return n.Operator == ":" && isEmpty
		case "set":
			isSet := task.TaskMetadata.DueDate != ""
			return n.Operator == ":" && isSet
		case "overdue":
			isOverdue := denote.IsOverdue(task.TaskMetadata.DueDate)
			return n.Operator == ":" && isOverdue
		case "today":
			daysUntil := denote.DaysUntilDue(task.TaskMetadata.DueDate)
			isToday := daysUntil == 0
			return n.Operator == ":" && isToday
		case "week":
			isThisWeek := denote.IsDueThisWeek(task.TaskMetadata.DueDate)
			return n.Operator == ":" && isThisWeek
		case "soon":
			isSoon := denote.IsDueSoon(task.TaskMetadata.DueDate, cfg.SoonHorizon)
			return n.Operator == ":" && isSoon
		default:
			// Compare as date string (YYYY-MM-DD)
			return compareString(task.TaskMetadata.DueDate, n.Operator, value)
		}

	case "start", "start_date":
		if value == "empty" {
			isEmpty := task.TaskMetadata.StartDate == ""
			return n.Operator == ":" && isEmpty
		}
		if value == "set" {
			isSet := task.TaskMetadata.StartDate != ""
			return n.Operator == ":" && isSet
		}
		return compareString(task.TaskMetadata.StartDate, n.Operator, value)

	case "title":
		return compareString(strings.ToLower(task.TaskMetadata.Title), n.Operator, value)

	case "tag", "tags":
		// Check if any tag matches
		for _, tag := range task.TaskMetadata.Tags {
			if compareString(strings.ToLower(tag), n.Operator, value) {
				return true
			}
		}
		return false

	case "content", "body", "text":
		// Search in file content (case-insensitive substring match)
		if n.Operator == ":" || n.Operator == "=" {
			return strings.Contains(strings.ToLower(task.Content), value)
		} else if n.Operator == "!=" {
			return !strings.Contains(strings.ToLower(task.Content), value)
		}
		return false

	default:
		// Unknown field always returns false
		return false
	}
}

// BooleanNode represents a boolean operation (AND, OR, NOT)
type BooleanNode struct {
	Op    string // "AND", "OR", "NOT"
	Left  Node
	Right Node // nil for NOT
}

func (n *BooleanNode) String() string {
	if n.Op == "NOT" {
		return fmt.Sprintf("NOT %s", n.Left)
	}
	return fmt.Sprintf("(%s %s %s)", n.Left, n.Op, n.Right)
}

func (n *BooleanNode) Evaluate(task *denote.Task, cfg *config.Config) bool {
	switch n.Op {
	case "AND":
		return n.Left.Evaluate(task, cfg) && n.Right.Evaluate(task, cfg)
	case "OR":
		return n.Left.Evaluate(task, cfg) || n.Right.Evaluate(task, cfg)
	case "NOT":
		return !n.Left.Evaluate(task, cfg)
	default:
		return false
	}
}

// Helper functions for comparison

func compareString(actual, operator, expected string) bool {
	switch operator {
	case ":":
		return actual == expected
	case "=":
		return actual == expected
	case "!=":
		return actual != expected
	default:
		return false
	}
}

func compareInt(actual int, operator, expectedStr string) bool {
	expected, err := strconv.Atoi(expectedStr)
	if err != nil {
		return false
	}

	switch operator {
	case ":", "=":
		return actual == expected
	case ">":
		return actual > expected
	case "<":
		return actual < expected
	case "!=":
		return actual != expected
	default:
		return false
	}
}
