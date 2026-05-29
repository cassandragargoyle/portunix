//go:build unit
// +build unit

/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Issue #019: `portunix docker install` must exist as an alias and route to
// the same ptx-installer code path as `portunix install docker`.

func TestDockerInstallCmd_IsRegistered(t *testing.T) {
	var found *cobraLike
	for _, sub := range dockerCmd.Commands() {
		if sub.Name() == "install" {
			found = newCobraLike(sub.Use, sub.Short, sub.DisableFlagParsing)
			break
		}
	}

	assert.NotNil(t, found, "dockerCmd must have an 'install' subcommand (Issue #019)")
	if found == nil {
		return
	}

	assert.True(t, strings.HasPrefix(found.Use, "install"),
		"docker install Use should start with 'install', got %q", found.Use)
	assert.True(t, found.DisableFlagParsing,
		"docker install must have DisableFlagParsing=true so flags pass through to ptx-installer")
	assert.NotEmpty(t, found.Short, "docker install must have a Short description")
}

// cobraLike captures the subset of cobra.Command fields the alias test
// inspects. Using a local struct keeps the test independent of cobra
// internals and makes the assertions explicit.
type cobraLike struct {
	Use                string
	Short              string
	DisableFlagParsing bool
}

func newCobraLike(use, short string, disableFlagParsing bool) *cobraLike {
	return &cobraLike{Use: use, Short: short, DisableFlagParsing: disableFlagParsing}
}
