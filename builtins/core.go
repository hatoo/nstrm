package builtins

import (
	"fmt"
	"reflect"

	"../vm"
)

func getString(v reflect.Value) (string, error) {
	if v.Kind() == reflect.String {
		return v.String(), nil
	}
	return "", fmt.Errorf("it's not string. it is %s", v.Kind().String())
}

func helper(fun func(reflect.Value, reflect.Value) (reflect.Value, error)) reflect.Value {
	return reflect.ValueOf(vm.NewBuiltinFunction(func(args ...reflect.Value) (reflect.Value, error) {
		if len(args) == 2 {
			return fun(args[0], args[1])
		}
		return reflect.ValueOf(nil), fmt.Errorf("wrong number of argments")
	}))
}

//LoadCore defines core function to env
func LoadCore(env *vm.Env) {
	LoadIO(env)
	LoadUtil(env)

	env.DefineBuiltin("append", helper(func(arr, elem reflect.Value) (reflect.Value, error) {
		switch a := arr.Interface().(type) {
		case []reflect.Value:
			return reflect.ValueOf(append(a, elem)), nil
		}
		return reflect.ValueOf(nil), fmt.Errorf("wrong type")
	}))

	env.DefineBuiltin("==", helper(func(a, b reflect.Value) (reflect.Value, error) {
		return reflect.ValueOf(vm.Equal(a, b)), nil
	}))

	env.DefineBuiltin("!=", helper(func(a, b reflect.Value) (reflect.Value, error) {
		return reflect.ValueOf(!vm.Equal(a, b)), nil
	}))

	env.DefineBuiltin("MOD", helper(vm.ModV))

	env.DefineBuiltin("ADD", helper(vm.AddV))

	env.DefineBuiltin("SUB", helper(vm.SubV))

	env.DefineBuiltin("MUL", helper(vm.MulV))

	env.DefineBuiltin("DIV", helper(vm.DivV))

	env.DefineBuiltin("or", helper(func(a, b reflect.Value) (reflect.Value, error) {
		if a.Kind() == reflect.Bool && b.Kind() == reflect.Bool {
			return reflect.ValueOf(a.Bool() || b.Bool()), nil
		}
		return reflect.ValueOf(nil), fmt.Errorf("type mismatch")
	}))

	env.DefineBuiltin("and", helper(func(a, b reflect.Value) (reflect.Value, error) {
		if a.Kind() == reflect.Bool && b.Kind() == reflect.Bool {
			return reflect.ValueOf(a.Bool() && b.Bool()), nil
		}
		return reflect.ValueOf(nil), fmt.Errorf("type mismatch")
	}))

	env.DefineBuiltin("<=", helper(func(a, b reflect.Value) (reflect.Value, error) {
		if cmp, err := vm.CmpV(a, b); err == nil {
			return reflect.ValueOf(cmp <= 0), nil
		}
		return reflect.ValueOf(nil), fmt.Errorf("type mismatch")
	}))

	env.DefineBuiltin(">=", helper(func(a, b reflect.Value) (reflect.Value, error) {
		if cmp, err := vm.CmpV(a, b); err == nil {
			return reflect.ValueOf(cmp >= 0), nil
		}
		return reflect.ValueOf(nil), fmt.Errorf("type mismatch")
	}))

	env.DefineBuiltin("<", helper(func(a, b reflect.Value) (reflect.Value, error) {
		if cmp, err := vm.CmpV(a, b); err == nil {
			return reflect.ValueOf(cmp < 0), nil
		}
		return reflect.ValueOf(nil), fmt.Errorf("type mismatch")
	}))

	env.DefineBuiltin(">", helper(func(a, b reflect.Value) (reflect.Value, error) {
		if cmp, err := vm.CmpV(a, b); err == nil {
			return reflect.ValueOf(cmp > 0), nil
		}
		return reflect.ValueOf(nil), fmt.Errorf("type mismatch")
	}))
}
