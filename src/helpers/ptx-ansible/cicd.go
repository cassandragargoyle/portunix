package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// PipelineStage represents a stage in a CI/CD pipeline
type PipelineStage string

const (
	StageBuild    PipelineStage = "build"
	StageTest     PipelineStage = "test"
	StageDeploy   PipelineStage = "deploy"
	StageValidate PipelineStage = "validate"
	StageCleanup  PipelineStage = "cleanup"
)

// PipelineStatus represents the status of a pipeline execution
type PipelineStatus string

const (
	StatusPending   PipelineStatus = "pending"
	StatusRunning   PipelineStatus = "running"
	StatusSuccess   PipelineStatus = "success"
	StatusFailed    PipelineStatus = "failed"
	StatusCancelled PipelineStatus = "cancelled"
	StatusSkipped   PipelineStatus = "skipped"
)

// CICDProvider represents different CI/CD systems
type CICDProvider string

const (
	ProviderGitHubActions CICDProvider = "github-actions"
	ProviderGitLabCI      CICDProvider = "gitlab-ci"
	ProviderJenkins       CICDProvider = "jenkins"
	ProviderAzureDevOps   CICDProvider = "azure-devops"
	ProviderGeneric       CICDProvider = "generic"
)

// PipelineDefinition defines a CI/CD pipeline
type PipelineDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Provider    CICDProvider           `json:"provider"`
	Triggers    []PipelineTrigger      `json:"triggers"`
	Stages      []PipelineStageConfig  `json:"stages"`
	Environment map[string]string      `json:"environment,omitempty"`
	Secrets     []string               `json:"secrets,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Timeout     time.Duration          `json:"timeout"`
	CreatedAt   time.Time              `json:"created_at"`
	CreatedBy   string                 `json:"created_by"`
}

// PipelineTrigger defines when a pipeline should be triggered
type PipelineTrigger struct {
	Type       string            `json:"type"`        // "push", "pr", "schedule", "manual"
	Branches   []string          `json:"branches,omitempty"`
	Paths      []string          `json:"paths,omitempty"`
	Schedule   string            `json:"schedule,omitempty"` // Cron format
	Conditions map[string]string `json:"conditions,omitempty"`
}

// PipelineStageConfig configures a pipeline stage
type PipelineStageConfig struct {
	Name         string                 `json:"name"`
	Stage        PipelineStage          `json:"stage"`
	PlaybookPath string                 `json:"playbook_path"`
	Environment  string                 `json:"environment"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	Depends      []string               `json:"depends,omitempty"`
	Condition    string                 `json:"condition,omitempty"`
	Timeout      time.Duration          `json:"timeout,omitempty"`
	AllowFailure bool                   `json:"allow_failure"`
	Parallel     bool                   `json:"parallel"`
}

// PipelineExecution represents a pipeline execution instance
type PipelineExecution struct {
	ID          string                    `json:"id"`
	PipelineID  string                    `json:"pipeline_id"`
	Status      PipelineStatus            `json:"status"`
	StartedAt   time.Time                 `json:"started_at"`
	FinishedAt  *time.Time                `json:"finished_at,omitempty"`
	Duration    time.Duration             `json:"duration"`
	Trigger     PipelineTrigger           `json:"trigger"`
	Stages      []PipelineStageExecution  `json:"stages"`
	Variables   map[string]interface{}    `json:"variables,omitempty"`
	Artifacts   []PipelineArtifact        `json:"artifacts,omitempty"`
	Logs        []string                  `json:"logs,omitempty"`
	User        string                    `json:"user"`
	Branch      string                    `json:"branch,omitempty"`
	Commit      string                    `json:"commit,omitempty"`
	Environment string                    `json:"environment"`
}

