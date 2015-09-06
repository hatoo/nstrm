package vm

import (
	"log"
	"reflect"
	"sync"

	"../pipe"
)

//Condition converts value to bool
func Condition(v reflect.Value) bool {
	if v == NIL {
		return false
	}
	switch t := v.Interface().(type) {
	case bool:
		return t
	case pipe.Terminal:
		var wg sync.WaitGroup
		log.Println("<<<<<")
		t.Run(&wg)
		t.NotifyExit()
		wg.Wait()
		ret := t.Result()
		log.Println(">>>>>")
		return Condition(ret)
	}
	panic("cant be condition")
}

func Equal(a, b reflect.Value) bool {
	if cmp, err := CmpV(a, b); err == nil {
		return cmp == 0
	}
	if !a.IsValid() || !b.IsValid() {
		return !a.IsValid() && !b.IsValid()
	}
	switch arr1 := a.Interface().(type) {
	case []reflect.Value:
		switch arr2 := b.Interface().(type) {
		case []reflect.Value:
			if len(arr1) != len(arr2) {
				return false
			}
			for i := 0; i < len(arr1); i++ {
				if !Equal(arr1[i], arr2[i]) {
					return false
				}
			}
			return true
		}
	}
	return a.Interface() == b.Interface()
}
