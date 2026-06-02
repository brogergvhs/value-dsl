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
make build       # standalone ./bin/dsl
make build-full  # combined ./bin/dsl-full with `dsl-full lsp`
make lsp         # standalone ./bin/dsl-lsp
```

For a combined binary literally named `dsl`, run `make build-full FULL_BIN=./bin/dsl`.

Run the CLI against the included example:

```bash
./bin/dsl validate examples/dsl/01_minimal_valid.dsl
./bin/dsl analyze examples/dsl/01_minimal_valid.dsl
./bin/dsl format examples/dsl/01_minimal_valid.dsl
```

Or use `go run` while developing:

```bash
go run ./cmd/dsl-cli validate examples/dsl/01_minimal_valid.dsl
go run ./cmd/dsl-cli analyze examples/dsl/01_minimal_valid.dsl
```

Run the test suite:

```bash
make test
```

Build and use the Docker image:

```bash
make docker-build
make docker-validate FILE=examples/dsl/01_minimal_valid.dsl
make docker-analyze FILE=examples/dsl/01_minimal_valid.dsl
```

The image contains `/usr/local/bin/dsl`, `/usr/local/bin/dsl-full`, and
`/usr/local/bin/dsl-lsp`. For LSP-over-stdio usage, keep stdin open and mount
the workspace:

```bash
docker run --rm -i -v "$PWD:/workspace" -w /workspace value-dsl:local /usr/local/bin/dsl-lsp
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
make build             # build standalone ./bin/dsl
make build-full        # build combined ./bin/dsl-full
make lsp               # build ./bin/dsl-lsp
make docker-build      # build value-dsl:local Docker image
make test              # run Go tests
make parse FILE=...    # parse a DSL file
make validate FILE=... # validate a DSL file
make analyze FILE=...  # report value conflicts
make full              # tests, binaries, tree-sitter pipeline
```

## Releases

Pushes and pull requests run the Go checks and build all three binaries. Tagged
pushes create GitHub Releases with GoReleaser:

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

Release artifacts include `dsl`, `dsl-full`, and `dsl-lsp` for Linux, macOS,
and Windows.

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