// PipelineStageExecution represents a stage execution within a pipeline
type PipelineStageExecution struct {
	Name       string                 `json:"name"`
	Status     PipelineStatus         `json:"status"`
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
	Duration   time.Duration          `json:"duration"`
	Output     string                 `json:"output,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
	Artifacts  []PipelineArtifact     `json:"artifacts,omitempty"`
}

// PipelineArtifact represents an output artifact from a pipeline
type PipelineArtifact struct {
	Name     string            `json:"name"`
	Path     string            `json:"path"`
	Type     string            `json:"type"`
	Size     int64             `json:"size"`
	Hash     string            `json:"hash"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CICDManager manages CI/CD pipeline integrations
type CICDManager struct {
	config       *CICDConfig
	pipelines    map[string]*PipelineDefinition
	executions   map[string]*PipelineExecution
	secretMgr    *SecretManager
	auditMgr     *AuditManager
	rbacMgr      *RBACManager
	dataDir      string
	pipelinesDir string
}

// CICDConfig represents CI/CD system configuration
type CICDConfig struct {
	Enabled         bool              `json:"enabled"`
	DataDir         string            `json:"data_dir"`
	DefaultProvider CICDProvider      `json:"default_provider"`
	MaxConcurrent   int               `json:"max_concurrent"`
	DefaultTimeout  time.Duration     `json:"default_timeout"`
	RetentionDays   int               `json:"retention_days"`
	WebhookSecret   string            `json:"webhook_secret,omitempty"`
	Providers       map[string]string `json:"providers,omitempty"`
}

// WebhookPayload represents a webhook payload from CI/CD systems
type WebhookPayload struct {
	Provider    CICDProvider           `json:"provider"`
	Event       string                 `json:"event"`
	Repository  string                 `json:"repository"`
	Branch      string                 `json:"branch"`
	Commit      string                 `json:"commit"`
	Author      string                 `json:"author"`
	Message     string                 `json:"message"`
	Timestamp   time.Time              `json:"timestamp"`
	Payload     map[string]interface{} `json:"payload"`
}

// NewCICDManager creates a new CI/CD manager
func NewCICDManager(config *CICDConfig, secretMgr *SecretManager, auditMgr *AuditManager, rbacMgr *RBACManager) (*CICDManager, error) {
	if err := os.MkdirAll(config.DataDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create CI/CD data directory: %w", err)
	}

	pipelinesDir := filepath.Join(config.DataDir, "pipelines")
	if err := os.MkdirAll(pipelinesDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create pipelines directory: %w", err)
	}

	mgr := &CICDManager{
		config:       config,
		pipelines:    make(map[string]*PipelineDefinition),
		executions:   make(map[string]*PipelineExecution),
		secretMgr:    secretMgr,
		auditMgr:     auditMgr,
		rbacMgr:      rbacMgr,
		dataDir:      config.DataDir,
		pipelinesDir: pipelinesDir,
	}

	// Load existing pipelines
	if err := mgr.loadPipelines(); err != nil {
		return nil, fmt.Errorf("failed to load pipelines: %w", err)
	}

	return mgr, nil
}

// CreatePipeline creates a new CI/CD pipeline
func (cicd *CICDManager) CreatePipeline(user string, def *PipelineDefinition) error {
	if !cicd.config.Enabled {
		return fmt.Errorf("CI/CD system is disabled")
	}

	// Check RBAC permissions
	accessResult := cicd.rbacMgr.CheckAccess(&AccessRequest{
		User:        user,
		Permission:  PermissionCICDWrite,
		Environment: "cicd",
	})

	if !accessResult.Granted {
		cicd.auditMgr.LogSystemEvent(AuditLevelWarning, "cicd.pipeline.create.denied", user, "cicd", map[string]interface{}{
			"pipeline_name": def.Name,
			"reason":        accessResult.Reason,
		})
		return fmt.Errorf("access denied: %s", accessResult.Reason)
	}

	// Validate pipeline definition
	if err := cicd.validatePipelineDefinition(def); err != nil {
		return fmt.Errorf("invalid pipeline definition: %w", err)
	}

	// Set metadata
	def.CreatedAt = time.Now()
	def.CreatedBy = user

	// Store pipeline
	cicd.pipelines[def.Name] = def
	if err := cicd.savePipeline(def); err != nil {
		return fmt.Errorf("failed to save pipeline: %w", err)
	}

	cicd.auditMgr.LogSystemEvent(AuditLevelInfo, "cicd.pipeline.create", user, "cicd", map[string]interface{}{
		"pipeline_name": def.Name,
		"provider":      def.Provider,
		"stages":        len(def.Stages),
	})

	return nil
}

// ExecutePipeline manually executes a pipeline
func (cicd *CICDManager) ExecutePipeline(user, pipelineName, environment, branch, commit string, variables map[string]interface{}) (*PipelineExecution, error) {
	if !cicd.config.Enabled {
		return nil, fmt.Errorf("CI/CD system is disabled")
	}

	// Check RBAC permissions
	accessResult := cicd.rbacMgr.CheckAccess(&AccessRequest{
		User:        user,
		Permission:  PermissionCICDExecute,
		Environment: environment,
	})

	if !accessResult.Granted {
		return nil, fmt.Errorf("access denied: %s", accessResult.Reason)
	}

	pipeline, exists := cicd.pipelines[pipelineName]
	if !exists {
		return nil, fmt.Errorf("pipeline '%s' not found", pipelineName)
	}

	// Create execution
	execution := &PipelineExecution{
		ID:         generateExecutionID(),
		PipelineID: pipelineName,
		Status:     StatusPending,
		StartedAt:  time.Now(),
		User:       user,
		Branch:     branch,
		Commit:     commit,
		Environment: environment,
		Variables:  variables,
		Trigger: PipelineTrigger{
			Type: "manual",
		},
	}

	cicd.executions[execution.ID] = execution

	// Start execution asynchronously
	go cicd.runPipelineExecution(pipeline, execution)

	cicd.auditMgr.LogSystemEvent(AuditLevelInfo, "cicd.pipeline.execute", user, environment, map[string]interface{}{
		"pipeline_name": pipelineName,
		"execution_id":  execution.ID,
		"branch":        branch,
		"commit":        commit,
	})

	return execution, nil
}

// HandleWebhook processes webhooks from CI/CD systems
func (cicd *CICDManager) HandleWebhook(payload *WebhookPayload) error {
	if !cicd.config.Enabled {
		return fmt.Errorf("CI/CD system is disabled")
	}

	// Find matching pipelines
	for _, pipeline := range cicd.pipelines {
		if cicd.shouldTriggerPipeline(pipeline, payload) {
			// Create execution from webhook
			execution := &PipelineExecution{
				ID:         generateExecutionID(),
				PipelineID: pipeline.Name,
				Status:     StatusPending,
				StartedAt:  time.Now(),
				User:       payload.Author,
				Branch:     payload.Branch,
				Commit:     payload.Commit,
				Environment: cicd.determineEnvironment(payload),
				Trigger: PipelineTrigger{
					Type: payload.Event,
				},
			}

			cicd.executions[execution.ID] = execution

			// Start execution
			go cicd.runPipelineExecution(pipeline, execution)

			cicd.auditMgr.LogSystemEvent(AuditLevelInfo, "cicd.webhook.trigger", payload.Author, execution.Environment, map[string]interface{}{
				"pipeline_name": pipeline.Name,
				"execution_id":  execution.ID,
				"event":         payload.Event,
				"repository":    payload.Repository,
			})
		}
	}

	return nil
}

// GetPipelineExecution retrieves a pipeline execution
func (cicd *CICDManager) GetPipelineExecution(user, executionID string) (*PipelineExecution, error) {
	// Check RBAC permissions
	accessResult := cicd.rbacMgr.CheckAccess(&AccessRequest{
		User:       user,
		Permission: PermissionCICDRead,
		Environment: "cicd",
	})

	if !accessResult.Granted {
		return nil, fmt.Errorf("access denied: %s", accessResult.Reason)
	}

	execution, exists := cicd.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution '%s' not found", executionID)
	}

	return execution, nil
}

