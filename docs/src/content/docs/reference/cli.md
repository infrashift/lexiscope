---
title: CLI
description: The commands, flags and exit statuses of lexis.
sidebar:
  order: 0
---

<!-- internal/cli/cli_test.go checks that this page lists each command and each flag. -->

The commands below exist from Phase 2. `lexis scan` and the `lexis rules` commands check their flags now. They do their work from Phase 4 and Phase 3. Until then they exit with status 2 and name the phase.

## Exit status

| Status | Meaning |
|---|---|
| 0 | The command succeeded. A scan found no error findings. |
| 1 | A scan found one or more error findings (from Phase 4). |
| 2 | A usage error, a failure, or a command that a later phase delivers. |

lexis writes each failure to stderr as `lexis: <message>`.

## `lexis`

| Flag | Meaning |
|---|---|
| `--version`, `-v` | Write the version line and exit. |
| `--help`, `-h` | Write the help text and exit. |

## `lexis version`

Writes the version line, the same line as `lexis --version`:

```text
lexis version v0.1.0 (commit: 1a2b3c4d5e6f, built: 2026-11-18T06:00:00Z)
```

The `zmake build` target sets the version from `git describe` and the date from the time of the last commit. A build without these values, for example `go install`, reads the version and the commit that Go stamps into the binary.

## `lexis scan`

```text
lexis scan [flags] [paths...]
```

Checks Markdown files and the comments of Go, Zig, Python, C and Rust files. Without paths, it checks the current directory.

| Flag | Default | Meaning |
|---|---|---|
| `--ruleset`, `-r` | `ste-100` | A built-in ruleset name, or one or more CUE files or directories. Repeat the flag or separate values with commas. |
| `--enable-rule` | none | Turn on these rule IDs. |
| `--omit-rule`, `--disable-rule` | none | Turn off these rule IDs. |
| `--set` | none | Change a rule parameter: `rule_id.param=value`, for example `STE-100-01.max_words=18`. Repeat the flag for each change. |
| `--allow-term` | none | Add approved technical terms. |
| `--format` | `text` | The report format: `text`, `json` or `sarif`. |
| `--sarif-version` | `2.1.0` | The SARIF version: `2.1.0` or `2.2`. It needs `--format sarif`. |
| `--output`, `-o` | `-` | The report file. `-` is stdout. |

A rule ID has capital letters, digits and hyphens, for example `STE-100-01` or `DOC-CUSTOM-01`. A parameter name has lowercase letters, digits and underscores. lexis rejects a rule that `--enable-rule` and `--omit-rule` both name.

## `lexis rules list`

Lists the active rules, the omitted rules and the word limits of the resolved ruleset.

| Flag | Default | Meaning |
|---|---|---|
| `--ruleset`, `-r` | `ste-100` | A built-in ruleset name, or one or more CUE files or directories. |

## `lexis rules inspect`

```text
lexis rules inspect <ruleset>
```

Shows each rule, the lexicon and the word limits of one ruleset.
