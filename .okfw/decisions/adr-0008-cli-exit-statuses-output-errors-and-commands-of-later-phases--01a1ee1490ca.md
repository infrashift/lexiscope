---
x-okfw-id: "01a1ee14-90ca-7915-81c8-97ec7e7ccc53"
title: "ADR-0008: CLI exit statuses, output errors and commands of later phases"
type: "ADR"
tags:
  - "adr"
  - "cli"
x-okfw-created: "2026-11-18T03:54:47-06:00"
x-okfw-timestamp: "2026-11-18T03:54:47-06:00"
description: "lexis exits with 0, 1 or 2, keeps the first output error in one writer, and its scan and rules commands check flags before their engines exist."
---

# ADR-0008: CLI exit statuses, output errors and commands of later phases

**Status:** Accepted
**Date:** 2026-11-18
**Source:** Phase 2, [ISSUE-002](/issues/issue-002-cli-version-and-ruleset-overrides-and-the-go-port-of-nasa-lint--01a1ed42b88b.md).

## Context
lexis runs in CI jobs, so its exit status has to be stable. Rule R7 forbids discarded results, and errcheck rejects an unchecked write. A check after each write makes the command code long. `PLAN.md` puts the `lexis scan` and `lexis rules` commands in Phase 2, but the ruleset engine arrives in Phase 3 and the rule engine in Phase 4.

## Decision
- **Exit status:** 0 when the command succeeds without error findings, 1 when a scan finds errors, and 2 on a usage error or a failure. A failed assertion also gives 2. Phase 5 adds `--fail-on <error|warning|note>`: status 1 means a finding at that severity or higher. The default is `error` ([ADR-0009](/decisions/adr-0009-the-self-check-fails-on-each-finding-with-a-repository-vocabulary--01a1ee772cc7.md)).
- **Messages:** lexis writes each failure to stderr as `lexis: <message>`. Cobra does not print its own errors or the usage text after an error.
- **Boundary:** `cli.execute` defers `assert.Recover(&code, &failure)`. `assert.Recover` does no output. The caller writes the message.
- **Output errors:** `internal/output.Writer` writes formatted text and keeps the first error. A command checks `Err()` once at its end and exits with status 2 when the output failed.
- **Commands of later phases:** `lexis scan` and the `lexis rules` commands check their flags in Phase 2. Then they return an error that wraps `errNotAvailable` and names the phase that delivers the work.
- **Flag rules:** `--disable-rule` is an alias of `--omit-rule` through the pflag name function. `--sarif-version` needs `--format sarif`. One rule ID in both `--enable-rule` and `--omit-rule` is an error. `--set` takes `rule_id.param=value`. The value stays text, and Phase 3 checks its type.
- **Version:** `zmake build` sets the version from `git describe`, the commit from `git rev-parse` and the date from the time of the last commit. Two builds of one commit give the same version line. A build without `-ldflags` reads the version and the VCS stamp of the binary.
- **Docs:** a test checks that the CLI reference page lists each command and each flag.

## Consequences
- A CI job can tell findings (1) from a broken setup (2).
- Phase 3 and Phase 4 replace the `errNotAvailable` returns. The flags stay the same. Phase 5 removes `errNotAvailable`, because each command works.

## Alternatives considered
- **A check after each write.** Rejected: it doubles the length of the output code.
- **No `scan` and `rules` commands before their engines exist.** Rejected: the flag checks and the docs can start in Phase 2, and the tests then cover the flags early.
