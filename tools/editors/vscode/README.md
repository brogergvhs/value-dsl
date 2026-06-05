# Value DSL for VS Code

VS Code language support for Value DSL.

The extension provides syntax highlighting for `.dsl` files and starts the
`dsl-lsp` language server for diagnostics, symbols, completion, and related
editor features.

It is not currently available on the marketplace, so please run it from local source.

## Run From Source

Build the grammar spec and LSP server from the repository root:

```bash
make grammar-export lsp
```

Install and build the extension:

```bash
cd tools/editors/vscode
bun ci
bun run build
```

Open the extension folder in VS Code:

```bash
code tools/editors/vscode
```

Then start an Extension Development Host:

1. Open the Run and Debug view.
2. Press `F5`, or alternatively search for `Debug: Start Debugging`.
3. In the new VS Code window, open a `.dsl` file from the repository or a DSL workspace.

The development extension automatically looks for the server at:

```text
../../../bin/dsl-lsp
```

relative to `tools/editors/vscode`, so `make lsp` is enough for the normal repo
layout.

## Configure a Custom LSP Path

If the server is somewhere else, set:

```json
{
  "dsl.lsp.path": "/absolute/path/to/dsl-lsp"
}
```

The extension resolves the server in this order:

1. `dsl.lsp.path`
2. bundled `bin/dsl-lsp` or `bin/dsl-lsp.exe`
3. repository development binary at `../../../bin/dsl-lsp`
4. `dsl-lsp` from `PATH`

## Local VSIX

You can also build a local VSIX:

```bash
cd tools/editors/vscode
bun run build
npx vsce package --out /tmp/value-dsl.vsix
```

This does not automatically bundle `dsl-lsp`. After installing the VSIX, either
put `dsl-lsp` on `PATH` or set `dsl.lsp.path`.
