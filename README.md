# symbolc

`symbolc` is a Go CLI that analyzes a codebase and generates a compact,
structured map of its packages, symbols, and relationships.

It is designed for AI-assisted development: an agent can discover the relevant
parts of a project from generated YAML artifacts before reading the source
tree. The artifacts are navigation aids, not replacements for source code.

## What it generates

For a project with one or more executables, `symbolc` generates an output
directory like this:

```text
symbolc/
├── index.yml
└── <executable-path>/
    ├── index.yml
    ├── flux-entrypoints/
    │   └── <package>.yml
    └── contracts/
        └── <package>/<symbol>.yml
```

The root `index.yml` is the atlas of executables. Each executable has an
isolated session, preventing packages with the same name from colliding across
different applications.

An executable session contains:

- `index.yml`: references to its flux-entrypoints and contracts;
- `flux-entrypoints/`: dependency-oriented maps starting at packages imported
  directly by the executable package;
- `contracts/`: one YAML contract per analyzed symbol.

Contracts contain canonical symbol IDs, package identity, source location,
kind, method relationships, function/method signatures, and struct fields when
available. Paths inside the artifacts are relative to the configured project
directory.

## Requirements

- Go 1.25 or newer;
- a valid Go module in each configured source root;
- permission to read the source tree and write the configured output path.

## Installation

### Go tool

Add the following to your `go.mod`:
```sh
tool github.com/Paulooo0/symbolc/cmd/symbolc
```

Update your `go.mod`:
```sh
go mod tidy
```

Then run with `go tool`:
```sh
go tool symbolc
```

### Go install

Install a published Go module version with:
```sh
go install github.com/Paulooo0/symbolc/cmd/symbolc@latest
```

### From source

Build from this repository:
```sh
go build -o bin/symbolc ./cmd/symbolc
```

Then run `bin/symbolc`.

## Quick start

Create `symbolc.yml` at the root of the project to analyze:

```yaml
version: 1

source:
  roots:
    - .

output:
  format: yaml
  path: symbolc

contracts:
  include:
    - internal/**
  exclude:
    - "**/*_test.go"
    - "**/generated/**"
```

Run:

```sh
symbolc sync
```

This reconstructs the configured output directory from the current source
tree. Generated files are derived artifacts and should not be edited manually.
Add the output directory to `.gitignore` when the context is intended to remain
local:

```gitignore
symbolc
```

## Configuration

### `version`

The configuration schema version. The current supported value is `1`. If it is
omitted, version `1` is used.

### `source.roots`

Directories from which Go packages are loaded. Relative paths are resolved from
the directory containing `symbolc.yml`.

```yaml
source:
  roots:
    - .
    - tools
```

### `output.format`

The output format. The current implementation supports only `yaml`.

### `output.path`

The output directory. Relative paths are resolved from the directory containing
`symbolc.yml`.

If the path is empty or `-`, `sync` writes only the root `index.yml` to stdout.
This is useful for discovery, but a directory output is recommended when
multiple agents need to reuse the same generated context.

### `contracts.include` and `contracts.exclude`

These patterns control which symbol contracts are materialized. They use
slash-separated paths and support `*`, `?`, and `**` patterns.

Filtering contracts does not change the source analysis itself. It only limits
which contracts are written to the generated context.

For example, to include application code while excluding tests and generated
files:

```yaml
contracts:
  include:
    - internal/**
    - cmd/**
  exclude:
    - "**/*_test.go"
    - "**/generated/**"
```

## CLI usage

### Synchronize artifacts

```sh
symbolc sync
```

### Use another configuration file

```sh
symbolc sync --config path/to/symbolc.yml
```

### Override the output directory

```sh
symbolc sync --output path/to/context
```

The command treats this path as a directory and replaces its managed contents
atomically.

### Write the root atlas to stdout

```sh
symbolc sync --output -
```

### Override the format

```sh
symbolc sync --format yaml
```

Only YAML is currently supported.

## Recommended workflow for AI agents

When a project contains generated `symbolc` artifacts, an AI agent should use
them as an atlas before exploring the source tree.

Recommended order:

1. Read `<output>/index.yml` to identify the relevant executable.
2. Read that executable's `index.yml`.
3. Select the relevant flux-entrypoint for the task.
4. Use the `symbols` references in the flux-entrypoint to locate contracts for
   the relevant packages.
5. Read only the contracts needed to understand the change.
6. Follow each contract's `source.file` and line information into the source
   code for implementation details.
7. Read related tests and mocks when changing behavior or adding coverage.
8. If the source code changes, run `symbolc sync` again before subsequent
   navigation or review.

The generated context should answer “where should I look?” quickly. The source
code remains authoritative for exact behavior, implementation details, and
tests.

### Suggested agent instruction

Projects can include an instruction similar to this in their agent context:

```text
Before exploring the source tree, inspect the generated symbolc context.
Start at symbolc/index.yml, follow the executable index, then select the
relevant flux-entrypoint. Use its package and symbol references to locate the
smallest set of contracts needed for the task. For each selected contract,
follow source.file and source line information into the code. Do not read the
entire generated directory indiscriminately, and do not treat contracts as a
replacement for source code. After changing the code, regenerate the context
with symbolc sync before continuing navigation or review.
```

### Multiple agents

A useful division of responsibility is:

- an architect reads the root and executable indexes plus the relevant
  flux-entrypoint;
- an implementer reads the selected contracts and then the referenced source
  files;
- a reviewer regenerates the context after implementation and checks the
  changed symbols, relationships, and source locations.

Because the generated directory is reusable, agents that only need navigation
can consume the same artifacts without rerunning the analyzer or requiring
shell access.

## Determinism and synchronization

`sync` performs a full reconstruction of the managed output directory. It
removes stale generated artifacts and publishes the new directory atomically,
so a failed generation does not intentionally replace a previously valid
context with a partial one.

Repeated synchronization without relevant source or configuration changes is
expected to produce the same logical artifact set.

## Current scope

The first version focuses on structural navigation:

- Go packages and symbols;
- canonical symbol IDs;
- package imports and reachable package maps;
- methods associated with their receiver types;
- function and method signatures;
- struct fields;
- source locations;
- deterministic, per-executable YAML sessions.

It does not attempt to replace source reading, infer arbitrary behavior, or
build a complete call graph. Those capabilities should be added only when a
measured navigation task demonstrates that they are necessary.

## Development

Run the test suite with:

```sh
go test ./...
```

Run the materialization benchmarks with:

```sh
go test ./internal/output -run '^$' -bench 'Benchmark(Materialize|AnalyzeAndMaterialize)$' -benchmem
```
