package builtins

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"

	"../pipe"
	"../vm"
)

//LoadIO defines standard IO function
func LoadIO(env *vm.Env) {
	stdin := pipe.NewValve()

	go func() {
		defer stdin.Close()
		reader := bufio.NewReader(os.Stdin)
		for {
			r, _, err := reader.ReadLine()
			if err != nil {
				break
			}
			if !stdin.Send(reflect.ValueOf(string(r))) {
				break
			}
			/*
				select {
				case stdin.Out <- reflect.ValueOf(string(r)):
				case <-stdin.Done:
					break
				}
			*/
		}
	}()

	env.DefineBuiltin("STDIN", reflect.ValueOf(pipe.NewProducer(stdin)))

	env.DefineBuiltin("STDOUT", reflect.ValueOf(vm.NewBuiltinFunction(func(args ...reflect.Value) (reflect.Value, error) {
		for _, v := range args {
			fmt.Println(v.Interface())
		}
		return vm.NIL, nil
	})))

	env.DefineBuiltin("upper", reflect.ValueOf(vm.NewBuiltinFunction(func(args ...reflect.Value) (reflect.Value, error) {
		switch t := args[0].Interface().(type) {
		case rune:
			return reflect.ValueOf(unicode.ToUpper(t)), nil
		case string:
			return reflect.ValueOf(strings.ToUpper(t)), nil
		default:
			return args[0], nil
		}
	})))
}
