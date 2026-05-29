/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// ptx-database is a Portunix helper binary that orchestrates database
// installation, lifecycle, backup/restore, and introspection. It delegates
// installation to ptx-installer and container lifecycle to ptx-container; this
// helper owns the engine-agnostic command surface and the driver layer.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"portunix.ai/ptx-database/drivers"
)

var version = "dev"

// Shared flags. Most commands accept --engine + --mode + --instance.
var (
	flagEngine   string
	flagMode     string
	flagInstance string
	flagFormat   string
)

var rootCmd = &cobra.Command{
	Use:   "ptx-database",
	Short: "Portunix Database Helper",
	Long: `ptx-database is a helper binary for the portunix dispatcher that manages
database systems: install, configure, lifecycle, backup/restore, and introspection.

Phase 1 ships PostgreSQL (native + container) and SQLite (embedded). The helper
delegates installation to ptx-installer and container lifecycle to ptx-container.

This binary is invoked through 'portunix database ...' (alias 'portunix db ...').`,
	Version: version,
}

// dbCmd is the user-facing top-level command. The dispatcher routes both
// "database" and "db" to this binary; cobra needs only the canonical name.
var dbCmd = &cobra.Command{
	Use:     "database",
	Aliases: []string{"db"},
	Short:   "Manage database engines (install, lifecycle, backup, introspection)",
}

// ── install / uninstall ─────────────────────────────────────────────────────

var installCmd = &cobra.Command{
	Use:   "install <engine>",
	Short: "Install a database engine via ptx-installer",
	Long: `Install a database engine. Delegates package resolution and download
to ptx-installer using the manifests in assets/packages/.

Examples:
  portunix database install postgresql --version 16
  portunix database install postgresql --mode container
  portunix database install sqlite`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		engine := args[0]
		ver, _ := cmd.Flags().GetString("version")
		variant, _ := cmd.Flags().GetString("variant")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Resolve package name. Engine name maps directly to a JSON manifest.
		pkg := engine

		installArgs := []string{"install", pkg}
		if variant != "" {
			installArgs = append(installArgs, "--variant", variant)
		} else if flagMode == string(drivers.ModeContainer) {
			installArgs = append(installArgs, "--variant", "container")
		}
		if ver != "" {
			installArgs = append(installArgs, "--version", ver)
		}
		if dryRun {
			installArgs = append(installArgs, "--dry-run")
		}
		return runPortunix(installArgs...)
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <engine>",
	Short: "Uninstall a database engine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Phase 1: just delegate to ptx-installer's uninstall flow if present.
		// We surface a clear error if it is not yet wired up there.
		purge, _ := cmd.Flags().GetBool("purge")
		uninstallArgs := []string{"package", "uninstall", args[0]}
		if purge {
			uninstallArgs = append(uninstallArgs, "--purge")
		}
		return runPortunix(uninstallArgs...)
	},
}

// ── lifecycle ────────────────────────────────────────────────────────────────

var startCmd = &cobra.Command{
	Use:   "start [engine|instance]",
	Short: "Start a database engine or named instance",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		d, opts, err := resolveDriver(args)
		if err != nil {
			return err
		}
		return d.Start(opts)
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop [engine|instance]",
	Short: "Stop a database engine or named instance",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		d, opts, err := resolveDriver(args)
		if err != nil {
			return err
		}
		return d.Stop(opts)
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart [engine|instance]",
	Short: "Restart a database engine or named instance",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		d, opts, err := resolveDriver(args)
		if err != nil {
			return err
		}
		return d.Restart(opts)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status [engine|instance]",
	Short: "Show status of a database engine or named instance",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		d, opts, err := resolveDriver(args)
		if err != nil {
			return err
		}
		st, err := d.Status(opts)
		if err != nil {
			return err
		}
		return emit(st)
	},
}

var healthCmd = &cobra.Command{
	Use:   "health [engine|instance]",
	Short: "Ping the engine and report liveness",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		d, opts, err := resolveDriver(args)
		if err != nil {
			return err
		}
		rep, err := d.Health(opts)
		if err != nil {
			return err
		}
		return emit(rep)
	},
}

// ── introspection ────────────────────────────────────────────────────────────

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List databases on an instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		d, opts, err := resolveDriver(nil)
		if err != nil {
			return err
		}
		dbs, err := d.ListDatabases(opts)
		if err != nil {
			return err
		}
		return emit(dbs)
	},
}

var tablesCmd = &cobra.Command{
	Use:   "tables <database>",
	Short: "List tables in a database",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		d, opts, err := resolveDriver(nil)
		if err != nil {
			return err
		}
		tables, err := d.ListTables(opts, args[0])
		if err != nil {
			return err
		}
		return emit(tables)
	},
}

var schemaCmd = &cobra.Command{
	Use:   "schema <database>",
	Short: "Dump the schema of a database (DDL or JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		d, opts, err := resolveDriver(nil)
		if err != nil {
			return err
		}
		fmtFlag, _ := cmd.Flags().GetString("schema-format")
		if fmtFlag == "" {
			fmtFlag = "sql"
		}
		s, err := d.DumpSchema(opts, args[0], fmtFlag)
		if err != nil {
			return err
		}
		// Schema dumps are big; print body verbatim regardless of --format
		// because most users want raw SQL/JSON, not wrapped envelopes.
		fmt.Print(s.Body)
		if len(s.Body) > 0 && s.Body[len(s.Body)-1] != '\n' {
			fmt.Println()
		}
		return nil
	},
}

