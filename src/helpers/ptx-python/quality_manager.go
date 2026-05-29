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
	"strings"
)

// QualityManager handles Python code quality operations:
// syntax check, lint, format, typecheck, test.
type QualityManager struct {
	venvManager *VenvManager
}

// NewQualityManager creates a new quality manager
func NewQualityManager() (*QualityManager, error) {
	vm, err := NewVenvManager()
	if err != nil {
		return nil, err
	}
	return &QualityManager{venvManager: vm}, nil
}

// LintOptions controls Lint behavior.
type LintOptions struct {
	Target     *VenvTarget
	Paths      []string
	Linter     string // "ruff" (default), "pylint", "flake8"
	OutputFile string
	Format     string // "text" (default), "json", "html"
	ExtraArgs  []string
}

// FormatOptions controls Format behavior.
type FormatOptions struct {
	Target    *VenvTarget
	Paths     []string
	Formatter string // "black" (default), "autopep8"
	Check     bool   // do not modify, exit code 1 if changes would be made
	ExtraArgs []string
}

// TypecheckOptions controls Typecheck behavior.
type TypecheckOptions struct {
	Target    *VenvTarget
	Paths     []string
	ExtraArgs []string
}

// TestOptions controls Test behavior.
type TestOptions struct {
	Target    *VenvTarget
	Paths     []string
	Coverage  bool
	Watch     bool
	ExtraArgs []string
}

// resolvePythonExe returns the python executable for the given target,
// falling back to system python3 (or python on Windows) when target is nil.
func (qm *QualityManager) resolvePythonExe(target *VenvTarget) string {
	if target != nil && target.Path != "" && qm.venvManager.VenvExistsAtPath(target.Path) {
		return qm.venvManager.getPythonExecutable(target.Path)
	}
	if runtime.GOOS == "windows" {
		return "python"
	}
	return "python3"
}

// resolveToolExe returns absolute path to a tool binary inside the target venv
// when available, otherwise just the tool name (resolved via PATH).
func (qm *QualityManager) resolveToolExe(target *VenvTarget, tool string) string {
	if target == nil || target.Path == "" || !qm.venvManager.VenvExistsAtPath(target.Path) {
		return tool
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(target.Path, "Scripts", tool+".exe")
	}
	return filepath.Join(target.Path, "bin", tool)
}

// ensureToolInstalled installs the given pip package into the target venv (or
// system python) when "pip show" reports it missing. distName is the pip
// package name; importName is the module/binary name expected post-install
// (often identical to distName, but e.g. "pytest-cov" -> "pytest_cov").
func (qm *QualityManager) ensureToolInstalled(target *VenvTarget, distName string) error {
	pythonExe := qm.resolvePythonExe(target)

	check := exec.Command(pythonExe, "-m", "pip", "show", distName)
	output, err := check.CombinedOutput()
	if err == nil && strings.Contains(strings.ToLower(string(output)), "name:") {
		return nil
	}

	fmt.Printf("Installing %s...\n", distName)
	install := exec.Command(pythonExe, "-m", "pip", "install", distName)
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr
	return install.Run()
}

// CheckSyntax validates Python syntax via the AST parser. Walks directories
// recursively. Returns an error summarising failures, but always reports
// every file (no fail-fast) so the user sees the full picture.
func (qm *QualityManager) CheckSyntax(target *VenvTarget, paths []string) error {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	pythonExe := qm.resolvePythonExe(target)
	files, err := collectPythonFiles(paths)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Println("No Python files found.")
		return nil
	}

	fmt.Printf("Checking syntax of %d file(s)...\n", len(files))
	failed := 0
	for _, f := range files {
		// python -c 'import ast; ast.parse(open(...).read())'
		// Use a small inline snippet for portability across 3.x.
		script := "import ast,sys; ast.parse(open(sys.argv[1],'rb').read(), filename=sys.argv[1])"
		cmd := exec.Command(pythonExe, "-c", script, f)
		out, err := cmd.CombinedOutput()
		if err != nil {
			failed++
			fmt.Printf("❌ %s\n%s\n", f, strings.TrimSpace(string(out)))
			continue
		}
	}
	if failed > 0 {
		return fmt.Errorf("syntax check failed for %d file(s)", failed)
	}
	fmt.Printf("✅ Syntax OK (%d file(s) checked)\n", len(files))
	return nil
}

