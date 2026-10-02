/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// =============================================================================
// PTX-PFT MCP Integration (Issue #111)
//
// MCP tools and resources exposing Product Feedback Tool (ptx-pft)
// functionality to AI assistants. Tools are thin wrappers that invoke the
// ptx-pft binary via exec and return its output (Option A architecture).
// =============================================================================

// pftCommandTimeout limits how long a single ptx-pft invocation may run
const pftCommandTimeout = 120 * time.Second

// runPftCommand executes ptx-pft with the given arguments and a timeout
func (s *Server) runPftCommand(args ...string) (string, error) {
	pftBinary, err := s.getPtxPftBinary()
	if err != nil {
		return "", fmt.Errorf("ptx-pft is not available: %w (ensure ptx-pft is installed alongside ptx-mcp)", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), pftCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, pftBinary, args...)
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(output), fmt.Errorf("ptx-pft command timed out after %s", pftCommandTimeout)
	}
	if err != nil {
		return string(output), fmt.Errorf("ptx-pft command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// getStringArg extracts an optional string argument from tool arguments
func getStringArg(args map[string]interface{}, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

// getBoolArg extracts an optional boolean argument from tool arguments
func getBoolArg(args map[string]interface{}, key string) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return false
}

// appendVoiceFlags maps the voice parameter (voc, vos, all) to CLI flags
func appendVoiceFlags(cmdArgs []string, voice string) []string {
	switch voice {
	case "voc":
		cmdArgs = append(cmdArgs, "--voc")
	case "vos":
		cmdArgs = append(cmdArgs, "--vos")
	}
	// "all" or empty: no flag, ptx-pft defaults to both voices
	return cmdArgs
}

// appendPathFlag adds the optional --path flag for config discovery
func appendPathFlag(cmdArgs []string, args map[string]interface{}) []string {
	if path := getStringArg(args, "path"); path != "" {
		cmdArgs = append(cmdArgs, "--path", path)
	}
	return cmdArgs
}

// =============================================================================
// CLI argument builders (separated from execution for testability)
// =============================================================================

func buildPftListArgs(args map[string]interface{}) ([]string, error) {
	cmdArgs := []string{"pft", "list"}
	cmdArgs = appendVoiceFlags(cmdArgs, getStringArg(args, "voice"))
	if getBoolArg(args, "all") {
		cmdArgs = append(cmdArgs, "--all")
	}
	if category := getStringArg(args, "category"); category != "" {
		cmdArgs = append(cmdArgs, "--category", category)
	}
	if getBoolArg(args, "uncategorized") {
		cmdArgs = append(cmdArgs, "--uncategorized")
	}
	if format := getStringArg(args, "format"); format != "" {
		cmdArgs = append(cmdArgs, "--format", format)
	}
	cmdArgs = appendPathFlag(cmdArgs, args)
	return cmdArgs, nil
}

func buildPftShowArgs(args map[string]interface{}) ([]string, error) {
	id := getStringArg(args, "id")
	if id == "" {
		return nil, fmt.Errorf("id parameter is required")
	}
	cmdArgs := []string{"pft", "show", id}
	cmdArgs = appendPathFlag(cmdArgs, args)
	return cmdArgs, nil
}

func buildPftSyncArgs(args map[string]interface{}) ([]string, error) {
	var subcommand string
	switch getStringArg(args, "direction") {
	case "pull":
		subcommand = "pull"
	case "push":
		subcommand = "push"
	case "bidirectional", "":
		subcommand = "sync"
	default:
		return nil, fmt.Errorf("direction must be one of: pull, push, bidirectional")
	}
	cmdArgs := []string{"pft", subcommand}
	cmdArgs = appendVoiceFlags(cmdArgs, getStringArg(args, "voice"))
	if getBoolArg(args, "dry_run") {
		cmdArgs = append(cmdArgs, "--dry-run")
	}
	return cmdArgs, nil
}

func buildPftLinkArgs(args map[string]interface{}) ([]string, error) {
	feedbackID := getStringArg(args, "feedback_id")
	if feedbackID == "" {
		return nil, fmt.Errorf("feedback_id parameter is required")
	}
	issueID := getStringArg(args, "issue_id")
	if issueID == "" {
		return nil, fmt.Errorf("issue_id parameter is required")
	}
	return []string{"pft", "link", feedbackID, issueID}, nil
}

func buildPftReportArgs(args map[string]interface{}) ([]string, error) {
	cmdArgs := []string{"pft", "report"}
	if reportType := getStringArg(args, "type"); reportType != "" {
		cmdArgs = append(cmdArgs, "--type", reportType)
	}
	if output := getStringArg(args, "output"); output != "" {
		cmdArgs = append(cmdArgs, "--output", output)
	}
	return cmdArgs, nil
}

func buildPftExportArgs(args map[string]interface{}) ([]string, error) {
	format := getStringArg(args, "format")
	if format == "" {
		return nil, fmt.Errorf("format parameter is required")
	}
	cmdArgs := []string{"pft", "export", "--format", format}
	if output := getStringArg(args, "output"); output != "" {
		cmdArgs = append(cmdArgs, "--output", output)
	}
	cmdArgs = appendVoiceFlags(cmdArgs, getStringArg(args, "voice"))
	return cmdArgs, nil
}

func buildPftNotifyArgs(args map[string]interface{}) ([]string, error) {
	feedbackID := getStringArg(args, "feedback_id")
	if feedbackID == "" {
		return nil, fmt.Errorf("feedback_id parameter is required")
	}
	notifyType := getStringArg(args, "notification_type")
	if notifyType == "" {
		return nil, fmt.Errorf("notification_type parameter is required")
	}
	cmdArgs := []string{"pft", "notify", feedbackID, "--type", notifyType}
	switch getStringArg(args, "recipients") {
	case "all-voc":
		cmdArgs = append(cmdArgs, "--all-voc")
	case "all-vos":
		cmdArgs = append(cmdArgs, "--all-vos")
	case "specific", "":
		email := getStringArg(args, "email")
		if email == "" {
			return nil, fmt.Errorf("email parameter is required when recipients is 'specific'")
		}
		cmdArgs = append(cmdArgs, "--user", email)
	default:
		return nil, fmt.Errorf("recipients must be one of: all-voc, all-vos, specific")
	}
	if getBoolArg(args, "dry_run") {
		cmdArgs = append(cmdArgs, "--dry-run")
	}
	return cmdArgs, nil
}

func buildPftStatusArgs(args map[string]interface{}) ([]string, error) {
	return []string{"pft", "status"}, nil
}

func buildPftUserListArgs(args map[string]interface{}) ([]string, error) {
	cmdArgs := []string{"pft", "user", "list"}
	switch category := getStringArg(args, "category"); category {
	case "voc", "vos", "vob", "voe":
		cmdArgs = append(cmdArgs, "--"+category)
	case "all", "":
		// no flag, lists all users
	default:
		return nil, fmt.Errorf("category must be one of: voc, vos, vob, voe, all")
	}
	cmdArgs = appendPathFlag(cmdArgs, args)
	return cmdArgs, nil
}

func buildPftUserAddArgs(args map[string]interface{}) ([]string, error) {
	id := getStringArg(args, "id")
	if id == "" {
		return nil, fmt.Errorf("id parameter is required")
	}
	name := getStringArg(args, "name")
	if name == "" {
		return nil, fmt.Errorf("name parameter is required")
	}
	cmdArgs := []string{"pft", "user", "add", "--id", id, "--name", name}
	if org := getStringArg(args, "org"); org != "" {
		cmdArgs = append(cmdArgs, "--org", org)
	}
	cmdArgs = appendPathFlag(cmdArgs, args)
	return cmdArgs, nil
}

func buildPftUserShowArgs(args map[string]interface{}) ([]string, error) {
	id := getStringArg(args, "id")
	if id == "" {
		return nil, fmt.Errorf("id parameter is required")
	}
	cmdArgs := []string{"pft", "user", "show", id}
	cmdArgs = appendPathFlag(cmdArgs, args)
	return cmdArgs, nil
}

func buildPftUserUpdateArgs(args map[string]interface{}) ([]string, error) {
	id := getStringArg(args, "id")
	if id == "" {
		return nil, fmt.Errorf("id parameter is required")
	}
	name := getStringArg(args, "name")
	org := getStringArg(args, "org")
	if name == "" && org == "" {
		return nil, fmt.Errorf("at least one of name or org parameters is required")
	}
	cmdArgs := []string{"pft", "user", "update", id}
	if name != "" {
		cmdArgs = append(cmdArgs, "--name", name)
	}
	if org != "" {
		cmdArgs = append(cmdArgs, "--org", org)
	}
	cmdArgs = appendPathFlag(cmdArgs, args)
	return cmdArgs, nil
}

func buildPftUserRoleArgs(args map[string]interface{}) ([]string, error) {
	id := getStringArg(args, "id")
	if id == "" {
		return nil, fmt.Errorf("id parameter is required")
	}
	category := getStringArg(args, "category")
	switch category {
	case "voc", "vos", "vob", "voe":
		// valid category
	default:
		return nil, fmt.Errorf("category must be one of: voc, vos, vob, voe")
	}
	remove := getBoolArg(args, "remove")
	role := getStringArg(args, "role")
	if !remove && role == "" {
		return nil, fmt.Errorf("role parameter is required unless remove is true")
	}
	cmdArgs := []string{"pft", "user", "role", id, "--" + category}
	if remove {
		cmdArgs = append(cmdArgs, "--remove")
	} else {
		cmdArgs = append(cmdArgs, role)
		if getBoolArg(args, "proxy") {
			cmdArgs = append(cmdArgs, "--proxy")
		}
	}
	cmdArgs = appendPathFlag(cmdArgs, args)
	return cmdArgs, nil
}

func buildPftUserLinkArgs(args map[string]interface{}) ([]string, error) {
	id := getStringArg(args, "id")
	if id == "" {
		return nil, fmt.Errorf("id parameter is required")
	}
	externalID := getStringArg(args, "external_id")
	if externalID == "" {
		return nil, fmt.Errorf("external_id parameter is required")
	}
	if provider := getStringArg(args, "provider"); provider != "" && provider != "fider" {
		return nil, fmt.Errorf("provider %q is not supported yet; only fider user linking is available", provider)
	}
	cmdArgs := []string{"pft", "user", "link", id, "--fider", externalID}
	cmdArgs = appendPathFlag(cmdArgs, args)
	return cmdArgs, nil
}

func buildPftUserRemoveArgs(args map[string]interface{}) ([]string, error) {
	id := getStringArg(args, "id")
	if id == "" {
		return nil, fmt.Errorf("id parameter is required")
	}
	cmdArgs := []string{"pft", "user", "remove", id}
	cmdArgs = appendPathFlag(cmdArgs, args)
	return cmdArgs, nil
}

func buildPftRoleListArgs(args map[string]interface{}) ([]string, error) {
	cmdArgs := []string{"pft", "role", "list"}
	switch category := getStringArg(args, "category"); category {
	case "voc", "vos", "vob", "voe":
		cmdArgs = append(cmdArgs, "--"+category)
	case "all", "":
		// no flag, lists roles for all categories
	default:
		return nil, fmt.Errorf("category must be one of: voc, vos, vob, voe, all")
	}
	cmdArgs = appendPathFlag(cmdArgs, args)
	return cmdArgs, nil
}

// pftToolBuilders maps MCP tool names to their CLI argument builders
var pftToolBuilders = map[string]func(map[string]interface{}) ([]string, error){
	"pft_list":        buildPftListArgs,
	"pft_show":        buildPftShowArgs,
	"pft_sync":        buildPftSyncArgs,
	"pft_link":        buildPftLinkArgs,
	"pft_report":      buildPftReportArgs,
	"pft_export":      buildPftExportArgs,
	"pft_notify":      buildPftNotifyArgs,
	"pft_status":      buildPftStatusArgs,
	"pft_user_list":   buildPftUserListArgs,
	"pft_user_add":    buildPftUserAddArgs,
	"pft_user_show":   buildPftUserShowArgs,
	"pft_user_update": buildPftUserUpdateArgs,
	"pft_user_role":   buildPftUserRoleArgs,
	"pft_user_link":   buildPftUserLinkArgs,
	"pft_user_remove": buildPftUserRemoveArgs,
	"pft_role_list":   buildPftRoleListArgs,
}

// handlePftToolCall dispatches pft_* tool calls to the matching builder
// and executes the resulting ptx-pft command
func (s *Server) handlePftToolCall(name string, args map[string]interface{}) (interface{}, error) {
	builder, ok := pftToolBuilders[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	if args == nil {
		args = map[string]interface{}{}
	}
	cmdArgs, err := builder(args)
	if err != nil {
		return nil, err
	}
	output, err := s.runPftCommand(cmdArgs...)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status": "success",
		"output": output,
	}, nil
}

// pftToolDefinitions returns MCP tool descriptors for PTX-PFT integration
func pftToolDefinitions() []map[string]interface{} {
	voiceProperty := map[string]interface{}{
		"type":        "string",
		"enum":        []string{"voc", "vos", "all"},
		"description": "Voice type: voc (customer), vos (stakeholder), or all (default)",
	}
	pathProperty := map[string]interface{}{
		"type":        "string",
		"description": "Path to document directory (where .pft-config.json is located)",
	}
	userCategoryProperty := map[string]interface{}{
		"type":        "string",
		"enum":        []string{"voc", "vos", "vob", "voe", "all"},
		"description": "Voice category: voc (customer), vos (stakeholder), vob (business), voe (engineer), or all",
	}

	return []map[string]interface{}{
		{
			"name":        "pft_list",
			"description": "List feedback items from the product feedback system",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"voice": voiceProperty,
					"all": map[string]interface{}{
						"type":        "boolean",
						"description": "Show full details for each item",
					},
					"category": map[string]interface{}{
						"type":        "string",
						"description": "Filter items by category ID",
					},
					"uncategorized": map[string]interface{}{
						"type":        "boolean",
						"description": "Show only uncategorized items",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"table", "json"},
						"description": "Output format (default: table)",
					},
					"path": pathProperty,
				},
			},
		},
		{
			"name":        "pft_show",
			"description": "Show details of a specific feedback item",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Feedback item ID (e.g., UC001, REQ001)",
					},
					"path": pathProperty,
				},
				"required": []string{"id"},
			},
		},
		{
			"name":        "pft_sync",
			"description": "Synchronize local documentation with the external feedback system",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"direction": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"pull", "push", "bidirectional"},
						"description": "Sync direction (default: bidirectional)",
					},
					"voice": voiceProperty,
					"dry_run": map[string]interface{}{
						"type":        "boolean",
						"description": "Preview changes without applying them",
					},
				},
			},
		},
		{
			"name":        "pft_link",
			"description": "Link a feedback item to a local issue",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"feedback_id": map[string]interface{}{
						"type":        "string",
						"description": "Feedback item ID (e.g., UC001, REQ001)",
					},
					"issue_id": map[string]interface{}{
						"type":        "string",
						"description": "Local issue reference (e.g., #107, ISSUE-42, docs/issues/internal/042-feature.md)",
					},
				},
				"required": []string{"feedback_id", "issue_id"},
			},
		},
		{
			"name":        "pft_report",
			"description": "Generate a feedback report",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"summary", "detailed", "status"},
						"description": "Report type (default: summary)",
					},
					"output": map[string]interface{}{
						"type":        "string",
						"description": "Output file path (prints to stdout when omitted)",
					},
				},
			},
		},
		{
			"name":        "pft_export",
			"description": "Export feedback data to various formats",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"format": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"md", "json", "csv"},
						"description": "Export format",
					},
					"output": map[string]interface{}{
						"type":        "string",
						"description": "Output file path",
					},
					"voice": voiceProperty,
				},
				"required": []string{"format"},
			},
		},
		{
			"name":        "pft_notify",
			"description": "Send notification to users about a feedback item",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"feedback_id": map[string]interface{}{
						"type":        "string",
						"description": "Feedback item ID",
					},
					"notification_type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"vote", "description", "acceptance", "status_change"},
						"description": "Type of notification",
					},
					"recipients": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"all-voc", "all-vos", "specific"},
						"description": "Recipient group (default: specific)",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"description": "Specific email if recipients is 'specific'",
					},
					"dry_run": map[string]interface{}{
						"type":        "boolean",
						"description": "Preview notification without sending",
					},
				},
				"required": []string{"feedback_id", "notification_type"},
			},
		},
		{
			"name":        "pft_status",
			"description": "Check status of the feedback tool deployment",
			"inputSchema": map[string]interface{}{
				"type":        "object",
				"properties":  map[string]interface{}{},
				"description": "No parameters required",
			},
		},
		{
			"name":        "pft_user_list",
			"description": "List users in the feedback system",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"category": userCategoryProperty,
					"path":     pathProperty,
				},
			},
		},
		{
			"name":        "pft_user_add",
			"description": "Add a new user to the feedback system",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "User ID (typically email address)",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "User's display name",
					},
					"org": map[string]interface{}{
						"type":        "string",
						"description": "User's company/organization",
					},
					"path": pathProperty,
				},
				"required": []string{"id", "name"},
			},
		},
		{
			"name":        "pft_user_show",
			"description": "Show details of a specific user",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "User ID or email",
					},
					"path": pathProperty,
				},
				"required": []string{"id"},
			},
		},
		{
			"name":        "pft_user_update",
			"description": "Update user information (name and/or organization)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "User ID or email",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "New display name",
					},
					"org": map[string]interface{}{
						"type":        "string",
						"description": "New company/organization",
					},
					"path": pathProperty,
				},
				"required": []string{"id"},
			},
		},
		{
			"name":        "pft_user_role",
			"description": "Assign, change, or remove a user's role in a voice category",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "User ID or email",
					},
					"category": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"voc", "vos", "vob", "voe"},
						"description": "Voice category for the role assignment",
					},
					"role": map[string]interface{}{
						"type":        "string",
						"description": "Role to assign (required unless remove is true)",
					},
					"proxy": map[string]interface{}{
						"type":        "boolean",
						"description": "Mark role as proxy (acting on behalf of someone)",
					},
					"remove": map[string]interface{}{
						"type":        "boolean",
						"description": "Remove the role from the category instead of assigning",
					},
					"path": pathProperty,
				},
				"required": []string{"id", "category"},
			},
		},
		{
			"name":        "pft_user_link",
			"description": "Link a local user to an external feedback system user ID",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Local user ID or email",
					},
					"external_id": map[string]interface{}{
						"type":        "string",
						"description": "External system user ID",
					},
					"provider": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"fider"},
						"description": "External provider name (default: fider)",
					},
					"path": pathProperty,
				},
				"required": []string{"id", "external_id"},
			},
		},
		{
			"name":        "pft_user_remove",
			"description": "Remove a user from the feedback system",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "User ID or email",
					},
					"path": pathProperty,
				},
				"required": []string{"id"},
			},
		},
		{
			"name":        "pft_role_list",
			"description": "List available roles and their permissions",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"category": userCategoryProperty,
					"path":     pathProperty,
				},
			},
		},
	}
}

