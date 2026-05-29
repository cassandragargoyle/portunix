/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package drivers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// postgresqlDriver implements Driver for PostgreSQL native + container modes.
//
// Native mode shells out to systemctl on Linux (or pg_ctl on Windows), and
// uses local psql/pg_dump/pg_restore that the user installed via
// `portunix database install postgresql`. Container mode delegates lifecycle
// to `portunix container` and runs psql via `container exec`.
type postgresqlDriver struct{}

func newPostgresqlDriver() Driver { return postgresqlDriver{} }

func init() { Register(newPostgresqlDriver()) }

func (postgresqlDriver) Engine() string         { return "postgresql" }
func (postgresqlDriver) SupportedModes() []Mode { return []Mode{ModeNative, ModeContainer} }

func (d postgresqlDriver) Status(opts LifecycleOpts) (*Status, error) {
	st := &Status{Engine: d.Engine(), Instance: opts.Instance, Mode: opts.Mode, Port: 5432}
	switch opts.Mode {
	case ModeContainer:
		name := containerName(opts.Instance, "postgres")
		running, err := containerRunning(name)
		if err != nil {
			return nil, err
		}
		st.Running = running
		if running {
			if v, _ := d.versionFromPsql(opts); v != "" {
				st.Version = v
			}
		}
		return st, nil
	case ModeNative, "":
		if runtime.GOOS == "linux" {
			out, err := exec.Command("systemctl", "is-active", "postgresql").CombinedOutput()
			st.Running = err == nil && strings.TrimSpace(string(out)) == "active"
			st.Message = strings.TrimSpace(string(out))
		} else {
			st.Message = "native status check not implemented for this OS"
		}
		if v, _ := d.versionFromPsql(opts); v != "" {
			st.Version = v
		}
		return st, nil
	default:
		return nil, fmt.Errorf("postgresql: unsupported mode %q", opts.Mode)
	}
}

func (d postgresqlDriver) Start(opts LifecycleOpts) error {
	switch opts.Mode {
	case ModeContainer:
		name := containerName(opts.Instance, "postgres")
		return runContainer("start", name)
	default:
		if runtime.GOOS != "linux" {
			return fmt.Errorf("postgresql: native start only implemented on linux (use --mode container)")
		}
		return runCmd("systemctl", "start", "postgresql")
	}
}

func (d postgresqlDriver) Stop(opts LifecycleOpts) error {
	switch opts.Mode {
	case ModeContainer:
		name := containerName(opts.Instance, "postgres")
		return runContainer("stop", name)
	default:
		if runtime.GOOS != "linux" {
			return fmt.Errorf("postgresql: native stop only implemented on linux (use --mode container)")
		}
		return runCmd("systemctl", "stop", "postgresql")
	}
}

func (d postgresqlDriver) Restart(opts LifecycleOpts) error {
	if err := d.Stop(opts); err != nil {
		return err
	}
	return d.Start(opts)
}

func (d postgresqlDriver) Health(opts LifecycleOpts) (*HealthReport, error) {
	rep := &HealthReport{Engine: d.Engine()}
	start := time.Now()
	out, err := d.psql(opts, "", "SELECT version();")
	rep.Latency = strconv.FormatInt(time.Since(start).Milliseconds(), 10)
	if err != nil {
		rep.OK = false
		rep.Message = strings.TrimSpace(out)
		return rep, nil
	}
	rep.OK = true
	rep.Message = firstNonEmptyLine(out)
	return rep, nil
}

func (d postgresqlDriver) ListDatabases(opts LifecycleOpts) ([]Database, error) {
	out, err := d.psql(opts, "postgres",
		`SELECT datname, pg_get_userbyid(datdba), pg_size_pretty(pg_database_size(datname)) `+
			`FROM pg_database WHERE datistemplate = false ORDER BY datname;`)
	if err != nil {
		return nil, fmt.Errorf("psql failed: %w (output: %s)", err, out)
	}
	var dbs []Database
	for _, line := range nonEmptyLines(out) {
		parts := splitPipe(line)
		if len(parts) < 1 {
			continue
		}
		db := Database{Name: parts[0]}
		if len(parts) > 1 {
			db.Owner = parts[1]
		}
		if len(parts) > 2 {
			db.Size = parts[2]
		}
		dbs = append(dbs, db)
	}
	return dbs, nil
}

func (d postgresqlDriver) ListTables(opts LifecycleOpts, db string) ([]Table, error) {
	if db == "" {
		return nil, fmt.Errorf("database name required")
	}
	// pg_stat_user_tables exposes the table name as `relname` (not `tablename` —
	// `tablename` is a column on pg_tables, a different view).
	out, err := d.psql(opts, db,
		`SELECT schemaname, relname, `+
			`COALESCE(n_live_tup, 0)::bigint, `+
			`pg_size_pretty(pg_total_relation_size(schemaname||'.'||relname)) `+
			`FROM pg_stat_user_tables ORDER BY schemaname, relname;`)
	if err != nil {
		return nil, fmt.Errorf("psql failed: %w (output: %s)", err, out)
	}
	var tables []Table
	for _, line := range nonEmptyLines(out) {
		parts := splitPipe(line)
		if len(parts) < 2 {
			continue
		}
		t := Table{Schema: parts[0], Name: parts[1]}
		if len(parts) > 2 {
			if n, e := strconv.ParseInt(parts[2], 10, 64); e == nil {
				t.Rows = n
			}
		}
		if len(parts) > 3 {
			t.Size = parts[3]
		}
		tables = append(tables, t)
	}
	return tables, nil
}

