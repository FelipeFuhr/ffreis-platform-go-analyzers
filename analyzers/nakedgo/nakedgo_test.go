package nakedgo_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/FelipeFuhr/ffreis-platform-go-analyzers/analyzers/nakedgo"
)

// TestNakedgo runs the analyzer against testdata/src/a, which contains both
// flagged and unflagged goroutine constructs annotated with `// want` lines.
// analysistest matches the comments against the diagnostics the analyzer
// reports.
func TestNakedgo(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), nakedgo.Analyzer, "a")
}
