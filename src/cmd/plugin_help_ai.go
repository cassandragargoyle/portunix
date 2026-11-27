/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"portunix.ai/app/plugins/manager"
)

// pluginHelpAITimeout is how long a plugin may take to answer --help-ai
var pluginHelpAITimeout = 10 * time.Second

// pluginHelpAIEnv is set to "1" in the plugin environment when the plugin
// is asked for --help-ai
const pluginHelpAIEnv = "PORTUNIX_HELP_AI"

// pluginHelpAIWaitDelay bounds the wait for output pipes held open by
// processes left behind after the plugin was killed
const pluginHelpAIWaitDelay = time.Second

// pluginAIHelp is the registry fallback answer to `<plugin> --help-ai`;
// it follows the root --help-ai output where it fits
type pluginAIHelp struct {
	Tool        string                `json:"tool"`
	Version     string                `json:"version"`
	Description string                `json:"description"`
	Usage       string                `json:"usage"`
	Mode        string                `json:"mode"`
	Runtime     string                `json:"runtime"`
	Source      string                `json:"source"`
	Commands    []pluginAIHelpCommand `json:"commands"`
}

// pluginAIHelpCommand is one manifest command in the fallback answer
type pluginAIHelpCommand struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  []ParameterInfo `json:"parameters"`
	Examples    []string        `json:"examples"`
	SubCommands []string        `json:"subcommands"`
}

// runPluginHelpAI answers --help-ai for a plugin command (issue #257): the
// plugin is asked first, the registry entry is the fallback when the plugin
// does not answer with JSON; a timeout is an error, not a fallback
func runPluginHelpAI(plugin *manager.RegistryPlugin, args []string, stdout io.Writer, timeout time.Duration) error {
	if plugin.Enabled {
		if state, _ := plugin.CheckInstall(); state == manager.InstallStateOK {
			answer, err := askPluginHelpAI(plugin, args, timeout)
			if err != nil {
				return err
			}
			if answer != nil {
				_, err := stdout.Write(answer)
				return err
			}
		}
	}

	fallback, err := GeneratePluginAIHelp(plugin)
	if err != nil {
		return fmt.Errorf("failed to generate AI help for plugin '%s': %w", plugin.Name, err)
	}
	_, err = fmt.Fprintln(stdout, fallback)
	return err
}

// askPluginHelpAI runs the plugin with the original arguments and returns
// its stdout when it exits with code 0 and prints a JSON object; nil means
// the plugin did not answer and the fallback applies. stdin is closed and
// plugin stderr is dropped, so a "No such option" message does not leak
func askPluginHelpAI(plugin *manager.RegistryPlugin, args []string, timeout time.Duration) ([]byte, error) {
	callerDir, err := os.Getwd()
	if err != nil {
		callerDir = ""
	}

	cmd := buildPluginCommand(plugin, args, callerDir)
	cmd.Env = append(cmd.Env, pluginHelpAIEnv+"=1")
	cmd.Stdin = nil
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = io.Discard
	cmd.WaitDelay = pluginHelpAIWaitDelay
	// Own process group, so a timeout kills wrapper children too
	preparePluginProcessTree(cmd)

	if err := cmd.Start(); err != nil {
		return nil, nil // plugin cannot be started (e.g. missing runtime)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	// The own process group no longer gets the terminal Ctrl+C, forward it
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	select {
	case sig := <-interrupt:
		killPluginProcessTree(cmd)
		<-done
		return nil, fmt.Errorf("plugin '%s' --help-ai interrupted by %s", plugin.Name, sig)
	case err := <-done:
		if err != nil {
			return nil, nil
		}
		answer := bytes.TrimSpace(out.Bytes())
		if len(answer) == 0 || answer[0] != '{' || !json.Valid(answer) {
			return nil, nil
		}
		return append(answer, '\n'), nil
	case <-timer.C:
		killPluginProcessTree(cmd)
		<-done
		return nil, fmt.Errorf("plugin '%s' did not answer --help-ai within %s", plugin.Name, timeout)
	}
}

// GeneratePluginAIHelp builds the --help-ai JSON of a plugin from its
// registry entry, without running the plugin
func GeneratePluginAIHelp(plugin *manager.RegistryPlugin) (string, error) {
	output := pluginAIHelp{
		Tool:        plugin.Name,
		Version:     plugin.Version,
		Description: plugin.Description,
		Usage:       fmt.Sprintf("portunix %s <command> [args...]", plugin.Name),
		Mode:        getPluginMode(plugin.Mode),
		Runtime:     getPluginRuntime(plugin.Runtime),
		Source:      "registry",
		Commands:    make([]pluginAIHelpCommand, 0, len(plugin.Commands)),
	}

	for _, c := range plugin.Commands {
		command := pluginAIHelpCommand{
			Name:        c.Name,
			Description: c.Description,
			Parameters:  make([]ParameterInfo, 0, len(c.Parameters)),
			Examples:    nonNilStrings(c.Examples),
			SubCommands: nonNilStrings(c.Subcommands),
		}
		for _, p := range c.Parameters {
			command.Parameters = append(command.Parameters, ParameterInfo{
				Name:        p.Name,
				Type:        p.Type,
				Required:    p.Required,
				Description: p.Description,
				Default:     p.DefaultValue,
			})
		}
		output.Commands = append(output.Commands, command)
	}

	// No HTML escaping: `<command>` in usage stays readable for clients
	var data bytes.Buffer
	encoder := json.NewEncoder(&data)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return "", err
	}
	return strings.TrimSuffix(data.String(), "\n"), nil
}

// nonNilStrings returns an empty slice for nil, so JSON shows [] instead of null
func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

// reloadPluginForHelpAI returns the current registry entry of the plugin,
// or the entry captured at startup when the registry cannot be read
func reloadPluginForHelpAI(plugin *manager.RegistryPlugin) *manager.RegistryPlugin {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return plugin
	}
	registry, err := manager.NewRegistry(filepath.Join(homeDir, ".portunix", "plugins", "registry.json"))
	if err != nil {
		return plugin
	}
	fresh, err := registry.GetPluginRegistryData(plugin.Name)
	if err != nil {
		return plugin
	}
	return fresh
}
