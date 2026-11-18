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
- lexis unifies the schema, the built-in ruleset, each user file and then the CLI flags (`--enable-rule`, `--omit-rule`, `--set`, `--allow-term`). CUE unification is commutative, so the order of the files does not change the result. A test proves this.
- Each ruleset file compiles in the scope of the schema, so it can refer to `#Severity` and the other definitions.
- `rules/ste100/` embeds the `ste-100` ruleset in the binary.

**Details that Phase 3 settled:**
- **Defaults:** a built-in value that a user can change is a CUE default, for example `max_words: *20 | int` or `severity: *"warning" | #Severity`. A user file or a flag replaces the default. Two concrete values for one field fail with the path of the field and the position in each file. The schema itself has no defaults for `severity` and `description`, because two defaults for one field do not unify. A user rule must state its severity.
- **Closed parameters:** `PLAN.md` asked for an open `params` struct. The schema closes `#Params` instead, so `--set STE-100-01.max_wrods=18` fails. Each check states its required parameters with an `if check == ...` clause.
- **Sets:** each list of terms that a user can extend is a struct with `true` values, for example `tokens: {should: true}`. CUE lists do not merge, but structs do. A file can add a term to a set, but it can not remove a term. The lexicon keeps lists, as `PLAN.md` asked. A user file can add a concept but can not change the lists of a built-in concept.
- **Projection:** `ste100.cue` makes the `swap` maps of STE-100-05 and STE-100-06 with comprehensions over `lexicon`. The comprehension sees the unified lexicon, so a user concept changes both rules.
- **Conflicts:** some terms have two roles: approved in the lexicon or the vocabulary, and also unapproved or banned. For each such term, `#Ruleset` adds a field `conflict.<term>` with two values that do not unify. The load then fails, and the error names the term. `--allow-term should` fails in the same way.
- **CLI values:** `--set` reads its value as CUE: `18` is a number, `["numbered","bullet"]` a list, `{"x": true}` a set. Text that is not a concrete CUE value is a string.
- **Rule IDs:** `#RuleID` is `^[A-Z][A-Z0-9]*(-[A-Z0-9]+)*$`. The CLI checks the same pattern before the load.
- **Keyword:** `try` is a keyword in CUE v0.17, so `ste100.cue` quotes it as a field name.
- **Limits:** 64 sources, 256 CUE files in one directory, 1 MiB for each file and 1024 rules.

## Consequences
- A user ruleset is data, not code. A new kind of check needs a change to lexis.
- The schema closes the `check` field, so a typo in a check name fails at load time.

## Alternatives considered
- **YAML rulesets.** Rejected: no unification, no constraints and no defaults.
- **Rules in Go code.** Rejected: users cannot change them without a build.
