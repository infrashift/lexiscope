---
id: "01a1f475-0b43-784c-bd6f-77d518600e6c"
title: "ISSUE-007: C++ comments"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "closed"
  - "phase-7"
  - "parser"
created: "2026-11-19T09:37:53-06:00"
timestamp: "2026-11-19T09:37:53-06:00"
description: "lexis does not read the comments of C++ files, although ztree has a C++ grammar."
---

# ISSUE-007: C++ comments

**Status:** closed · **Phase:** 7

## Description
ztree reads C++ files (`ztree.LangCPP`), but lexis does not. `parser.LangOf` maps no C++ file extension, so a scan skips each C++ file. A `.h` file stays C.

## Proposed change
- Map `.cc`, `.cpp`, `.cxx`, `.hh`, `.hpp` and `.hxx` to `ztree.LangCPP` in `internal/parser/comments.go`. semscope uses the same set (`internal/parser/language.go` in semscope).
- C++ comments have the same forms as C comments, so `commentLines` needs no change. Add a test with `//`, `/* */` and a raw string.
- Add C++ to the docs pages that list the languages: `docs/src/content/docs/guides/cli.md` and `docs/src/content/docs/reference/cli.md`.
- Add C++ to the architecture overview and to the ztree line of [ADR-0002](/decisions/adr-0002-go-1-26-a-small-set-of-dependencies-and-comments-from-ztree-through-cgo--01a1ed42b868.md).
- Add the term `C++` to `vocabulary.cue` if the self-check flags it.

## Acceptance Criteria
- [x] lexis reads the comments of the six C++ file extensions.
- [x] lexis scans real C and C++ trees without an internal error.
- [x] The docs pages and the wiki list C++.
- [x] `zmake check` reports zero errors and zero warnings.

## Resolution
`parser.LangOf` maps the six extensions to `ztree.LangCPP`. `TestGrammarComments` reads a C++ file with `//`, `/* */` and a raw string, and one comment for each extension.

A scan of V8 found a defect that also affects C files. In C and C++, a backslash at the end of a `//` comment continues the comment on the next line. V8 draws diagrams that way. ztree then gives one comment span with more lines. `appendProse` failed its assertion "the content is one line", and lexis exited with status 2. `commentLines` now reads each line of a line comment separately and removes the backslash. `TestContinuedLineComment` holds the case for `.c`, `.h` and `.cpp` files. After the fix, lexis scans V8, OpenSSL, ICU, the libstdc++ headers and the Linux headers without an internal error.

The CLI help, the CLI guide, the CLI reference, the landing page, ADR-0002 and the architecture overview list C++.

## Related
- [ISSUE-006](/issues/issue-006-ztree-comment-coverage--01a1f474ef49.md)
