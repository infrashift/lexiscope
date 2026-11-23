---
x-okfw-id: "01a1f474-ef49-79bb-b1f1-e76fc33d0a35"
title: "ISSUE-006: ztree comment coverage"
type: "Issue"
x-okfw-kind: "issue"
tags:
  - "issue"
  - "closed"
  - "phase-6"
  - "parser"
x-okfw-created: "2026-11-19T09:37:46-06:00"
x-okfw-timestamp: "2026-11-19T09:37:46-06:00"
description: "Test the comments of each ztree grammar that lexis reads, and record what lexis does when ztree cannot parse a file."
---

# ISSUE-006: ztree comment coverage

**Status:** closed · **Phase:** 6

## Description
lexis reads comments with ztree ([ADR-0002](/decisions/adr-0002-go-1-26-a-small-set-of-dependencies-and-comments-from-ztree-through-cgo--01a1ed42b868.md)). The tests read the comments of Go and Python only. No test reads a Zig, C or Rust comment, or C++ code in a `.h` file. No decision records what lexis does when ztree cannot parse a file.

## Acceptance Criteria
- [x] A test reads the comments of Zig, C and Rust files.
- [x] A test reads the comment of a `.h` file with C++ code that needs more than 64 stack heads in ztree.
- [x] A test proves that a ztree error names the file and wraps the ztree error.
- [x] An ADR records that a ztree error fails the scan.
- [x] lexis reads a Rust doc comment (`///` or `//!`) without a failed assertion.
- [x] The daily note records whether the Go build cache sees a new `libztree.a`.
- [x] `zmake check` and `zmake -f Makefile check` report zero errors and zero warnings.

## Resolution
- `TestGrammarComments` reads the comments of Zig, C and Rust files. A comment marker in a string gives no finding.
- The Rust case found a defect. ztree ends the span of a Rust doc comment after the line break. So `appendProse` failed its assertion "the content is one line", and lexis exited with status 2. `parser.Comments` now removes the line break from the end of each line comment. Two adjacent `///` lines form one comment again.
- `TestManyStackHeads` reads a `.h` file with C++ code that needs 71 stack heads in ztree. ztree failed on it before ztree ADR-0013.
- `TestParseLimit` proves that a ztree error names the file and wraps `ztree.ErrLimitReached`. [ADR-0010](/decisions/adr-0010-a-ztree-error-fails-the-scan--01a1f4789ef0.md) records that the error fails the scan. The message has one `ztree:` prefix now, not two.
- The Go build cache does not see a new `libztree.a`. After a debug build of the library, `go test -c` gave the same test binary as before. ztree ISSUE-018 records it. ztree Phase 17 (ztree ADR-0016) fixes it: `zig build lib` writes the SHA-256 of `libztree.a` into the binding. The next Go build of lexis links the new library, so `go clean -cache` is not necessary.
- C++ files stay out of scope: [ISSUE-007](/issues/issue-007-c-comments--01a1f4750b43.md).

## Related
- [ADR-0008](/decisions/adr-0008-cli-exit-statuses-output-errors-and-commands-of-later-phases--01a1ee1490ca.md)
