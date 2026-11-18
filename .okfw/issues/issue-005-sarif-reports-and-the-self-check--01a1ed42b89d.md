---
id: "01a1ed42-b89d-7750-9844-16bccaa789d0"
title: "ISSUE-005: SARIF reports and the self-check"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "closed"
  - "phase-5"
  - "report"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T00:05:35-06:00"
description: "Add the SARIF 2.1.0 and 2.2 reports and run the self-check on the repository."
---

# ISSUE-005: SARIF reports and the self-check

**Status:** closed · **Phase:** 5

## Description
Write SARIF reports and prove that lexis passes its own checks.

## Acceptance Criteria
- [x] `--sarif-version` selects SARIF 2.1.0 or 2.2.
- [x] A test proves that the validation rejects a broken 2.2 report, or the docs mark 2.2 as experimental.
- [x] The self-check writes both reports for this repository with zero results.
- [x] The docs site has a SARIF guide.
- [x] `zmake check` reports zero errors and zero warnings.

## Resolution
- `internal/report/sarif.go` builds a SARIF 2.1.0 report with go-sarif v3.3.2. A strict `convert.ToV22` makes the 2.2 report. lexis validates each report before it writes it, and writes nothing when validation fails.
- `tool.driver.rules` holds the active rules. Each result has a relative URI, the start and end line and column, the rule ID, the rule index and the level. A STE-100-05 result has a fix with the approved term.
- `TestValidateRejectsBroken22` and `TestSchema22NeedsNoDraft202012` prove that go-sarif validates 2.2 reports. SARIF 2.2 is not experimental.
- `--fail-on` sets the lowest severity that gives exit status 1.
- `zmake self-check` loads `vocabulary.cue`, fails on each finding, and writes `lexis-self.sarif` and `lexis-self-v22.sarif` with zero results ([ADR-0009](/decisions/adr-0009-the-self-check-fails-on-each-finding-with-a-repository-vocabulary--01a1ee772cc7.md)).
- The docs site has the guide "Write a SARIF report".

## Related
- [ADR-0006](/decisions/adr-0006-sarif-2-1-0-and-2-2-reports-with-go-sarif-v3--01a1ed42b87b.md).
- [ADR-0009](/decisions/adr-0009-the-self-check-fails-on-each-finding-with-a-repository-vocabulary--01a1ee772cc7.md).
