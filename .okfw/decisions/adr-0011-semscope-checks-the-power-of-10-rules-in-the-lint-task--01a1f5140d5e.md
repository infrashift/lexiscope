---
x-okfw-id: "01a1f514-0d5e-7c08-8ece-1b6c888559fe"
title: "ADR-0011: semscope checks the Power of 10 rules in the lint task"
type: "ADR"
tags:
  - "adr"
  - "power-of-10"
x-okfw-created: "2026-11-19T12:31:34-06:00"
x-okfw-timestamp: "2026-11-19T12:31:34-06:00"
description: "zmake lint runs semscope with the nasa-p10 ruleset and the .semscope.cue overlay when semscope is on the PATH, after cmd/nasa-lint."
---

# ADR-0011: semscope checks the Power of 10 rules in the lint task

**Status:** Accepted
**Date:** 2026-11-19
**Source:** User request. zmake ADR-0026, which runs semscope in the zmake lint task.

## Context
[ADR-0003](/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1ed42b86d.md) adapts the Power of 10 rules to Go. `cmd/nasa-lint` checks R1, R2, R4, R5, R7 and R9. semscope checks the same rules with the `nasa-p10` ruleset, and it reads the code with a syntax tree of ztree. A second tool finds defects of the first tool. zmake runs semscope in its lint task for the same reason.

The `nasa-p10` ruleset also checks R3 and R6. ADR-0003 makes R3 a deviation and gives R6 to the compiler and to review. So semscope needs an overlay.

## Decision
- `.semscope.cue` is the overlay. It turns off R3 and R6 and excludes `scripts/`, the Python tools from zmake.
- `zmake lint` runs `cmd/nasa-lint` first. Then it runs `semscope scan -r nasa-p10 -r .semscope.cue --fail-level warning .` when `semscope` is on the PATH. Set `SEMSCOPE` for another command.
- When `semscope` is not on the PATH, the lint task writes a note and continues. `cmd/nasa-lint` then checks the rules alone. CI has no semscope install.
- When the two tools give different findings, one tool has a defect. Fix lexis, or file an issue in lexis or in semscope.

On 2026-11-19, semscope `74e703e` and `cmd/nasa-lint` gave 0 findings, 174 functions and 352 assertions.

## Consequences
- A developer with semscope gets two independent checks of the Power of 10 rules.
- CI checks the rules with `cmd/nasa-lint` only.
- A change to ADR-0003 needs a change to `.semscope.cue`.

## Alternatives considered
- **Replace `cmd/nasa-lint` with semscope.** Rejected: CI cannot install semscope, and the port has its own tests.
- **Fail the lint task without semscope.** Rejected: a new checkout then needs a second repository to pass `zmake check`.
