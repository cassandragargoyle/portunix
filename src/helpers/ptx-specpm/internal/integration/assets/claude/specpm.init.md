---
description: Run the spec-kit-pm Initiation gate and draft the Project Charter
argument-hint: <project context, source authority, project manager, sponsor>
---

# /specpm.init — Initiation gate

You are the spec-kit-pm **PM agent**. Run the Initiation gate per
`@.specpm/workflows/project-initiation.md`.

Before anything else, follow the preflight in `@.specpm/agents/pm.md`.

If the user supplied $ARGUMENTS, treat them as the source authority,
project manager, and sponsor inputs. Otherwise, read `CLAUDE.local.md` for
the user's `## Identity` block and run the authoring-party inference
algorithm from `@.specpm/agents/pm.md` before asking the user for missing
fields.

Output: a fully populated `specs/project/<project-id>/project.md` Charter
based on `@.specpm/templates/project/charter.md`. Set `status: active`
only after the Sponsor sign-off step is met. Update the governance memory
in `@.specpm/memory/governance.md` with the new project ID.

$ARGUMENTS
