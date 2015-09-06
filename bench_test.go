package main

import (
	"sync"
	"testing"

	"io/ioutil"
	"log"

	"./builtins"
	"./parser"
	"./vm"
	//"github.com/mattn/anko/vm"
)

func bParse(expr string, t *testing.B) {
	p := &parser.Nstrm{Buffer: expr}
	p.Init()
	p.MyParser.Init()
	if err := p.Parse(); err != nil {
		t.Errorf("Parser Error")
		return
	}
	p.Execute()
}

func bRun(expr string, t *testing.B) {
	var wg sync.WaitGroup
	p := &parser.Nstrm{Buffer: expr}
	p.Init()
	p.MyParser.Init()
	if err := p.Parse(); err != nil {
		t.Errorf("Parser Error")
		return
	}
	p.Execute()
	env := vm.NewEnv(&wg)
	builtins.LoadCore(env)
	log.SetOutput(ioutil.Discard)
	if r, err := p.Run(env); err == nil {
		env.Decref()
		env.RunWait(r)
		wg.Wait()
	} else {
		t.Fatal(err)
	}
}

func BenchmarkEmit(b *testing.B) {
	//prog := "seq(10) | { x-> if x%2==0 { emit x,x }else{ emit x } } | STDOUT"
	prog := "if x%2==0 { emit x,x}"
	bParse(prog, b)
}

func BenchmarkSum(b *testing.B) {
	prog := `
	fold = { f,init ->
	  ret = init
	  | {x->
	    ret = f(ret,x)
	  } | last()
	}

	sum = {->
	  fold(ADD,0)
	}
	seq(1000) | sum() | STDOUT
	`
	bRun(prog, b)
}

func BenchmarkFib(b *testing.B) {
	prog := `
	takewhile = { f ->
	  { x ->
	    if f(x) {
	      x
	    }else{
	      close
	    }
	  }
	}

	fib = { ->
	  a = 0
	  b = 1
	  {->
	    c = b
	    b = a+b
	    a = c
	  }
	}

	fold = { f,init ->
	  ret = init
	  | {x->
	    ret = f(ret,x)
	  } | last()
	}

	sum = {->
	  fold(ADD,0)
	}

	fib() | { x-> if x%2==0 {x} } | takewhile({x-> x <= 4000000}) | sum() | STDOUT
	`
	bRun(prog, b)
}

func BenchmarkSumOf10000(b *testing.B) {
	prog := `
	fold = { f,init ->
	  ret = init
	  | {x->
	    ret = f(ret,x)
	  } | last()
	}

	sum = {->
	  fold(ADD,0)
	}

	seq(10000) | sum() | STDOUT
	`
	bRun(prog, b)
}

func BenchmarkPrime(b *testing.B) {
	prog := `
	takewhile = { f ->
	  { x ->
	    if f(x) {
	      x
	    }else{
	      close
	    }
	  }
	}

	all = {f->
	  | f | {cond ->
	    emit cond
	    if cond==false {
	      close
	    }
	  } | last(true)
	}

	N = 1000

	primes = {->
	  ps = []
	  seq(2,N) | {n ->
	    if ps | takewhile({x->n>=x*x}) | all({x-> n%x!=0 } ) {
	      ps = append(ps,n)
	      n
	    }
	  } |
	}

	primes() | last() | STDOUT
	`
	bRun(prog, b)
}
