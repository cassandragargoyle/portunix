/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// handleMetaArg processes the --help-ai / --help-expert flags before Cobra
// parses the arguments. Returning true means the flag was handled.
func handleMetaArg(arg string) bool {
	switch arg {
	case "--help-ai":
		showHelpAI()
		return true
	case "--help-expert":
		showHelpExpert()
		return true
	}
	return false
}

// handleMetaArgs handles the meta flags both for a direct call
// (ptx-<helper> --help-ai) and for a dispatcher call, where the top-level
// command name comes first (ptx-<helper> <command> --help-ai)
func handleMetaArgs(args []string) bool {
	switch len(args) {
	case 1:
		return handleMetaArg(args[0])
	case 2:
		if isTopLevelCommand(args[0]) {
			return handleMetaArg(args[1])
		}
	}
	return false
}

// isTopLevelCommand reports whether name is a direct subcommand of rootCmd
// or one of its aliases
func isTopLevelCommand(name string) bool {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name || c.HasAlias(name) {
			return true
		}
	}
	return false
}

// showHelpAI emits a machine-readable command summary (per Issues #163, #244)
func showHelpAI() {
	type CommandInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	type AIHelp struct {
		Tool        string        `json:"tool"`
		Version     string        `json:"version"`
		Description string        `json:"description"`
		Commands    []CommandInfo `json:"commands"`
	}
	help := AIHelp{
		Tool:        "ptx-database",
		Version:     version,
		Description: "Database engine management: install, lifecycle, backup/restore, introspection (PostgreSQL, SQLite)",
		Commands: []CommandInfo{
			{Name: "database engines", Description: "List supported database engines"},
			{Name: "database install", Description: "Install a database engine via ptx-installer"},
			{Name: "database uninstall", Description: "Uninstall a database engine"},
			{Name: "database start", Description: "Start a database engine or named instance"},
			{Name: "database stop", Description: "Stop a database engine or named instance"},
			{Name: "database restart", Description: "Restart a database engine or named instance"},
			{Name: "database status", Description: "Show status of a database engine or named instance"},
			{Name: "database health", Description: "Ping the engine and report liveness"},
			{Name: "database backup", Description: "Backup a database to a file"},
			{Name: "database restore", Description: "Restore a database from a backup file"},
			{Name: "database list", Description: "List databases on an instance"},
			{Name: "database tables", Description: "List tables in a database"},
			{Name: "database schema", Description: "Dump the schema of a database (DDL or JSON)"},
		},
	}
	data, _ := json.MarshalIndent(help, "", "  ")
	fmt.Println(string(data))
}

// showHelpExpert prints detailed help with flags, modes and examples
func showHelpExpert() {
	fmt.Printf("PTX-DATABASE v%s - Database engine management\n", strings.TrimPrefix(version, "v"))
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  Engine-agnostic command surface for database install, lifecycle,")
	fmt.Println("  backup/restore and introspection. Installation is delegated to")
	fmt.Println("  ptx-installer, container lifecycle to ptx-container. Invoked as")
	fmt.Println("  'portunix database ...' (alias 'portunix db ...').")
	fmt.Println()
	fmt.Println("ENGINES:")
	fmt.Println("  postgresql                         Native or container mode")
	fmt.Println("  sqlite                             Embedded mode (instance = path to .db file)")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  engines                            List supported engines")
	fmt.Println("  install <engine>                   Install an engine via ptx-installer")
	fmt.Println("  uninstall <engine>                 Uninstall an engine")
	fmt.Println("  start|stop|restart [engine|inst]   Engine or named instance lifecycle")
	fmt.Println("  status [engine|instance]           Show engine or instance status")
	fmt.Println("  health [engine|instance]           Ping the engine and report liveness")
	fmt.Println("  backup <database>                  Backup a database to a file")
	fmt.Println("  restore <file>                     Restore a database from a backup file")
	fmt.Println("  list                               List databases on an instance")
	fmt.Println("  tables <database>                  List tables in a database")
	fmt.Println("  schema <database>                  Dump the schema (DDL or JSON)")
	fmt.Println()
	fmt.Println("GLOBAL FLAGS:")
	fmt.Println("  -e, --engine <name>                postgresql|sqlite (or positional argument)")
	fmt.Println("      --instance <name>              Instance name, or .db path for sqlite")
	fmt.Println("      --mode <mode>                  native|container|embedded")
	fmt.Println("      --format json                  Machine-readable output")
	fmt.Println()
	fmt.Println("INSTALL FLAGS:")
	fmt.Println("      --version <ver>                Engine version (e.g. 16)")
	fmt.Println("      --variant <name>               Package variant override (e.g. container)")
	fmt.Println("      --dry-run                      Preview without changing the system")
	fmt.Println()
	fmt.Println("BACKUP FLAGS:")
	fmt.Println("      --destination <dir>            Backup destination directory")
	fmt.Println("      --compress                     Compress the backup output")
	fmt.Println()
	fmt.Println("RESTORE FLAGS:")
	fmt.Println("      --target-db <name>             Target database (postgres only)")
	fmt.Println("      --yes                          Confirm destructive restore")
	fmt.Println()
	fmt.Println("SCHEMA FLAGS:")
	fmt.Println("      --schema-format sql|json       Schema format (default sql)")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  portunix database install postgresql --version 16")
	fmt.Println("  portunix database install postgresql --variant container --dry-run")
	fmt.Println("  portunix db status postgresql --format json")
	fmt.Println("  portunix db backup mydb -e postgresql --destination ./backups --compress")
	fmt.Println("  portunix db restore ./backups/mydb.sql -e postgresql --target-db mydb --yes")
	fmt.Println("  portunix db tables main -e sqlite --instance ./app.db")
}
