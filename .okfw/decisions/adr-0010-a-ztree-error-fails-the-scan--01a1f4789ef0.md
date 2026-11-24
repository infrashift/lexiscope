---
x-okfw-id: "01a1f478-9ef0-782e-a81e-f19838cee835"
title: "ADR-0010: A ztree error fails the scan"
type: "ADR"
tags:
  - "adr"
  - "parser"
  - "cli"
x-okfw-created: "2026-11-19T09:41:48-06:00"
x-okfw-timestamp: "2026-11-19T09:41:48-06:00"
description: "When ztree cannot parse a source file, lexis stops the scan with status 2 and names the file and the ztree error."
---

# ADR-0010: A ztree error fails the scan

**Status:** Accepted
**Date:** 2026-11-19
**Source:** User decision, [ISSUE-006](/issues/issue-006-ztree-comment-coverage--01a1f474ef49.md). ztree ADR-0013 and ADR-0014.

## Context
lexis reads the comments of a source file with `ztree.ExtractComments` ([ADR-0002](/decisions/adr-0002-go-1-26-a-small-set-of-dependencies-and-comments-from-ztree-through-cgo--01a1ed42b868.md)). ztree returns an error when a parse reaches one of its hard limits, for example `ztree.ErrLimitReached`. A Go table of 1.5 million items reaches a ztree limit. A syntax error is not a ztree error: ztree still gives the comments of a file with syntax errors.

The ztree limit of 64 stack heads failed on real C++ headers, such as `stl_algobase.h` of libstdc++. ztree ADR-0013 raised the limit to 256. ztree ADR-0014 removed a counter that overflowed on tables of more than 65,535 items. On 73,701 files, the largest head count is 90, so none of these files reaches the new limit of 256.

## Decision
- A ztree error fails the scan. `lexis scan` stops at the file and exits with status 2, the failure status of [ADR-0008](/decisions/adr-0008-cli-exit-statuses-output-errors-and-commands-of-later-phases--01a1ee1490ca.md).
- The message names the file and the ztree error, for example `lexis: big.go: ztree: parser limit reached`. `parser.Comments` returns the ztree error without a second `ztree:` prefix.
- `TestParseLimit` holds the behavior. It checks the file name and `errors.Is(err, ztree.ErrLimitReached)`.

## Consequences
- A CI job cannot pass when lexis did not read a file.
- One file that ztree cannot parse stops the scan of each later file. Use `--exclude` to skip the file.

## Alternatives considered
- **Skip the file and write a warning to stderr.** Rejected: a scan with status 0 must mean that lexis read each file.
- **Report a parse finding for the file.** Rejected: it adds a rule that is not a documentation rule, and it changes the report formats.
