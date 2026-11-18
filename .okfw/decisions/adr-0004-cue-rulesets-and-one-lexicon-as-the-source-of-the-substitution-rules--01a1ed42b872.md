---
id: "01a1ed42-b872-70ed-807e-3f6ca3893b88"
title: "ADR-0004: CUE rulesets and one lexicon as the source of the substitution rules"
type: "ADR"
tags:
  - "adr"
  - "ruleset"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T00:05:35-06:00"
description: "lexis defines each ruleset in CUE, keys rules by stable ID, and makes the substitution and nominalization rules from one lexicon."
---

# ADR-0004: CUE rulesets and one lexicon as the source of the substitution rules

**Status:** Accepted
**Date:** 2026-11-18
**Source:** User request, `PLAN.md` Phase 3.

## Context
Users need to add rules, turn off rules, change limits and add approved terms. A synonym list and a nominalization list for the same concept drift apart when they live in two places.

## Decision
- Each ruleset is a CUE file. `internal/ruleset/schema.cue` defines `#Rule` and the lexicon. `cuelang.org/go` loads and unifies the files.
- The `rules` map has the stable ID as its key: `rules: [ID=string]: #Rule & {id: ID}`.
- A rule has `enabled`, `severity`, `scope`, `title`, `message`, `params` and `check` fields. `check` selects one of a fixed set of check functions.
- The lexicon maps each concept to one approved term, its unapproved synonyms and its nominalizations. lexis makes the substitution rule (STE-100-05) and the nominalization rule (STE-100-06) from the lexicon.
- CUE unification fails when an approved term is also unapproved or banned.
- The order of changes: the built-in ruleset, then each user overlay file, then the CLI flags (`--omit-rule`, `--enable-rule`, `--set`, `--allow-term`).
- `rules/ste100/` embeds the `ste-100` ruleset in the binary.

## Consequences
- A user ruleset is data, not code. A new kind of check needs a change to lexis.
- The schema closes the `check` field, so a typo in a check name fails at load time.

## Alternatives considered
- **YAML rulesets.** Rejected: no unification, no constraints and no defaults.
- **Rules in Go code.** Rejected: users cannot change them without a build.
