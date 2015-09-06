package gc

import (
	"reflect"
	"sync"
)

//GcThing is data that is controlled by GC
type GcThing interface {
	Incref()
	Decref()
	Wait()
}

//Decif decrements refcount if v is GcThing
func Decif(v reflect.Value) {
	if v.IsValid() {
		switch t := v.Interface().(type) {
		case GcThing:
			t.Decref()
		}
	}
}

//Incif increments refcount if v is GcThing
func Incif(v reflect.Value) {
	if v.IsValid() {
		switch t := v.Interface().(type) {
		case GcThing:
			t.Incref()
		}
	}
}

//Waitif calls Wait() if v is GcThing
func Waitif(v reflect.Value) {
	if v.IsValid() {
		switch t := v.Interface().(type) {
		case GcThing:
			t.Wait()
		}
	}
}

//Ref is refcount
type Ref sync.WaitGroup

//Incref increments refcount
func (r *Ref) Incref() {
	(*sync.WaitGroup)(r).Add(1)
}

//Decref decrements refcount
func (r *Ref) Decref() {
	(*sync.WaitGroup)(r).Done()
}

//Wait wait untill refcount is zero
func (r *Ref) Wait() {
	(*sync.WaitGroup)(r).Wait()
}