// Lint runs the selected linter against given paths. Defaults to ruff.
func (qm *QualityManager) Lint(opts LintOptions) error {
	linter := opts.Linter
	if linter == "" {
		linter = "ruff"
	}

	if err := qm.ensureToolInstalled(opts.Target, linter); err != nil {
		return fmt.Errorf("failed to install linter %s: %v", linter, err)
	}

	paths := opts.Paths
	if len(paths) == 0 {
		paths = []string{"."}
	}

	args := buildLinterArgs(linter, opts, paths)
	tool := qm.resolveToolExe(opts.Target, linter)

	fmt.Printf("Running %s on %s...\n", linter, strings.Join(paths, " "))
	cmd := exec.Command(tool, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// buildLinterArgs constructs CLI arguments for the selected linter, including
// the optional --format / --output flags. Output and format mappings are
// linter-specific.
func buildLinterArgs(linter string, opts LintOptions, paths []string) []string {
	args := []string{}
	switch linter {
	case "ruff":
		args = append(args, "check")
		switch opts.Format {
		case "json":
			args = append(args, "--output-format=json")
		case "github":
			args = append(args, "--output-format=github")
		}
		if opts.OutputFile != "" {
			args = append(args, "--output-file", opts.OutputFile)
		}
	case "flake8":
		if opts.Format == "json" {
			args = append(args, "--format=json")
		}
		if opts.OutputFile != "" {
			args = append(args, "--output-file="+opts.OutputFile)
		}
	case "pylint":
		switch opts.Format {
		case "json":
			args = append(args, "--output-format=json")
		case "html":
			args = append(args, "--output-format=html")
		}
		if opts.OutputFile != "" {
			args = append(args, "--output="+opts.OutputFile)
		}
	}
	args = append(args, opts.ExtraArgs...)
	args = append(args, paths...)
	return args
}

// Format runs black (default) or autopep8 against given paths. With Check=true
// no files are modified; black exits non-zero if changes are needed (used by
// CI). autopep8 with Check=true uses --diff to print proposed changes.
func (qm *QualityManager) Format(opts FormatOptions) error {
	formatter := opts.Formatter
	if formatter == "" {
		formatter = "black"
	}

	if err := qm.ensureToolInstalled(opts.Target, formatter); err != nil {
		return fmt.Errorf("failed to install formatter %s: %v", formatter, err)
	}

	paths := opts.Paths
	if len(paths) == 0 {
		paths = []string{"."}
	}

	args := []string{}
	switch formatter {
	case "black":
		if opts.Check {
			args = append(args, "--check", "--diff")
		}
	case "autopep8":
		if opts.Check {
			args = append(args, "--diff")
		} else {
			args = append(args, "--in-place", "--recursive")
		}
	}
	args = append(args, opts.ExtraArgs...)
	args = append(args, paths...)

	tool := qm.resolveToolExe(opts.Target, formatter)
	fmt.Printf("Running %s on %s...\n", formatter, strings.Join(paths, " "))
	cmd := exec.Command(tool, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Typecheck runs mypy against the given paths.
func (qm *QualityManager) Typecheck(opts TypecheckOptions) error {
	if err := qm.ensureToolInstalled(opts.Target, "mypy"); err != nil {
		return fmt.Errorf("failed to install mypy: %v", err)
	}

	paths := opts.Paths
	if len(paths) == 0 {
		paths = []string{"."}
	}

	args := append([]string{}, opts.ExtraArgs...)
	args = append(args, paths...)

	tool := qm.resolveToolExe(opts.Target, "mypy")
	fmt.Printf("Running mypy on %s...\n", strings.Join(paths, " "))
	cmd := exec.Command(tool, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Test runs pytest, optionally with coverage and watch mode.
// Watch mode requires pytest-watch (auto-installed).
func (qm *QualityManager) Test(opts TestOptions) error {
	if err := qm.ensureToolInstalled(opts.Target, "pytest"); err != nil {
		return fmt.Errorf("failed to install pytest: %v", err)
	}
	if opts.Coverage {
		if err := qm.ensureToolInstalled(opts.Target, "pytest-cov"); err != nil {
			return fmt.Errorf("failed to install pytest-cov: %v", err)
		}
	}
	if opts.Watch {
		if err := qm.ensureToolInstalled(opts.Target, "pytest-watch"); err != nil {
			return fmt.Errorf("failed to install pytest-watch: %v", err)
		}
	}

	var tool string
	args := []string{}
	if opts.Watch {
		tool = qm.resolveToolExe(opts.Target, "ptw")
		if opts.Coverage {
			args = append(args, "--", "--cov")
		}
	} else {
		tool = qm.resolveToolExe(opts.Target, "pytest")
		if opts.Coverage {
			args = append(args, "--cov")
		}
	}
	args = append(args, opts.ExtraArgs...)
	args = append(args, opts.Paths...)

	cmd := exec.Command(tool, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// collectPythonFiles walks the given paths (files or directories) and returns
// all .py files. Files are passed through verbatim; directories are scanned
// recursively. Hidden dirs (".venv", "__pycache__", ".git") are skipped.
func collectPythonFiles(paths []string) ([]string, error) {
	var files []string
	skipDirs := map[string]bool{
		".venv":       true,
		"venv":        true,
		"__pycache__": true,
		".git":        true,
		".tox":        true,
		".mypy_cache": true,
		".ruff_cache": true,
	}
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("cannot stat %s: %v", p, err)
		}
		if !info.IsDir() {
			if strings.HasSuffix(p, ".py") {
				files = append(files, p)
			}
			continue
		}
		walkErr := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			if !info.IsDir() && strings.HasSuffix(path, ".py") {
				files = append(files, path)
			}
			return nil
		})
		if walkErr != nil {
			return nil, walkErr
		}
	}
	return files, nil
}
