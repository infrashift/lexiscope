---
title: Contributing
description: The work loop and the gates of the lexis repository.
sidebar:
  order: 2
---

## Work loop

1. Pick the next task in `PLAN.md`.
2. Do the task.
3. Run `zmake check`.
4. Fix each error and each warning, then run `zmake check` again.
5. Go to the next task when the check reports zero errors and zero warnings.

At the end of a phase, update the wiki in `.okfw/` and the pages of this site. Then make one commit for the phase.

## Writing style

All documentation follows ASD-STE100: the Markdown files, the wiki pages, these pages, the CLI help text and the code comments. `zmake docs-style` checks the Markdown files and the wiki pages. From Phase 4, lexis also checks its own documentation and comments.

## Code rules

The Go source follows the NASA Power of 10 rules. See [Rules](../../reference/rules/) for the checks.
