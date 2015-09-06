package vm

import (
	"reflect"
	"sync"

	"../ast"
	"../gc"
	"../pipe"
)

type SpecialValue interface {
	specialvalue()
}

type SpecialValueImpl struct{}

func (*SpecialValueImpl) specialvalue() {}

type Skip struct {
	SpecialValueImpl
}

type Close struct {
	SpecialValueImpl
}

type Error struct {
	SpecialValueImpl
	Pos     ast.Position
	Message string
}

func Eval(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return v
	}
	switch t := v.Interface().(type) {
	case pipe.Terminal:
		var wg sync.WaitGroup
		t.Run(&wg)
		t.NotifyExit()
		wg.Wait()
		return Eval(t.Result())
	default:
		return v
	}
}

type Function interface {
	Call(ast.Pos, []reflect.Value, pipe.Valve) (reflect.Value, SpecialValue)
	gc.GcThing
}

//User Defined Function
type UserFunction struct {
	FormalArgments []string
	Body           []ast.Expr
	Captured       *Env
	gc.Ref
	Gone bool
}

type BuiltinFunction struct {
	funbody func(...reflect.Value) (reflect.Value, error)
	gc.Ref
	Gone bool
}

func NewBuiltinFunction(f func(...reflect.Value) (reflect.Value, error)) Function {
	ret := &BuiltinFunction{funbody: f, Gone: false}
	ret.Incref()
	go func() {
		ret.Wait()
		ret.Gone = true
	}()
	return ret
}

func NewUserFunction(fargs []string, body []ast.Expr, captured *Env) Function {
	u := &UserFunction{
		FormalArgments: fargs,
		Body:           body,
		Captured:       captured,
		Gone:           false,
	}
	u.Incref()
	go func() {
		u.Wait()
		u.Gone = true
		captured.Decref()
	}()
	return u
}

func (f *BuiltinFunction) Call(context ast.Pos, args []reflect.Value, out pipe.Valve) (reflect.Value, SpecialValue) {
	if f.Gone {
		//panic("called released builtinfunction")
		return NIL, Errorf(context, "called released builtinfunction %s", f)
	}
	for _, v := range args {
		gc.Incif(v)
	}
	defer func() {
		for _, v := range args {
			gc.Decif(v)
		}
	}()
	ret, err := f.funbody(args...)
	if err != nil {
		return ret, Errorf(context, err.Error())
	} else {
		return ret, nil
	}
}

func (this *UserFunction) Call(context ast.Pos, args []reflect.Value, out pipe.Valve) (reflect.Value, SpecialValue) {
	if this.Gone {
		return NIL, Errorf(context, "called released function %s", this)
	}
	if len(this.FormalArgments) != len(args) {
		return NIL, Errorf(context, "invalid number of argments")
	}
	env := this.Captured.ChildEnv()
	env.SetOut(out)

	var ret reflect.Value
	var err SpecialValue

	defer func() {
		gc.Incif(ret)
		env.Run(ret)
		env.Decref()
	}()

	for i, name := range this.FormalArgments {
		env.Define(name, args[i])
	}

	for _, expr := range this.Body {
		if ret, err = Run(expr, env); err != nil {
			return ret, err
		}
	}

	return ret, err
}
