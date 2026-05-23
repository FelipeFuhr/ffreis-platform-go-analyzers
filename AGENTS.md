# Agent Context

**This repo:** `ffreis-platform-go-analyzers` — custom Go static analyzers
for the ffreis fleet, packaged as standalone binaries plus importable
`*analysis.Analyzer` values.

## Non-obvious facts

- **Each analyzer is independent.** Adding a new analyzer means a new
  package under `analyzers/` + a new `cmd/<name>/main.go` driver.
- **Tests use `golang.org/x/tools/go/analysis/analysistest`** with
  fixtures at `analyzers/<name>/testdata/src/<pkg>/`. Use `// want
  "<regex>"` comments inline with the offending construct; the regex is
  matched as a substring of the diagnostic message.
- **`inspector.Preorder` is the right traversal**, not `WithStack`.
  Empirically, `WithStack` interacts poorly with analysistest's
  want-comment matching (diagnostics get attributed to wrong positions).
- **Skip rules in `nakedgo`** are file-level by filename: `_test.go`,
  any path under `vendor/`, files with "DO NOT EDIT" in their first
  comment block. Per-position skips are not implemented.

## Structure

```
analyzers/
  nakedgo/             ← go-stmt panic-recovery analyzer
    nakedgo.go
    nakedgo_test.go
    testdata/src/a/    ← analysistest fixture
cmd/
  nakedgo/             ← standalone CLI wrapping the analyzer
go.mod
Makefile
```

## Build/run

```bash
make build              # → ./bin/nakedgo
./bin/nakedgo ./...     # run against the current module
```

Exit code is non-zero when at least one diagnostic is reported.

## Tests

```bash
make test     # runs with -race -shuffle=on per the workspace invariant
```

## Public repo — private-repo hygiene

If this becomes a **public** GitHub repository, follow the standard
convention: never name private repos in commit messages, PR titles, or
descriptions. Use generic terms ("the orchestrator", "a private
consumer", "internal infra").

## Keeping this file current

- If you add a new analyzer, document its detection rules and skip
  rules here under "Analyzers".
- If you change the traversal approach (e.g. swap Preorder for
  WithStack), update the "Non-obvious facts" section so the next agent
  doesn't trip on the same analysistest gotcha.
