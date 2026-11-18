---
title: Write a SARIF Report
description: Write SARIF 2.1.0 and 2.2 reports for code scanning tools, and know what each report holds.
sidebar:
  order: 3
---

SARIF is the Static Analysis Results Interchange Format of OASIS. Code scanning tools, for example GitHub code scanning, read SARIF 2.1.0. lexis also writes SARIF 2.2.

## Write a report

1. Run `zmake build`.
2. Run `bin/lexis scan --format sarif -o lexis.sarif docs/`.

For SARIF 2.2, add `--sarif-version 2.2`:

```sh
bin/lexis scan --format sarif --sarif-version 2.2 -o lexis-v22.sarif docs/
```

The exit status does not change with the format. Use `--fail-on` to fail a build on warnings or notes too:

```sh
bin/lexis scan --format sarif --fail-on warning -o lexis.sarif docs/
```

## What the report holds

Each report has one run.

- `tool.driver` has the name `lexis`, the version of the binary and `informationUri`.
- `tool.driver.rules` holds the active rules of the resolved ruleset only. Each rule has its ID, its title as `shortDescription`, its severity as `defaultConfiguration.level`, and a `helpUri` to the [rules reference](../../reference/rules/).
- Each result has `ruleId`, `ruleIndex`, `level` and the message.
- The location of each result has a relative `artifactLocation.uri` and a region: `startLine`, `startColumn`, `endLine` and `endColumn`.
- A STE-100-05 result has a fix. The fix replaces the term with the approved term. The replacement keeps a capital first letter, so `Utilize` becomes `Use`.
- `columnKind` is `utf16CodeUnits`. `endColumn` is the column after the last character of the finding.
- `newlineSequences` holds `\r\n`, `\n` and `\r`, because lexis counts a lone `\r` as a line break.

The URI of each file is relative to the current directory, with slashes. lexis escapes characters that a URI can not hold, for example a space becomes `%20`. Run lexis from the repository root, so that a code scanning tool finds each file.

A report of a scan without findings has an empty `results` array. SARIF reads an absent array as "the tool did not run".

## Validation

lexis validates each report against the schema of its SARIF version before it writes the report. A report that fails validation is a failure: lexis writes no file and exits with status 2.

lexis builds a SARIF 2.1.0 report, and go-sarif converts it to 2.2. The conversion is strict: it fails when the 2.2 report can not hold a field of the 2.1.0 report.

The SARIF 2.2 schema declares JSON Schema draft 2020-12. The validator of go-sarif reads drafts up to draft 07. The tests of lexis prove that the validation of 2.2 still works:

- The 2.2 schema uses no keyword that only the newer drafts have, and no `$ref` with a sibling constraint. A test reads the schema and fails when this changes.
- A test gives the validator broken 2.2 reports, for example a bad `level`, a `startLine` of 0 and a bad `guid`. The validator rejects each one.

SARIF 2.2 is not an OASIS standard yet. Code scanning tools can reject it. Use 2.1.0 for GitHub code scanning.

## Upload to GitHub code scanning

The `upload-sarif` action of `github/codeql-action` uploads a SARIF 2.1.0 file. Run the scan from the root of the checkout, so that each URI is relative to the repository root. See the GitHub documentation about SARIF uploads for the workflow syntax.
