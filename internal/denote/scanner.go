package denote

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Scanner finds and loads Denote files
type Scanner struct {
	BaseDir string
}

// NewScanner creates a new scanner for the given directory
func NewScanner(dir string) *Scanner {
	return &Scanner{BaseDir: dir}
}

// FindAllTaskAndProjectFiles finds all task and project files in the directory
// This is primarily used for completion and scanning operations
func (s *Scanner) FindAllTaskAndProjectFiles() ([]File, error) {
	var allFiles []File
	
	// Find task files
	taskPattern := filepath.Join(s.BaseDir, "*__task*.md")
	taskPaths, err := filepath.Glob(taskPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob task files: %w", err)
	}
	
	// Find project files
	projectPattern := filepath.Join(s.BaseDir, "*__project*.md")
	projectPaths, err := filepath.Glob(projectPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob project files: %w", err)
	}
	
	// Combine paths
	allPaths := append(taskPaths, projectPaths...)
	
	parser := NewParser()
	for _, path := range allPaths {
		// Parse filename
		file, err := parser.ParseFilename(filepath.Base(path))
		if err != nil {
			// Skip non-Denote files
			continue
		}
		
		file.Path = path
		
		// Get file modification time
		if info, err := os.Stat(path); err == nil {
			file.ModTime = info.ModTime()
		}
		
		// Try to get title from frontmatter
		if metadata, err := parser.ParseFrontmatter(path); err == nil && metadata != nil {
			if title, ok := metadata["title"].(string); ok && title != "" {
				file.Title = title
			}
		}
		
		allFiles = append(allFiles, *file)
	}
	
	return allFiles, nil
}

// FindAllNotes is deprecated - use FindAllTaskAndProjectFiles instead
// Kept for backward compatibility during refactoring
func (s *Scanner) FindAllNotes() ([]File, error) {
	return s.FindAllTaskAndProjectFiles()
}

