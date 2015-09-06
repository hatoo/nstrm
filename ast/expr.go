package ast

import "reflect"

// Expr represents expression. Note: there is no Statement.
type Expr interface {
	Pos
	expr()
}

//ExprImpl provides implementation for Expr
type ExprImpl struct {
	PosImpl
}

func (*ExprImpl) expr() {}

//Literal is hardcoded value. eg Number String
type Literal struct {
	ExprImpl
	Value reflect.Value
}

//RefVar is reference of variable
type RefVar struct {
	ExprImpl
	Identifer string
}

//BindVar is Variable Binding. eg a=1
type BindVar struct {
	ExprImpl
	Identifer string
	Expr      Expr
}

//Funcall is function call
type Funcall struct {
	ExprImpl
	Identifer string
	Args      []Expr
}

//Pipe is pipe expression. eg STDOUT | STDIN
type Pipe struct {
	ExprImpl
	Args        []Expr
	FirstFilter bool
	LastFilter  bool
}

//Block is block. eg {-> }
type Block struct {
	ExprImpl
	FormalArgments []string
	Body           []Expr
}

//If is if expression
type If struct {
	ExprImpl
	Cond []Expr
	True []Expr
	Else []Expr
}

//While is while expression
type While struct {
	ExprImpl
	Cond []Expr
	Body []Expr
}

// Array is array.
type Array struct {
	ExprImpl
	Elements []Expr
}

// Wait is wait expression
type Wait struct {
	ExprImpl
}

// Close is close expression
type Close struct {
	ExprImpl
	Ret []Expr
}

// Emit is emit expression
type Emit struct {
	ExprImpl
	Elements []Expr
}

//Skip is skip expression
type Skip struct {
	ExprImpl
}
