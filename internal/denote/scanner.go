package denote

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mph-llm-experiments/acore"
)

// Scanner finds and loads task/project files
type Scanner struct {
	BaseDir string
}

// NewScanner creates a new scanner for the given directory
func NewScanner(dir string) *Scanner {
	return &Scanner{BaseDir: dir}
}

// FindAllTaskAndProjectFiles finds all task and project files and returns File views.
func (s *Scanner) FindAllTaskAndProjectFiles() ([]File, error) {
	var allFiles []File
	sc := &acore.Scanner{Dir: s.BaseDir}

	// Find task files
	taskPaths, err := sc.FindByType("task")
	if err != nil {
		return nil, err
	}
	for _, path := range taskPaths {
		task, err := ParseTaskFile(path)
		if err != nil {
			continue
		}
		allFiles = append(allFiles, FileFromTask(task))
	}

	// Find project files
	projectPaths, err := sc.FindByType("project")
	if err != nil {
		return nil, err
	}
	for _, path := range projectPaths {
		project, err := ParseProjectFile(path)
		if err != nil {
			continue
		}
		allFiles = append(allFiles, FileFromProject(project))
	}

	return allFiles, nil
}

// FindAllNotes is deprecated - use FindAllTaskAndProjectFiles instead
func (s *Scanner) FindAllNotes() ([]File, error) {
	return s.FindAllTaskAndProjectFiles()
}

// FindTasks finds all task files in the directory
func (s *Scanner) FindTasks() ([]*Task, error) {
	sc := &acore.Scanner{Dir: s.BaseDir}
	paths, err := sc.FindByType("task")
	if err != nil {
		return nil, err
	}

	var tasks []*Task
	for _, path := range paths {
		task, err := ParseTaskFile(path)
		if err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// FindProjects finds all project files in the directory
func (s *Scanner) FindProjects() ([]*Project, error) {
	sc := &acore.Scanner{Dir: s.BaseDir}
	paths, err := sc.FindByType("project")
	if err != nil {
		return nil, err
	}

	var projects []*Project
	for _, path := range paths {
		project, err := ParseProjectFile(path)
		if err != nil {
			continue
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// FindActions finds all action files in the queue/ subdirectory
func (s *Scanner) FindActions() ([]*Action, error) {
	queueDir := filepath.Join(s.BaseDir, "queue")

	// Ensure queue dir exists
	if _, err := os.Stat(queueDir); os.IsNotExist(err) {
		return nil, nil
	}

	sc := &acore.Scanner{Dir: queueDir}
	paths, err := sc.FindByType("action")
	if err != nil {
		return nil, err
	}

	var actions []*Action
	for _, path := range paths {
		action, err := ParseActionFile(path)
		if err != nil {
			continue
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// FindArchivedActions finds action files in the queue/archive/ subdirectory
func (s *Scanner) FindArchivedActions() ([]*Action, error) {
	archiveDir := filepath.Join(s.BaseDir, "queue", "archive")

	if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
		return nil, nil
	}

	sc := &acore.Scanner{Dir: archiveDir}
	paths, err := sc.FindByType("action")
	if err != nil {
		return nil, err
	}

	var actions []*Action
	for _, path := range paths {
		action, err := ParseActionFile(path)
		if err != nil {
			continue
		}
		actions = append(actions, action)
	}

	return actions, nil
}

// SortTasks sorts tasks by various criteria
func SortTasks(tasks []*Task, sortBy string, reverse bool) {
	switch sortBy {
	case "priority":
		sort.Slice(tasks, func(i, j int) bool {
			pi := priorityValue(tasks[i].Priority)
			pj := priorityValue(tasks[j].Priority)
			if pi != pj {
				return pi < pj
			}
			return tasks[i].DueDate < tasks[j].DueDate
		})

	case "due":
		sort.Slice(tasks, func(i, j int) bool {
			if tasks[i].DueDate == "" && tasks[j].DueDate != "" {
				return false
			}
			if tasks[i].DueDate != "" && tasks[j].DueDate == "" {
				return true
			}
			return tasks[i].DueDate < tasks[j].DueDate
		})

	case "status":
		sort.Slice(tasks, func(i, j int) bool {
			si := statusValue(tasks[i].Status)
			sj := statusValue(tasks[j].Status)
			if si != sj {
				return si < sj
			}
			return priorityValue(tasks[i].Priority) < priorityValue(tasks[j].Priority)
		})

	case "id":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].IndexID < tasks[j].IndexID
		})

	case "created":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].ID < tasks[j].ID
		})

	case "modified":
		fallthrough
	default:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].ModTime.After(tasks[j].ModTime)
		})
	}

	if reverse {
		reverseTaskSlice(tasks)
	}
}

