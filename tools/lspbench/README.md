# LSP Bench

`lspbench` is a small stdio LSP benchmark harness. It starts a language server,
opens a file through JSON-RPC/LSP messages, and measures startup,
first-open readiness, and configured warm requests.

Full setup, config format, metrics, graph interpretation, and comparison notes
are documented in
[`docs/content/docs/dev/lsp-benchmarking.mdx`](../../docs/content/docs/dev/lsp-benchmarking.mdx).

## Quick Start

Build the value-dsl LSP server first:

```bash
make lsp
```

Run the bundled value-dsl benchmark:

```bash
go run ./tools/lspbench -config tools/lspbench/configs/value-dsl.json -runs 5 -count 30
```

Useful variants:

```bash
go run ./tools/lspbench -config tools/lspbench/configs/value-dsl.json -runs 5 -count 30 -format json
go run ./tools/lspbench -config tools/lspbench/configs/value-dsl.json -runs 5 -count 30 -format csv
go run ./tools/lspbench -config tools/lspbench/configs -runs 5 -count 30
go run ./tools/lspbench -config tools/lspbench/configs -runs 5 -count 30 -graph lspbench.png
go run ./tools/lspbench -config tools/lspbench/configs/value-dsl.json -warmup 0 -runs 5 -count 30
```

## Configs

`-config` accepts either one JSON config file or a directory of `*.json`
configs. Directory mode runs configs in sorted order and continues if one config
fails.

Bundled configs and templates live in [`configs`](configs):

- `value-dsl.json`
- `gopls.template.json`
- `rust-analyzer.template.json`
- `pyright.template.json`
- `gleam.template.json`

To benchmark another language server, copy a template and set `command`, `root`,
`file`, `languageId`, and request cursor positions for a representative project.
