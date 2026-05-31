SHELL := /bin/sh

APP := dsl
CMD := ./cmd/dsl-cli
BIN := ./bin/$(APP)

LSP_CMD := ./cmd/dsl-lsp
LSP_BIN := ./bin/dsl-lsp

TS_PARSER_DIR ?= $(HOME)/.local/share/nvim/lazy/nvim-treesitter/parser
TS_PARSER_SO  := $(TS_PARSER_DIR)/value_dsl.so

FILE ?= examples/basic.dsl
OUTPUT ?= conflicts.csv
FORMAT ?= csv

.PHONY: help deps build lsp run test fmt check clean parse validate analyze analyze-stdin grammar-export langspec-export tree-sitter-sync tree-sitter-test tree-sitter tree-sitter-install full

help:
	@echo "Usage: make <target> [FILE=examples/basic.dsl] [OUTPUT=conflicts.csv] [FORMAT=csv]"
	@echo ""
	@echo "Targets:"
	@echo "  deps          Download Go module dependencies"
	@echo "  build         Build CLI binary to $(BIN)"
	@echo "  lsp           Build LSP binary to $(LSP_BIN)"
	@echo "  run           Run CLI root command (shows help)"
	@echo "  test          Run all tests"
	@echo "  fmt           Format Go code"
	@echo "  check         Check Go formatting without modifying files (for CI)"
	@echo "  clean         Remove built artifacts"
	@echo "  parse         Parse DSL file"
	@echo "  validate      Validate DSL file"
	@echo "  analyze       Analyze DSL file"
	@echo "  analyze-stdin Analyze DSL from stdin: make analyze-stdin FILE=examples/basic.dsl"
	@echo "  grammar-export  Write generated/grammar.json from the current DSL grammar"
	@echo "  tree-sitter-sync Refresh generated/grammar.json, queries, and parser artifacts"
	@echo "  tree-sitter-test Run Tree-sitter drift checks and corpus tests"
	@echo "  tree-sitter-install Compile parser.so into TS_PARSER_DIR ($(TS_PARSER_DIR))"
	@echo "  tree-sitter   Run the full Tree-sitter generation, verification, and install pipeline"
	@echo "  full          Build CLI, LSP, regenerate Tree-sitter artifacts, and run tests"

deps:
	go mod download

build:
	@mkdir -p ./bin
	go build -o $(BIN) $(CMD)

lsp:
	@mkdir -p ./bin
	go build -o $(LSP_BIN) $(LSP_CMD)
	@$(LSP_BIN) --version >/dev/null

run:
	go run $(CMD)

test:
	go test ./...

fmt:
	gofmt -w $(shell find . -name '*.go' -type f)

check:
	@out=$$(gofmt -l $(shell find . -name '*.go' -type f)); \
	if [ -n "$$out" ]; then echo "needs formatting:"; echo "$$out"; exit 1; fi

clean:
	rm -rf ./bin

parse:
	go run $(CMD) parse "$(FILE)"

validate:
	go run $(CMD) validate "$(FILE)"

analyze:
	go run $(CMD) analyze "$(FILE)"

analyze-stdin:
	cat "$(FILE)" | go run $(CMD) analyze -

grammar-export:
	go run ./cmd/dsl-langspec-export

tree-sitter-sync:
	@command -v npm >/dev/null 2>&1 || { echo "error: npm not found"; exit 1; }
	go run ./cmd/dsl-langspec-export
	cd ./tools/tree-sitter-value-dsl && npm run generate

tree-sitter-test:
	@command -v npm >/dev/null 2>&1 || { echo "error: npm not found"; exit 1; }
	cd ./tools/tree-sitter-value-dsl && npm test

tree-sitter-install:
	@if [ ! -d "$(TS_PARSER_DIR)" ]; then \
		echo "skip: $(TS_PARSER_DIR) not found (set TS_PARSER_DIR to override)"; \
	else \
		cd ./tools/tree-sitter-value-dsl && tree-sitter build -o "$(TS_PARSER_SO)"; \
		echo "installed parser -> $(TS_PARSER_SO)"; \
	fi

tree-sitter: grammar-export
	@command -v npm >/dev/null 2>&1 || { echo "error: npm not found"; exit 1; }
	@command -v tree-sitter >/dev/null 2>&1 || { echo "error: tree-sitter CLI not found"; exit 1; }
	cd ./tools/tree-sitter-value-dsl && npm run generate
	cd ./tools/tree-sitter-value-dsl && node scripts/check-drift.mjs
	cd ./tools/tree-sitter-value-dsl && tree-sitter test
	$(MAKE) tree-sitter-install

full: test build lsp tree-sitter
