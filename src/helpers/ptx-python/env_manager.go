/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// EnvManager handles Python env reporting, project scaffolding and security
// audit. It bundles a handful of conceptually-related features that don't
// warrant their own files.
type EnvManager struct {
	venvManager *VenvManager
}

// NewEnvManager creates a new env manager.
func NewEnvManager() (*EnvManager, error) {
	vm, err := NewVenvManager()
	if err != nil {
		return nil, err
	}
	return &EnvManager{venvManager: vm}, nil
}

// ShowEnv prints Python-relevant environment configuration: PYTHON*, VIRTUAL_ENV,
// and the active venv (auto-detected ./.venv when not set), plus PATH highlights.
func (em *EnvManager) ShowEnv() error {
	fmt.Println("Python Environment:")
	fmt.Println()

	pyVars := []string{
		"VIRTUAL_ENV",
		"PYTHONPATH",
		"PYTHONHOME",
		"PYTHONSTARTUP",
		"PYTHONDONTWRITEBYTECODE",
		"PYTHONUNBUFFERED",
		"PIP_INDEX_URL",
		"PIP_EXTRA_INDEX_URL",
	}
	sort.Strings(pyVars)
	for _, v := range pyVars {
		val := os.Getenv(v)
		if val == "" {
			val = "(unset)"
		}
		fmt.Printf("  %-25s = %s\n", v, val)
	}

	// Active venv: $VIRTUAL_ENV or auto-detected ./.venv
	fmt.Println()
	active := os.Getenv("VIRTUAL_ENV")
	if active == "" {
		cwd, _ := os.Getwd()
		local := filepath.Join(cwd, ".venv")
		if em.venvManager.VenvExistsAtPath(local) {
			active = local + " (auto-detected, not activated)"
		}
	}
	if active == "" {
		fmt.Println("Active venv: (none)")
	} else {
		fmt.Printf("Active venv: %s\n", active)
	}

	return nil
}

// SetEnv emits a shell snippet the user can eval to set a Python-related env
// variable. We deliberately do NOT mutate the parent shell (impossible from a
// helper binary) — instead we print POSIX and Windows variants so the user
// chooses what fits.
func (em *EnvManager) SetEnv(name, value string) error {
	if name == "" {
		return fmt.Errorf("variable name required")
	}
	fmt.Println("# Run one of the following in your shell:")
	fmt.Println()
	fmt.Println("# POSIX (bash/zsh):")
	fmt.Printf("export %s=%s\n", name, shellQuote(value))
	fmt.Println()
	fmt.Println("# Windows PowerShell:")
	fmt.Printf("$env:%s = %s\n", name, psQuote(value))
	fmt.Println()
	fmt.Println("# Windows cmd.exe:")
	fmt.Printf("set %s=%s\n", name, value)
	return nil
}

func shellQuote(s string) string {
	if s == "" {
		return "\"\""
	}
	if strings.ContainsAny(s, " \t\"$'`\\") {
		return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
	}
	return s
}

func psQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// InitProjectOptions controls InitProject behavior.
type InitProjectOptions struct {
	Name     string
	Template string // console|library|web|package
	Path     string // optional target directory (defaults to ./<Name>)
}

// InitProject scaffolds a new Python project. Templates differ in the
// boilerplate generated but all produce pyproject.toml, a src/<pkg>/ package,
// tests/, README.md and .gitignore.
func (em *EnvManager) InitProject(opts InitProjectOptions) error {
	if opts.Name == "" {
		return fmt.Errorf("project name required")
	}
	template := opts.Template
	if template == "" {
		template = "console"
	}
	if !isValidTemplate(template) {
		return fmt.Errorf("unknown template %q (valid: console, library, web, package)", template)
	}

	root := opts.Path
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		root = filepath.Join(cwd, opts.Name)
	}

	if _, err := os.Stat(root); err == nil {
		return fmt.Errorf("target %s already exists", root)
	}

	pkg := normalizePackageName(opts.Name)

	// Directory layout
	dirs := []string{
		filepath.Join(root, "src", pkg),
		filepath.Join(root, "tests"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create %s: %v", d, err)
		}
	}

	files := map[string]string{
		filepath.Join(root, "pyproject.toml"):          renderPyproject(opts.Name, pkg, template),
		filepath.Join(root, "README.md"):               renderReadme(opts.Name, template),
		filepath.Join(root, ".gitignore"):              renderGitignore(),
		filepath.Join(root, "src", pkg, "__init__.py"): renderInit(pkg),
		filepath.Join(root, "src", pkg, "main.py"):     renderMain(template),
		filepath.Join(root, "tests", "__init__.py"):    "",
		filepath.Join(root, "tests", "test_smoke.py"):  renderSmokeTest(pkg),
	}

	if template == "web" {
		files[filepath.Join(root, "src", pkg, "app.py")] = renderWebApp(pkg)
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %v", path, err)
		}
	}

	fmt.Printf("✅ Project '%s' scaffolded at %s\n", opts.Name, root)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", opts.Name)
	fmt.Println("  portunix python init             # create ./.venv")
	fmt.Println("  portunix python pip install -e . # install in editable mode")
	return nil
}

