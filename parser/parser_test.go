package parser

import (
	"testing"
)

func parse(text string, t *testing.T) *MyParser {
	p := &Nstrm{Buffer: text}
	p.Init()
	p.MyParser.Init()
	if err := p.Parse(); err != nil {
		t.Fatal(err)
	}
	p.Execute()
	return &p.MyParser
}

func bparse(text string, t *testing.B) *MyParser {
	p := &Nstrm{Buffer: text}
	p.Init()
	p.MyParser.Init()
	if err := p.Parse(); err != nil {
		t.Fatal(err)
	}
	p.Execute()
	return &p.MyParser
}

func Test_Parse_Int(t *testing.T) {
	parse("1+1", t)
}

func Test_BlockBlock(t *testing.T) {
	expr := `
	{x->
		{
			y->
		}
	}
	`
	parse(expr, t)
}

func Test_Fold(t *testing.T) {
	expr := `
	fold = {f,init->
		ret = init
		{x->
			ret = f(ret,init)
		} | last()
	}
	`
	parse(expr, t)
}
