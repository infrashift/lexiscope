---
id: "01a1ed42-b891-79de-a788-6bdbe48a56cb"
title: "ISSUE-003: CUE ruleset engine and the ste-100 ruleset"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "open"
  - "phase-3"
  - "ruleset"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T00:05:35-06:00"
description: "Add the CUE schema, the loader, the lexicon projection, the overrides and the ste-100 ruleset."
---

# ISSUE-003: CUE ruleset engine and the ste-100 ruleset

**Status:** open · **Phase:** 3

## Description
Load rulesets from CUE and define the eleven `ste-100` rules.

## Acceptance Criteria
- [ ] `internal/ruleset/schema.cue` defines `#Rule` and the lexicon.
- [ ] CUE unification fails when an approved term is also unapproved or banned.
- [ ] `rules/ste100/ste100.cue` defines STE-100-01 to STE-100-11 with the severities of ADR-0005.
- [ ] Tests show that a user overlay can add a rule, omit a rule, change `max_words` and add approved terms.
- [ ] The docs site has a rules reference.
- [ ] `zmake check` reports zero errors and zero warnings.

## Related
- [ADR-0004](/decisions/adr-0004-cue-rulesets-and-one-lexicon-as-the-source-of-the-substitution-rules--01a1ed42b872.md) and [ADR-0005](/decisions/adr-0005-the-ste-100-rule-ids-with-a-core-group-that-matches-the-oracle-and-an-extended-group--01a1ed42b876.md).
