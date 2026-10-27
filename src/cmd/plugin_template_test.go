/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPluginTemplateProtoInSync fails when the embedded template protocol
// differs from src/app/plugins/proto, e.g. after the proto was regenerated
func TestPluginTemplateProtoInSync(t *testing.T) {
	files := map[string]string{
		"plugintemplate/plugin.proto":           "plugin.proto",
		"plugintemplate/plugin.pb.go.tmpl":      "plugin.pb.go",
		"plugintemplate/plugin_grpc.pb.go.tmpl": "plugin_grpc.pb.go",
	}
	for embedded, source := range files {
		want, err := os.ReadFile(filepath.Join("..", "app", "plugins", "proto", source))
		if err != nil {
			t.Fatal(err)
		}
		got, err := pluginTemplateProto.ReadFile(embedded)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("src/cmd/%s differs from src/app/plugins/proto/%s; copy it: cp src/app/plugins/proto/%s src/cmd/%s",
				embedded, source, source, embedded)
		}
	}
}

func TestCreatePluginTemplateWritesHostProtocol(t *testing.T) {
	outputDir := t.TempDir()
	if err := createPluginTemplate("demo-grpc", "test", "", outputDir); err != nil {
		t.Fatal(err)
	}
	pluginDir := filepath.Join(outputDir, "demo-grpc")

	for _, name := range []string{"plugin.proto", "plugin.pb.go", "plugin_grpc.pb.go"} {
		if _, err := os.Stat(filepath.Join(pluginDir, "proto", name)); err != nil {
			t.Errorf("proto/%s not written: %v", name, err)
		}
	}

	mainGo, err := os.ReadFile(filepath.Join(pluginDir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`pb "demo-grpc/proto"`, "pb.RegisterPluginServiceServer", `pluginName    = "demo-grpc"`} {
		if !strings.Contains(string(mainGo), want) {
			t.Errorf("main.go does not contain %q", want)
		}
	}
	if strings.Contains(string(mainGo), "__PLUGIN_NAME__") {
		t.Error("main.go still contains the name placeholder")
	}
}
