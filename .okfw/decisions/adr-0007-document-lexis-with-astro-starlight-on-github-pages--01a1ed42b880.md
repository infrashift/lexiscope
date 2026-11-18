---
id: "01a1ed42-b880-7af0-b709-7c3f22eb52d5"
title: "ADR-0007: Document lexis with Astro Starlight on GitHub Pages"
type: "ADR"
tags:
  - "adr"
  - "docs"
created: "2026-11-18T00:05:35-06:00"
timestamp: "2026-11-18T00:05:35-06:00"
description: "An Astro Starlight site in docs/, built with bun and deployed to GitHub Pages by GitHub Actions; the docs toolchain stays outside the Go gates."
---

# ADR-0007: Document lexis with Astro Starlight on GitHub Pages

**Status:** Accepted
**Date:** 2026-11-18
**Source:** User request, Phase 1, [ISSUE-001](/issues/issue-001-bootstrap-the-repository-the-gates-the-wiki-and-the-docs-site--01a1ed42b886.md). zmake ADR-0009.

## Context
The wiki describes lexis for its developers. Users need a public site with guides and a reference. zmake publishes an Astro Starlight site to GitHub Pages with bun 1.3.9 and a GitHub Actions workflow.

## Decision
- The site lives in `docs/`. It uses Astro 7.4.11, Starlight 0.41.11, sharp 0.35.5 and bun 1.3.9, the same versions as zmake. `docs/bun.lock` pins the packages.
- The Astro `base` is `/lexis`, so the site is at https://infrashift.github.io/lexis/.
- `zmake docs-dev` runs the site locally. `zmake docs-build` builds it into `docs/dist/`. Both install packages with `--frozen-lockfile`.
- `.github/workflows/docs-release.yml` builds the site with `withastro/action` and deploys it with `actions/deploy-pages`. A pull request only builds the site.
- The Pages source of the repository is "GitHub Actions".
- `scripts/ste-check.py` checks the docs pages against ASD-STE100.
- `zmake check` does not build the site, so the Go gates do not need bun.

**Differences from zmake:**
1. The home page is `src/content/docs/index.mdx` with the Starlight `splash` template. zmake has a custom `index.astro` page, so its Starlight pages live under `/docs/`. The splash page needs no custom CSS, and `ste-check.py` reads its text.
2. `docs-dev` also installs with `--frozen-lockfile`. A plain `bun install` can change `bun.lock` during local work. To change a package, edit `docs/package.json` and run `bun install` in `docs/`.
3. Each sidebar group lists the pages of its directory (`autogenerate`). A new page needs no change to `astro.config.mjs`. Starlight shows an empty heading for a group with no pages. The Guides group starts in Phase 2 with the first guide.
4. `docs/go.mod` declares a separate Go module. Without it, `go list ./...` walks `docs/node_modules`. A `.go` file in an npm package then joins the vet and the tests. A test with a probe file showed this.

## Consequences
- Each phase updates the docs pages. The [Development workflow](/development-workflow--01a1ed42b859.md) has a step for it.
- The build prints two warnings from Starlight: the `i18n` collection is empty, and the `404` docs entry does not exist. The zmake build prints the same two warnings. A custom `404.md` page gives a route conflict warning instead.
- Astro uses the next free port when port 4321 is in use. When it detects an agent, `astro dev` runs in the background and `astro dev stop` stops it.

## Alternatives considered
- **A custom landing page as in zmake.** Rejected by the user: more code, and `ste-check.py` does not read `.astro` files.
- **The Jekyll build of GitHub Pages.** Rejected: no search and no sidebar. It also conflicts with the Actions deploy.
