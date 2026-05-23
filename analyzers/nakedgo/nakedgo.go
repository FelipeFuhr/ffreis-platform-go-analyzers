// Package nakedgo provides a Go static analyzer that flags `go ...`
// statements whose function bodies do not start with `defer recover()`.
//
// Rationale: a panic in a goroutine that lacks recover() is fatal to the
// entire process — Go's runtime terminates the program. In long-running
// services (orchestrators, scanners, server pools) a single bad input can
// kill the whole worker fleet. Convention: every goroutine that runs
// untrusted or third-party code must defer a recover() at its entry.
//
// Detection rules:
//
//   - `go func() { ... }()` is flagged unless the literal body begins with a
//     deferred recover() statement (either inline `defer func() {
//     recover() }()` or `defer someRecoverHelper(...)` whose callee name
//     contains "recover" case-insensitively).
//
//   - `go someFn(args)` (call to a named function, not a literal) is
//     flagged with an informational message noting that the analyzer
//     cannot verify the called function's recovery.
//
// Skipped files: _test.go, anything under vendor/, and files whose first
// comment block contains "DO NOT EDIT" (the generated-code convention).
package nakedgo

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer is the singleton instance loaded by the cmd/nakedgo driver and
// by golangci-lint's custom-analyzer plugin path.
var Analyzer = &analysis.Analyzer{
	Name:     "nakedgo",
	Doc:      "flags goroutines lacking a deferred recover() at their entry",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Build the set of file positions to skip up front (test, vendor,
	// generated). We key by token.Pos of the file's first byte so the
	// per-statement check stays cheap.
	skippedFile := make(map[string]bool, len(pass.Files))
	for _, f := range pass.Files {
		if shouldSkipFile(pass, f) {
			skippedFile[pass.Fset.Position(f.Pos()).Filename] = true
		}
	}

	insp.Preorder([]ast.Node{(*ast.GoStmt)(nil)}, func(n ast.Node) {
		gs := n.(*ast.GoStmt)
		fname := pass.Fset.Position(gs.Pos()).Filename
		if skippedFile[fname] {
			return
		}

		switch gs.Call.Fun.(type) {
		case *ast.FuncLit:
			lit := gs.Call.Fun.(*ast.FuncLit)
			if !startsWithDeferredRecover(lit.Body) {
				pass.Report(analysis.Diagnostic{
					Pos:     gs.Pos(),
					Message: "goroutine literal does not begin with `defer recover()` — a panic here will crash the process",
				})
			}
		case *ast.SelectorExpr, *ast.Ident:
			pass.Report(analysis.Diagnostic{
				Pos:     gs.Pos(),
				Message: "goroutine calls a named function; ensure that function defers a recover() at its entry (analyzer cannot verify)",
			})
		}
	})

	return nil, nil
}

// startsWithDeferredRecover returns true iff the first statement in body is
// a DeferStmt whose called function either calls recover() inline or whose
// callee name contains "recover" (case-insensitive). Either form is a
// strong indication of intentional panic handling.
func startsWithDeferredRecover(body *ast.BlockStmt) bool {
	if body == nil || len(body.List) == 0 {
		return false
	}
	first, ok := body.List[0].(*ast.DeferStmt)
	if !ok {
		return false
	}
	switch fn := first.Call.Fun.(type) {
	case *ast.FuncLit:
		return containsRecoverCall(fn.Body)
	case *ast.Ident:
		return strings.Contains(strings.ToLower(fn.Name), "recover")
	case *ast.SelectorExpr:
		return strings.Contains(strings.ToLower(fn.Sel.Name), "recover")
	}
	return false
}

// containsRecoverCall reports whether the block contains a call to a
// function named "recover". Used inside the inline-defer form.
func containsRecoverCall(block *ast.BlockStmt) bool {
	if block == nil {
		return false
	}
	found := false
	ast.Inspect(block, func(n ast.Node) bool {
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		id, ok := ce.Fun.(*ast.Ident)
		if !ok {
			return true
		}
		if id.Name == "recover" {
			found = true
			return false
		}
		return true
	})
	return found
}

// shouldSkipFile reports whether f should be excluded from analysis.
func shouldSkipFile(pass *analysis.Pass, f *ast.File) bool {
	fname := pass.Fset.Position(f.Pos()).Filename
	if strings.HasSuffix(fname, "_test.go") {
		return true
	}
	if strings.Contains(fname, "/vendor/") {
		return true
	}
	// Generated-code convention: a comment containing "DO NOT EDIT" anywhere
	// near the file head.
	for _, cg := range f.Comments {
		// Only check the top of the file (before package decl) — that's the
		// canonical location for generator markers.
		if cg.Pos() >= f.Package {
			break
		}
		for _, c := range cg.List {
			if strings.Contains(c.Text, "DO NOT EDIT") {
				return true
			}
		}
	}
	return false
}
