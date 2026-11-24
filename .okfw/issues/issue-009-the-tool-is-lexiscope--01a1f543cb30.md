---
x-okfw-id: "01a1f543-cb30-7631-b61a-04f58f1c6cf2"
title: "ISSUE-009: The tool is lexiscope"
type: "Issue"
x-okfw-kind: "issue"
tags:
  - "issue"
  - "closed"
  - "phase-9"
  - "cli"
x-okfw-created: "2026-11-19T13:23:43-06:00"
x-okfw-timestamp: "2026-11-19T13:23:43-06:00"
description: "The tool gets the official name lexiscope, and lexis stays as a short alias."
---

# ISSUE-009: The tool is lexiscope

**Status:** closed · **Phase:** 9

## Description
The user gives the tool the official name `lexiscope`. The short name `lexis` must still start the CLI. The module path, the SARIF URLs, the docs site and the current documents use the old name.

## Acceptance Criteria
- [x] The help, the version line, the error prefix, the summary line and the SARIF driver name say `lexiscope`.
- [x] The module path is `github.com/infrashift/lexiscope`, and the entry point is `cmd/lexiscope`.
- [x] `zmake build` and `zmake install` make the link `lexis` to `lexiscope`. Both names start the CLI.
- [x] The docs site, CLAUDE.md and the current concepts of the wiki use `lexiscope` and name the alias. Records keep the old name.
- [x] Record the decision in [ADR-0012](/decisions/adr-0012-the-tool-is-lexiscope-with-lexis-as-a-short-alias--01a1f543cb28.md).
- [x] `zmake check`, `zmake -f Makefile check` and `zmake docs-build` pass.

## Resolution
The module path is `github.com/infrashift/lexiscope`, and `cmd/lexiscope` is the entry point. The help, the version line, the error prefix, the summary line and the SARIF driver name say `lexiscope`. `zmake build` makes `bin/lexis`, and `zmake install` makes `$(PREFIX)/bin/lexis`. Both are links to `lexiscope`, and the self-check starts `bin/lexis --version` to check the alias.

The docs site has the base path `/lexiscope/`. The SARIF `informationUri` and the rules help URI use the new repository name. They work after the user renames the GitHub repository.
