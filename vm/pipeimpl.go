package vm

import (
	"log"
	"reflect"

	//	"github.com/k0kubun/pp"

	"../ast"
	"../pipe"
)

func producerFunction(p ast.Pos, f Function) pipe.Producer {
	out := pipe.NewValve()
	f.Incref()
	go func() {
		defer func() {
			f.Decref()
			out.Close()
		}()
		for {
			if ret, err := f.Call(p, []reflect.Value{}, out); err == nil {
				if pipe.IsEOF(ret) {
				} else {
					if !out.Send(ret) {
						return
					}
				}
			} else {
				switch E := err.(type) {
				case *Skip:
				case *Close:
					if pipe.IsEOF(ret) {
					} else {
						if !out.Send(ret) {
							return
						}
					}
					return
				case *Error:
					panic(E.Message)
				default:
					panic("unimplemented")
				}
			}
		}
	}()
	ret := pipe.NewProducer(out)
	return ret
}

func filterFunction(p ast.Pos, f Function) pipe.Filter {
	f.Incref()
	fun := func(read <-chan reflect.Value, write pipe.Valve) {
		defer func() {
			f.Decref()
			//write.Close()
		}()
		for value := range read {
			if ret, err := f.Call(p, []reflect.Value{Eval(value)}, write); err == nil {
				if pipe.IsEOF(ret) {
				} else {
					if !write.Send(ret) {
						return
					}
				}
			} else {
				log.Println(err)
				switch E := err.(type) {
				case *Skip:
				case *Close:
					if pipe.IsEOF(ret) {
					} else {
						write.Send(ret)
					}
					return
				case *Error:
					log.Println("err filterfunction")
					panic(E.Message)
				default:
					panic("unimplemented")
				}
			}
		}
	}
	return pipe.NewFilter(fun)
}

func consumerFunction(p ast.Pos, f Function) pipe.Consumer {
	f.Incref()
	fun := func(r <-chan reflect.Value) reflect.Value {
		defer func() {
			f.Decref()
		}()
		for value := range r {
			if ret, err := f.Call(p, []reflect.Value{Eval(value)}, pipe.NilValve()); err != nil {
				switch E := err.(type) {
				case *Skip:
				case *Close:
					return ret
				case *Error:
					panic(E.Message)
				default:
					panic("unimplemented")
				}
			}
		}
		return NIL
	}
	return pipe.NewConsumer(fun)
}

func connect(l pipe.Pipe, r pipe.Pipe, env *Env) (pipe.Pipe, bool) {
	switch lp := l.(type) {
	case pipe.Filter:
		switch rp := r.(type) {
		case pipe.Filter:
			return pipe.ConnectFF(lp, rp), true
		case pipe.Consumer:
			return pipe.ConnectFC(lp, rp), true
		}
	case pipe.Producer:
		switch rp := r.(type) {
		case pipe.Filter:
			return pipe.ConnectPF(lp, rp), true
		case pipe.Consumer:
			return pipe.ConnectPC(lp, rp), true
		}
	}
	return nil, false
}

func asProducer(pos ast.Pos, v reflect.Value, env *Env) (pipe.Producer, bool) {
	if v.IsValid() {
		switch t := v.Interface().(type) {
		case pipe.Producer:
			return t, true
		case pipe.Terminal:
			valve := pipe.NewValve()
			env.RunLater(t)
			go func() {
				result := t.Result()
				valve.Send(result)
				valve.Close()
			}()
			ret := pipe.NewProducer(valve)
			env.DecrefLater(ret)
			return ret, true
		case Function:
			ret := producerFunction(pos, t)
			env.DecrefLater(ret)
			return ret, true
		case []reflect.Value:
			log.Println(t)
			valve := pipe.NewValve()
			go func() {
				defer valve.Close()
				for _, v := range t {
					if !valve.Send(v) {
						return
					}
				}
			}()
			ret := pipe.NewProducer(valve)
			env.DecrefLater(ret)
			return ret, true
		default:
			return nil, false
		}
	} else {
		return nil, false
	}
}

