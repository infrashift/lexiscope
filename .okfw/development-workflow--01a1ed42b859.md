---
id: "01a1ed42-b859-7667-a4f2-d6cbd596a2d5"
title: "Development workflow"
type: "Guide"
tags:
  - "conventions"
  - "workflow"
  - "meta"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T00:05:35-06:00"
description: "The loop for each lexis phase: do one task, run zmake check until green, update the wiki and the docs site, then commit."
---

# Development workflow

Every change follows this loop. Do not report or commit work while a gate fails.

1. **Pick the phase.** Read `PLAN.md`. Take the first unchecked phase. Find its mirror issue with `okfw search --path issues --tag open`.
2. **Tag the issue.** Run `okfw remove-tags <id> open && okfw add-tags <id> in-progress`.
3. **Do one task of that phase.** Write an ADR before you add a dependency.
4. **Apply the Power of 10 rules.** Read [ADR-0003](/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1ed42b86d.md). Use explicit bounded stacks, not recursion.
5. **Run every gate.** Run `zmake check`. It runs gofmt, `go vet`, nasa-lint, the tests, `okfw validate` and `ste-check.py`.
6. **Repeat until green.** Fix each error and each warning. Run `zmake check` again. Do not weaken or delete tests.
7. **Update the wiki during the work.** Record each decision as an ADR when you make it. File each bug or design debt as an issue.
8. **Go to the next task.** Repeat steps 3 to 7 for each task of the phase.
9. **Close the phase.** Tick the `PLAN.md` checkboxes and the acceptance criteria. Add a **Resolution** section. Swap the tag `in-progress` for `closed`.
10. **Update the docs site.** Add or change the `docs/` pages for the delivered features. Update the roadmap page. Run `zmake docs-build`.
11. **Journal.** Add an entry to today's `okfw daily` note: what changed, what you learned, what is next.
12. **Commit.** Commit the code, the wiki, the docs and `PLAN.md` together. Make one commit per phase. Do not push.

`ZMakefile` (zmake, modern mode) and `Makefile` (GNU Make) have the same targets. Change both files together. Verify that `zmake -n check` and `make -n check` print the same commands.

See [Wiki conventions](/wiki-conventions--01a1ed42b854.md).
