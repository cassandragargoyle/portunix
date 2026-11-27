/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateString_ASCII(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"shorter than max", "hello", 10, "hello"},
		{"equal to max", "hello", 5, "hello"},
		{"longer than max", "hello world", 8, "hello..."},
		{"empty string", "", 10, ""},
		{"maxLen exactly 3 (truncates without ellipsis)", "abcdef", 3, "abc"},
		{"maxLen 1 (truncates without ellipsis)", "abcdef", 1, "a"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := truncateString(tc.input, tc.maxLen)
			if got != tc.want {
				t.Errorf("truncateString(%q, %d) = %q; want %q", tc.input, tc.maxLen, got, tc.want)
			}
		})
	}
}

func TestTruncateString_UTF8_NoBrokenRunes(t *testing.T) {
	// Czech / accented chars — historically the byte-based version split runes.
	// "Universální" is 11 runes but more than 11 bytes (accents are 2-byte).
	input := "Universální nástroj pro vývoj"
	got := truncateString(input, 15)
	if !utf8.ValidString(got) {
		t.Fatalf("output contains invalid UTF-8 runes: %q", got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected truncated output to end with ellipsis, got %q", got)
	}
	// 15 runes total = 12 prefix runes + "..."
	if utf8.RuneCountInString(got) != 15 {
		t.Errorf("expected 15 runes, got %d (%q)", utf8.RuneCountInString(got), got)
	}
}

func TestTruncateString_Emoji_NoBrokenRunes(t *testing.T) {
	// Each emoji is a multi-byte rune; byte slicing would corrupt them.
	input := "🚀🚀🚀🚀🚀🚀🚀🚀🚀🚀 rocket plugin description"
	got := truncateString(input, 12)
	if !utf8.ValidString(got) {
		t.Fatalf("output contains invalid UTF-8 runes: %q", got)
	}
	if utf8.RuneCountInString(got) != 12 {
		t.Errorf("expected 12 runes, got %d (%q)", utf8.RuneCountInString(got), got)
	}
}

func TestWrapForTerminal_ShortText(t *testing.T) {
	lines := wrapForTerminal("short text", 0)
	if len(lines) != 1 || lines[0] != "short text" {
		t.Errorf("expected single line 'short text', got %v", lines)
	}
}

func TestWrapForTerminal_EmptyText(t *testing.T) {
	lines := wrapForTerminal("", 4)
	if len(lines) != 1 || lines[0] != "    " {
		t.Errorf("expected single line with indent only, got %v", lines)
	}
}

func TestWrapForTerminal_WordBoundaryWrap(t *testing.T) {
	// Force a tight width by leaving most of the terminal taken up by the
	// indent. terminalWidth() falls back to 80 in non-TTY tests, so indent
	// of 70 leaves a 10-column wrapping budget.
	lines := wrapForTerminal("alpha beta gamma delta", 70)
	if len(lines) < 2 {
		t.Fatalf("expected text to wrap across multiple lines, got %v", lines)
	}
	// Every line should start with the 70-space indent.
	for _, line := range lines {
		if !strings.HasPrefix(line, strings.Repeat(" ", 70)) {
			t.Errorf("line missing indent prefix: %q", line)
		}
	}
	// The joined content must contain all original words in order.
	joined := strings.Join(lines, " ")
	for _, w := range []string{"alpha", "beta", "gamma", "delta"} {
		if !strings.Contains(joined, w) {
			t.Errorf("word %q missing from wrapped output: %v", w, lines)
		}
	}
}

func TestWrapForTerminal_HardBreakLongToken(t *testing.T) {
	// A 30-char unbreakable token with a tight wrapping budget must be split.
	long := strings.Repeat("x", 30)
	lines := wrapForTerminal(long, 70) // budget = 80 - 70 = 10
	if len(lines) < 3 {
		t.Errorf("expected long token to be split into >=3 lines, got %d: %v", len(lines), lines)
	}
}

func TestWrapForTerminal_UTF8_NoBrokenRunes(t *testing.T) {
	// Multi-byte runes must not be split mid-rune during wrapping.
	long := strings.Repeat("á", 30) // 30 runes, 60 bytes
	lines := wrapForTerminal(long, 70)
	for _, line := range lines {
		if !utf8.ValidString(line) {
			t.Errorf("wrapped line contains invalid UTF-8: %q", line)
		}
	}
}

func TestTruncateString_DescriptionWidth(t *testing.T) {
	if pluginListDescriptionWidth != 45 {
		t.Fatalf("pluginListDescriptionWidth = %d; want 45", pluginListDescriptionWidth)
	}

	exact := strings.Repeat("d", pluginListDescriptionWidth)
	if got := truncateString(exact, pluginListDescriptionWidth); got != exact {
		t.Errorf("description of exactly %d runes must not be truncated, got %q", pluginListDescriptionWidth, got)
	}

	long := "Universal plugin providing a very long description that exceeds the column"
	got := truncateString(long, pluginListDescriptionWidth)
	if utf8.RuneCountInString(got) != pluginListDescriptionWidth {
		t.Errorf("expected %d runes, got %d (%q)", pluginListDescriptionWidth, utf8.RuneCountInString(got), got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected truncated description to end with ellipsis, got %q", got)
	}
}

func TestResolvePluginListFormat(t *testing.T) {
	cases := []struct {
		name          string
		jsonFlag      bool
		output        string
		outputChanged bool
		want          string
		wantErr       bool
	}{
		{"default table", false, "table", false, "table", false},
		{"explicit yaml", false, "yaml", true, "yaml", false},
		{"json flag alone", true, "table", false, "json", false},
		{"json flag with output json", true, "json", true, "json", false},
		{"json flag with output yaml", true, "yaml", true, "", true},
		{"json flag with output table", true, "table", true, "", true},
		{"uppercase output json", false, "JSON", true, "json", false},
		{"mixed case output yaml", false, "Yaml", true, "yaml", false},
		{"json flag with uppercase output json", true, "JSON", true, "json", false},
		{"unknown output format", false, "xml", true, "", true},
		{"json flag with unknown output format", true, "xml", true, "", true},
		{"empty output format", false, "", true, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolvePluginListFormat(tc.jsonFlag, tc.output, tc.outputChanged)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v; wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("format = %q; want %q", got, tc.want)
			}
		})
	}
}
