package parser

import (
	"reflect"

	"../ast"
	"../vm"
)

type scope struct {
	Parent         *scope
	Stack          []ast.Expr
	Identifer      string
	FormalArgments []string
	IfCond         []ast.Expr
	IfTrue         []ast.Expr
	IfElse         []ast.Expr
	WhileCond      []ast.Expr
	FirstFilter    bool
	LastFilter     bool
	Pipe           *ast.Pipe
}

//MyParser is parser for this language
type MyParser struct {
	Current *scope
}

//Init initializes parser
func (p *MyParser) Init() {
	p.Current = newScope(nil)
}

func newScope(parent *scope) *scope {
	return &scope{
		Parent: parent,
		Stack:  []ast.Expr{},
	}
}

func (p *MyParser) pushScope() {
	p.Current = newScope(p.Current)
}

func (p *MyParser) prepare(id string) {
	p.pushScope()
	p.Current.Identifer = id
}

func (p *MyParser) popScope(expr ast.Expr) {
	p.Current = p.Current.Parent
	p.addExpr(expr)
}

func (p *MyParser) addNumber(str string, begin int, end int) {
	n, _ := vm.SscanNumber(str)
	ex := &ast.Literal{Value: reflect.ValueOf(n)}
	ex.SetPosition(ast.Position{Begin: begin, End: end})
	p.Current.Stack = append(p.Current.Stack, ex)
}

func (p *MyParser) addExpr(expr ast.Expr) {
	p.Current.Stack = append(p.Current.Stack, expr)
}

func (p *MyParser) addArgment(id string) {
	p.Current.FormalArgments = append(p.Current.FormalArgments, id)
}

func (p *MyParser) literal(lit interface{}, begin int, end int) {
	ex := ast.Literal{Value: reflect.ValueOf(lit)}
	ex.SetPosition(ast.Position{Begin: begin, End: end})
	p.Current.Stack = append(p.Current.Stack, &ex)
}

func (p *MyParser) refVar(id string, begin int, end int) {
	ex := ast.RefVar{Identifer: id}
	ex.SetPosition(ast.Position{Begin: begin, End: end})
	p.Current.Stack = append(p.Current.Stack, &ex)
}

func (p *MyParser) funcall(begin int, end int) {
	ex := ast.Funcall{Identifer: p.Current.Identifer, Args: p.Current.Stack}
	ex.SetPosition(ast.Position{Begin: begin, End: end})
	p.popScope(&ex)
}

func (p *MyParser) block() {
	ex := ast.Block{
		FormalArgments: p.Current.FormalArgments,
		Body:           p.Current.Stack,
	}
	p.popScope(&ex)
}
func (p *MyParser) ifexpr() {
	ex := ast.If{
		Cond: p.Current.IfCond,
		True: p.Current.IfTrue,
		Else: p.Current.IfElse,
	}
	if ex.Else == nil {
		ex.Else = []ast.Expr{}
	}
	p.popScope(&ex)
}

func (p *MyParser) whileexpr() {
	ex := ast.While{
		Cond: p.Current.WhileCond,
		Body: p.Current.Stack,
	}
	p.popScope(&ex)
}

func (p *MyParser) ifCond() {
	p.Current.IfCond = p.Current.Stack
	p.Current.Stack = []ast.Expr{}
}

func (p *MyParser) whileCond() {
	p.Current.WhileCond = p.Current.Stack
	p.Current.Stack = []ast.Expr{}
}

func (p *MyParser) ifTrue() {
	p.Current.IfTrue = p.Current.Stack
	p.Current.Stack = []ast.Expr{}
}

func (p *MyParser) ifElse() {
	p.Current.IfElse = p.Current.Stack
	p.Current.Stack = []ast.Expr{}
}

func (p *MyParser) array() {
	ex := ast.Array{
		Elements: p.Current.Stack,
	}
	p.popScope(&ex)
}

func (p *MyParser) emit() {
	ex := ast.Emit{
		Elements: p.Current.Stack,
	}
	p.popScope(&ex)
}

func (p *MyParser) skip() {
	p.Current.Stack = append(p.Current.Stack, &ast.Skip{})
}

func (p *MyParser) bind() {
	p.popScope(&ast.BindVar{Identifer: p.Current.Identifer, Expr: p.Current.Stack[0]})
}

func (p *MyParser) addOp2(id string, begin int, end int) {
	s := p.Current.Stack
	p.Current.Stack = make([]ast.Expr, len(s)-1)
	copy(p.Current.Stack, s[0:len(s)-2])
	ex := ast.Funcall{
		Identifer: id,
		Args:      s[len(s)-2:],
	}
	ex.SetPosition(ast.Position{Begin: ex.Args[0].GetPosition().Begin, End: ex.Args[1].GetPosition().End})
	p.Current.Stack[len(s)-2] = &ex
}

func (p *MyParser) pipeStart(begin int, end int) {
	p.Current.Pipe = &ast.Pipe{
		Args:        []ast.Expr{},
		FirstFilter: p.Current.FirstFilter,
	}
	p.pipePush(begin, end)
}

func (p *MyParser) pipeEnd() {
	p.Current.Pipe.LastFilter = p.Current.LastFilter
	p.Current.Stack = append(p.Current.Stack, p.Current.Pipe)
	p.Current.Pipe = nil
	p.Current.FirstFilter = false
	p.Current.LastFilter = false
}

func (p *MyParser) pipePush(begin int, end int) {
	s := p.Current.Stack
	ex := s[len(s)-1]
	ex.SetPosition(ast.Position{Begin: begin, End: end})
	p.Current.Pipe.Args = append(p.Current.Pipe.Args, ex)
	p.Current.Stack = make([]ast.Expr, len(s)-1)
	copy(p.Current.Stack, s[0:len(s)-1])

}

func (p *MyParser) wait() {
	p.Current.Stack = append(p.Current.Stack, &ast.Wait{})
}

func (p *MyParser) close() {
	//p.Current.Stack = append(p.Current.Stack, &ast.Close{})
	p.popScope(&ast.Close{Ret: p.Current.Stack})
}

//Run run parsed ast
func (p *MyParser) Run(env *vm.Env) (reflect.Value, vm.SpecialValue) {
	var ret reflect.Value
	var err vm.SpecialValue
	for _, expr := range p.Current.Stack {
		if ret, err = vm.Run(expr, env); err != nil {
			return ret, err
		}
	}
	return ret, err
}