func asFilter(pos ast.Pos, v reflect.Value, env *Env) (pipe.Filter, bool) {
	if v.IsValid() {
		switch t := v.Interface().(type) {
		case pipe.Filter:
			return t, true
		case Function:
			ret := filterFunction(pos, t)
			env.DecrefLater(ret)
			return ret, true
		case pipe.Consumer:
			log.Println("convert consumer to filter")
			env.RunLater(t)
			cinput := t.NewR()
			filter := func(r <-chan reflect.Value, w pipe.Valve) {
				defer w.Close()
				done := false
				for v := range r {
					if pipe.IsEOF(v) {
						panic("")
					}
					if !cinput.Send(v) {
						done = true
						break
					}
				}
				if !done {
					cinput.Send(pipe.EOF)
					log.Println("converted consumer sent EOF")
				}
				w.Send(t.Result())
			}
			ret := pipe.NewFilter(filter)
			env.DecrefLater(ret)
			return ret, true
		default:
			return nil, false
		}
	} else {
		return nil, false
	}
}

func asConsumer(pos ast.Pos, v reflect.Value, env *Env) (pipe.Consumer, bool) {
	if v.IsValid() {
		switch t := v.Interface().(type) {
		case pipe.Consumer:
			return t, true
		case Function:
			ret := consumerFunction(pos, t)
			env.DecrefLater(ret)
			return ret, true
		default:
			return nil, false
		}
	} else {
		return nil, false
	}
}

func RunPipeExpr(expr *ast.Pipe, env *Env) (reflect.Value, SpecialValue) {
	args := []reflect.Value{}
	for _, ex := range expr.Args {
		if ret, err := Run(ex, env); err == nil {
			args = append(args, ret)
		} else {
			return ret, err
		}
	}
	if expr.FirstFilter {
		if f, ok := asFilter(expr.Args[0], args[0], env); ok {
			for i := 1; i < len(args)-1; i++ {
				if r, ok := asFilter(expr.Args[i], args[i], env); ok {
					f = pipe.ConnectFF(f, r)
					env.DecrefLater(f)
				} else {
					return NIL, Errorf(expr, "not filter")
				}
			}
			if expr.LastFilter {
				if r, ok := asFilter(expr.Args[len(args)-1], args[len(args)-1], env); ok {
					ret := pipe.ConnectFF(f, r)
					env.DecrefLater(ret)
					//env.RunLater(ret)
					return reflect.ValueOf(ret), nil
				}
				return NIL, Errorf(expr, "not filter")
			} else {
				if r, ok := asConsumer(expr.Args[len(args)-1], args[len(args)-1], env); ok {
					ret := pipe.ConnectFC(f, r)
					env.DecrefLater(ret)
					//env.RunLater(ret)
					return reflect.ValueOf(ret), nil
				}
				return NIL, Errorf(expr, "not consumer %v", args[len(args)-1])
			}
		} else {
			return NIL, Errorf(expr, "not filter")
		}
	} else {
		if p, ok := asProducer(expr.Args[0], args[0], env); ok {
			for i := 1; i < len(args)-1; i++ {
				if r, ok := asFilter(expr.Args[i], args[i], env); ok {
					p = pipe.ConnectPF(p, r)
					env.DecrefLater(p)
				} else {
					return NIL, Errorf(expr, "not filter")
				}
			}
			if expr.LastFilter {
				if r, ok := asFilter(expr.Args[len(args)-1], args[len(args)-1], env); ok {
					ret := pipe.ConnectPF(p, r)
					env.DecrefLater(ret)
					//env.RunLater(ret)
					return reflect.ValueOf(ret), nil
				}
				return NIL, Errorf(expr, "not filter")
			} else {
				if r, ok := asConsumer(expr.Args[len(args)-1], args[len(args)-1], env); ok {
					ret := pipe.ConnectPC(p, r)
					env.DecrefLater(ret)
					env.RunLater(ret)
					return reflect.ValueOf(ret), nil
				}
				return NIL, Errorf(expr, "not consumer", args[len(args)-1])
			}
		} else {
			return NIL, Errorf(expr, "not producer")
		}
	}
	return NIL, Errorf(expr, "pipeimpl error")
}
