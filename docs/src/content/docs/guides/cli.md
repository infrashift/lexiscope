---
title: Use the CLI
description: Check the version of lexis and prepare a scan command line.
sidebar:
  order: 1
---

This guide shows the commands that work in Phase 3. The [CLI reference](../../reference/cli/) lists each flag.

## Check the version

1. Run `zmake build` in the repository root.
2. Run `bin/lexis --version`.

The output has the version, the commit and the build date:

```text
lexis version v0.0.0-dev (commit: 04c24b78cf5f, built: 2026-11-18T09:46:48Z)
```

## Prepare a scan

`lexis scan` checks each flag and loads the ruleset before it starts. Use this to test a command line for a CI job:

```sh
bin/lexis scan --format sarif --sarif-version 2.2 -o lexis.sarif \
  --set STE-100-01.max_words=18 --omit-rule STE-100-07 docs/
```

Phase 4 adds the scan itself. Until then, a correct command line exits with status 2 and this message:

```text
lexis: scan docs/: not available yet: ruleset ste-100 has 10 active rules, but the rule engine arrives in Phase 4 of PLAN.md
```

An incorrect command line also exits with status 2. The message names the flag and the problem:

```text
lexis: --sarif-version needs --format sarif
```

## Turn off a rule

`--omit-rule` and `--disable-rule` do the same thing. Give one rule ID, or a list with commas:

```sh
bin/lexis scan --omit-rule STE-100-07,STE-100-08 .
```

## See the active rules

`lexis rules list` takes the same ruleset flags as `lexis scan`. It shows the effect of each change before a scan:

```sh
bin/lexis rules list --omit-rule STE-100-07 --set STE-100-01.max_words=18
```

`lexis rules inspect ste-100` shows each parameter, the lexicon and the vocabulary.