// FindTasks finds all task files in the directory
func (s *Scanner) FindTasks() ([]*Task, error) {
	pattern := filepath.Join(s.BaseDir, "*__task*.md")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob task files: %w", err)
	}

	var tasks []*Task
	for _, file := range files {
		task, err := ParseTaskFile(file)
		if err != nil {
			// Skip files that fail to parse
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// FindProjects finds all project files in the directory
func (s *Scanner) FindProjects() ([]*Project, error) {
	pattern := filepath.Join(s.BaseDir, "*__project*.md")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob project files: %w", err)
	}

	var projects []*Project
	for _, file := range files {
		project, err := ParseProjectFile(file)
		if err != nil {
			// Skip files that fail to parse
			continue
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// SortTasks sorts tasks by various criteria
func SortTasks(tasks []*Task, sortBy string, reverse bool) {
	switch sortBy {
	case "priority":
		sort.Slice(tasks, func(i, j int) bool {
			// P1 < P2 < P3 < no priority
			pi := priorityValue(tasks[i].Priority)
			pj := priorityValue(tasks[j].Priority)
			if pi != pj {
				return pi < pj
			}
			// Secondary sort by due date
			return tasks[i].DueDate < tasks[j].DueDate
		})
	
	case "due":
		sort.Slice(tasks, func(i, j int) bool {
			// Tasks with due dates come before those without
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
			// Open < Paused < Delegated < Done < Dropped
			si := statusValue(tasks[i].Status)
			sj := statusValue(tasks[j].Status)
			if si != sj {
				return si < sj
			}
			// Secondary sort by priority
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
// This function is kept for backward compatibility during refactoring
func SortFiles(files []File, sortBy string, reverse bool) {
	// Just delegate to SortTaskFiles with empty metadata maps
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
			// Get priorities for both files
			pi, pj := getPriority(files[i], taskMeta, projectMeta), getPriority(files[j], taskMeta, projectMeta)
			
			// Convert priority strings to numbers for comparison (p1=1, p2=2, p3=3, empty=4)
			piNum, pjNum := priorityToNumber(pi), priorityToNumber(pj)
			
			// Sort by priority first, then by date
			if piNum != pjNum {
				return piNum < pjNum
			}
			return files[i].ID < files[j].ID
		})
	case "due":
		sort.Slice(files, func(i, j int) bool {
			// Get due dates for both files
			di, dj := getDueDate(files[i], taskMeta, projectMeta), getDueDate(files[j], taskMeta, projectMeta)
			
			// Empty dates go to the end
			if di == "" && dj == "" {
				return files[i].ID < files[j].ID
			}
			if di == "" {
				return false
			}
			if dj == "" {
				return true
			}
			
			// Compare dates
			if di != dj {
				return di < dj
			}
			return files[i].ID < files[j].ID
		})
	case "project":
		sort.Slice(files, func(i, j int) bool {
			// Get project names for both files
			pi, pj := getProjectName(files[i], taskMeta, projectMeta), getProjectName(files[j], taskMeta, projectMeta)
			
			// Tasks without projects go to the end
			if pi == "" && pj == "" {
				return files[i].ID < files[j].ID
			}
			if pi == "" {
				return false
			}
			if pj == "" {
				return true
			}
			
			// Compare project names
			if pi != pj {
				return strings.ToLower(pi) < strings.ToLower(pj)
			}
			// Within same project, sort by due date
			di, dj := getDueDate(files[i], taskMeta, projectMeta), getDueDate(files[j], taskMeta, projectMeta)
			if di != dj && di != "" && dj != "" {
				return di < dj
			}
			return files[i].ID < files[j].ID
		})
	case "estimate":
		sort.Slice(files, func(i, j int) bool {
			// Get estimates for both files
			ei, ej := getEstimate(files[i], taskMeta), getEstimate(files[j], taskMeta)
			
			// Sort by estimate first, then by date
			if ei != ej {
				return ei < ej
			}
			return files[i].ID < files[j].ID
		})
	case "modified":
		sort.Slice(files, func(i, j int) bool {
			// If both have zero time, fall back to ID
			if files[i].ModTime.IsZero() && files[j].ModTime.IsZero() {
				return files[i].ID < files[j].ID
			}
			// Zero times go to the end
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

// Helper functions for sorting
func getPriority(file File, taskMeta map[string]*Task, projectMeta map[string]*Project) string {
	// Check cache first if available
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
	
	// No cache, read from disk
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
	// Check cache first if available
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
	
	// No cache, read from disk
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
	// Check cache first if available
	if taskMeta != nil {
		if task, ok := taskMeta[file.Path]; ok {
			return task.TaskMetadata.Estimate
		}
	}
	
	// No cache, read from disk
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
		return 4 // No priority goes last
	}
}

func getProjectName(file File, taskMeta map[string]*Task, projectMeta map[string]*Project) string {
	// If it's a project, return its own title
	if project, ok := projectMeta[file.Path]; ok {
		return project.ProjectMetadata.Title
	}
	
	// If it's a task, get the project it belongs to
	if task, ok := taskMeta[file.Path]; ok && task.TaskMetadata.ProjectID != "" {
		// Find the project by index_id
		for _, proj := range projectMeta {
			if strconv.Itoa(proj.IndexID) == task.TaskMetadata.ProjectID {
				return proj.ProjectMetadata.Title
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
		// Open, paused, or delegated tasks
		for _, task := range tasks {
			if task.Status == TaskStatusOpen || 
			   task.Status == TaskStatusPaused || 
			   task.Status == TaskStatusDelegated {
				filtered = append(filtered, task)
			}
		}
		
	case "area":
		// Filter by specific area
		for _, task := range tasks {
			if task.Area == filterValue {
				filtered = append(filtered, task)
			}
		}
		
	case "project":
		// Filter by specific project using Denote ID (v2.0.0)
		for _, task := range tasks {
			if task.ProjectID == filterValue {
				filtered = append(filtered, task)
			}
		}
		
	case "overdue":
		// Tasks with due dates in the past
		for _, task := range tasks {
			if task.DueDate != "" && IsOverdue(task.DueDate) && task.Status != TaskStatusDone {
				filtered = append(filtered, task)
			}
		}
		
	case "today":
		// Tasks due today
		today := time.Now().Format("2006-01-02")
		for _, task := range tasks {
			if task.DueDate == today && task.Status != TaskStatusDone {
				filtered = append(filtered, task)
			}
		}
		
	case "week":
		// Tasks due this week
		for _, task := range tasks {
			if task.DueDate != "" && IsDueThisWeek(task.DueDate) && task.Status != TaskStatusDone {
				filtered = append(filtered, task)
			}
		}
		
	case "priority":
		// Filter by specific priority
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