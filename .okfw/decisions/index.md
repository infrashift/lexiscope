# decisions

<!-- okfw:generated — maintained by okfw; edits here are overwritten -->

## Concepts

* [ADR-0001: Adopt an OKF wiki (okfw) for project knowledge](adr-0001-adopt-an-okf-wiki-okfw-for-project-knowledge--01a1ed42b864.md) - lexis keeps decisions, specs and tracked work in an OKF wiki in .okfw/, managed with okfw and its MCP server.
* [ADR-0002: Go 1.26, a small set of dependencies, and comments from ztree through cgo](adr-0002-go-1-26-a-small-set-of-dependencies-and-comments-from-ztree-through-cgo--01a1ed42b868.md) - lexis needs Go 1.26 or later, uses cobra, CUE and go-sarif as its only Go dependencies, and reads source comments with ztree through cgo.
* [ADR-0003: Adapt the NASA Power of 10 rules to Go](adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1ed42b86d.md) - How lexis applies each Power of 10 rule to Go code, which rules the Go port of nasa-lint checks, and the deviation for R3.
* [ADR-0004: CUE rulesets and one lexicon as the source of the substitution rules](adr-0004-cue-rulesets-and-one-lexicon-as-the-source-of-the-substitution-rules--01a1ed42b872.md) - lexis defines each ruleset in CUE, keys rules by stable ID, and makes the substitution and nominalization rules from one lexicon.
* [ADR-0005: The ste-100 rule IDs, with a core group that matches the oracle and an extended group](adr-0005-the-ste-100-rule-ids-with-a-core-group-that-matches-the-oracle-and-an-extended-group--01a1ed42b876.md) - STE-100-01 to 04 match scripts/ste-check.py exactly in both directions; STE-100-05 to 11 are warnings with exception lists and their own fixtures.
* [ADR-0006: SARIF 2.1.0 and 2.2 reports with go-sarif v3](adr-0006-sarif-2-1-0-and-2-2-reports-with-go-sarif-v3--01a1ed42b87b.md) - lexis writes SARIF 2.1.0 by default and SARIF 2.2 on request with go-sarif v3, and proves that validation rejects a broken 2.2 report.
* [ADR-0007: Document lexis with Astro Starlight on GitHub Pages](adr-0007-document-lexis-with-astro-starlight-on-github-pages--01a1ed42b880.md) - An Astro Starlight site in docs/, built with bun and deployed to GitHub Pages by GitHub Actions; the docs toolchain stays outside the Go gates.
* [ADR-0008: CLI exit statuses, output errors and commands of later phases](adr-0008-cli-exit-statuses-output-errors-and-commands-of-later-phases--01a1ee1490ca.md) - lexis exits with 0, 1 or 2, keeps the first output error in one writer, and its scan and rules commands check flags before their engines exist.
* [ADR-0009: The self-check fails on each finding, with a repository vocabulary](adr-0009-the-self-check-fails-on-each-finding-with-a-repository-vocabulary--01a1ee772cc7.md) - zmake self-check fails on each finding, and vocabulary.cue holds the technical terms of lexis that STE-100-06 and STE-100-07 do not report.
* [ADR-0010: A ztree error fails the scan](adr-0010-a-ztree-error-fails-the-scan--01a1f4789ef0.md) - When ztree cannot parse a source file, lexis stops the scan with status 2 and names the file and the ztree error.

