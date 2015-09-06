package vm

import (
	"fmt"
	"log"
	"reflect"

	"../ast"
	"../gc"
	"../pipe"
)

var NIL = reflect.ValueOf(nil)

func isInt(v reflect.Value) bool {
	return v.Kind() == reflect.Int || v.Kind() == reflect.Int64
}

func isFloat(v reflect.Value) bool {
	return v.Kind() == reflect.Float32 || v.Kind() == reflect.Float64
}

func getInt(v reflect.Value) (int64, error) {
	if isInt(v) {
		return int64(v.Int()), nil
	} else {
		return 0, fmt.Errorf("it's not Int %s", v.Kind().String())
	}
}

func getFloat(v reflect.Value) (float64, error) {
	if isFloat(v) {
		return v.Float(), nil
	} else if isInt(v) {
		i, e := getInt(v)
		return float64(i), e
	} else {
		return 0.0, fmt.Errorf("it's not Float. it is %s", v.Kind().String())
	}
}

func getString(v reflect.Value) (string, error) {
	if v.Kind() == reflect.String {
		return v.String(), nil
	} else {
		return "", fmt.Errorf("it's not string. it is %s", v.Kind().String())
	}
}

func RunList(exprs []ast.Expr, env *Env) (reflect.Value, SpecialValue) {
	ret := reflect.ValueOf(nil)
	var err SpecialValue = nil
	for _, expr := range exprs {
		if ret, err = Run(expr, env); err != nil {
			return ret, err
		}
	}
	return ret, err
}

func Run(expr ast.Expr, env *Env) (reflect.Value, SpecialValue) {
	switch E := expr.(type) {
	case *ast.Literal:
		return E.Value, nil
	case *ast.BindVar:
		if value, err := Run(E.Expr, env); err == nil {
			env.Define(E.Identifer, value)
			return value, nil
		} else {
			return value, err
		}
	case *ast.RefVar:
		if value, ok := env.Lookup(E.Identifer); ok {
			gc.Incif(value)
			env.DecrefLaterV(value)
			return value, nil
		} else {
			return NIL, Errorf(E, "%s is undefined", E.Identifer)
		}
	case *ast.Funcall:
		if fbody, ok := env.Lookup(E.Identifer); ok {
			gc.Incif(fbody)
			env.DecrefLaterV(fbody)
			args := []reflect.Value{}
			for _, e := range E.Args {
				if arg, err := Run(e, env); err == nil {
					args = append(args, Eval(arg))
				} else {
					return arg, err
				}
			}
			switch fun := fbody.Interface().(type) {
			case Function:
				ret, err := fun.Call(E, args, pipe.NilValve())
				env.DecrefLaterV(ret)
				return ret, err
			default:
				return NIL, Errorf(E, "%s is not Function", E.Identifer)
			}
		} else {
			return NIL, Errorf(E, "%s is undefined", E.Identifer)
		}
	case *ast.Pipe:
		return RunPipeExpr(E, env)
	case *ast.Wait:
		env.Wait()
		return NIL, nil
	case *ast.Block:
		ret := NewUserFunction(E.FormalArgments, E.Body, env.ChildEnv())
		env.DecrefLater(ret)
		return reflect.ValueOf(ret), nil
	case *ast.If:
		log.Println("ifcond run <<<")
		cond, err := RunList(E.Cond, env)
		log.Println("ifcond run >>>")
		if err != nil {
			return cond, err
		}
		log.Println("ifcond eval <<<")
		b := Condition(cond)
		log.Println("ifcond eval >>>")
		if b {
			return RunList(E.True, env)
		} else {
			return RunList(E.Else, env)
		}
	case *ast.While:
		ret := NIL
		cap := env.ChildEnv()
		defer func() {
			cap.Decref()
		}()
		for {
			child := cap.ChildEnv()
			if cond, err := RunList(E.Cond, cap); err == nil {
				b := Condition(cond)
				if b {
					if v, err := RunList(E.Body, child); err == nil {
						gc.Decif(ret)
						ret = v
						child.Run(v)
						child.Decref()
					} else {
						child.Run(v)
						child.Decref()
						return v, err
					}
				} else {
					child.Decref()
					child.Run(NIL)
					return ret, nil
				}
			} else {
				return cond, err
			}
		}
		return ret, nil
	case *ast.Array:
		arr := []reflect.Value{}
		for _, el := range E.Elements {
			if ret, err := Run(el, env); err == nil {
				arr = append(arr, ret)
			} else {
				return ret, err
			}
		}
		return reflect.ValueOf(arr), nil
	case *ast.Emit:
		for _, el := range E.Elements {
			if ret, err := Run(el, env); err == nil {
				log.Println("Emit", ret, err)

				if !env.Send(ret) {
					return NIL, &Close{}
				}
			} else {
				return ret, err
			}
		}
		return NIL, nil
	case *ast.Skip:
		return NIL, &Skip{}
	case *ast.Close:
		log.Println("Close")
		if len(E.Ret) == 0 {
			return NIL, &Close{}
		} else {
			if ret, err := Run(E.Ret[0], env); err == nil {
				return ret, &Close{}
			} else {
				return ret, err
			}
		}
	default:
		return NIL, Errorf(expr, "unimplemented Expr")
	}
}
