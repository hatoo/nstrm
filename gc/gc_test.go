package gc

import (
	"testing"
)

type A struct {
	a int
	Ref
}

func TestSimpleRefCount(t *testing.T) {
	a := A{a: 10}
	a.Incref()
	a.Decref()
	a.Wait()
}

func TestSimpleRefCountParallel(t *testing.T) {
	a := A{a: 10}
	a.Incref()
	go a.Decref()
	a.Wait()
}
