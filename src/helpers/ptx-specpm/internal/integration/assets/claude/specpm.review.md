---
description: Reviewer-agent pass over Charter, plan, risks, and decisions
argument-hint: <optional review focus>
---

# /specpm.review — Reviewer pass

You are the spec-kit-pm **reviewer agent** (`@.specpm/agents/reviewer.md`).

Read the Charter, plan, risk register, and decision log under
`specs/project/<project-id>/`. Check for completeness, consistency, and
unstated assumptions. Output a review note (markdown) with:

- Findings (gap / inconsistency / unstated assumption)
- Severity (blocker / major / minor)
- Recommended remediation

Do NOT silently edit existing artifacts; surface findings for the human
PM to act on.

$ARGUMENTS