// =============================================================================
// MCP Resources (pft:// scheme)
// =============================================================================

// pftResourceDefinitions returns MCP resource descriptors for PTX-PFT
func pftResourceDefinitions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"uri":         "pft://config",
			"name":        "PFT configuration",
			"description": "Current Product Feedback Tool configuration",
			"mimeType":    "text/plain",
		},
		{
			"uri":         "pft://cache",
			"name":        "PFT sync cache",
			"description": "Local cache status of synchronized feedback items",
			"mimeType":    "text/plain",
		},
		{
			"uri":         "pft://users",
			"name":        "PFT users",
			"description": "List of all users in the feedback system",
			"mimeType":    "text/plain",
		},
		{
			"uri":         "pft://roles",
			"name":        "PFT roles",
			"description": "Available roles and their permissions",
			"mimeType":    "text/plain",
		},
		{
			"uri":         "pft://voc/{id}",
			"name":        "Voice of Customer item",
			"description": "Voice of Customer feedback item by ID",
			"mimeType":    "text/markdown",
		},
		{
			"uri":         "pft://vos/{id}",
			"name":        "Voice of Stakeholder item",
			"description": "Voice of Stakeholder requirement by ID",
			"mimeType":    "text/markdown",
		},
	}
}

// pftResourceCommand maps a pft:// resource URI to a ptx-pft command and MIME type
func pftResourceCommand(uri string) ([]string, string, error) {
	switch uri {
	case "pft://config":
		return []string{"pft", "configure", "--show"}, "text/plain", nil
	case "pft://cache":
		return []string{"pft", "cache", "status"}, "text/plain", nil
	case "pft://users":
		return []string{"pft", "user", "list"}, "text/plain", nil
	case "pft://roles":
		return []string{"pft", "role", "list"}, "text/plain", nil
	}

	// Item resources: pft://voc/{id} and pft://vos/{id}
	for _, prefix := range []string{"pft://voc/", "pft://vos/"} {
		if strings.HasPrefix(uri, prefix) {
			id := strings.TrimPrefix(uri, prefix)
			if id == "" {
				return nil, "", fmt.Errorf("missing item ID in resource URI: %s", uri)
			}
			return []string{"pft", "show", id}, "text/markdown", nil
		}
	}

	return nil, "", fmt.Errorf("unknown resource URI: %s", uri)
}

// handleResourcesList handles the MCP resources/list method
func (s *Server) handleResourcesList(params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"resources": pftResourceDefinitions(),
	}, nil
}

// handleResourcesRead handles the MCP resources/read method
func (s *Server) handleResourcesRead(params json.RawMessage) (interface{}, error) {
	var request struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid resource read parameters: %w", err)
	}
	if request.URI == "" {
		return nil, fmt.Errorf("uri parameter is required")
	}

	cmdArgs, mimeType, err := pftResourceCommand(request.URI)
	if err != nil {
		return nil, err
	}

	output, err := s.runPftCommand(cmdArgs...)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"uri":      request.URI,
				"mimeType": mimeType,
				"text":     output,
			},
		},
	}, nil
}
