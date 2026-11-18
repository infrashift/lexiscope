---
id: "01a1ed42-b886-7297-9310-4d9b13524d64"
title: "ISSUE-001: Bootstrap the repository, the gates, the wiki and the docs site"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "phase-1"
  - "bootstrap"
  - "closed"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T00:06:50-06:00"
description: "Create the Go module, the package layout, the build files, the OKF wiki, the docs site and the Pages workflow."
---

# ISSUE-001: Bootstrap the repository, the gates, the wiki and the docs site

**Status:** closed · **Phase:** 1

## Description
Set up the lexis repository so that the gates run before the first rule exists.

## Acceptance Criteria
- [x] `PLAN.md` follows ASD-STE100, and the user approved it.
- [x] The git repository has the branch `main` and a `.gitignore`.
- [x] `scripts/ste-check.py` and `scripts/nasa-lint.py` are copies of the zmake scripts without changes.
- [x] The module `github.com/infrashift/lexis` sets `go 1.26` and has the package layout of `PLAN.md`.
- [x] `ZMakefile` and `Makefile` have the same targets, and `zmake -n` and `make -n` print the same commands.
- [x] `zmake docs-build` builds the Starlight site in `docs/`, and `zmake docs-dev` serves it.
- [x] `.github/workflows/docs-release.yml` builds the site and deploys it to GitHub Pages.
- [x] The wiki has the guides, the architecture overview, ADR-0001 to ADR-0007 and ISSUE-001 to ISSUE-005.
- [x] `zmake check` reports zero errors and zero warnings.

## Resolution
- `PLAN.md` follows ASD-STE100. The user approved the rewrite. It adds the settled decisions, Go 1.26, the docs site and the R3 deviation.
- `zmake check` and `make check` pass: gofmt, `go vet`, nasa-lint, `go test -race`, okfw and ste-check, with zero errors and zero warnings.
- Mutation checks: a badly formatted Go file makes `zmake lint` fail. A banned word in a docs page makes `zmake docs-style` fail.
- `scripts/nasa-lint.py` reads Zig only, so it reports 0 functions. Phase 2 ports it to Go.
- `zmake docs-build` builds 7 pages under `/lexis/`. `zmake docs-dev` serves the home page and the rules page. `bun.lock` keeps the zmake versions, with the package name `lexis-docs`.
- `go list ./...` walks `docs/node_modules`. A probe file proved it. `docs/go.mod` stops the walk ([ADR-0007](/decisions/adr-0007-document-lexis-with-astro-starlight-on-github-pages--01a1ed42b880.md)).
- The deploy needs three steps from the user: create `infrashift/lexis` on GitHub, set the Pages source to "GitHub Actions", and push.

## Related
- [ADR-0001](/decisions/adr-0001-adopt-an-okf-wiki-okfw-for-project-knowledge--01a1ed42b864.md), [ADR-0002](/decisions/adr-0002-go-1-26-a-small-set-of-dependencies-and-comments-from-ztree-through-cgo--01a1ed42b868.md) and [ADR-0007](/decisions/adr-0007-document-lexis-with-astro-starlight-on-github-pages--01a1ed42b880.md).