func isValidTemplate(t string) bool {
	switch t {
	case "console", "library", "web", "package":
		return true
	}
	return false
}

// normalizePackageName turns "My-Cool App" into "my_cool_app" so it's a valid
// Python package identifier.
func normalizePackageName(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + 32)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	out := b.String()
	if out == "" || (out[0] >= '0' && out[0] <= '9') {
		out = "pkg_" + out
	}
	return out
}

func renderPyproject(name, pkg, template string) string {
	deps := ""
	if template == "web" {
		deps = "    \"flask>=3.0\",\n"
	}
	return fmt.Sprintf(`[build-system]
requires = ["setuptools>=61.0"]
build-backend = "setuptools.build_meta"

[project]
name = "%s"
version = "0.1.0"
description = "%s project (%s template)"
requires-python = ">=3.9"
dependencies = [
%s]

[project.scripts]
%s = "%s.main:main"

[tool.setuptools.packages.find]
where = ["src"]
`, name, name, template, deps, pkg, pkg)
}

func renderReadme(name, template string) string {
	return fmt.Sprintf("# %s\n\nGenerated with `portunix python init project` (%s template).\n\n## Quick start\n\n```bash\nportunix python init        # create local venv\nportunix python pip install -e .\nportunix python test\n```\n", name, template)
}

func renderGitignore() string {
	return `__pycache__/
*.py[cod]
.venv/
build/
dist/
*.egg-info/
.coverage
.pytest_cache/
.mypy_cache/
.ruff_cache/
`
}

func renderInit(pkg string) string {
	return fmt.Sprintf("\"\"\"%s package.\"\"\"\n\n__version__ = \"0.1.0\"\n", pkg)
}

func renderMain(template string) string {
	switch template {
	case "web":
		return `"""Entry point for the web app."""
from .app import app


def main() -> None:
    app.run(host="127.0.0.1", port=8000, debug=True)


if __name__ == "__main__":
    main()
`
	case "library":
		return `"""Library entry point (typically unused for libraries)."""


def main() -> None:
    print("This is a library; import it from another project.")


if __name__ == "__main__":
    main()
`
	default:
		return `"""Console entry point."""


def main() -> None:
    print("Hello from portunix python project!")


if __name__ == "__main__":
    main()
`
	}
}

func renderWebApp(pkg string) string {
	return fmt.Sprintf(`"""Minimal Flask app for %s."""
from flask import Flask

app = Flask(__name__)


@app.route("/")
def index() -> str:
    return "Hello from %s!"
`, pkg, pkg)
}

func renderSmokeTest(pkg string) string {
	return fmt.Sprintf(`"""Smoke test ensuring the package imports."""
import %s


def test_package_imports() -> None:
    assert %s.__version__
`, pkg, pkg)
}

// Audit runs pip-audit in the resolved venv (or system python).
// Falls back to "safety" if pip-audit fails to install.
func (em *EnvManager) Audit(target *VenvTarget, extraArgs []string) error {
	pythonExe := em.resolvePythonExe(target)

	if err := em.ensureAuditTool(target, "pip-audit"); err != nil {
		fmt.Println("⚠️  pip-audit unavailable, trying safety...")
		if err2 := em.ensureAuditTool(target, "safety"); err2 != nil {
			return fmt.Errorf("no audit tool could be installed: %v / %v", err, err2)
		}
		args := append([]string{"-m", "safety", "check"}, extraArgs...)
		cmd := exec.Command(pythonExe, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	args := append([]string{"-m", "pip_audit"}, extraArgs...)
	cmd := exec.Command(pythonExe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (em *EnvManager) resolvePythonExe(target *VenvTarget) string {
	if target != nil && target.Path != "" && em.venvManager.VenvExistsAtPath(target.Path) {
		return em.venvManager.getPythonExecutable(target.Path)
	}
	if runtime.GOOS == "windows" {
		return "python"
	}
	return "python3"
}

func (em *EnvManager) ensureAuditTool(target *VenvTarget, tool string) error {
	pythonExe := em.resolvePythonExe(target)
	check := exec.Command(pythonExe, "-m", "pip", "show", tool)
	out, err := check.CombinedOutput()
	if err == nil && strings.Contains(strings.ToLower(string(out)), "name:") {
		return nil
	}
	fmt.Printf("Installing %s...\n", tool)
	install := exec.Command(pythonExe, "-m", "pip", "install", tool)
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr
	return install.Run()
}
