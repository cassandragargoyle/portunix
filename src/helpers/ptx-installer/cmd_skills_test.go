/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"reflect"
	"testing"
)

func TestParseSelection(t *testing.T) {
	tests := []struct {
		answer   string
		defaults []int
		want     []int
		wantErr  bool
	}{
		{answer: "", defaults: []int{1, 3}, want: []int{1, 3}},
		{answer: "  \n", defaults: nil, want: []int{}},
		{answer: "all", want: []int{0, 1, 2, 3}},
		{answer: "NONE", defaults: []int{0}, want: []int{}},
		{answer: "2,4", want: []int{1, 3}},
		{answer: "3 1 3", want: []int{2, 0}},
		{answer: "0", wantErr: true},
		{answer: "5", wantErr: true},
		{answer: "x", wantErr: true},
	}
	for _, tt := range tests {
		got, err := parseSelection(tt.answer, 4, tt.defaults)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseSelection(%q) expected error", tt.answer)
			}
			continue
		}
		if err != nil || !reflect.DeepEqual(got, tt.want) {
			t.Errorf("parseSelection(%q) = %v, %v; want %v", tt.answer, got, err, tt.want)
		}
	}
}

func TestOutcomeError(t *testing.T) {
	if err := outcomeError(0, 0); err != nil {
		t.Errorf("no problems must not fail: %v", err)
	}
	if err := outcomeError(2, 1); err == nil || err.Error() != "1 skill target(s) failed, 2 skill target(s) skipped due to conflicts" {
		t.Errorf("unexpected error %v", err)
	}
}

func TestParseSkillsFlags(t *testing.T) {
	f, err := parseSkillsFlags([]string{"obra/superpowers", "--skill", "a,b", "-s=c", "--agent=claude", "-g", "--ref", "v1", "--yes"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(f.positional, []string{"obra/superpowers"}) || !reflect.DeepEqual(f.skills, []string{"a", "b", "c"}) ||
		!reflect.DeepEqual(f.agents, []string{"claude"}) || !f.global || f.ref != "v1" || !f.yes {
		t.Errorf("parsed %+v", f)
	}
	if f, _ := parseSkillsFlags([]string{"x"}); f.agents != nil {
		t.Error("agents must stay nil without --agent (enables detection / prompt)")
	}
	if _, err := parseSkillsFlags([]string{"--ref"}); err == nil {
		t.Error("missing flag value must fail")
	}
	if _, err := parseSkillsFlags([]string{"--bogus"}); err == nil {
		t.Error("unknown flag must fail")
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("žluťoučký kůň", 8); got != "žluťo..." {
		t.Errorf("truncate = %q", got)
	}
	if got := truncate("short", 8); got != "short" {
		t.Errorf("truncate = %q", got)
	}
}
