# CLAUDE.md — lexiscope

## Purpose (WHY)
lexiscope is a documentation linter in Go. It checks Markdown files and source code comments against rulesets in CUE. The first ruleset is `ste-100`, a subset of ASD-STE100. lexiscope writes SARIF, text and JSON reports. `PLAN.md` holds the phases.

`lexiscope` is the official name. `lexis` is a short alias: a symbolic link to `lexiscope` ([ADR-0012](.okfw/decisions/adr-0012-the-tool-is-lexiscope-with-lexis-as-a-short-alias--01a1f543cb28.md)). Use `lexiscope` in new text. ADRs, closed issues and daily notes keep the old name as records.

## Tech Stack (WHAT)
- **Language:** Go 1.26.10 or later (`go.mod`; go-sarif needs it). The host has Go 1.27.3.
- **Dependencies:** cobra, `cuelang.org/go` and go-sarif v3. The ztree Go binding through cgo ([ADR-0002](.okfw/decisions/adr-0002-go-1-26-a-small-set-of-dependencies-and-comments-from-ztree-through-cgo--01a1ed42b868.md)). Write an ADR before you add a dependency.
- **Docs site:** Astro Starlight in `docs/`, built with bun 1.3.9 and deployed to GitHub Pages ([ADR-0007](.okfw/decisions/adr-0007-document-lexis-with-astro-starlight-on-github-pages--01a1ed42b880.md)).
- **Wiki:** okfw (`.okfw/`), [ADR-0001](.okfw/decisions/adr-0001-adopt-an-okf-wiki-okfw-for-project-knowledge--01a1ed42b864.md). See the okfw block below.

## Layout
- `cmd/lexiscope/` holds the entry point. `internal/cli` holds the cobra commands and the exit statuses ([ADR-0008](.okfw/decisions/adr-0008-cli-exit-statuses-output-errors-and-commands-of-later-phases--01a1ee1490ca.md)).
- `cmd/nasa-lint/` and `internal/nasalint/` hold the Go port of `scripts/nasa-lint.py`.
- Write command output through `internal/output.Writer`. Do not discard a write result.
- `internal/report` writes text, JSON and SARIF. It builds SARIF 2.1.0 and converts it to 2.2 with go-sarif.
- `internal/` holds the packages. `internal/parser` and `internal/engine` copy the oracle for the core rules: read the [Known differences from ste-check](.okfw/specs/known-differences-from-ste-check--01a1ee352a3c.md) spec before you change them. The [architecture overview](.okfw/architecture/lexiscope-architecture-overview--01a1ed42b85f.md) lists them.
- `vocabulary.cue` holds the technical terms of this repository for the self-check. Add a term only when it names a concept of the project, a tool or a standard. Rewrite other text. Put a code name in a comment in backticks.
- `rules/ste100/` holds the CUE source of the `ste-100` ruleset. `internal/ruleset/schema.cue` is the schema. Run `bin/lexiscope rules inspect ste-100` to see the resolved ruleset.
- In CUE, write each value that a user can change as a default (`*value | type`). Write each list of terms as a struct of `true` values.
- `docs/` holds the docs site. `docs/go.mod` keeps it out of the Go module. Do not delete that file.
- `.github/workflows/docs-release.yml` builds the docs site and deploys it to GitHub Pages.
- `scripts/` holds the lint scripts from zmake: `nasa-lint.py` (Power of 10) and `ste-check.py` (ASD-STE100, the oracle). Do not change them.

## Verification & Workflow Commands (HOW)
`ZMakefile` (zmake, with `@name:` tasks) and `Makefile` (GNU Make) have the same targets. Change both files together.