// Helper functions for sorting

func priorityValue(p string) int {
	switch p {
	case PriorityP1:
		return 1
	case PriorityP2:
		return 2
	case PriorityP3:
		return 3
	default:
		return 4
	}
}

func statusValue(s string) int {
	switch s {
	case TaskStatusOpen:
		return 1
	case TaskStatusPaused:
		return 2
	case TaskStatusDelegated:
		return 3
	case TaskStatusDone:
		return 4
	case TaskStatusDropped:
		return 5
	default:
		return 6
	}
}

func reverseTaskSlice(tasks []*Task) {
	for i, j := 0, len(tasks)-1; i < j; i, j = i+1, j-1 {
		tasks[i], tasks[j] = tasks[j], tasks[i]
	}
}

// SortFiles is deprecated - use SortTaskFiles instead
func SortFiles(files []File, sortBy string, reverse bool) {
	SortTaskFiles(files, sortBy, reverse, make(map[string]*Task), make(map[string]*Project))
}

// SortTaskFiles sorts files with task metadata by various criteria
func SortTaskFiles(files []File, sortBy string, reverse bool, taskMeta map[string]*Task, projectMeta map[string]*Project) {
	switch sortBy {
	case "title":
		sort.Slice(files, func(i, j int) bool {
			return strings.ToLower(files[i].Title) < strings.ToLower(files[j].Title)
		})
	case "priority":
		sort.Slice(files, func(i, j int) bool {
			pi, pj := getPriority(files[i], taskMeta, projectMeta), getPriority(files[j], taskMeta, projectMeta)
			piNum, pjNum := priorityToNumber(pi), priorityToNumber(pj)
			if piNum != pjNum {
				return piNum < pjNum
			}
			return files[i].ID < files[j].ID
		})
	case "due":
		sort.Slice(files, func(i, j int) bool {
			di, dj := getDueDate(files[i], taskMeta, projectMeta), getDueDate(files[j], taskMeta, projectMeta)
			if di == "" && dj == "" {
				return files[i].ID < files[j].ID
			}
			if di == "" {
				return false
			}
			if dj == "" {
				return true
			}
			if di != dj {
				return di < dj
			}
			return files[i].ID < files[j].ID
		})
	case "project":
		sort.Slice(files, func(i, j int) bool {
			pi, pj := getProjectName(files[i], taskMeta, projectMeta), getProjectName(files[j], taskMeta, projectMeta)
			if pi == "" && pj == "" {
				return files[i].ID < files[j].ID
			}
			if pi == "" {
				return false
			}
			if pj == "" {
				return true
			}
			if pi != pj {
				return strings.ToLower(pi) < strings.ToLower(pj)
			}
			di, dj := getDueDate(files[i], taskMeta, projectMeta), getDueDate(files[j], taskMeta, projectMeta)
			if di != dj && di != "" && dj != "" {
				return di < dj
			}
			return files[i].ID < files[j].ID
		})
	case "estimate":
		sort.Slice(files, func(i, j int) bool {
			ei, ej := getEstimate(files[i], taskMeta), getEstimate(files[j], taskMeta)
			if ei != ej {
				return ei < ej
			}
			return files[i].ID < files[j].ID
		})
	case "modified":
		sort.Slice(files, func(i, j int) bool {
			if files[i].ModTime.IsZero() && files[j].ModTime.IsZero() {
				return files[i].ID < files[j].ID
			}
			if files[i].ModTime.IsZero() {
				return false
			}
			if files[j].ModTime.IsZero() {
				return true
			}
			return files[i].ModTime.After(files[j].ModTime)
		})
	case "created":
		fallthrough
	case "date":
		fallthrough
	default:
		sort.Slice(files, func(i, j int) bool {
			return files[i].ID < files[j].ID
		})
	}

	if reverse {
		reverseFileSlice(files)
	}
}

// Helper functions for file-based sorting
func getPriority(file File, taskMeta map[string]*Task, projectMeta map[string]*Project) string {
	if taskMeta != nil {
		if task, ok := taskMeta[file.Path]; ok {
			return task.TaskMetadata.Priority
		}
	}
	if projectMeta != nil {
		if project, ok := projectMeta[file.Path]; ok {
			return project.ProjectMetadata.Priority
		}
	}
	if file.IsTask() {
		if task, err := ParseTaskFile(file.Path); err == nil {
			return task.TaskMetadata.Priority
		}
	} else if file.IsProject() {
		if project, err := ParseProjectFile(file.Path); err == nil {
			return project.ProjectMetadata.Priority
		}
	}
	return ""
}

