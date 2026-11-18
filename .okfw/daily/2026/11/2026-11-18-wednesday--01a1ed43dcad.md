---
id: "01a1ed43-dcad-7ef2-847d-2520b7758dda"
title: "2026-11-18 Wednesday"
type: "daily-note"
kind: "daily-note"
tags:
  - "daily"
created: "2026-11-18T00:06:50-06:00"
timestamp: "2026-11-18T00:06:50-06:00"
---

# 2026-11-18 Wednesday

## Morning

-

## Tasks

- [x] Phase 1: bootstrap, gates, wiki and docs site
- [x] Phase 2: Go port of nasa-lint, golangci-lint, version and CLI
- [x] Phase 3: CUE ruleset engine and the ste-100 ruleset

## Notes

- **Changed:** Phase 1. The repo, the Go module, the build files, the lint scripts, the wiki (ADR-0001 to ADR-0007, ISSUE-001 to ISSUE-005), the Starlight site and the Pages workflow exist.
- **Learned:** `go list ./...` walks `docs/node_modules`. `docs/go.mod` stops it.
- **Learned:** Starlight 0.41 needs `items: [{ autogenerate }]` inside a group. A group with no pages shows an empty heading.
- **Learned:** When an agent runs it, `astro dev` runs in the background and takes the next free port. The semscope dev server holds port 4321 on this host.
- **Changed:** Phase 2. `cmd/nasa-lint` checks R1, R2, R4, R5, R7 and R9 on Go code. golangci-lint 2.15.0 runs in `zmake lint`. `lexis version`, `scan` and `rules` exist. ADR-0008 records the exit statuses.
- **Learned:** errcheck and R7 conflict on writes. A writer that keeps the first error satisfies both.
- **Learned:** funlen counts the lines between the braces. Its limit of 58 equals R4.
- **Learned:** `recover` works in a named function that the code defers directly, so the command boundary needs no function literal.
- **Changed:** Phase 3. `internal/ruleset` loads CUE rulesets. `ste100.cue` defines 11 rules and a lexicon of 20 concepts. `lexis rules list` and `lexis rules inspect` work.
- **Learned:** two CUE defaults for one field do not unify, so the schema has no default for `severity`. A comprehension in a built-in file sees the lexicon of each user file. `try` is a CUE keyword.
- **Learned:** a test that reads the constants of `scripts/ste-check.py` keeps rules 01 to 04 equal to the oracle before the engine exists.
- **Next:** Phase 4. Add the Markdown extractor, the comment extractor with ztree, the rule checks and the parity test.

## End of Day



## Links Created

-