// ── backup / restore ─────────────────────────────────────────────────────────

var backupCmd = &cobra.Command{
	Use:   "backup <database>",
	Short: "Backup a database to a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		d, opts, err := resolveDriver(nil)
		if err != nil {
			return err
		}
		dest, _ := cmd.Flags().GetString("destination")
		compress, _ := cmd.Flags().GetBool("compress")
		res, err := d.Backup(opts, drivers.BackupOpts{
			Database:    args[0],
			Destination: dest,
			Compress:    compress,
		})
		if err != nil {
			return err
		}
		return emit(res)
	},
}

var restoreCmd = &cobra.Command{
	Use:   "restore <file>",
	Short: "Restore a database from a backup file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			return fmt.Errorf("restore is destructive — re-run with --yes to confirm")
		}
		d, opts, err := resolveDriver(nil)
		if err != nil {
			return err
		}
		target, _ := cmd.Flags().GetString("target-db")
		return d.Restore(opts, drivers.RestoreOpts{File: args[0], TargetDB: target})
	},
}

// ── meta commands ────────────────────────────────────────────────────────────

var enginesCmd = &cobra.Command{
	Use:   "engines",
	Short: "List supported database engines",
	RunE: func(cmd *cobra.Command, args []string) error {
		return emit(drivers.List())
	},
}

// ── helpers ──────────────────────────────────────────────────────────────────

// resolveDriver returns the driver and lifecycle options based on positional
// arg + global flags. positional arg (if any) overrides --engine.
func resolveDriver(args []string) (drivers.Driver, drivers.LifecycleOpts, error) {
	engine := flagEngine
	if len(args) > 0 && args[0] != "" {
		engine = args[0]
	}
	if engine == "" {
		return nil, drivers.LifecycleOpts{}, fmt.Errorf(
			"engine required: pass as argument or --engine (supported: %v)",
			drivers.List())
	}
	d, err := drivers.Get(engine)
	if err != nil {
		return nil, drivers.LifecycleOpts{}, err
	}
	mode := drivers.Mode(flagMode)
	if mode == "" {
		// Pick a sensible default: native for service engines, embedded for SQLite.
		modes := d.SupportedModes()
		if len(modes) == 1 {
			mode = modes[0]
		} else {
			mode = drivers.ModeNative
		}
	}
	return d, drivers.LifecycleOpts{Instance: flagInstance, Mode: mode}, nil
}

// emit prints v as JSON when --format json is set; otherwise prints a
// human-readable representation. JSON is the canonical wire format for
// MCP tools and machine consumers.
func emit(v interface{}) error {
	if flagFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	// Default: pretty-print as JSON too. ptx-trace and ptx-credential follow
	// the same rule — table layouts can come later when a clear UX is agreed.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// runPortunix shells out to the parent dispatcher to delegate work to
// other helpers (ptx-installer, ptx-container). Streaming I/O so the user
// sees progress for long-running installs.
func runPortunix(args ...string) error {
	cmd := exec.Command("portunix", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() {
	// Persistent flags inherited by all sub-commands.
	dbCmd.PersistentFlags().StringVarP(&flagEngine, "engine", "e", "",
		"Database engine (postgresql|sqlite); also accepted as positional argument")
	dbCmd.PersistentFlags().StringVar(&flagMode, "mode", "",
		"Execution mode: native|container|embedded")
	dbCmd.PersistentFlags().StringVar(&flagInstance, "instance", "",
		"Instance name (or path to .db file for sqlite)")
	dbCmd.PersistentFlags().StringVar(&flagFormat, "format", "",
		"Output format: json (default behaves the same)")

	installCmd.Flags().String("version", "", "Engine version (e.g. 16)")
	installCmd.Flags().String("variant", "", "Package variant override (e.g. container)")
	installCmd.Flags().Bool("dry-run", false, "Preview without changing the system")

	uninstallCmd.Flags().Bool("purge", false, "Also remove data directories and config")

	schemaCmd.Flags().String("schema-format", "sql", "Schema format: sql|json")

	backupCmd.Flags().String("destination", "", "Backup destination directory")
	backupCmd.Flags().Bool("compress", false, "Compress the backup output")

	restoreCmd.Flags().String("target-db", "", "Target database to restore into (postgres only)")
	restoreCmd.Flags().Bool("yes", false, "Confirm destructive restore")

	dbCmd.AddCommand(
		installCmd, uninstallCmd,
		startCmd, stopCmd, restartCmd, statusCmd, healthCmd,
		listCmd, tablesCmd, schemaCmd,
		backupCmd, restoreCmd,
		enginesCmd,
	)
	rootCmd.AddCommand(dbCmd)
	rootCmd.SetVersionTemplate("ptx-database version {{.Version}}\n")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
