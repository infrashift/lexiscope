# PLAN.md — lexis

lexis is a documentation linter in Go. It checks Markdown files and source code comments against rulesets in CUE. The first ruleset is `ste-100`, a subset of ASD-STE100 (Simplified Technical English). lexis writes text, JSON and SARIF reports.

Work on one phase at a time. Inside a phase, work on one task at a time. Each phase has a mirror issue in the OKF wiki (`.okfw/issues/`). At the end of a phase, tick this checklist and the acceptance criteria of the issue.

Loop: do one task, run `zmake check`, and repeat until it reports zero errors and zero warnings. Then go to the next task. At the end of a phase, update the wiki and the docs site, then make one git commit. Do not push.

- Architecture: [lexis architecture overview](.okfw/architecture/lexis-architecture-overview--01a1ed42b85f.md)
- Decisions: [ADR-0001](.okfw/decisions/adr-0001-adopt-an-okf-wiki-okfw-for-project-knowledge--01a1ed42b864.md) · [ADR-0002](.okfw/decisions/adr-0002-go-1-26-a-small-set-of-dependencies-and-comments-from-ztree-through-cgo--01a1ed42b868.md) · [ADR-0003](.okfw/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1ed42b86d.md) · [ADR-0004](.okfw/decisions/adr-0004-cue-rulesets-and-one-lexicon-as-the-source-of-the-substitution-rules--01a1ed42b872.md) · [ADR-0005](.okfw/decisions/adr-0005-the-ste-100-rule-ids-with-a-core-group-that-matches-the-oracle-and-an-extended-group--01a1ed42b876.md) · [ADR-0006](.okfw/decisions/adr-0006-sarif-2-1-0-and-2-2-reports-with-go-sarif-v3--01a1ed42b87b.md) · [ADR-0007](.okfw/decisions/adr-0007-document-lexis-with-astro-starlight-on-github-pages--01a1ed42b880.md) · [ADR-0008](.okfw/decisions/adr-0008-cli-exit-statuses-output-errors-and-commands-of-later-phases--01a1ee1490ca.md) · [ADR-0009](.okfw/decisions/adr-0009-the-self-check-fails-on-each-finding-with-a-repository-vocabulary--01a1ee772cc7.md) · [ADR-0010](.okfw/decisions/adr-0010-a-ztree-error-fails-the-scan--01a1f4789ef0.md)
- Guides: [Development workflow](.okfw/development-workflow--01a1ed42b859.md) · [Wiki conventions](.okfw/wiki-conventions--01a1ed42b854.md)

## Rules

### Quality gate

`zmake check` runs the gate. `make check` runs the same targets from the `Makefile`. The gate grows with each phase:

| Check | Tool | From phase |
|---|---|---|
| Format and vet | `gofmt -l`, `go vet` | 1 |
| Power of 10 | `scripts/nasa-lint.py`, then its Go port | 1, port in 2 |
| Tests | `go test -race -count=1 ./...` | 1 |
| Static analysis | `golangci-lint run` (v2, pinned version) | 2 |
| Wiki | `okfw validate --all --okf` | 1 |
| Writing style | `scripts/ste-check.py` | 1 |
| Self-check | `lexis scan --ruleset ste-100 --ruleset vocabulary.cue --fail-on note` | 4, zero results from 5 |

The gate does not build the docs site, so the Go checks do not need bun. `zmake docs-build` builds the site, and the Pages workflow builds it for each pull request.

### NASA Power of 10 for Go

ADR-0003 adapts the rules to Go. The Go port of `nasa-lint.py` checks R1, R2, R4, R5, R7 and R9. Test files (`_test.go`) are exempt.

