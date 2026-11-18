---
id: "01a1ed42-b85f-748e-ba6a-49ec6bae2e7d"
title: "lexis architecture overview"
type: "Architecture"
tags:
  - "architecture"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T00:05:35-06:00"
description: "The packages of lexis, the data flow from a file to a report, and the boundaries to CUE, ztree and SARIF."
---

# lexis architecture overview

lexis is a documentation linter in Go. It checks Markdown files and source code comments against rulesets in CUE. This overview describes the plan. Each phase updates it when the code exists.

## Data flow
1. `cmd/lexis` reads the flags and the paths.
2. `internal/ruleset` loads the ruleset, applies the user overlays and the CLI overrides, and makes rules from the lexicon.
3. `internal/parser` reads each file. It gives the prose of a Markdown file, or the comments of a source file.
4. `internal/engine` splits the prose into sentences and runs each active rule. Each rule gives findings.
5. `internal/report` writes the findings as SARIF, text or JSON.

## Packages
| Package | Contents | Phase |
|---|---|---|
| `cmd/lexis` | The entry point. It calls `internal/cli`. | 2 |
| `internal/cli` | The cobra commands, the flag checks and the exit statuses ([ADR-0008](/decisions/adr-0008-cli-exit-statuses-output-errors-and-commands-of-later-phases--01a1ee1490ca.md)). | 2 |
| `internal/output` | A writer that keeps the first write error. | 2 |
| `cmd/nasa-lint` and `internal/nasalint` | The Go port of `scripts/nasa-lint.py`. | 2 |
| `internal/assert` | `assert.True` for rule R5, and `assert.Recover` for the command boundary. | 1 |
| `internal/version` | The version, commit and build date, and the SARIF versions. | 2 |
| `internal/ruleset` | The CUE schema, the loader, the lexicon projection and the overrides. | 3 |
| `rules/ste100` | The CUE source of the `ste-100` ruleset, embedded in the binary. | 3 |
| `internal/parser` | The Markdown extractor and the comment extractor. | 4 |
| `internal/engine` | The tokenizer, the sentence splitter and the rule checks. | 4 |
| `internal/report` | The SARIF, text and JSON reports. | 5 |

## Boundaries
- **CUE:** `cuelang.org/go` loads and unifies the rulesets ([ADR-0004](/decisions/adr-0004-cue-rulesets-and-one-lexicon-as-the-source-of-the-substitution-rules--01a1ed42b872.md)).
- **ztree:** the Go binding `github.com/infrashift/ztree/bindings/go` gives the comment spans of Go, Zig, Python, C and Rust files through cgo ([ADR-0002](/decisions/adr-0002-go-1-26-a-small-set-of-dependencies-and-comments-from-ztree-through-cgo--01a1ed42b868.md)).
- **SARIF:** `github.com/owenrumney/go-sarif/v3` writes SARIF 2.1.0 and 2.2 ([ADR-0006](/decisions/adr-0006-sarif-2-1-0-and-2-2-reports-with-go-sarif-v3--01a1ed42b87b.md)).
- **Docs site:** `docs/` is a separate bun project. The Go module does not include it ([ADR-0007](/decisions/adr-0007-document-lexis-with-astro-starlight-on-github-pages--01a1ed42b880.md)).
