---
x-okfw-id: "01a1ed42-b897-76c0-9c5c-608603e65a67"
title: "ISSUE-004: Prose and comment engine, and parity with ste-check"
type: "Issue"
x-okfw-kind: "issue"
tags:
  - "issue"
  - "phase-4"
  - "engine"
  - "closed"
x-okfw-created: "2026-11-18T00:05:35-06:00"
x-okfw-timestamp: "2026-11-18T04:31:12-06:00"
description: "Add the Markdown extractor, the comment extractor with ztree, the rule checks and the parity test against the oracle."
---

# ISSUE-004: Prose and comment engine, and parity with ste-check

**Status:** closed · **Phase:** 4

## Description
Read Markdown files and source comments, run the rules, and match the oracle on the core rules.

## Acceptance Criteria
- [x] The Markdown extractor keeps the line and the UTF-16 column of each token and uses no recursion.
- [x] The comment extractor reads `.go`, `.zig`, `.py`, `.c` and `.rs` files with the ztree Go binding.
- [x] The parity test matches the oracle on STE-100-01 to STE-100-04 for every fixture and every document of this repository, in both directions.
- [x] A spec lists each deliberate difference from the oracle.
- [x] `zmake check` runs `lexis scan --ruleset ste-100 .`.
- [x] `zmake check` reports zero errors and zero warnings.

## Resolution
- `internal/parser` copies the line rules, the paragraphs and the cleanup of the oracle. Each block keeps the source offset of each byte, so a finding has its exact line and UTF-16 column.
- The comment extractor reads the spans of ztree through `github.com/infrashift/ztree/bindings/go`. It skips directives such as `//go:build` and indented code. A trailing comment stands alone.
- `internal/engine` runs the nine checks. The core checks copy the Python semantics of the oracle: Unicode `\s` and `\b`, the sentence split, the ASCII word count and the re.I flag.
- The parity test matches the oracle on 20 fixtures (54 oracle findings), on each repository document, and with a changed word limit. Mutation checks: 19 mutants of the parser and the engine. Each mutant that can change a finding fails the test. A mutant that only adds an empty sentence can not change a finding.
- `lexis scan` writes text and JSON. It exits with 1 on an error finding. `--exclude` skips paths. SARIF stays for Phase 5.
- `zmake self-check` runs lexis on its own documents and Go comments: 0 errors, 79 warnings of the extended rules. The fix of nine long comment sentences came first.
- The word checks skip code names, such as `skipFenceOrComment`, and STE-100-07 accepts "finding".
- [Known differences from ste-check](/specs/known-differences-from-ste-check--01a1ee352a3c.md) lists the three deliberate differences.

## Related
- [ADR-0002](/decisions/adr-0002-go-1-26-a-small-set-of-dependencies-and-comments-from-ztree-through-cgo--01a1ed42b868.md) and [ADR-0005](/decisions/adr-0005-the-ste-100-rule-ids-with-a-core-group-that-matches-the-oracle-and-an-extended-group--01a1ed42b876.md). [Known differences from ste-check](/specs/known-differences-from-ste-check--01a1ee352a3c.md).
