---
x-okfw-id: "01a1ed42-b854-79cc-8016-b5f49fdac3e1"
title: "Wiki conventions"
type: "Guide"
tags:
  - "conventions"
  - "meta"
x-okfw-created: "2026-11-18T00:05:35-06:00"
x-okfw-timestamp: "2026-11-18T00:05:35-06:00"
description: "Layout, writing style and authoring rules for architecture notes, specs, ADRs, issues and daily notes in this wiki."
---

# Wiki conventions

This guide defines how the lexiscope OKF wiki (`.okfw/`) stores project knowledge.

- Create concepts with `okfw create` or the okfw MCP `create` tool. Do not write frontmatter by hand.
- Search before you write. Link to related concepts.
- Keep one idea per concept. Give each concept a one-sentence `description`.
- Do not start a `--body` with `# Title`. okfw writes the H1 from the title.

## Layout
| Path | Type | Contents |
|---|---|---|
| `architecture/` | `Architecture` | System overview, package boundaries, data flow. |
| `specs/` | `Spec` | Contracts, for example the ruleset schema and the known differences from the oracle. |
| `decisions/` | `ADR` | Architecture decision records. |
| `issues/` | `Issue` (kind `issue`) | Tracked work: one issue per `PLAN.md` phase, plus bugs and design debt. |
| `daily/` | daily note | `okfw daily`: what changed, what you learned, what is next. |

## PLAN.md and issues
`PLAN.md` is the phase checklist. Each phase from 1 to 5 has a mirror issue (ISSUE-001 to ISSUE-005). Each phase heading in `PLAN.md` has an `Issue:` line that links to its issue. At the end of a phase, tick the `PLAN.md` checkboxes and the acceptance criteria of the issue.

## ADRs (`decisions/`)
- Title: `ADR-NNNN: <decision>`. Number ADRs in sequence. Do not reuse a number.
- Command: `okfw create "ADR-NNNN: <title>" --path decisions --type ADR --tags adr,<topic> -d "<one sentence>" --body "<markdown>"`
- Body: **Status** (Proposed, Accepted, or Superseded by ADR-XXXX), **Date**, **Source**, then **Context**, **Decision**, **Consequences**, **Alternatives considered**.
- To supersede an ADR, write the new ADR with a link to the old ADR. Update the Status line of the old ADR. Run `okfw set-lifecycle <id> deprecated`.
- Write an ADR before you add a Go module dependency or a docs package.

## Issues (`issues/`)
- Title: `ISSUE-NNN: <summary>`. Number issues in sequence.
- Command: `okfw create "ISSUE-NNN: <title>" --kind issue --path issues --type Issue --tags issue,open,phase-<n>,<area> -d "<summary>" --body "<markdown>"`
- Body: a **Status/Phase** line, **Description**, **Acceptance Criteria** (checkboxes), **Notes**, **Related**.
- okfw has no status field for issues. Tags carry the state: `open`, then `in-progress`, then `closed`. Use `okfw remove-tags` and `okfw add-tags`. Keep the Status line the same as the tags. Add a **Resolution** section when you close an issue.
- Find open work with `okfw search --path issues --tag open`.

## Writing style
All documentation follows ASD-STE100 Simplified Technical English. This includes the wiki, the docs site, `README.md`, `PLAN.md`, `CLAUDE.md`, the CLI help text and the code comments.

- Use active voice and simple present tense: "The parser reads the document."
- Start each procedural step with an imperative verb: Run, Read, Inspect, Verify, Add, Set.
- Keep procedural steps to 20 words or fewer. Keep descriptive sentences to 25 words or fewer.
- Use a verb, not a noun form: "lexiscope checks the file", not "The file check occurs".
- Do not use `your`. Write "the ruleset file".
- Use one term for one concept. This wiki uses these terms:
  - *rule*: one check with a stable ID, for example `STE-100-01`.
  - *ruleset*: a set of rules in CUE, for example `ste-100`.
  - *lexicon*: the map from each concept to its approved term, its unapproved synonyms and its nominalizations.
  - *finding*: one result of one rule at one position.
  - *oracle*: `scripts/ste-check.py`. The parity test compares lexiscope with it.
  - *core rule* and *extended rule*: the two groups of `ste-100` rules ([ADR-0005](/decisions/adr-0005-the-ste-100-rule-ids-with-a-core-group-that-matches-the-oracle-and-an-extended-group--01a1ed42b876.md)).
  - *step*: a numbered list item. *descriptive sentence*: each other sentence.
  - *comment*: the text of one source code comment. ztree gives its span.
  - *phase*: one unit of work in `PLAN.md`.
- Do not use the words `should`, `will`, `would`, `please`, `simply`, `just`, `very`, `really` or `easily`. Do not use marketing language.
- Run `zmake docs-style`. The check (`scripts/ste-check.py`) reports banned words and long sentences as errors. It reports possible passive voice as warnings. Fix the warnings too.

## Housekeeping
- Use `okfw move` and `okfw delete`. Do not use `mv` or `rm`. The okfw commands repair links and update the `index.md` listings.
- `status` is the OKF lifecycle field (`draft`, `stable` or `deprecated`). It is not the task state. Mark superseded knowledge `deprecated`. Do not delete it.
- Do not edit `index.md` or `log.md`. okfw writes them.
- Run `zmake wiki` before you commit wiki changes.

See [Development workflow](/development-workflow--01a1ed42b859.md).
