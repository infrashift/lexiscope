---
id: "01a1ed42-b87b-7b52-83d8-ba09fc8bf50f"
title: "ADR-0006: SARIF 2.1.0 and 2.2 reports with go-sarif v3"
type: "ADR"
tags:
  - "adr"
  - "report"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T00:05:35-06:00"
description: "lexis writes SARIF 2.1.0 by default and SARIF 2.2 on request with go-sarif v3, and proves that validation rejects a broken 2.2 report."
---

# ADR-0006: SARIF 2.1.0 and 2.2 reports with go-sarif v3

**Status:** Accepted
**Date:** 2026-11-18
**Source:** User request, `PLAN.md` Phase 5.

## Context
Code scanning tools read SARIF 2.1.0. SARIF 2.2 is newer. `github.com/owenrumney/go-sarif/v3` (v3.3.2) has the packages `pkg/report/v210/sarif` and `pkg/report/v22/sarif`. The 2.2 schema uses JSON Schema draft 2020-12. The validator of go-sarif, gojsonschema, reads drafts up to draft-07 only.

## Decision
- `--sarif-version 2.1.0` is the default. `--sarif-version 2.2` selects the 2.2 package.
- `tool.driver.rules` holds the active rules of the resolved ruleset only.
- Each result has a relative URI, the start and end line and column, the rule ID and the level. STE-100-05 results also have the replacement text.
- lexis calls `Validate()` before it writes a report.
- A test gives `Validate()` a broken 2.2 report. If `Validate()` accepts it, lexis uses a draft 2020-12 validator for 2.2. If no validator fits, lexis marks 2.2 as experimental.

## Consequences
- Phase 5 adds one Go dependency, and maybe a second one for the 2.2 validator. Each needs this ADR or a new ADR.

## Alternatives considered
- **SARIF 2.1.0 only.** Rejected: the user asked for 2.2 support.
