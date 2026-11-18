---
id: "01a1ed42-b864-7056-91b8-45af1d0e7b08"
title: "ADR-0001: Adopt an OKF wiki (okfw) for project knowledge"
type: "ADR"
tags:
  - "adr"
  - "meta"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T00:05:35-06:00"
description: "lexis keeps decisions, specs and tracked work in an OKF wiki in .okfw/, managed with okfw and its MCP server."
---

# ADR-0001: Adopt an OKF wiki (okfw) for project knowledge

**Status:** Accepted
**Date:** 2026-11-18
**Source:** User request. zmake and ztree use the same setup.

## Context
lexis makes design decisions that the code does not record. Examples are the rule IDs, the limits of the parity with the oracle and the Power of 10 deviations. An agent and a person both need a searchable record of these decisions.

## Decision
- lexis keeps an OKF wiki in `.okfw/` and manages it with `okfw`.
- `.mcp.json` registers `okfw mcp` as the MCP server for Claude Code.
- `okfw install --platform claude` writes the skill and the okfw block in `CLAUDE.md`.
- The wiki follows the [Wiki conventions](/wiki-conventions--01a1ed42b854.md) guide. The gate `zmake wiki` validates the wiki.

## Consequences
- Each phase of `PLAN.md` has a mirror issue.
- Each decision gets an ADR when the work makes it.
- The `okfw` CLI and the MCP server give the same operations. Work continues when the MCP server is not connected.
- The wiki is for developers. The docs site in `docs/` is for users ([ADR-0007](/decisions/adr-0007-document-lexis-with-astro-starlight-on-github-pages--01a1ed42b880.md)).

## Alternatives considered
- **Plain Markdown files in the repository.** Rejected: no frontmatter schema, no link repair and no search.
