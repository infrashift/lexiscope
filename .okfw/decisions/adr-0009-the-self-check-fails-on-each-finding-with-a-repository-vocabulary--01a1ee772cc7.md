---
x-okfw-id: "01a1ee77-2cc7-79ea-8041-20cd6e5f9eb7"
title: "ADR-0009: The self-check fails on each finding, with a repository vocabulary"
type: "ADR"
tags:
  - "adr"
  - "self-check"
x-okfw-created: "2026-11-18T05:42:30-06:00"
x-okfw-timestamp: "2026-11-18T05:42:30-06:00"
description: "zmake self-check fails on each finding, and vocabulary.cue holds the technical terms of lexis that STE-100-06 and STE-100-07 do not report."
---

# ADR-0009: The self-check fails on each finding, with a repository vocabulary

# ADR-0009: The self-check fails on each finding, with a repository vocabulary

**Status:** Accepted
**Date:** 2026-11-18
**Source:** User decision before Phase 5, [ISSUE-005](/issues/issue-005-sarif-reports-and-the-self-check--01a1ed42b89d.md).

## Context
In Phase 4, `zmake self-check` failed on an error only. It reported 82 warnings of the extended rules in 69 files. `PLAN.md` requires zero errors and zero warnings for each gate, and Phase 5 requires SARIF reports with zero results. Most warnings named technical terms of the project, for example "unification" (CUE) and "decision" (ADR). A few were real faults of style: `therefore`, `modern`, `is done` and `your`.

## Decision
- `vocabulary.cue` in the repository root holds the technical terms of lexis. Each term has a comment that names its concept. The self-check loads it with `--ruleset ste-100 --ruleset vocabulary.cue`.
- A term goes into the vocabulary only when it names a concept of the project, a tool or a standard. The authors rewrite other text. The fix of the 82 warnings put 24 terms in the vocabulary and rewrote 16 sentences. The `Tracking:` label of `PLAN.md` became `Issue:`.
- `lexis scan --fail-on <error|warning|note>` sets the lowest severity that gives exit status 1. The default stays `error` ([ADR-0008](/decisions/adr-0008-cli-exit-statuses-output-errors-and-commands-of-later-phases--01a1ee1490ca.md)).
- `zmake self-check` uses `--fail-on note`, so each finding fails it. It also writes `lexis-self.sarif` and `lexis-self-v22.sarif`. Each scan must have zero results.

## Consequences
- New prose and new comments must pass the extended rules. In Phase 5, the self-check found eight faults in the new SARIF code and docs before the commit.
- A code name in a comment goes in backticks, for example `Validate`, when the identifier skip does not cover it.
- The vocabulary applies to STE-100-06 and STE-100-07 only. The schema rejects a vocabulary term that the lexicon marks as unapproved or that a `banned_terms` rule bans.

## Alternatives considered
- **Keep the errors-only gate and accept the warnings as a baseline.** Rejected: it breaks the zero-warnings rule, and new warnings hide in the old ones.
- **Put each flagged word in the vocabulary.** Rejected: a word such as `therefore` or `your` is a fault, not a term.
- **A suppression comment in the prose.** Rejected for now: lexis has no such syntax, and the vocabulary covers each case.
