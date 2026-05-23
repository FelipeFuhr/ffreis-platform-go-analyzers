// Command nakedgo runs the nakedgo analyzer as a standalone tool.
//
// Usage:
//
//	nakedgo ./...
//	nakedgo ./internal/...
//
// Exits non-zero when at least one diagnostic is reported.
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/FelipeFuhr/ffreis-platform-go-analyzers/analyzers/nakedgo"
)

func main() {
	singlechecker.Main(nakedgo.Analyzer)
}
