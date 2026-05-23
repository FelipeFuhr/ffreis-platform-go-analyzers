// Test fixture for the nakedgo analyzer. Each `go ...` site is annotated
// with the expected diagnostic (or absence thereof) for analysistest.
package a

import "fmt"

func bareLiteral() {
	go func() { // want "goroutine literal does not begin with"
		fmt.Println("naked")
	}()
}

func bareNamedFunc() {
	go workerFn() // want "goroutine calls a named function"
}

func protectedInlineRecover() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				_ = r
			}
		}()
		fmt.Println("safe")
	}()
}

func protectedNamedRecover() {
	go func() {
		defer recoverFn()
		fmt.Println("safe via helper")
	}()
}

func protectedSelectorRecover() {
	r := &recoverer{}
	go func() {
		defer r.Recover()
		fmt.Println("safe via method")
	}()
}

func emptyLiteralIsFlagged() {
	go func() {}() // want "goroutine literal does not begin with"
}

func deferButNotRecoverIsFlagged() {
	go func() { // want "goroutine literal does not begin with"
		defer fmt.Println("cleanup but not recovery")
		fmt.Println("naked")
	}()
}

func workerFn() {}

func recoverFn() {
	if r := recover(); r != nil {
		_ = r
	}
}

type recoverer struct{}

func (r *recoverer) Recover() {
	if rec := recover(); rec != nil {
		_ = rec
	}
}
