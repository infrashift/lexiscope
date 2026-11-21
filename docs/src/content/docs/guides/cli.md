---
title: Use the CLI
description: Check the version of lexiscope and prepare a scan command line.
sidebar:
  order: 1
---

This guide shows the common commands. The [CLI reference](../../reference/cli/) lists each flag.

## Check the version

1. Run `zmake build` in the repository root.
2. Run `bin/lexiscope --version`.

The output has the version, the commit and the build date:

```text
lexiscope version v0.0.0-dev (commit: 04c24b78cf5f, built: 2026-11-18T09:46:48Z)
```

## Scan the documentation

1. Run `zmake build`.
2. Run `bin/lexiscope scan docs/`.

lexiscope writes one line for each finding and a summary. The exit status is 1 when lexiscope finds an error, and 0 when it finds only warnings or nothing. `--fail-on warning` also gives status 1 for a warning.

Write a JSON report to a file:

```sh
bin/lexiscope scan --format json -o lexiscope.json docs/
```

Write a SARIF report for a code scanning tool. [Write a SARIF report](../sarif/) tells more:

```sh
bin/lexiscope scan --format sarif -o lexiscope.sarif docs/
```

Skip generated files with `--exclude`:

```sh
bin/lexiscope scan --exclude index.md --exclude .okfw/daily .okfw
```

## Scan source comments

lexiscope reads the comments of `.go`, `.zig`, `.py`, `.c`, `.h`, `.rs` and `.cs` files with ztree. It also reads the C++ files `.cc`, `.cpp`, `.cxx`, `.hh`, `.hpp` and `.hxx`. A `.h` file is C. A C# doc comment (`///`) can start with XML doc tags, such as `<summary>` and `<param name="x">`. lexiscope removes these tags and checks the text in them. It skips the lines of a `<code>` block. It skips tool directives such as `//go:build`, and indented code examples. A comment at the end of a code line stands alone.

```sh
bin/lexiscope scan cmd/ internal/
```

## Turn off a rule

`--omit-rule` and `--disable-rule` do the same thing. Give one rule ID, or a list with commas:

```sh
bin/lexiscope scan --omit-rule STE-100-07,STE-100-08 .
```

## See the active rules

`lexiscope rules list` takes the same ruleset flags as `lexiscope scan`. It shows the effect of each change before a scan:

```sh
bin/lexiscope rules list --omit-rule STE-100-07 --set STE-100-01.max_words=18
```

`lexiscope rules inspect ste-100` shows each parameter, the lexicon and the vocabulary.
