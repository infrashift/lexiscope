---
x-okfw-id: "01a1f543-cb28-7d36-93d5-6239939d237f"
title: "ADR-0012: The tool is lexiscope, with lexis as a short alias"
type: "ADR"
tags:
  - "adr"
  - "cli"
x-okfw-created: "2026-11-19T13:23:43-06:00"
x-okfw-timestamp: "2026-11-19T13:23:43-06:00"
description: "The official name of the tool is lexiscope; lexis is a symbolic link to it, and the module, the URLs and the current documents use the new name."
---

# ADR-0012: The tool is lexiscope, with lexis as a short alias

**Status:** Accepted
**Date:** 2026-11-19
**Source:** User request.

## Context
The tool, the Go module, the GitHub repository and the docs site had the name `lexis`. The user gives the tool the official name `lexiscope`. The short name `lexis` must still start the CLI. The GitHub repository becomes `infrashift/lexiscope`. The local checkout has no git remote, so no module imports `github.com/infrashift/lexis`.

## Decision
- The official name is `lexiscope`. The help, the version line, the error prefix, the summary line and the SARIF driver name use it.
- The module path is `github.com/infrashift/lexiscope`. The entry point is `cmd/lexiscope`, so `go install` writes a `lexiscope` binary.
- `lexis` is a symbolic link to `lexiscope`. `zmake build` makes `bin/lexis`, and `zmake install` makes `$(PREFIX)/bin/lexis`. The binary has no code for the alias, and the help shows `lexiscope` for both names.
- The SARIF `informationUri` is `https://github.com/infrashift/lexiscope`. The docs site has the base path `/lexiscope/`, at `https://infrashift.github.io/lexiscope/`.
- The self-check writes `lexiscope-self.sarif` and `lexiscope-self-v22.sarif`, and it starts `bin/lexis --version` to check the alias.
- Current documents use `lexiscope`: CLAUDE.md, the docs site, the architecture overview, the specs and the guides. ADRs, closed issues, daily notes and completed phases of `PLAN.md` are records. They keep the old name.
- ztree and semscope start the tool as `lexiscope` in their `docs-lexiscope` tasks.

## Consequences
- Old scripts and the gates of other repositories that start `lexis` continue to work.
- A SARIF report has a new driver name, so a code scanning service shows the results of `lexiscope` as a new tool.
- The user renames the GitHub repository. The links of the docs site and the `informationUri` work after that rename.
- The local folder can keep the name `lexis`. No file reads the folder name.

## Alternatives considered
- **Two binaries, `cmd/lexis` and `cmd/lexiscope`.** Rejected: two builds give the same program, and the two copies can differ.
- **Keep the module path and the URLs.** Rejected: the user renames the repository, and nothing imports the module yet.
- **Change the records too.** Rejected: a record tells what happened under the name of that time.
