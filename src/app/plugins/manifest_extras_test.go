/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package plugins

import (
	"reflect"
	"strings"
	"testing"
)

// extrasTestManifest returns a valid Python wheel manifest with the given extras
func extrasTestManifest(extras ...WheelExtra) *PluginManifest {
	m := buildManifest()
	m.Plugin.Runtime = "python"
	m.Plugin.Wheel = "ptx_demo-1.0.0-py3-none-any.whl"
	m.Plugin.OptionalExtras = extras
	return m
}

func TestValidateManifest_OptionalExtras(t *testing.T) {
	tests := []struct {
		name    string
		extras  []WheelExtra
		noWheel bool
		wantErr string
	}{
		{name: "none"},
		{name: "valid", extras: []WheelExtra{{Name: "train", Description: "Training"}, {Name: "gpu.cuda-12_x"}}},
		{name: "without wheel", extras: []WheelExtra{{Name: "train"}}, noWheel: true, wantErr: "requires wheel"},
		{name: "empty name", extras: []WheelExtra{{Name: ""}}, wantErr: "invalid optional_extras name"},
		{name: "leading dash", extras: []WheelExtra{{Name: "-train"}}, wantErr: "invalid optional_extras name"},
		{name: "bracket", extras: []WheelExtra{{Name: "train]"}}, wantErr: "invalid optional_extras name"},
		{name: "space", extras: []WheelExtra{{Name: "my extra"}}, wantErr: "invalid optional_extras name"},
		{name: "duplicate", extras: []WheelExtra{{Name: "train"}, {Name: "train"}}, wantErr: "duplicate"},
		{name: "duplicate normalized", extras: []WheelExtra{{Name: "gpu_cuda"}, {Name: "GPU-cuda"}}, wantErr: "duplicate"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := extrasTestManifest(tt.extras...)
			if tt.noWheel {
				m.Plugin.Wheel = ""
			}
			err := ValidateManifest(m)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestResolveWheelExtras(t *testing.T) {
	m := extrasTestManifest(WheelExtra{Name: "train"}, WheelExtra{Name: "gpu_cuda"})

	got, err := ResolveWheelExtras(m, []string{" train ", "GPU-cuda", "train", ""})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"gpu_cuda", "train"}; !reflect.DeepEqual(got, want) {
		t.Errorf("resolved = %v, want %v", got, want)
	}

	if got, err := ResolveWheelExtras(m, nil); err != nil || got != nil {
		t.Errorf("no extras = %v, %v; want nil, nil", got, err)
	}

	_, err = ResolveWheelExtras(m, []string{"cuda"})
	if err == nil || !strings.Contains(err.Error(), "declared extras: train, gpu_cuda") {
		t.Errorf("undeclared extra error = %v", err)
	}

	_, err = ResolveWheelExtras(extrasTestManifest(), []string{"train"})
	if err == nil || !strings.Contains(err.Error(), "declares no optional extras") {
		t.Errorf("no declared extras error = %v", err)
	}
}
