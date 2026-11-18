---
title: CLI
description: The commands, flags and exit statuses of lexis.
sidebar:
  order: 0
---

<!-- internal/cli/cli_test.go checks that this page lists each command and each flag. -->

The `lexis rules` commands work from Phase 3. `lexis scan` checks its flags and loads the ruleset now. It checks files from Phase 4. Until then it exits with status 2 and names the phase.

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

Lists each rule of the resolved ruleset after each change: its status, its severity, its check, its word limit and its title.

| Flag | Default | Meaning |
|---|---|---|
| `--ruleset`, `-r` | `ste-100` | A built-in ruleset name, or one or more CUE files or directories. |
| `--enable-rule` | none | Turn on these rule IDs. |
| `--omit-rule`, `--disable-rule` | none | Turn off these rule IDs. |
| `--set` | none | Change a rule parameter: `rule_id.param=value`. |
| `--allow-term` | none | Add approved technical terms. |

```text
$ lexis rules list --omit-rule STE-100-07 --set STE-100-01.max_words=18
ruleset ste-100: 11 rules, 10 active, 1 omitted
ID          STATUS   SEVERITY  CHECK                         LIMIT  TITLE
STE-100-01  active   error     max_words                     18     Limit the length of a procedural step
...
STE-100-07  omitted  warning   ambiguous_gerunds             -      Avoid unclear -ing forms
```

## `lexis rules inspect`

```text
lexis rules inspect <ruleset>
```

Shows each rule with its parameters, then the lexicon and the vocabulary of one ruleset. The argument is a built-in name, a CUE file or a directory.

## Values of `--set`

lexis reads the value of `--set` as CUE. The value of a parameter in a CUE file has the same syntax. Text that is not valid CUE is a string.

| Flag | Effect |
|---|---|
| `--set STE-100-01.max_words=18` | Sets the word limit to the number 18. |
| `--set 'STE-100-01.procedural_markers=["numbered","bullet"]'` | Makes bullet items steps too. |
| `--set 'STE-100-04.tokens={"utilise": true}'` | Adds a banned term. A set only grows. |
| `--set STE-100-09.pattern=x+` | Sets a string. |

A value that does not fit the parameter fails with the path of the field, for example `rules."STE-100-01".params.max_words: conflicting values "many" and int`.
