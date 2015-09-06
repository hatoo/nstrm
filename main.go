package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"sync"

	"./builtins"
	"./parser"
	"./vm"

	//	"github.com/k0kubun/pp"
)

func main() {
	e := flag.String("e", "", "evaluate line")
	v := flag.Bool("v", false, "print version")
	d := flag.Bool("d", false, "debug mode")
	numprocs := flag.Int("p", 0, "number of processes")

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	if *numprocs != 0 {
		runtime.GOMAXPROCS(*numprocs)
	}

	if *v {
		fmt.Println("0.0.0")
		return
	}

	expression := ""
	if *e != "" {
		expression = *e
	} else {
		fname := flag.Args()
		if buffer, err := ioutil.ReadFile(fname[0]); err == nil {
			expression = string(buffer)
		} else {
			log.Fatal(err)
			return
		}
	}
	p := &parser.Nstrm{Buffer: expression}
	p.Init()
	p.MyParser.Init()
	if err := p.Parse(); err != nil {
		fmt.Println(err)
		return
	}
	p.Execute()

	if !(*d) {
		log.SetOutput(ioutil.Discard)
	}

	var wg sync.WaitGroup
	env := vm.NewEnv(&wg)
	builtins.LoadCore(env)
	builtins.LoadNet(env)

	if _, err := p.Run(env); err == nil {
		env.Decref()
		env.RunWait(vm.NIL)
		wg.Wait()
	} else {
		switch E := err.(type) {
		case *vm.Error:
			E.Fatal(p.Buffer)
		}
	}

}
