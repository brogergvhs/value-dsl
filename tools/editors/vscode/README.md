# Value DSL for VS Code

VS Code language support for Value DSL.

The extension provides syntax highlighting for `.dsl` files and starts the
`dsl-lsp` language server for diagnostics, symbols, completion, and related
editor features.

Platform-specific Marketplace packages include a bundled `dsl-lsp` executable.
The universal package falls back to `dsl-lsp` on `PATH`, or to the path set in
`dsl.lsp.path`.
