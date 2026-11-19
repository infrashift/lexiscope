---
id: "01a1f475-0b43-784c-bd6f-77d518600e6c"
title: "ISSUE-007: C++ comments"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "open"
  - "parser"
created: "2026-11-19T09:37:53-06:00"
timestamp: "2026-11-19T09:37:53-06:00"
description: "lexis does not read the comments of C++ files, although ztree has a C++ grammar."
---

# ISSUE-007: C++ comments

**Status:** open · **Phase:** none

## Description
ztree reads C++ files (`ztree.LangCPP`), but lexis does not. `parser.LangOf` maps no C++ file extension, so a scan skips each C++ file. A `.h` file stays C.

## Proposed change
- Map `.cc`, `.cpp`, `.cxx`, `.hh`, `.hpp` and `.hxx` to `ztree.LangCPP` in `internal/parser/comments.go`. semscope uses the same set (`internal/parser/language.go` in semscope).
- C++ comments have the same forms as C comments, so `commentLines` needs no change. Add a test with `//`, `/* */` and a raw string.
- Add C++ to the docs pages that list the languages: `docs/src/content/docs/guides/cli.md` and `docs/src/content/docs/reference/cli.md`.
- Add C++ to the architecture overview and to the ztree line of [ADR-0002](/decisions/adr-0002-go-1-26-a-small-set-of-dependencies-and-comments-from-ztree-through-cgo--01a1ed42b868.md).
- Add the term `C++` to `vocabulary.cue` if the self-check flags it.

## Acceptance Criteria
- [ ] lexis reads the comments of the six C++ file extensions.
- [ ] The docs pages and the wiki list C++.
- [ ] `zmake check` reports zero errors and zero warnings.

## Related
- [ISSUE-006](/issues/issue-006-ztree-comment-coverage--01a1f474ef49.md)
