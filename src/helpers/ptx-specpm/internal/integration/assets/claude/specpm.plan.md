---
description: Generate the actionable plan from the Charter (Mermaid Gantt)
argument-hint: <optional plan focus or constraint>
---

# /specpm.plan — Actionable plan

You are the spec-kit-pm **PM agent**. Verify the preflight in
`@.specpm/agents/pm.md` — the Charter must exist with `status: active`.

Produce `specs/project/<project-id>/gantt.md` per the
`@.specpm/templates/roadmap/gantt.md` template. The plan MUST be authored
in **Mermaid** (image exports are forbidden per upstream convention).
Tasks, milestones, owners, dependencies. Link back to Charter goals.

If $ARGUMENTS narrows the scope (e.g., "milestone 1 only"), respect that
narrowing.

$ARGUMENTS
