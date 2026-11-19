---
title: CLI
description: The commands, flags and exit statuses of lexis.
sidebar:
  order: 0
---

<!-- internal/cli/cli_test.go checks that this page lists each command and each flag. -->

This page lists each command and each flag of lexis.

## Exit status

| Status | Meaning |
|---|---|
| 0 | The command succeeded. A scan found no finding at the `--fail-on` severity or higher. |
| 1 | A scan found one or more findings at the `--fail-on` severity or higher. The default is `error`, so warnings and notes do not change the status. |
| 2 | A usage error or a failure, for example a SARIF report that fails validation. |

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

Checks Markdown files (`.md`, `.mdx`, `.markdown`) and the comments of Go, Zig, Python, C, C++ and Rust files. The source extensions are `.go`, `.zig`, `.py`, `.c`, `.h`, `.rs` and the C++ extensions `.cc`, `.cpp`, `.cxx`, `.hh`, `.hpp`, `.hxx`. A `.h` file is C. Without paths, it checks the current directory.

A path can be a file or a directory. In a directory, lexis checks each file of a known type. It does not enter hidden directories, `node_modules`, `vendor`, `testdata` or `dist`. To check a hidden directory, name it as a path, for example `.okfw`. A named file of another type is an error.

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
| `--exclude` | none | Skip each path that matches the pattern. Repeat the flag for each pattern. |
| `--fail-on` | `error` | The lowest severity that gives exit status 1: `error`, `warning` or `note`. |

An exclude pattern uses the syntax of Go's `path.Match`. It matches a path, each parent directory of the path, or the base name. For example, `--exclude .okfw/daily` skips the directory and its files, and `--exclude index.md` skips each file with that name.

The text report has one line for each finding, then a summary:

```text
docs/guide.md:12:9: error: STE-100-04: Do not use "will".
lexis: 1 error(s), 0 warning(s), 0 note(s) in 14 file(s), ruleset ste-100
```

The line and the column start at 1. The column counts UTF-16 code units, as SARIF does. The JSON report has a `summary` object and a `findings` list with the same fields. [Write a SARIF report](../../guides/sarif/) tells what a SARIF report holds.

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
