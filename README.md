# value-dsl

`value-dsl` is a small Go-based DSL for describing stakeholders, values,
requirements, and stakeholder-value assignments. The CLI can validate DSL files,
format them, export/import JSON, and analyze requirements for value conflicts.

The repository also includes an LSP server, tree-sitter grammar assets, and a
Fumadocs documentation site.

Detailed language, CLI, workspace, LSP, and implementation notes live in
[`docs/`](docs/README.md).

## Requirements

- Go toolchain matching [`go.mod`](go.mod)
- `make`
- For the docs site: Bun or Node.js/npm
- Optional tree-sitter tooling for editor grammar work

## Quick Start

Build the CLI and LSP server:

```bash
make build
make lsp
```

Run the CLI against the included example:

```bash
./bin/dsl validate examples/basic.dsl
./bin/dsl analyze examples/basic.dsl
./bin/dsl format examples/basic.dsl
```

Or use `go run` while developing:

```bash
go run ./cmd/dsl-cli validate examples/basic.dsl
go run ./cmd/dsl-cli analyze examples/basic.dsl
```

Run the test suite:

```bash
make test
```

## A Minimal DSL File

```text
stakeholder Worker
stakeholder Manager

value privacy_custom = 1.57, 0.92

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker, Manager
```

For complete syntax and semantics, see the language docs under
[`docs/content/docs/(usage)/language`](<docs/content/docs/(usage)/language>).

## Common Targets

```bash
make help              # list available targets
make build             # build ./bin/dsl
make lsp               # build ./bin/dsl-lsp
make test              # run Go tests
make parse FILE=...    # parse a DSL file
make validate FILE=... # validate a DSL file
make analyze FILE=...  # report value conflicts
make full              # tests, binaries, tree-sitter pipeline
```

## Documentation Site

The docs site is in [`docs/`](docs/README.md):

```bash
cd docs
bun install
bun run dev
```

Then open <http://localhost:3000>.

## Repository Map

- [`cmd/dsl-cli`](cmd/dsl-cli): CLI entry point and commands
- [`cmd/dsl-lsp`](cmd/dsl-lsp): LSP server binary
- [`internal`](internal): parser, semantic model, validation, analysis, LSP, and formatting packages
- [`examples`](examples): sample DSL files and workspaces
- [`docs`](docs): documentation
- [`tools/tree-sitter-value-dsl`](tools/tree-sitter-value-dsl): generated tree-sitter grammar and queries
- [`tools/lspbench`](tools/lspbench): LSP benchmarking helper
