/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
// Package kit resolves the spec-kit-pm content tree at runtime according to
// the user's --source choice. ADR-040 D2/D3: the kit is NOT embedded in the
// helper binary; it is fetched from upstream (default) or read from a local
// path (forks / air-gapped).
package kit

// Provider supplies a populated kit directory tree. Implementations are
// chosen by the init command based on --source / --ref flags.
type Provider interface {
	// Fetch returns the absolute path to a directory holding the kit tree
	// (templates/, agents/, workflows/, profiles/). The boolean indicates
	// whether the result was served from cache (no network call needed).
	// Implementations are responsible for caching and idempotency.
	Fetch() (kitDir string, fromCache bool, err error)
}
