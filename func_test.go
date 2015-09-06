package main

import "testing"

func TestBlock(t *testing.T) {
	assertNum("a={x->x};a(10)", "10", t)
}
