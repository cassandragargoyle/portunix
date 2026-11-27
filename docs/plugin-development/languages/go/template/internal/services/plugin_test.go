package services

import (
	"context"
	"testing"

	"go-plugin-template/internal/config"
)

func newService() *PluginService {
	return NewPluginService(&config.Config{
		Plugin:  config.PluginConfig{Name: "my-go-plugin", Version: "1.0.0"},
		Timeout: 30,
	})
}

func TestExecute(t *testing.T) {
	tests := []struct {
		command string
		args    []string
		want    string
		wantErr bool
	}{
		{command: "hello", want: "Hello, World! This is my-go-plugin v1.0.0"},
		{command: "hello", args: []string{"Portunix"}, want: "Hello, Portunix! This is my-go-plugin v1.0.0"},
		{command: "echo", args: []string{"a", "b"}, want: "a b"},
		{command: "echo", wantErr: true},
		{command: "uppercase", args: []string{"abc"}, want: "ABC"},
		{command: "unknown", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got, err := newService().Execute(context.Background(), tt.command, tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
