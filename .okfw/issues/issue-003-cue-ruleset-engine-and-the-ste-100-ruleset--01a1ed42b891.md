---
id: "01a1ed42-b891-79de-a788-6bdbe48a56cb"
title: "ISSUE-003: CUE ruleset engine and the ste-100 ruleset"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "phase-3"
  - "ruleset"
  - "closed"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T04:10:02-06:00"
description: "Add the CUE schema, the loader, the lexicon projection, the overrides and the ste-100 ruleset."
---

# ISSUE-003: CUE ruleset engine and the ste-100 ruleset

**Status:** closed · **Phase:** 3

## Description
Load rulesets from CUE and define the eleven `ste-100` rules.

## Acceptance Criteria
- [x] `internal/ruleset/schema.cue` defines `#Rule` and the lexicon.
- [x] CUE unification fails when an approved term is also unapproved or banned.
- [x] `rules/ste100/ste100.cue` defines STE-100-01 to STE-100-11 with the severities of ADR-0005.
- [x] Tests show that a user overlay can add a rule, omit a rule, change `max_words` and add approved terms.
- [x] The docs site has a rules reference.
- [x] `zmake check` reports zero errors and zero warnings.

## Resolution
- `internal/ruleset/schema.cue` defines `#Ruleset`, `#Rule`, `#Params` and `#Concept`. `ruleset.Load` unifies the schema, the sources and the CLI changes, then decodes the result.
- `rules/ste100/ste100.cue` defines STE-100-01 to STE-100-11, a lexicon of 20 concepts and an empty vocabulary. STE-100-05 and STE-100-06 come from the lexicon.
- A test reads `scripts/ste-check.py` and compares the word limits, the 13 banned terms, the passive auxiliaries and the `-ed` suffix with rules 01 to 04. Four mutants of the ruleset each fail it.
- Tests prove that a user file can add a rule, omit a rule, change `max_words`, change a severity, add vocabulary and add a lexicon concept. The order of the files does not change the result.
- `lexis rules list` and `lexis rules inspect` work. `lexis scan` loads the ruleset and then names Phase 4.
- The docs site has the guide "Write a Ruleset" and the reference "Ruleset Schema".
- [ADR-0004](/decisions/adr-0004-cue-rulesets-and-one-lexicon-as-the-source-of-the-substitution-rules--01a1ed42b872.md) records the details: defaults, closed parameters, sets as structs, the conflict fields and the CLI values.

## Related
- [ADR-0004](/decisions/adr-0004-cue-rulesets-and-one-lexicon-as-the-source-of-the-substitution-rules--01a1ed42b872.md) and [ADR-0005](/decisions/adr-0005-the-ste-100-rule-ids-with-a-core-group-that-matches-the-oracle-and-an-extended-group--01a1ed42b876.md).
