/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package kit

// DefaultPinnedRef is the upstream spec-kit-pm git ref the helper points at
// when --source git is used without --ref. Bumping this constant is how
// the kit version moves; see ADR-040 D6.
//
// Use a tag rather than a branch — no floating refs in the default path.
// Until upstream cuts its first tag, "main" is acceptable as a transitional
// pin (documented in ADR-040 Revisions).
const DefaultPinnedRef = "main"

// UpstreamURL is the canonical spec-kit-pm repository.
const UpstreamURL = "https://github.com/CassandraGargoyle/spec-kit-pm.git"