- **R1 (simple control flow):** No recursion, direct or indirect. No `goto`. Walk the Markdown tree and the comment list with explicit, bounded stacks.
- **R2 (fixed loop bounds):** Each loop has a static upper bound, for example `MaxFiles`, `MaxTokensPerFile`, `MaxSentencesPerFile` or `MaxASTNodes`. A loop that reaches its bound returns an error. A bare `for {}` is an error.
- **R3 (allocation):** This rule is a deviation, as in zmake and ztree. Go allocates on the heap and has a garbage collector. lexis sets hard caps on each buffer and allocates each buffer once per file.
- **R4 (short functions):** A function has 60 lines or fewer.
- **R5 (assertions):** The average is two or more `assert.` calls per function. `internal/assert` panics on a programmer error. The CLI recovers the panic at its boundary and exits with status 2.
- **R6 (smallest scope):** Declare each variable in the smallest scope.
- **R7 (check all returns):** Do not discard a value with `_ =`. Do not discard an error with `, _ :=` on a call.
- **R8 (code generation):** Keep build tags and generated code to a minimum.
- **R9 (indirection):** Do not declare function types or function variables. Each func literal that cobra needs has an allow comment.
- **R10 (zero warnings):** `go vet`, golangci-lint and nasa-lint report zero warnings.

### Writing style

All Markdown files, wiki pages, docs pages, CLI help text and code comments follow ASD-STE100. `scripts/ste-check.py` is the gate and the test oracle. The oracle has four checks:

1. A numbered step has 20 words or fewer.
2. Each other sentence has 25 words or fewer.
3. A be-verb before a word that ends in `-ed` is a passive warning.
4. The oracle bans 13 words and phrases, for example `should`, `will` and `just`.

lexis adds more checks. They apply to lexis output, and the authors of this repository follow them too:

- Do not use marketing words, for example `seamless`, `robust`, `powerful` or `effortless`.
- Do not use the second-person possessive pronoun. Write "the ruleset file".
- Use the active voice. Name the actor: "The parser reads the document."
- Use one term for one concept. Avoid nominalizations (`-tion`, `-ment`, `-ance`) and unclear `-ing` forms.
- Write one instruction in each sentence.

## Phase 1 — Bootstrap, gates, wiki and docs site

Issue: [ISSUE-001](.okfw/issues/issue-001-bootstrap-the-repository-the-gates-the-wiki-and-the-docs-site--01a1ed42b886.md)

- [x] Task 0: Rewrite this plan in STE, then get approval from the user
- [x] Task 1.0: Create the git repo on `main` and write `.gitignore`
- [x] Task 1.0: Copy `scripts/ste-check.py` and `scripts/nasa-lint.py` from zmake without changes
- [x] Task 1.1: Create the Go module `github.com/infrashift/lexis` with Go 1.26 (go-sarif v3 needs Go 1.26)
- [x] Task 1.1: Create the package layout below, with a `doc.go` in each package
- [x] Task 1.2: Write `ZMakefile` and `Makefile` with the same targets
- [x] Task 1.3: Create the Astro Starlight site in `docs/` with bun, and the `docs-dev` and `docs-build` targets
- [x] Task 1.4: Add `.github/workflows/docs-release.yml` to publish the site to GitHub Pages
- [x] Task 1.5: Create the OKF wiki (`okfw init`, `okfw install --platform claude`) and `.mcp.json`
- [x] Task 1.5: Write `CLAUDE.md`, the wiki guides, the architecture overview, the ADRs and one issue per phase

Package layout:

- `cmd/lexis/`: the entry point and the cobra commands.
- `internal/assert/`: assertions for R5.
- `internal/version/`: the version, commit and build date, and the SARIF versions.
- `internal/ruleset/`: the CUE loader, the lexicon projection, rule overrides and the embedded `ste-100` ruleset.
- `internal/parser/`: the Markdown extractor and the comment extractor.
- `internal/engine/`: the tokenizer, the sentence splitter and the rule checks.
- `internal/report/`: the SARIF, text and JSON reports.
- `rules/ste100/`: the CUE source of the `ste-100` ruleset and its lexicon.
- `scripts/`: the linter scripts from zmake.
- `docs/`: the documentation site.

First ADRs:

- [ADR-0001](.okfw/decisions/adr-0001-adopt-an-okf-wiki-okfw-for-project-knowledge--01a1ed42b864.md): an OKF wiki for project knowledge.
- [ADR-0002](.okfw/decisions/adr-0002-go-1-26-a-small-set-of-dependencies-and-comments-from-ztree-through-cgo--01a1ed42b868.md): Go 1.26, and comments from ztree through cgo.
- [ADR-0003](.okfw/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1ed42b86d.md): the Power of 10 rules for Go.
- [ADR-0004](.okfw/decisions/adr-0004-cue-rulesets-and-one-lexicon-as-the-source-of-the-substitution-rules--01a1ed42b872.md): CUE rulesets and one lexicon as the source of the substitution rules.
- [ADR-0005](.okfw/decisions/adr-0005-the-ste-100-rule-ids-with-a-core-group-that-matches-the-oracle-and-an-extended-group--01a1ed42b876.md): the `ste-100` rule IDs, with a core that matches the oracle and extended rules.
- [ADR-0006](.okfw/decisions/adr-0006-sarif-2-1-0-and-2-2-reports-with-go-sarif-v3--01a1ed42b87b.md): SARIF 2.1.0 and 2.2 reports.
- [ADR-0007](.okfw/decisions/adr-0007-document-lexis-with-astro-starlight-on-github-pages--01a1ed42b880.md): an Astro Starlight docs site on GitHub Pages.
- [ADR-0008](.okfw/decisions/adr-0008-cli-exit-statuses-output-errors-and-commands-of-later-phases--01a1ee1490ca.md): CLI exit statuses, output errors and commands of later phases (Phase 2).

## Phase 2 — CLI, version and ruleset overrides

Issue: [ISSUE-002](.okfw/issues/issue-002-cli-version-and-ruleset-overrides-and-the-go-port-of-nasa-lint--01a1ed42b88b.md)

- [x] Port `scripts/nasa-lint.py` to Go (`cmd/nasa-lint`) in a separate commit, and use the port in `zmake lint`
- [x] Install golangci-lint v2 at a pinned version, with errcheck, funlen and govet, and add it to `zmake lint`
- [x] Add `internal/version` with the `-ldflags` variables `Version`, `Commit` and `BuildDate`, and the SARIF versions `2.1.0` and `2.2`
- [x] Add the cobra CLI in `cmd/lexis`
- [x] Add the CLI guide and the CLI reference to the docs site

CLI commands:

- `lexis --version`, `lexis -v` and `lexis version` print `lexis version <semver> (commit: <sha>, built: <date>)`.
- `lexis scan [paths...]` takes these flags:
  - `--ruleset`, `-r <name-or-path>`: a built-in ruleset (the default is `ste-100`), or one or more `.cue` files or directories.
  - `--enable-rule <id[,id...]>`: turn on rules.
  - `--omit-rule`, `--disable-rule <id[,id...]>`: turn off rules.
  - `--set <rule_id.param=value>`: change a rule parameter, for example `--set STE-100-01.max_words=18`.
  - `--allow-term <word[,word...]>`: add approved technical terms.
  - `--format <sarif|text|json>`: the report format.
  - `--sarif-version <2.1.0|2.2>`: the SARIF version. The default is `2.1.0`.
  - `--output`, `-o <file>`: the report file. The default is stdout.
- `lexis rules list` and `lexis rules inspect <ruleset>` show the active rules, the omitted rules, the lexicon and the word limits.

In Phase 2, `lexis scan` and the `lexis rules` commands check their flags. Then they exit with status 2 and name the phase that delivers the work: Phase 3 for `rules`, Phase 4 for `scan` ([ADR-0008](.okfw/decisions/adr-0008-cli-exit-statuses-output-errors-and-commands-of-later-phases--01a1ee1490ca.md)).

## Phase 3 — CUE ruleset engine and the ste-100 ruleset

Issue: [ISSUE-003](.okfw/issues/issue-003-cue-ruleset-engine-and-the-ste-100-ruleset--01a1ed42b891.md)

- [x] Write the CUE schema in `internal/ruleset/schema.cue` and load it with `cuelang.org/go`
- [x] Write the `ste-100` ruleset in `rules/ste100/ste100.cue`
- [x] Test user overlays: add a rule, omit a rule, change `max_words`, and add approved terms
- [x] Add the rules reference to the docs site

The schema:

- A lexicon maps each concept to one approved term, its unapproved synonyms and its nominalizations. lexis makes the substitution and nominalization rules from the lexicon. CUE unification fails when an approved term is also unapproved or banned.
- The `rules` map has a stable ID as its key (`rules: [ID=string]: #Rule & {id: ID}`).
- Each `#Rule` has these fields:
  - `id`, for example `"STE-100-01"`.
  - `enabled`: `bool | *true`. A user overlay or `--omit-rule` turns a rule off.
  - `severity`: `*"error" | "warning" | "note"`.
  - `scope`: `"procedural_sentence"`, `"descriptive_sentence"`, `"any_sentence"`, `"heading"`, `"comment"` or `"token"`.
  - `title` and `message`.
  - `params`: a closed struct of the known parameters, so a misspelled name fails ([ADR-0004](.okfw/decisions/adr-0004-cue-rulesets-and-one-lexicon-as-the-source-of-the-substitution-rules--01a1ed42b872.md)).
  - `check`: one of `"max_words"`, `"banned_terms"`, `"term_substitution"`, `"passive_voice"`, `"nominalizations"`, `"ambiguous_gerunds"`, `"possessive_pronouns"`, `"one_instruction_per_sentence"` or `"regex_ban"`.

The `ste-100` rules. Rules 01 to 04 are the core. They match `scripts/ste-check.py` exactly, and the parity test checks both directions. Rules 05 to 11 form the extended group. Each extended rule is a warning by default, has an exception list and has its own fixtures.

| ID | Check | Severity | Group |
|---|---|---|---|
| STE-100-01 | Numbered step: 20 words or fewer (`max_words`) | error | core |
| STE-100-02 | Other sentence: 25 words or fewer (`max_words`) | error | core |
| STE-100-03 | Be-verb and a word that ends in `-ed` (`passive_voice`) | warning | core |
| STE-100-04 | The 13 banned words and phrases of the oracle (`banned_terms`) | error | core |
| STE-100-05 | Synonyms from the lexicon (`term_substitution`) | warning | extended |
| STE-100-06 | Nominalizations (`nominalizations`) | warning | extended |
| STE-100-07 | Unclear `-ing` forms (`ambiguous_gerunds`) | warning | extended |
| STE-100-08 | `your` and `yours` (`possessive_pronouns`) | warning | extended |
| STE-100-09 | Marketing words (`banned_terms`) | warning | extended |
| STE-100-10 | Irregular passive participles (`passive_voice`) | warning | extended |
| STE-100-11 | One instruction per sentence (`one_instruction_per_sentence`) | warning | extended |

Procedural sentences: the oracle treats only numbered list items as steps. lexis does the same by default. The `procedural_markers` parameter can add bullet items. lexis finds imperative sentences with a list of verbs. It has no part-of-speech tagger.

## Phase 4 — Prose and comment engine, and parity with ste-check

Issue: [ISSUE-004](.okfw/issues/issue-004-prose-and-comment-engine-and-parity-with-ste-check--01a1ed42b897.md)

- [x] Add the Markdown extractor in `internal/parser`
- [x] Add the comment extractor in `internal/parser` with the ztree Go binding
- [x] Add the parity test in `internal/engine/ste_parity_test.go`
- [x] Add `lexis scan --ruleset ste-100 .` to `zmake check` (the `self-check` target; in Phase 4 an error fails it, and from Phase 5 each finding fails it)
- [x] Add the comment guide to the docs site

The Markdown extractor:

