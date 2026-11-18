---
title: Rules
description: The rules of the ste-100 ruleset, and the Power of 10 rules that the lexis source follows.
sidebar:
  order: 1
---

`rules/ste100/ste100.cue` defines these rules. Run `lexis rules inspect ste-100` to see each parameter, the lexicon and the vocabulary. To change a rule, see [Write a ruleset](../../guides/rulesets/).

## The ste-100 ruleset

The ruleset follows `scripts/ste-check.py`, which zmake and ztree use as their writing gate.
Rules 01 to 04 form the core group. They match the script exactly. A test compares their limits, their banned words and their passive pattern with the script. From Phase 4, a parity test also compares the findings in both directions.
Rules 05 to 11 form the extended group. Each extended rule gives a warning by default and has an exception list.

| ID | Check | Severity | Group |
|---|---|---|---|
| STE-100-01 | A numbered step has 20 words or fewer | error | core |
| STE-100-02 | Each other sentence has 25 words or fewer | error | core |
| STE-100-03 | A be-verb before a word that ends in `-ed` (passive voice) | warning | core |
| STE-100-04 | The 13 banned words and phrases of the script | error | core |
| STE-100-05 | A synonym of an approved term from the lexicon | warning | extended |
| STE-100-06 | A nominalization, for example `-tion`, `-ment` or `-ance` | warning | extended |
| STE-100-07 | An unclear `-ing` form | warning | extended |
| STE-100-08 | `your` or `yours` | warning | extended |
| STE-100-09 | A marketing word, for example `seamless` or `robust` | warning | extended |
| STE-100-10 | An irregular passive participle, for example `is written` | warning | extended |
| STE-100-11 | More than one instruction in a sentence | warning | extended |

The banned words and phrases of STE-100-04 are `should`, `will`, `won't`, `would`, `please`, `simply`, `just`, `very`, `really`, `easily`, `is expected to`, `are expected to` and `can be seen`.

A numbered list item is a step. Each other sentence is descriptive. The `procedural_markers` parameter can also make bullet items steps.

## Power of 10 rules for the lexis source

ADR-0003 in the wiki adapts the NASA Power of 10 rules to Go. `cmd/nasa-lint` is the Go port of `scripts/nasa-lint.py`. `zmake lint` runs it. The port checks these rules:

| Rule | Check |
|---|---|
| R1 | No recursion, direct or indirect, and no `goto` |
| R2 | No `for {}` loop without a bound |
| R4 | Each function has 60 lines or fewer |
| R5 | An average of two or more `assert.` calls per function |
| R7 | No `_ =` discard, and no blank result of a call, for example `x, _ := f()` |
| R9 | No function types, no function literals and no function values |

Test files (`_test.go`) are exempt. Rule R3 is a deviation: lexis sets hard caps on each buffer, but Go allocates on the heap.

To allow one finding, add a comment on the same line or the line above, for example `// nasa-lint: allow(R9) <reason>`. Give a reason in each comment.
