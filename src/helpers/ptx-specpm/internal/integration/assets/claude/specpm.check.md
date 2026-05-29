---
description: Surface ptx-specpm check output and reconcile with project artifacts
---

# /specpm.check — Project state surface

Run `portunix specpm check` and present the output to the user. If the
output indicates drift (e.g., the project pinned at a kit ref different
from the helper default), explain the difference and suggest next steps
(stay pinned, or run `portunix specpm upgrade --ref <ref>` once Phase 2
ships).

If the project is not yet initialised in this directory, instruct the
user to run `portunix specpm init . --integration claude` first.