// ListPipelineExecutions lists pipeline executions with filtering
func (cicd *CICDManager) ListPipelineExecutions(user, pipelineName string, limit int) ([]*PipelineExecution, error) {
	// Check RBAC permissions
	accessResult := cicd.rbacMgr.CheckAccess(&AccessRequest{
		User:       user,
		Permission: PermissionCICDRead,
		Environment: "cicd",
	})

	if !accessResult.Granted {
		return nil, fmt.Errorf("access denied: %s", accessResult.Reason)
	}

	var executions []*PipelineExecution
	for _, execution := range cicd.executions {
		if pipelineName == "" || execution.PipelineID == pipelineName {
			executions = append(executions, execution)
		}
	}

	// Sort by start time (newest first)
	// Apply limit if specified
	if limit > 0 && len(executions) > limit {
		executions = executions[:limit]
	}

	return executions, nil
}

// GenerateProviderConfig generates provider-specific configuration
func (cicd *CICDManager) GenerateProviderConfig(user, pipelineName string, provider CICDProvider) (string, error) {
	// Check RBAC permissions
	accessResult := cicd.rbacMgr.CheckAccess(&AccessRequest{
		User:       user,
		Permission: PermissionCICDRead,
		Environment: "cicd",
	})

	if !accessResult.Granted {
		return "", fmt.Errorf("access denied: %s", accessResult.Reason)
	}

	pipeline, exists := cicd.pipelines[pipelineName]
	if !exists {
		return "", fmt.Errorf("pipeline '%s' not found", pipelineName)
	}

	switch provider {
	case ProviderGitHubActions:
		return cicd.generateGitHubActionsConfig(pipeline), nil
	case ProviderGitLabCI:
		return cicd.generateGitLabCIConfig(pipeline), nil
	case ProviderJenkins:
		return cicd.generateJenkinsConfig(pipeline), nil
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

// Helper functions

func (cicd *CICDManager) runPipelineExecution(pipeline *PipelineDefinition, execution *PipelineExecution) {
	execution.Status = StatusRunning

	defer func() {
		if execution.FinishedAt == nil {
			now := time.Now()
			execution.FinishedAt = &now
			execution.Duration = now.Sub(execution.StartedAt)
		}
	}()

	// Execute stages
	for _, stageConfig := range pipeline.Stages {
		if !cicd.shouldExecuteStage(stageConfig, execution) {
			continue
		}

		stageExecution := &PipelineStageExecution{
			Name:      stageConfig.Name,
			Status:    StatusRunning,
			StartedAt: time.Now(),
		}

		execution.Stages = append(execution.Stages, *stageExecution)

		// Execute playbook for this stage
		success := cicd.executeStagePlaybook(stageConfig, execution, stageExecution)

		now := time.Now()
		stageExecution.FinishedAt = &now
		stageExecution.Duration = now.Sub(stageExecution.StartedAt)

		if success {
			stageExecution.Status = StatusSuccess
		} else {
			stageExecution.Status = StatusFailed
			if !stageConfig.AllowFailure {
				execution.Status = StatusFailed
				return
			}
		}
	}

	execution.Status = StatusSuccess
}

func (cicd *CICDManager) executeStagePlaybook(stageConfig PipelineStageConfig, execution *PipelineExecution, stageExecution *PipelineStageExecution) bool {
	// This would integrate with the main playbook execution system
	// For now, we'll simulate execution

	cicd.auditMgr.LogSystemEvent(AuditLevelInfo, "cicd.stage.execute", execution.User, execution.Environment, map[string]interface{}{
		"execution_id":   execution.ID,
		"stage_name":     stageConfig.Name,
		"playbook_path":  stageConfig.PlaybookPath,
	})

	// TODO: Integrate with actual playbook execution
	// This would call the main ptx-ansible execution system

	return true // Simulate success for now
}

func (cicd *CICDManager) shouldTriggerPipeline(pipeline *PipelineDefinition, payload *WebhookPayload) bool {
	for _, trigger := range pipeline.Triggers {
		if cicd.matchesTrigger(trigger, payload) {
			return true
		}
	}
	return false
}

func (cicd *CICDManager) matchesTrigger(trigger PipelineTrigger, payload *WebhookPayload) bool {
	if trigger.Type != payload.Event {
		return false
	}

	// Check branch patterns
	if len(trigger.Branches) > 0 {
		matched := false
		for _, pattern := range trigger.Branches {
			if matched, _ := regexp.MatchString(pattern, payload.Branch); matched {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func (cicd *CICDManager) shouldExecuteStage(stageConfig PipelineStageConfig, execution *PipelineExecution) bool {
	// Check dependencies
	for _, dep := range stageConfig.Depends {
		found := false
		for _, stage := range execution.Stages {
			if stage.Name == dep && stage.Status == StatusSuccess {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check condition
	if stageConfig.Condition != "" {
		// TODO: Implement condition evaluation
		return true
	}

	return true
}

func (cicd *CICDManager) determineEnvironment(payload *WebhookPayload) string {
	// Simple branch-based environment mapping
	switch payload.Branch {
	case "main", "master":
		return "production"
	case "develop", "dev":
		return "development"
	case "staging":
		return "staging"
	default:
		if strings.HasPrefix(payload.Branch, "feature/") {
			return "development"
		}
		return "development"
	}
}

func (cicd *CICDManager) validatePipelineDefinition(def *PipelineDefinition) error {
	if def.Name == "" {
		return fmt.Errorf("pipeline name is required")
	}

	if len(def.Stages) == 0 {
		return fmt.Errorf("pipeline must have at least one stage")
	}

	for _, stage := range def.Stages {
		if stage.Name == "" {
			return fmt.Errorf("stage name is required")
		}
		if stage.PlaybookPath == "" {
			return fmt.Errorf("stage playbook path is required")
		}
	}

	return nil
}

func (cicd *CICDManager) loadPipelines() error {
	files, err := filepath.Glob(filepath.Join(cicd.pipelinesDir, "*.json"))
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var pipeline PipelineDefinition
		if err := json.Unmarshal(data, &pipeline); err != nil {
			continue
		}

		cicd.pipelines[pipeline.Name] = &pipeline
	}

	return nil
}

func (cicd *CICDManager) savePipeline(pipeline *PipelineDefinition) error {
	data, err := json.MarshalIndent(pipeline, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(cicd.pipelinesDir, pipeline.Name+".json")
	return os.WriteFile(filename, data, 0640)
}

func (cicd *CICDManager) generateGitHubActionsConfig(pipeline *PipelineDefinition) string {
	config := fmt.Sprintf(`name: %s

on:
`, pipeline.Name)

	for _, trigger := range pipeline.Triggers {
		switch trigger.Type {
		case "push":
			config += "  push:\n"
			if len(trigger.Branches) > 0 {
				config += "    branches:\n"
				for _, branch := range trigger.Branches {
					config += fmt.Sprintf("      - %s\n", branch)
				}
			}
		case "pr":
			config += "  pull_request:\n"
		}
	}

	config += "\njobs:\n"

	for _, stage := range pipeline.Stages {
		config += fmt.Sprintf(`  %s:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Portunix Playbook
        run: |
          ./portunix playbook run %s --env %s
`, stage.Name, stage.PlaybookPath, stage.Environment)
	}

	return config
}

func (cicd *CICDManager) generateGitLabCIConfig(pipeline *PipelineDefinition) string {
	config := fmt.Sprintf("# Generated GitLab CI configuration for %s\n\n", pipeline.Name)

	config += "stages:\n"
	for _, stage := range pipeline.Stages {
		config += fmt.Sprintf("  - %s\n", string(stage.Stage))
	}

	config += "\n"

	for _, stage := range pipeline.Stages {
		config += fmt.Sprintf(`%s:
  stage: %s
  script:
    - ./portunix playbook run %s --env %s

`, stage.Name, string(stage.Stage), stage.PlaybookPath, stage.Environment)
	}

	return config
}

func (cicd *CICDManager) generateJenkinsConfig(pipeline *PipelineDefinition) string {
	config := fmt.Sprintf(`pipeline {
    agent any

    stages {
`)

	for _, stage := range pipeline.Stages {
		config += fmt.Sprintf(`        stage('%s') {
            steps {
                sh './portunix playbook run %s --env %s'
            }
        }
`, stage.Name, stage.PlaybookPath, stage.Environment)
	}

	config += `    }
}
`

	return config
}

func generateExecutionID() string {
	return fmt.Sprintf("exec-%d", time.Now().UnixNano())
}

// GetDefaultCICDConfig returns default CI/CD configuration
func GetDefaultCICDConfig() *CICDConfig {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".portunix", "cicd")

	return &CICDConfig{
		Enabled:         true,
		DataDir:         dataDir,
		DefaultProvider: ProviderGitHubActions,
		MaxConcurrent:   3,
		DefaultTimeout:  30 * time.Minute,
		RetentionDays:   30,
		Providers:       make(map[string]string),
	}
}