---
x-okfw-id: "01a1ed42-b87b-7b52-83d8-ba09fc8bf50f"
title: "ADR-0006: SARIF 2.1.0 and 2.2 reports with go-sarif v3"
type: "ADR"
tags:
  - "adr"
  - "report"
x-okfw-created: "2026-11-18T00:05:35-06:00"
x-okfw-timestamp: "2026-11-18T00:05:35-06:00"
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

**Phase 5 result:**
- lexis builds one 2.1.0 report. `convert.ToV22` of go-sarif makes the 2.2 report, with `WithStrictConversion`. So lexis has one report builder, not two.
- `Validate()` of go-sarif rejects broken 2.2 reports. The test `TestValidateRejectsBroken22` breaks eight constraints. Four are fields of the report and of a result: `version`, `guid`, `level` and `kind`. The others are a `startLine` of 0, the level of a rule, an absent driver name and a bad `columnKind`. Each case fails validation, and a mutant of the test that breaks nothing fails the test.
- `TestSchema22NeedsNoDraft202012` reads the 2.2 schema of go-sarif. The schema uses no keyword that only draft 2019-09 or 2020-12 has, and no `$ref` with a sibling constraint. So draft 07 semantics give the same result as draft 2020-12 for this schema. lexis needs no second validator, and SARIF 2.2 is not experimental in lexis. A go-sarif version with a changed schema fails the test.
- **A fault of go-sarif:** its package comment calls the 2.1.0 to 2.2 conversion lossless, but the conversion drops custom entries of a property bag. The strict conversion found this, so the rules of lexis have no `properties`.
- Fixes: a STE-100-05 result has a fix that replaces the term with the approved term. A STE-100-06 swap gives a verb for a noun, so it is not a text fix.
- Each 2.2 report has a new random `guid`, as SARIF 2.2 requires. Two 2.2 reports of one scan are not equal byte for byte.

## Consequences
- Phase 5 adds one Go dependency: go-sarif v3.3.2, with gojsonschema. go-sarif needs `go 1.26.10` in `go.mod`.
- The go-sarif structs write each empty array and the value -1 for each unset index, so a report has many fields with no data. The schema accepts them.

## Alternatives considered
- **SARIF 2.1.0 only.** Rejected: the user asked for 2.2 support.
