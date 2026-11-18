// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// Deployed to GitHub Pages by .github/workflows/docs-release.yml (ADR-0007).
// Each PLAN.md phase adds or updates the pages for the features it delivers
// (see the Development workflow guide in .okfw/). The sidebar groups list the
// pages of their directory, so a new page needs no change here. Phase 2 adds
// a Guides group with the first guide.
// https://astro.build/config
export default defineConfig({
	site: 'https://infrashift.github.io',
	base: '/lexis',
	integrations: [
		starlight({
			title: 'lexis',
			description:
				'A documentation linter in Go: it checks Markdown files and source code comments against CUE rulesets such as ASD-STE100.',
			favicon: '/favicon.svg',
			customCss: ['./src/styles/custom.css'],
			social: [
				{
					icon: 'github',
					label: 'GitHub',
					href: 'https://github.com/infrashift/lexis',
				},
			],
			editLink: {
				baseUrl: 'https://github.com/infrashift/lexis/edit/main/docs/',
			},
			sidebar: [
				{ label: 'Start Here', items: [{ autogenerate: { directory: 'start-here' } }] },
				{ label: 'Reference', items: [{ autogenerate: { directory: 'reference' } }] },
				{ label: 'Project', collapsed: true, items: [{ autogenerate: { directory: 'project' } }] },
			],
		}),
	],
});