- It removes front matter, fenced code (` ``` ` and `~~~`), HTML comments, inline code, URLs and link targets.
- It keeps the 1-based line and the UTF-16 column of each token.
- It tells numbered steps from other text with the line rules of the oracle. It does not use recursion.
- The spec [Known differences from ste-check](.okfw/specs/known-differences-from-ste-check--01a1ee352a3c.md) lists each deliberate difference from the oracle.

The comment extractor:

- It reads the comments of `.go`, `.zig`, `.py`, `.c` and `.rs` files.
- It imports `github.com/infrashift/ztree/bindings/go`. A `replace` directive in `go.mod` points to the local ztree checkout.
- Build ztree first with `zig build lib --release`.
- With `CC="zig cc"`, cgo on this host also needs `-Wno-gcc-install-dir-libstdcxx`.

The parity test runs `scripts/ste-check.py` and `lexis scan --ruleset ste-100` on these inputs:

1. Fixtures in `testdata/ste100/violations/`, one or more for each rule.
2. Clean fixtures in `testdata/ste100/compliant/`.
3. User ruleset fixtures in `testdata/custom_rulesets/`.
4. All Markdown files, wiki pages and comments of this repository.

The test copies each fixture into a temporary `docs/src/content/docs/` directory. Then it runs the unchanged oracle there. The oracle reports the first line of each paragraph, so the test maps each lexis position to the first line of its paragraph. Some differences are deliberate, for example `~~~` fences. Each deliberate difference has its own fixture and an entry in a wiki spec.

Done when the core rules match the oracle on every input, in both directions.

## Phase 5 — SARIF reports and the self-check

Issue: [ISSUE-005](.okfw/issues/issue-005-sarif-reports-and-the-self-check--01a1ed42b89d.md)

- [x] Add the SARIF report with `github.com/owenrumney/go-sarif/v3`
- [x] Prove that `Validate()` rejects a broken SARIF 2.2 report
- [x] Run the self-check on this repository
- [x] Add the SARIF guide to the docs site, and update the wiki

The SARIF report:

- `--sarif-version` selects `pkg/report/v210/sarif` or `pkg/report/v22/sarif`.
- The report sets `tool.driver.name` to `"lexis"`, and sets `tool.driver.version` and `tool.driver.informationUri`.
- `tool.driver.rules` holds the active rules of the resolved ruleset only.
- Each result has a relative `artifactLocation.uri`, `startLine`, `startColumn`, `endLine`, `endColumn`, `ruleId` and `level`. STE-100-05 results also have the replacement text.
- lexis calls `Validate()` before it writes the report.

The SARIF 2.2 schema uses JSON Schema draft 2020-12, and gojsonschema reads only drafts up to 07. If `Validate()` cannot reject a broken 2.2 report, use a draft 2020-12 validator. If none fits, mark SARIF 2.2 as experimental.

The self-check:

1. Run `zmake check`.
2. Run `./bin/lexis --version`.
3. Write a SARIF 2.1.0 report of this repository to `lexis-self.sarif`.
4. Write a SARIF 2.2 report of this repository to `lexis-self-v22.sarif`.
5. Make sure that both reports have zero results.

`zmake self-check` does steps 2 to 5. It loads `vocabulary.cue`, the technical terms of this repository, and uses `--fail-on note`, so each finding fails it ([ADR-0009](.okfw/decisions/adr-0009-the-self-check-fails-on-each-finding-with-a-repository-vocabulary--01a1ee772cc7.md)).

Result: go-sarif validates SARIF 2.2 reports. The 2.2 schema uses no keyword of the newer drafts, and the tests prove that `Validate()` rejects broken reports. So lexis needs no second validator, and 2.2 is not experimental.

## Phase 6 — ztree comment coverage

Issue: [ISSUE-006](.okfw/issues/issue-006-ztree-comment-coverage--01a1f474ef49.md)

- [x] Add comment tests for Zig, C and Rust, and for C++ code in a `.h` file
- [x] Test that a ztree error names the file, and record that it fails the scan
- [x] Check whether the Go build cache sees a new `libztree.a`
- [x] Update the wiki and the docs site

Result: the Rust test found a defect. The span of a Rust doc comment (`///` or `//!`) ends after the line break, so a failed assertion stopped lexis. The extractor now removes the line break. The Go build cache does not see a new `libztree.a`: ztree ISSUE-018 records it.

C++ files stay out of scope. [ISSUE-007](.okfw/issues/issue-007-c-comments--01a1f4750b43.md) records the change.

## Phase 7 — C++ comments

Issue: [ISSUE-007](.okfw/issues/issue-007-c-comments--01a1f4750b43.md)

- [x] Map `.cc`, `.cpp`, `.cxx`, `.hh`, `.hpp` and `.hxx` to the C++ grammar of ztree, and test each extension
- [x] Scan real C and C++ trees: V8, OpenSSL, ICU, the libstdc++ headers and the Linux headers
- [x] Update the CLI help, the wiki and the docs site

Result: C++ comments have the same forms as C comments. A raw string that holds `//` gives no finding. A `.h` file stays C.

The scan of V8 found a defect that also affects C files. A backslash at the end of a `//` comment continues the comment on the next line. Then one comment span holds more lines, and a failed assertion stopped lexis. The extractor now reads each line of a line comment separately.
