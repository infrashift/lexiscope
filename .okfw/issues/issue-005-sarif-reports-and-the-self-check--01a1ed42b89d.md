---
id: "01a1ed42-b89d-7750-9844-16bccaa789d0"
title: "ISSUE-005: SARIF reports and the self-check"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "open"
  - "phase-5"
  - "report"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T00:05:35-06:00"
description: "Add the SARIF 2.1.0 and 2.2 reports and run the self-check on the repository."
---

# ISSUE-005: SARIF reports and the self-check

**Status:** open · **Phase:** 5

## Description
Write SARIF reports and prove that lexis passes its own checks.

## Acceptance Criteria
- [ ] `--sarif-version` selects SARIF 2.1.0 or 2.2.
- [ ] A test proves that the validation rejects a broken 2.2 report, or the docs mark 2.2 as experimental.
- [ ] The self-check writes both reports for this repository with zero results.
- [ ] The docs site has a SARIF guide.
- [ ] `zmake check` reports zero errors and zero warnings.

## Related
- [ADR-0006](/decisions/adr-0006-sarif-2-1-0-and-2-2-reports-with-go-sarif-v3--01a1ed42b87b.md).
