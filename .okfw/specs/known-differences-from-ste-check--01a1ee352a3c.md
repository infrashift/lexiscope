---
id: "01a1ee35-2a3c-7eb6-852d-94d88b100402"
title: "Known differences from ste-check"
type: "Spec"
tags:
  - "spec"
  - "ste-100"
  - "parity"
created: "2026-11-18T04:30:24-06:00"
timestamp: "2026-11-18T04:30:24-06:00"
description: "The deliberate differences between lexiscope and scripts/ste-check.py, the oracle behavior that lexiscope copies, and the parity test."
---

# Known differences from ste-check

This spec lists each deliberate difference between lexiscope and `scripts/ste-check.py`, the oracle of the core rules STE-100-01 to STE-100-04 ([ADR-0005](/decisions/adr-0005-the-ste-100-rule-ids-with-a-core-group-that-matches-the-oracle-and-an-extended-group--01a1ed42b876.md)). Each difference has a fixture in `internal/engine/testdata/ste100/divergent/`. `TestDivergences` proves that the oracle reports each fixture and that lexiscope does not.

## Differences

| Input | Oracle | lexiscope | Fixture |
|---|---|---|---|
| A `~~~` fence | Checks the lines inside the fence as prose. | Skips the fence, as for a ``` fence. | `d_tilde.md` |
| A multi-line HTML comment (`<!--` without `-->` on the same line) | Skips the first line only, then checks the commented lines. | Skips each line up to the line with `-->`. | `d_comment.md` |
| TOML front matter (`+++`) | Checks the front matter as prose. | Skips it, as for YAML front matter. | `d_toml.md` |

## Behavior that lexiscope copies on purpose

These look like defects, but lexiscope copies them, so the core rules keep parity:

- A ``` line inside a fence closes the fence, even with a different length or an info string.
- Indented code blocks (four spaces) are prose.
- lexiscope skips each line that starts with `#`, also `#hashtag`.
- The okfw marker `<!-- BEGIN okfw -->` stops the check only in a file with the name `CLAUDE.md`. In each other file, the text after the marker is prose.
- `**` and `*` go before the sentence split and the word count, so `wi*ll` is the word `will`.
- A word counts when it has an ASCII letter or digit. A word in another script alone does not count.
- White space and word characters follow Python's Unicode rules for `str` patterns: `\s` includes the no-break space, and `\b` uses Unicode letters and digits.
- Line breaks follow Python's universal newlines: `\r\n`, `\r` and `\n`.

## Additions that do not change the core findings

- lexiscope reports the exact line and the UTF-16 column of each finding. The oracle reports the first line of the paragraph. The parity test compares the first line of the block (`BlockLine`).
- lexiscope checks headings with rules that have the scope `heading`. The core rules do not read headings, as in the oracle.
- lexiscope checks source code comments. The oracle reads Markdown only.

## Test

`internal/engine/ste_parity_test.go` runs the unchanged oracle and lexiscope on the fixtures and on each document of the repository. It also runs a copy of the oracle with a word limit of 15. It compares the findings as multisets, in both directions.
