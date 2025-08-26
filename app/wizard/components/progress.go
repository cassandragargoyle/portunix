package components

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/cassandragargoyle/portunix/app/wizard"
	"github.com/schollz/progressbar/v3"
)

// ProgressComponent displays progress during task execution
type ProgressComponent struct {
	Title     string
	Tasks     []wizard.Task
	Current   int
	TotalWork float64
	Progress  *progressbar.ProgressBar
	Spinner   *spinner.Spinner
}

// NewProgressComponent creates a new progress component
func NewProgressComponent(title string, tasks []wizard.Task) *ProgressComponent {
	totalWork := 0.0
	for _, task := range tasks {
		if task.Weight > 0 {
			totalWork += task.Weight
		} else {
			totalWork += 1.0
		}
	}

	return &ProgressComponent{
		Title:     title,
		Tasks:     tasks,
		Current:   0,
		TotalWork: totalWork,
	}
}

// Render executes tasks and displays progress
func (p *ProgressComponent) Render(ctx *wizard.WizardContext) error {
	if p.Title != "" {
		title := expandVariables(p.Title, ctx.Variables)
		fmt.Printf("\n%s\n", formatTitle(title, ctx.Theme))
		fmt.Println(strings.Repeat("=", len(title)))
	}

	// Create progress bar
	p.Progress = progressbar.NewOptions(100,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetDescription("[cyan][1/3][reset] Starting..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	completedWork := 0.0

	for i, task := range p.Tasks {
		// Check condition
		if task.Condition != "" && !evaluateCondition(task.Condition, ctx.Variables) {
			continue
		}

		// Update description
		desc := fmt.Sprintf("[cyan][%d/%d][reset] %s", i+1, len(p.Tasks), task.Label)
		p.Progress.Describe(desc)

		// Create spinner for current task
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" %s", task.Label)
		s.Start()

		// Execute command
		cmd := expandVariables(task.Command, ctx.Variables)
		err := executeCommand(cmd)
		
		s.Stop()

		if err != nil {
			fmt.Printf("❌ %s failed: %v\n", task.Label, err)
			return err
		}

		fmt.Printf("✅ %s completed\n", task.Label)

		// Update progress
		weight := task.Weight
		if weight == 0 {
			weight = 1.0
		}
		completedWork += weight
		progress := int((completedWork / p.TotalWork) * 100)
		p.Progress.Set(progress)
	}

	p.Progress.Finish()
	fmt.Println()
	return nil
}

// GetValue returns nil as progress doesn't collect values
func (p *ProgressComponent) GetValue() interface{} {
	return nil
}

// SetValue does nothing for progress component
func (p *ProgressComponent) SetValue(value interface{}) {}

// Validate always returns nil for progress
func (p *ProgressComponent) Validate() error {
	return nil
}

// executeCommand runs a shell command
func executeCommand(command string) error {
	var cmd *exec.Cmd
	if strings.Contains(command, "portunix") {
		// For portunix commands, use direct execution
		parts := strings.Fields(command)
		if len(parts) > 0 {
			cmd = exec.Command(parts[0], parts[1:]...)
		}
	} else {
		// For other commands, use shell
		if os.Getenv("OS") == "Windows_NT" {
			cmd = exec.Command("cmd", "/c", command)
		} else {
			cmd = exec.Command("sh", "-c", command)
		}
	}

	if cmd != nil {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return fmt.Errorf("invalid command: %s", command)
}

// expandVariables replaces {{variable}} with actual values
func expandVariables(str string, variables map[string]interface{}) string {
	result := str
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// evaluateCondition evaluates a simple condition
func evaluateCondition(condition string, variables map[string]interface{}) bool {
	// Simple implementation - can be extended
	parts := strings.Fields(condition)
	if len(parts) != 3 {
		return false
	}

	varName := parts[0]
	operator := parts[1]
	compareValue := parts[2]

	value, exists := variables[varName]
	if !exists {
		return false
	}

	switch operator {
	case "==":
		return fmt.Sprintf("%v", value) == compareValue
	case "!=":
		return fmt.Sprintf("%v", value) != compareValue
	case "true":
		if b, ok := value.(bool); ok {
			return b
		}
		return fmt.Sprintf("%v", value) == "true"
	case "false":
		if b, ok := value.(bool); ok {
			return !b
		}
		return fmt.Sprintf("%v", value) == "false"
	}

	return false
}