func (d postgresqlDriver) DumpSchema(opts LifecycleOpts, db string, format string) (*SchemaDef, error) {
	if db == "" {
		return nil, fmt.Errorf("database name required")
	}
	if format == "" {
		format = "sql"
	}
	switch format {
	case "sql":
		args := []string{"--schema-only", "--no-owner", "--no-acl", db}
		out, err := d.execTool(opts, "pg_dump", args...)
		if err != nil {
			return nil, fmt.Errorf("pg_dump failed: %w", err)
		}
		return &SchemaDef{Format: "sql", Body: out}, nil
	case "json":
		// Minimal JSON schema: list of tables with columns. Built via psql.
		query := `SELECT table_schema || '.' || table_name || '|' || column_name || '|' || data_type ` +
			`FROM information_schema.columns WHERE table_schema NOT IN ('pg_catalog','information_schema') ` +
			`ORDER BY table_schema, table_name, ordinal_position;`
		out, err := d.psql(opts, db, query)
		if err != nil {
			return nil, fmt.Errorf("psql failed: %w", err)
		}
		return &SchemaDef{Format: "json", Body: jsonifyColumns(out)}, nil
	default:
		return nil, fmt.Errorf("unsupported format %q (want sql|json)", format)
	}
}

func (d postgresqlDriver) Backup(opts LifecycleOpts, b BackupOpts) (*BackupResult, error) {
	if b.Database == "" {
		return nil, fmt.Errorf("database name required")
	}
	dest, err := resolveBackupDir(b.Destination, "postgres", opts.Instance)
	if err != nil {
		return nil, err
	}
	stamp := time.Now().UTC().Format("20060102T150405Z")
	ext := ".sql"
	if b.Compress {
		ext = ".sql.gz"
	}
	file := filepath.Join(dest, fmt.Sprintf("%s-%s%s", b.Database, stamp, ext))

	if opts.Mode == ModeContainer {
		// Run pg_dump inside the container, redirect via shell pipe.
		name := containerName(opts.Instance, "postgres")
		shell := fmt.Sprintf("pg_dump -U postgres %s", shellQuote(b.Database))
		if b.Compress {
			shell += " | gzip"
		}
		out, err := exec.Command("portunix", "container", "exec", name, "sh", "-c", shell).Output()
		if err != nil {
			return nil, fmt.Errorf("pg_dump (container) failed: %w", err)
		}
		if err := os.WriteFile(file, out, 0o600); err != nil {
			return nil, err
		}
	} else {
		args := []string{b.Database}
		if b.Compress {
			// pg_dump -Fc gives a compressed custom-format archive
			args = append([]string{"-Fc"}, args...)
		}
		out, err := exec.Command("pg_dump", args...).Output()
		if err != nil {
			return nil, fmt.Errorf("pg_dump (native) failed: %w", err)
		}
		if err := os.WriteFile(file, out, 0o600); err != nil {
			return nil, err
		}
	}

	info, _ := os.Stat(file)
	size := int64(0)
	if info != nil {
		size = info.Size()
	}
	return &BackupResult{
		Path: file, SizeBytes: size, Engine: "postgresql",
		Database: b.Database, Compress: b.Compress, Timestamp: stamp,
	}, nil
}

func (d postgresqlDriver) Restore(opts LifecycleOpts, r RestoreOpts) error {
	if r.File == "" {
		return fmt.Errorf("backup file required")
	}
	target := r.TargetDB
	if target == "" {
		return fmt.Errorf("--target-db is required for postgresql restore")
	}
	if opts.Mode == ModeContainer {
		name := containerName(opts.Instance, "postgres")
		data, err := os.ReadFile(r.File)
		if err != nil {
			return err
		}
		shell := fmt.Sprintf("psql -U postgres -d %s", shellQuote(target))
		cmd := exec.Command("portunix", "container", "exec", "-i", name, "sh", "-c", shell)
		cmd.Stdin = strings.NewReader(string(data))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	if strings.HasSuffix(r.File, ".dump") || strings.HasSuffix(r.File, ".pgcustom") {
		return runCmd("pg_restore", "-d", target, r.File)
	}
	cmd := exec.Command("psql", "-d", target, "-f", r.File)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// psql runs a query using psql (native) or `container exec` (container).
// db may be empty to connect to the default ("postgres") database.
func (d postgresqlDriver) psql(opts LifecycleOpts, db, query string) (string, error) {
	if db == "" {
		db = "postgres"
	}
	if opts.Mode == ModeContainer {
		name := containerName(opts.Instance, "postgres")
		cmd := exec.Command("portunix", "container", "exec", name,
			"psql", "-U", "postgres", "-d", db, "-Atc", query)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	cmd := exec.Command("psql", "-d", db, "-Atc", query)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (d postgresqlDriver) execTool(opts LifecycleOpts, tool string, args ...string) (string, error) {
	if opts.Mode == ModeContainer {
		name := containerName(opts.Instance, "postgres")
		// Inside the official postgres image the OS user is `root`, but the
		// only PostgreSQL role is `postgres`. Mirror what psql() already does
		// and prepend -U postgres for tools that accept it (pg_dump, pg_restore,
		// pg_isready, …) so they don't fail with role-does-not-exist errors.
		toolArgs := append([]string{"-U", "postgres"}, args...)
		full := append([]string{"container", "exec", name, tool}, toolArgs...)
		out, err := exec.Command("portunix", full...).CombinedOutput()
		return string(out), err
	}
	out, err := exec.Command(tool, args...).CombinedOutput()
	return string(out), err
}

func (d postgresqlDriver) versionFromPsql(opts LifecycleOpts) (string, error) {
	out, err := d.psql(opts, "", "SHOW server_version;")
	if err != nil {
		return "", err
	}
	return firstNonEmptyLine(out), nil
}
