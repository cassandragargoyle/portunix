package engine

import (
	"os"
	"path/filepath"
	"testing"

	"portunix.ai/app/wizard"
)

func TestRegisterWizardBytes(t *testing.T) {
	eng := NewWizardEngine()

	yaml := []byte(`wizard:
  id: "registered"
  name: "Registered Wizard"
  version: "1.0"
  pages:
    - id: "p1"
      type: "info"
      title: "Hello"
`)

	wiz, err := eng.RegisterWizardBytes(yaml)
	if err != nil {
		t.Fatalf("RegisterWizardBytes failed: %v", err)
	}
	if wiz.ID != "registered" {
		t.Errorf("expected id 'registered', got %q", wiz.ID)
	}

	loaded := eng.ListLoadedWizards()
	if _, ok := loaded["registered"]; !ok {
		t.Errorf("expected wizard 'registered' in ListLoadedWizards, got %v", loaded)
	}
}

func TestRegisterWizardBytes_MissingID(t *testing.T) {
	eng := NewWizardEngine()
	_, err := eng.RegisterWizardBytes([]byte(`wizard:
  name: "no-id"
  pages:
    - id: "p1"
      type: "info"
`))
	if err == nil {
		t.Fatal("expected error for wizard without id, got nil")
	}
}

func TestLoadVariablesFromConfig(t *testing.T) {
	tempDir := t.TempDir()
	cfg := filepath.Join(tempDir, "vars.yaml")
	if err := os.WriteFile(cfg, []byte("db_type: postgresql\nport: 5432\nfeatures:\n  - backup\n  - monitoring\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	eng := NewWizardEngine()
	if err := eng.LoadVariablesFromConfig(cfg); err != nil {
		t.Fatalf("LoadVariablesFromConfig failed: %v", err)
	}

	wiz := &wizard.Wizard{
		ID:    "t",
		Pages: []wizard.Page{{ID: "p1", Type: wizard.PageTypeInfo, Title: "Done"}},
	}
	eng.SetNonInteractive(true)
	res, err := eng.ExecuteWizard(wiz)
	if err != nil {
		t.Fatalf("ExecuteWizard failed: %v", err)
	}
	if res.Variables["db_type"] != "postgresql" {
		t.Errorf("expected db_type=postgresql, got %v", res.Variables["db_type"])
	}
	if res.Variables["port"] != 5432 {
		t.Errorf("expected port=5432, got %v", res.Variables["port"])
	}
}

func TestNonInteractive_SelectFromPreloadedVars(t *testing.T) {
	eng := NewWizardEngine()
	yaml := []byte(`wizard:
  id: "ni-select"
  name: "Non-interactive Select"
  version: "1.0"
  pages:
    - id: "pick"
      type: "select"
      title: "Pick"
      prompt: "DB:"
      options:
        - value: "pg"
          label: "PostgreSQL"
        - value: "my"
          label: "MySQL"
      variable: "db"
    - id: "done"
      type: "success"
      title: "Done"
`)
	wiz, err := eng.RegisterWizardBytes(yaml)
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	tempDir := t.TempDir()
	cfg := filepath.Join(tempDir, "v.yaml")
	if err := os.WriteFile(cfg, []byte("db: pg\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := eng.LoadVariablesFromConfig(cfg); err != nil {
		t.Fatal(err)
	}
	eng.SetNonInteractive(true)

	res, err := eng.ExecuteWizard(wiz)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !res.Completed {
		t.Errorf("expected completed=true, got %+v", res)
	}
	if res.Variables["db"] != "pg" {
		t.Errorf("expected db=pg, got %v", res.Variables["db"])
	}
}

func TestNonInteractive_SelectRejectsUnknownValue(t *testing.T) {
	eng := NewWizardEngine()
	yaml := []byte(`wizard:
  id: "ni-bad"
  name: "Bad value"
  version: "1.0"
  pages:
    - id: "pick"
      type: "select"
      title: "Pick"
      prompt: "DB:"
      options:
        - value: "pg"
          label: "PostgreSQL"
      variable: "db"
`)
	wiz, _ := eng.RegisterWizardBytes(yaml)

	tempDir := t.TempDir()
	cfg := filepath.Join(tempDir, "v.yaml")
	_ = os.WriteFile(cfg, []byte("db: oracle\n"), 0o644)
	_ = eng.LoadVariablesFromConfig(cfg)
	eng.SetNonInteractive(true)

	_, err := eng.ExecuteWizard(wiz)
	if err == nil {
		t.Fatal("expected error for value not in option list, got nil")
	}
}

func TestNonInteractive_MissingVariableErrors(t *testing.T) {
	eng := NewWizardEngine()
	yaml := []byte(`wizard:
  id: "ni-missing"
  name: "Missing var"
  version: "1.0"
  pages:
    - id: "pick"
      type: "select"
      title: "Pick"
      prompt: "DB:"
      options:
        - value: "pg"
          label: "PostgreSQL"
      variable: "db"
`)
	wiz, _ := eng.RegisterWizardBytes(yaml)
	eng.SetNonInteractive(true)

	_, err := eng.ExecuteWizard(wiz)
	if err == nil {
		t.Fatal("expected error for missing preloaded value, got nil")
	}
}
