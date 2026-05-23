# ffreis-platform-go-analyzers

Custom Go static analyzers for the ffreis platform fleet, each shipped as a
standalone binary plus an importable `*analysis.Analyzer` for use in
golangci-lint v2 custom-plugin configs.

## Analyzers

### `nakedgo`

Flags `go ...` statements whose function bodies do not start with `defer
recover()`. A panic in a goroutine without a recover is fatal to the entire
process — Go's runtime terminates the program. In long-running services
(orchestrators, scanners, server pools) a single bad input can therefore
kill the whole worker fleet.

Detection rules:

- `go func() { ... }()` is flagged unless the literal body begins with a
  deferred recover() — either inline `defer func() { recover() }()` or
  `defer someHelper()` whose callee name contains "recover"
  case-insensitively.
- `go someFn(args)` (named function, not a literal) gets an informational
  diagnostic noting that the analyzer cannot inspect the callee's body.

Skipped: `_test.go` files, anything under `vendor/`, files whose head
comments contain `DO NOT EDIT` (the standard generated-code convention).

## Build and run

```bash
make build              # → ./bin/nakedgo
./bin/nakedgo ./...     # in any Go module
./bin/nakedgo ./internal/runner/...
```

Exit code is non-zero when at least one diagnostic is reported.

## Use from golangci-lint v2 (planned)

golangci-lint v2 supports custom analyzers via its `linters.settings.custom`
mechanism with `path:` pointing at a `.so` plugin build. A
`make plugin` target will be added once an integrator needs it; meanwhile
the standalone binary is invokable from a Makefile or pre-commit hook:

```yaml
# lefthook.yml
pre-commit:
  commands:
    nakedgo:
      run: |
        go run github.com/FelipeFuhr/ffreis-platform-go-analyzers/cmd/nakedgo ./...
```

## Tests

```bash
make test
```

Tests use `golang.org/x/tools/go/analysis/analysistest` against the fixture
package at `analyzers/nakedgo/testdata/src/a/`. Each `go ...` site is
annotated with the expected diagnostic via `// want "<regex>"` comments.
