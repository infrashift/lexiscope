---
id: "01a1ed42-b876-7d06-b9ee-df3065fe5c9a"
title: "ADR-0005: The ste-100 rule IDs, with a core group that matches the oracle and an extended group"
type: "ADR"
tags:
  - "adr"
  - "ruleset"
  - "ste-100"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T00:05:35-06:00"
description: "STE-100-01 to 04 match scripts/ste-check.py exactly in both directions; STE-100-05 to 11 are warnings with exception lists and their own fixtures."
---

# ADR-0005: The ste-100 rule IDs, with a core group that matches the oracle and an extended group

**Status:** Accepted
**Date:** 2026-11-18
**Source:** User decision. Analysis of `scripts/ste-check.py`.

## Context
`PLAN.md` asks for full parity with `scripts/ste-check.py`. The oracle has four checks only:

1. A numbered list item has 20 words or fewer.
2. Each other sentence has 25 words or fewer.
3. A be-verb before a word that ends in `-ed` gives a passive warning.
4. 13 banned words and phrases give errors.

The oracle has no check for synonyms, nominalizations, `-ing` forms, `your`, marketing words or irregular participles. A rule that the oracle does not have cannot have parity with it.

## Decision
- **Core group:** STE-100-01 to STE-100-04 copy the four checks of the oracle. The parity test checks both directions: each oracle finding has a lexis finding, and each lexis finding has an oracle finding.
- **Extended group:** STE-100-05 to STE-100-11 give warnings by default. Each has an exception list and its own fixtures. The parity test does not compare them.
- A step is a numbered list item, as in the oracle. The `procedural_markers` parameter can add bullet items.
- lexis finds imperative sentences with a list of verbs. It has no part-of-speech tagger.

| ID | Check | Group |
|---|---|---|
| STE-100-01 | Step length, 20 words | core |
| STE-100-02 | Sentence length, 25 words | core |
| STE-100-03 | Passive voice with `-ed` | core |
| STE-100-04 | The 13 banned words and phrases | core |
| STE-100-05 | Synonyms from the lexicon | extended |
| STE-100-06 | Nominalizations | extended |
| STE-100-07 | Unclear `-ing` forms | extended |
| STE-100-08 | `your` and `yours` | extended |
| STE-100-09 | Marketing words | extended |
| STE-100-10 | Irregular passive participles | extended |
| STE-100-11 | One instruction per sentence | extended |

**Parity test:** the test copies each fixture into a temporary `docs/src/content/docs/` directory and runs the unchanged oracle there. The oracle reports the first line of each paragraph. The test maps each lexis position to the first line of its paragraph before it compares. A deliberate difference, for example a `~~~` fence, gets its own fixture and an entry in a spec.

**Phase 4 result:** the parity test runs the unchanged oracle on 20 fixtures and on each repository document. It also runs a copy of the oracle with a step limit of 15, against lexis with `--set STE-100-01.max_words=15`. The extractor copies the Python semantics of the oracle: Unicode `\s` and `\b`, universal newlines, and the order of the cleanup steps. [Known differences from ste-check](/specs/known-differences-from-ste-check--01a1ee352a3c.md) lists the three deliberate differences.

**Self-check:** `zmake self-check` runs `lexis scan` on the documents of the oracle and on the Go comments. In Phase 4, an error failed the gate and a warning did not. [ADR-0009](/decisions/adr-0009-the-self-check-fails-on-each-finding-with-a-repository-vocabulary--01a1ee772cc7.md) replaces this policy: each finding fails the gate, and `vocabulary.cue` holds the technical terms of lexis.

## Consequences
- Rules 01 to 04 keep the limits of the oracle, for example its word count.
- The extended rules can give false warnings. Their exception lists and fixtures control this.

## Alternatives considered
- **Parity for all rules.** Rejected: the oracle has no counterpart for rules 05 to 11.
- **Extended rules as errors.** Rejected: the heuristics give false findings, and the gate needs zero errors.
