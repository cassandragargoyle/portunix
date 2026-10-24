/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package manager

import (
	"archive/zip"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// pluginPythonVersionEnv overrides the interpreter minor used for wheel
// plugins that do not pin plugin.python_version
const pluginPythonVersionEnv = "PORTUNIX_PLUGIN_PYTHON_VERSION"

// defaultPluginPythonVersion returns the interpreter minor for wheel plugins
// without plugin.python_version: the env override, else a mature CPython minor
func defaultPluginPythonVersion() string {
	if v := strings.TrimSpace(os.Getenv(pluginPythonVersionEnv)); v != "" {
		return v
	}
	return matureCPythonMinor(time.Now())
}

// matureCPythonMinor returns the newest CPython minor released at least six
// months before now; per PEP 602 CPython 3.N.0 ships each October of year
// 2011+N (3.13 in 2024-10), so a just-released minor that native dependencies
// may not publish wheels for yet is skipped
func matureCPythonMinor(now time.Time) string {
	cutoff := now.AddDate(0, -6, 0)
	minor := cutoff.Year() - 2011
	if cutoff.Month() < time.October {
		minor--
	}
	return fmt.Sprintf("3.%d", minor)
}

// cpythonTagPattern matches a single CPython wheel tag such as "cp313"
var cpythonTagPattern = regexp.MustCompile(`^cp3(\d{1,2})$`)

// pycMagicRanges maps .pyc magic number ranges to the CPython minor version
// (ranges as reserved in CPython Lib/importlib/_bootstrap_external.py)
var pycMagicRanges = []struct {
	low, high uint16
	version   string
}{
	{3400, 3419, "3.8"},
	{3420, 3429, "3.9"},
	{3430, 3449, "3.10"},
	{3450, 3499, "3.11"},
	{3500, 3549, "3.12"},
	{3550, 3599, "3.13"},
	{3600, 3649, "3.14"},
	{3650, 3699, "3.15"},
}

// detectWheelPythonVersion returns the CPython "major.minor" a wheel is bound
// to, or empty string for a version-independent wheel; the wheel filename tag
// is used first, then the magic number of sourceless bytecode inside it
func detectWheelPythonVersion(wheelPath string) string {
	if v := wheelTagPythonVersion(filepath.Base(wheelPath)); v != "" {
		return v
	}
	return wheelBytecodePythonVersion(wheelPath)
}

// wheelTagPythonVersion reads the python tag of a wheel filename
// ({dist}-{version}(-{build})?-{python}-{abi}-{platform}.whl)
func wheelTagPythonVersion(filename string) string {
	parts := strings.Split(strings.TrimSuffix(filename, ".whl"), "-")
	if len(parts) < 5 {
		return ""
	}
	// Compressed tag sets (e.g. "cp312.cp313") are not bound to one version
	m := cpythonTagPattern.FindStringSubmatch(parts[len(parts)-3])
	if m == nil {
		return ""
	}
	return "3." + m[1]
}

// wheelBytecodePythonVersion reads the magic number of the first sourceless
// .pyc in the wheel; cached bytecode in __pycache__ is ignored because the
// interpreter recompiles it from the shipped sources
func wheelBytecodePythonVersion(wheelPath string) string {
	r, err := zip.OpenReader(wheelPath)
	if err != nil {
		return ""
	}
	defer r.Close()

	for _, f := range r.File {
		if !strings.HasSuffix(f.Name, ".pyc") || strings.Contains(f.Name, "__pycache__/") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		header := make([]byte, 4)
		_, err = io.ReadFull(rc, header)
		_ = rc.Close()
		if err != nil || header[2] != '\r' || header[3] != '\n' {
			continue
		}
		return pycMagicVersion(binary.LittleEndian.Uint16(header[:2]))
	}
	return ""
}

// pycMagicVersion maps a .pyc magic number to the CPython minor version
func pycMagicVersion(magic uint16) string {
	for _, r := range pycMagicRanges {
		if magic >= r.low && magic <= r.high {
			return r.version
		}
	}
	return ""
}
