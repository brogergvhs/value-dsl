# value-dsl docs

This directory contains the Fumadocs/Next.js documentation site for
`value-dsl`. The content covers user-facing DSL usage and developer notes for
the Go parser, semantic model, validation, conflict analysis, LSP server,
tree-sitter and lspbench tooling.

For a short project setup path, start with the repository-level
[`README.md`](../README.md).

## Run Locally

The repository currently includes a Bun lockfile, so Bun is the preferred
package manager for the docs app:

```bash
cd docs
bun install
bun run dev
```

Open <http://localhost:3000>.

## Useful Commands

```bash
bun run dev         # start the local docs server
bun run build       # build the Next.js app
bun run start       # serve a production build
bun run lint        # run ESLint
bun run types:check # generate Fumadocs metadata and run TypeScript checks
```

Use the equivalent `npm run ...` commands if you installed with npm.

## Content Layout

- [`content/docs/(usage)`](<content/docs/(usage)>): getting started, language reference, CLI, validation, formatting, import/export, LSP, tree-sitter, and workspaces
- [`content/docs/dev`](content/docs/dev): architecture and implementation notes for the Go packages and tooling
- [`app/(home)`](<app/(home)>): documentation landing page
- [`app/docs`](app/docs): Fumadocs documentation routes
- [`app/api/search/route.ts`](app/api/search/route.ts): search route
- [`lib/source.ts`](lib/source.ts): Fumadocs content source loader
- [`lib/layout.shared.tsx`](lib/layout.shared.tsx): shared layout options

## Editing Docs

Most pages are MDX files under `content/docs`. Navigation is controlled by the
nearby `meta.json` files. After adding or moving pages, run:

```bash
bun run types:check
```

The generated `/llms.txt`, `/llms-full.txt`, and MDX routes read from the same
Fumadocs source configuration.
