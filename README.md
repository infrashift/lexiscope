# lexiscope

lexiscope is a documentation linter in Go. It checks Markdown files and source code comments against rulesets in CUE. The first ruleset is `ste-100`, a subset of ASD-STE100 (Simplified Technical English). lexiscope writes text, JSON and SARIF reports.

`lexis` is a short alias of `lexiscope`.

Documentation: <https://infrashift.github.io/lexiscope/>

## Features

- **Rules in CUE.** Each ruleset is a CUE file. A user file can add a rule, turn off a rule or change a limit.
- **Prose and comments.** lexiscope reads Markdown files. It also reads the comments of Go, Zig, Python, C, C++, Rust and C# files with [ztree](https://github.com/infrashift/ztree).
- **SARIF output.** lexiscope writes SARIF 2.1.0 and 2.2 reports for code scanning tools. It also writes text and JSON.
- **Strict code rules.** The Go source follows the NASA Power of 10 rules. The documentation follows ASD-STE100.

## The ste-100 rules

| ID | Severity | Rule |
|---|---|---|
| STE-100-01 | error | A procedural step has 20 words or fewer. |
| STE-100-02 | error | A descriptive sentence has 25 words or fewer. |
| STE-100-03 | warning | Use the active voice. |
| STE-100-04 | error | Do not use the banned words and phrases. |
| STE-100-05 | warning | Use the approved term for each concept. |
| STE-100-06 | warning | Use a verb, not a noun form of the verb. |
| STE-100-07 | warning | Avoid unclear `-ing` forms. |
| STE-100-08 | warning | Do not use "your". |
| STE-100-09 | warning | Do not use marketing words. |
| STE-100-10 | warning | Use the active voice with irregular verbs. |
| STE-100-11 | warning | Write one instruction in each sentence. |

Rules STE-100-01 to 04 match `scripts/ste-check.py` exactly.

## Build

lexiscope has no binary release. Build it from source.

| Tool | Version |
|---|---|
| Go | 1.26.10 or later |
| Zig | 0.17.1 |
| C compiler | any |
| zmake or GNU Make | any |

lexiscope reads source comments with ztree through cgo. Put a ztree checkout next to the lexiscope checkout, or set `ZTREE_DIR`.

1. Clone `lexiscope` and `ztree` into the same directory.
2. Run `zmake build` to write `bin/lexiscope` and the link `bin/lexis`.
3. Run `bin/lexiscope --version`.
4. Run `zmake install` to copy the binary into `~/.local/bin`.

Set `PREFIX` for a different directory, for example `zmake install PREFIX=/usr/local`. GNU Make reads the `Makefile`, which has the same targets.

## Usage

Scan a directory and write one line for each finding:

```sh
lexiscope scan docs/
```

Write a SARIF report for a code scanning tool:

```sh
lexiscope scan --format sarif -o lexiscope.sarif docs/
```

Scan the comments of source files:

```sh
lexiscope scan cmd/ internal/
```

Turn off a rule, or change a limit:

```sh
lexiscope scan --omit-rule STE-100-07 --set STE-100-01.max_words=18 .
```

Show the active rules:

```sh
lexiscope rules list
lexiscope rules inspect ste-100
```

The exit status is 1 when lexiscope finds an error. The status is 0 when it finds only warnings or nothing. `--fail-on warning` also gives status 1 for a warning.

## Development

`zmake check` runs each gate: lint, test, wiki, docs-style and self-check. The self-check scans the documentation and the Go comments of lexiscope with lexiscope. Each finding fails it.

| Target | Action |
|---|---|
| `zmake lint` | gofmt, `go vet`, golangci-lint 2.15.0 and the Power of 10 checks |
| `zmake test` | `go test -race -count=1 ./...` |
| `zmake wiki` | Validate the wiki in `.okfw/` with okfw |
| `zmake docs-style` | Check the Markdown files with `scripts/ste-check.py` |
| `zmake self-check` | Scan this repository with lexiscope |
| `zmake docs-dev` | Run the docs site locally |
| `zmake docs-build` | Build the docs site into `docs/dist/` |

`PLAN.md` holds the phases of the project. The [work loop](https://infrashift.github.io/lexiscope/project/contributing/) page on the docs site tells more.

## Repository layout

| Path | Contents |
|---|---|
| `cmd/lexiscope/` | The entry point |
| `internal/` | The CLI, the parser, the engine, the rulesets and the reports |
| `rules/ste100/` | The CUE source of the `ste-100` ruleset |
| `vocabulary.cue` | The technical terms of this repository for the self-check |
| `docs/` | The Astro Starlight docs site |
| `.okfw/` | The project wiki: architecture, decisions, specs and issues |
| `scripts/` | `ste-check.py` and `nasa-lint.py`, the reference lint scripts |

## Docs site

The workflow `.github/workflows/docs-release.yml` builds the site in `docs/` with Astro Starlight. It deploys the site to GitHub Pages for each push to `main` that changes `docs/`. A pull request only builds the site.
