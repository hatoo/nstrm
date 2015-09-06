package main

import (
	"io/ioutil"
	"log"
	"reflect"
	"sync"
	"testing"

	"./builtins"
	"./parser"
	"./vm"
)

func assertNum(expr string, expected string, t *testing.T) {
	var wg sync.WaitGroup
	p := &parser.Nstrm{Buffer: expr}
	p.Init()
	p.MyParser.Init()
	if err := p.Parse(); err != nil {
		t.Errorf("Parser Error")
		return
	}
	p.Execute()
	log.SetOutput(ioutil.Discard)
	env := vm.NewEnv(&wg)
	builtins.LoadCore(env)
	n, _ := vm.SscanNumber(expected)
	if v, err := p.Run(env); err == nil {
		if c, e := vm.CmpV(v, reflect.ValueOf(n)); e != nil || c != 0 {
			t.Errorf("unexpected return got %v expected %v", v.Interface(), expected)
		}
	} else {
		t.Fatal(err)
	}
}

// test add

func TestADDii(t *testing.T) {
	assertNum("1+1", "2", t)
}

func TestADDif(t *testing.T) {
	assertNum("1+1.0", "2.0", t)
}

func TestADDfi(t *testing.T) {
	assertNum("1.0+1", "2.0", t)
}

func TestADDff(t *testing.T) {
	assertNum("1.0+1.0", "2.0", t)
}

// test sub

func TestSUBii(t *testing.T) {
	assertNum("1-1", "0", t)
}

func TestSUBif(t *testing.T) {
	assertNum("1-1.0", "0.0", t)
}

func TestSUBfi(t *testing.T) {
	assertNum("1.0-1", "0.0", t)
}

func TestSUBff(t *testing.T) {
	assertNum("1.0-1.0", "0.0", t)
}

// test mul

func TestMULii(t *testing.T) {
	assertNum("1*1", "1", t)
}

func TestMULif(t *testing.T) {
	assertNum("1*1.0", "1.0", t)
}

func TestMULfi(t *testing.T) {
	assertNum("1.0*1", "1.0", t)
}

func TestMULff(t *testing.T) {
	assertNum("1.0*1.0", "1.0", t)
}

// test div

func TestDIVii(t *testing.T) {
	assertNum("1/1", "1", t)
}

func TestDIVif(t *testing.T) {
	assertNum("1/1.0", "1.0", t)
}

func TestDIVfi(t *testing.T) {
	assertNum("1.0/1", "1.0", t)
}

func TestDIVff(t *testing.T) {
	assertNum("1.0/1.0", "1.0", t)
}

//

func TestMath1(t *testing.T) {
	assertNum("1+2+3-3", "3", t)
}

func TestMath2(t *testing.T) {
	assertNum("(1+2+3)*(4+5+6)", "90", t)
}

func TestRecursive(t *testing.T) {
	assertNum("ADD(ADD(1,2),SUB(3,2))", "4", t)
}

func TestBind1(t *testing.T) {
	assertNum("a=10;a+a", "20", t)
}

func TestMultipleBind(t *testing.T) {
	assertNum("a=b=10;a+b", "20", t)
}
