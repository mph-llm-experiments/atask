package task

import (
	"github.com/mph-llm-experiments/acore"
	"github.com/mph-llm-experiments/atask/internal/denote"
)

// UpdateTaskFile updates the task metadata in a file using acore.
func UpdateTaskFile(path string, task *denote.Task) error {
	task.Modified = acore.Now()
	return acore.UpdateFrontmatter(path, task)
}
