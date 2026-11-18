---
id: "01a1ed42-b897-76c0-9c5c-608603e65a67"
title: "ISSUE-004: Prose and comment engine, and parity with ste-check"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "open"
  - "phase-4"
  - "engine"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T00:05:35-06:00"
description: "Add the Markdown extractor, the comment extractor with ztree, the rule checks and the parity test against the oracle."
---

# ISSUE-004: Prose and comment engine, and parity with ste-check

**Status:** open · **Phase:** 4

## Description
Read Markdown files and source comments, run the rules, and match the oracle on the core rules.

## Acceptance Criteria
- [ ] The Markdown extractor keeps the line and the UTF-16 column of each token and uses no recursion.
- [ ] The comment extractor reads `.go`, `.zig`, `.py`, `.c` and `.rs` files with the ztree Go binding.
- [ ] The parity test matches the oracle on STE-100-01 to STE-100-04 for every fixture and every document of this repository, in both directions.
- [ ] A spec lists each deliberate difference from the oracle.
- [ ] `zmake check` runs `lexis scan --ruleset ste-100 .`.
- [ ] `zmake check` reports zero errors and zero warnings.

## Related
- [ADR-0002](/decisions/adr-0002-go-1-26-a-small-set-of-dependencies-and-comments-from-ztree-through-cgo--01a1ed42b868.md) and [ADR-0005](/decisions/adr-0005-the-ste-100-rule-ids-with-a-core-group-that-matches-the-oracle-and-an-extended-group--01a1ed42b876.md).
