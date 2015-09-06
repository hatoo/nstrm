package builtins

import (
	"log"
	"reflect"

	"../pipe"
	"../vm"
)

//LoadUtil defines utility function
func LoadUtil(env *vm.Env) {
	env.DefineBuiltin("seq", reflect.ValueOf(vm.NewBuiltinFunction(func(args ...reflect.Value) (reflect.Value, error) {
		valve := pipe.NewValve()
		var start int64
		var end int64
		if len(args) == 1 {
			start = 1
			end = args[0].Interface().(vm.Number).ToInt()
		} else if len(args) == 2 {
			start = args[0].Interface().(vm.Number).ToInt()
			end = args[1].Interface().(vm.Number).ToInt()
		}
		go func() {
			defer func() {
				log.Println("seq close")
				valve.Close()
			}()
			for i := start; i <= end; i++ {
				if !valve.Send(reflect.ValueOf(vm.NewInt(int64(i)))) {
					log.Println("seq Done")
					return
				}
			}
		}()
		return reflect.ValueOf(pipe.NewProducer(valve)), nil
	})))

	env.DefineBuiltin("chan", reflect.ValueOf(vm.NewBuiltinFunction(func(args ...reflect.Value) (reflect.Value, error) {
		return reflect.ValueOf(pipe.NewChan()), nil
	})))

	env.DefineBuiltin("last", reflect.ValueOf(vm.NewBuiltinFunction(func(args ...reflect.Value) (reflect.Value, error) {
		fun := func(r <-chan reflect.Value) reflect.Value {
			defer log.Println("last end")
			log.Println("last start")
			var ret reflect.Value
			if len(args) == 1 {
				ret = args[0]
			}
			for v := range r {
				ret = v
			}
			return ret
		}
		return reflect.ValueOf(pipe.NewConsumer(fun)), nil
	})))

	env.DefineBuiltin("collect", reflect.ValueOf(vm.NewBuiltinFunction(func(args ...reflect.Value) (reflect.Value, error) {
		fun := func(r <-chan reflect.Value) reflect.Value {
			ret := []reflect.Value{}
			for v := range r {
				ret = append(ret, v)
			}
			return reflect.ValueOf(ret)
		}
		return reflect.ValueOf(pipe.NewConsumer(fun)), nil
	})))
}