Run these targets with `zmake`:
- `zmake lint` runs gofmt, `go vet`, golangci-lint 2.15.0 (`.golangci.yml`) and `cmd/nasa-lint`, the Go port of `scripts/nasa-lint.py`. When `semscope` is on the PATH, it also runs `semscope scan -r nasa-p10 -r .semscope.cue` ([ADR-0011](.okfw/decisions/adr-0011-semscope-checks-the-power-of-10-rules-in-the-lint-task--01a1f5140d5e.md)). Run `go run ./cmd/nasa-lint --verbose` to list the functions with fewer than 2 assertions.
- `zmake test` runs `go test -race -count=1 ./...`.
- `zmake build` writes `bin/lexiscope` with the version, the commit and the commit date from git. It also makes the link `bin/lexis`.
- `zmake install` runs `build`, copies `bin/lexiscope` into `$(PREFIX)/bin` and makes the link `lexis` there (`PREFIX` is `~/.local` by default).
- `zmake fmt` formats the Go source.
- `zmake wiki` runs `okfw validate --all` and `okfw validate --okf`.
- `zmake docs-style` runs `scripts/ste-check.py`.
- `zmake docs-dev` runs the docs site locally. `zmake docs-build` builds it into `docs/dist/`. Neither target changes `docs/bun.lock`.
- `zmake ztree` builds `libztree.a` in `../ztree` (set `ZTREE_DIR` for another path). `build`, `test` and `lint` run it first.
- `zmake self-check` scans the documents and the Go comments of lexiscope with lexiscope and `vocabulary.cue`, then writes `lexiscope-self.sarif` and `lexiscope-self-v22.sarif`. Each finding fails it ([ADR-0009](.okfw/decisions/adr-0009-the-self-check-fails-on-each-finding-with-a-repository-vocabulary--01a1ee772cc7.md)).
- `zmake check` runs lint, test, wiki, docs-style and self-check. It is the default goal. It does not build the docs site.

When an agent runs it, `astro dev` runs in the background. Stop it with `bun --bun astro dev stop` in `docs/`. Do not kill other dev servers on this host.

## Mandatory Execution Rules
1. **One phase at a time.** Read `PLAN.md`. Do only the first unchecked phase. Follow the [Development workflow](.okfw/development-workflow--01a1ed42b859.md).
2. **Repeat until green.** After each task, run `zmake check`. Fix each error and each warning. Do not commit while a gate fails.
3. **Test integrity.** Do not weaken or delete tests to make a gate pass.
4. **Power of 10.** Follow the rules as [ADR-0003](.okfw/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1ed42b86d.md) adapts them.
5. **Parity.** Rules STE-100-01 to 04 match `scripts/ste-check.py` exactly ([ADR-0005](.okfw/decisions/adr-0005-the-ste-100-rule-ids-with-a-core-group-that-matches-the-oracle-and-an-extended-group--01a1ed42b876.md)).
6. **Minimal abstractions.** Add no unrequested interfaces or factories.
7. **Wiki and docs stay in sync.** Each `PLAN.md` phase has a mirror issue in `.okfw/issues/`. Record each decision as an ADR when you make it. Update the docs pages for each delivered feature.
8. **Documentation follows ASD-STE100.** Use active voice, simple present tense and short sentences. Read the [Wiki conventions](.okfw/wiki-conventions--01a1ed42b854.md). Run `zmake docs-style`.
9. **Git.** Make one commit per phase. Do not push.

<!-- BEGIN okfw -->
## Project knowledge base (okfw)

lexiscope keeps a local-first OKF wiki in `.okfw/`: markdown concepts with
YAML frontmatter, linked to each other. The files are the whole database —
search and the link graph are computed from them on demand, so there is no
index to build and nothing to sync. Each directory's `index.md` is generated
and gives a cheap overview before you open anything (OKF section 8).

- **Search before you derive.** `okfw search "<query>"` / `okfw_search`. The
  wiki holds what the code does not say — why things are built the way they
  are, what was considered and rejected, dependency and security findings,
  operational context, open work. Check it before explaining a design,
  assessing risk, or running an analysis someone may already have recorded and
  adjudicated.
- **Browse top-down.** Each directory's `index.md` lists its concepts with
  their descriptions (`okfw_read {"ref":"<dir>/"}` over MCP). Open only what fits.
- **Search before you write**, too, so a new concept links to what already
  exists instead of duplicating it.
- **Link actively** — `[Title](/dir/file.md)` from the wiki root, or
  `./file.md`. Links are what makes the wiki more than a folder.
- **One idea per concept**, with a one-sentence `description`.
- **Use `okfw move` / `okfw delete`, never `mv` / `rm`.** `move` repairs every
  link to the concept and its own relative links, and both keep the `index.md`
  listings current; the shell commands break the links and leave the listings
  stale.
- **`status` is the OKF lifecycle field** (`draft`/`stable`/`deprecated`), not
  task state. Mark superseded knowledge rather than deleting it.
- **Custom metadata goes under `x:`** — the schema is closed and rejects any
  other unrecognized key.
- Don't edit `index.md` or `log.md`; okfw generates them.

Results are capped and say so: `truncated`, or `… N more`, means you did not
see everything — raise the limit or page with `--offset`/`"offset"`.

`okfw mcp` serves all of this as MCP stdio tools named `okfw_*`.
<!-- END okfw -->