func getDueDate(file File, taskMeta map[string]*Task, projectMeta map[string]*Project) string {
	if taskMeta != nil {
		if task, ok := taskMeta[file.Path]; ok {
			return task.TaskMetadata.DueDate
		}
	}
	if projectMeta != nil {
		if project, ok := projectMeta[file.Path]; ok {
			return project.ProjectMetadata.DueDate
		}
	}
	if file.IsTask() {
		if task, err := ParseTaskFile(file.Path); err == nil {
			return task.TaskMetadata.DueDate
		}
	} else if file.IsProject() {
		if project, err := ParseProjectFile(file.Path); err == nil {
			return project.ProjectMetadata.DueDate
		}
	}
	return ""
}

func getEstimate(file File, taskMeta map[string]*Task) int {
	if taskMeta != nil {
		if task, ok := taskMeta[file.Path]; ok {
			return task.TaskMetadata.Estimate
		}
	}
	if file.IsTask() {
		if task, err := ParseTaskFile(file.Path); err == nil {
			return task.TaskMetadata.Estimate
		}
	}
	return 0
}

func priorityToNumber(priority string) int {
	switch priority {
	case "p1":
		return 1
	case "p2":
		return 2
	case "p3":
		return 3
	default:
		return 4
	}
}

func getProjectName(file File, taskMeta map[string]*Task, projectMeta map[string]*Project) string {
	if project, ok := projectMeta[file.Path]; ok {
		return project.Title
	}
	if task, ok := taskMeta[file.Path]; ok && task.TaskMetadata.ProjectID != "" {
		for _, proj := range projectMeta {
			if strconv.Itoa(proj.IndexID) == task.TaskMetadata.ProjectID {
				return proj.Title
			}
		}
	}
	return ""
}

func reverseFileSlice(files []File) {
	for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
		files[i], files[j] = files[j], files[i]
	}
}

// FilterTasks filters tasks based on various criteria
func FilterTasks(tasks []*Task, filterType string, filterValue string) []*Task {
	var filtered []*Task

	switch filterType {
	case "all":
		return tasks

	case "open":
		for _, task := range tasks {
			if task.Status == TaskStatusOpen {
				filtered = append(filtered, task)
			}
		}

	case "done":
		for _, task := range tasks {
			if task.Status == TaskStatusDone {
				filtered = append(filtered, task)
			}
		}

	case "active":
		for _, task := range tasks {
			if task.Status == TaskStatusOpen ||
				task.Status == TaskStatusPaused ||
				task.Status == TaskStatusDelegated {
				filtered = append(filtered, task)
			}
		}

	case "area":
		for _, task := range tasks {
			if task.Area == filterValue {
				filtered = append(filtered, task)
			}
		}

	case "project":
		for _, task := range tasks {
			if task.ProjectID == filterValue {
				filtered = append(filtered, task)
			}
		}

	case "overdue":
		for _, task := range tasks {
			if task.DueDate != "" && IsOverdue(task.DueDate) && task.Status != TaskStatusDone {
				filtered = append(filtered, task)
			}
		}

	case "today":
		today := time.Now().Format("2006-01-02")
		for _, task := range tasks {
			if task.DueDate == today && task.Status != TaskStatusDone {
				filtered = append(filtered, task)
			}
		}

	case "week":
		for _, task := range tasks {
			if task.DueDate != "" && IsDueThisWeek(task.DueDate) && task.Status != TaskStatusDone {
				filtered = append(filtered, task)
			}
		}

	case "priority":
		for _, task := range tasks {
			if task.Priority == filterValue {
				filtered = append(filtered, task)
			}
		}
	}

	return filtered
}

// GetUniqueAreas returns all unique areas from tasks
func GetUniqueAreas(tasks []*Task) []string {
	areaMap := make(map[string]bool)
	for _, task := range tasks {
		if task.Area != "" {
			areaMap[task.Area] = true
		}
	}

	var areas []string
	for area := range areaMap {
		areas = append(areas, area)
	}
	sort.Strings(areas)
	return areas
}

// GetUniqueProjectIDs returns all unique project IDs from tasks
func GetUniqueProjectIDs(tasks []*Task) []string {
	projectMap := make(map[string]bool)
	for _, task := range tasks {
		if task.ProjectID != "" {
			projectMap[task.ProjectID] = true
		}
	}

	var projectIDs []string
	for projectID := range projectMap {
		projectIDs = append(projectIDs, projectID)
	}
	sort.Strings(projectIDs)
	return projectIDs
}

// FindTaskFileByPath is a helper used by the scanner's FindAllTaskAndProjectFiles
// when constructing File view objects. It uses os.Stat to get modification time.
func findModTime(path string) time.Time {
	if info, err := os.Stat(path); err == nil {
		return info.ModTime()
	}
	return time.Time{}
}
