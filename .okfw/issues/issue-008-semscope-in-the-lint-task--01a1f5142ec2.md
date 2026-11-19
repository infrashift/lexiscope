---
id: "01a1f514-2ec2-76af-b5ef-2698aaf2f956"
title: "ISSUE-008: semscope in the lint task"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "closed"
  - "phase-8"
  - "power-of-10"
created: "2026-11-19T12:31:43-06:00"
timestamp: "2026-11-19T12:31:43-06:00"
description: "zmake lint checks the Power of 10 rules with cmd/nasa-lint only, and lexis has no semscope overlay."
---

# ISSUE-008: semscope in the lint task

**Status:** closed · **Phase:** 8

## Description
`zmake lint` checks the Power of 10 rules with `cmd/nasa-lint` only. semscope checks the same rules with the `nasa-p10` ruleset, and zmake runs it in its lint task. lexis has no semscope overlay, so a plain semscope scan gives R3, R6 and Python R5 findings that [ADR-0003](/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1ed42b86d.md) does not require.

## Acceptance Criteria
- [x] Add `.semscope.cue`. With it, semscope and `cmd/nasa-lint` give the same findings and the same R5 counts.
- [x] `zmake lint` and `make lint` run semscope when it is on the PATH. Without semscope, they write a note and pass.
- [x] Record the decision in [ADR-0011](/decisions/adr-0011-semscope-checks-the-power-of-10-rules-in-the-lint-task--01a1f5140d5e.md).
- [x] Update CLAUDE.md, the docs site and the wiki.
- [x] `zmake check` and `zmake -f Makefile check` pass.

## Resolution
`.semscope.cue` turns off R3 and R6 and excludes `scripts/`. The lint tasks run semscope after `cmd/nasa-lint`. Both tools give 0 findings, 174 functions and 352 assertions. A run with `SEMSCOPE=semscope-missing` writes the note and passes.

The installed semscope was from 2026-11-18, before semscope used ztree. It is now semscope `74e703e`.
