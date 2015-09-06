package vm

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
)

type Number struct {
	*big.Rat
	isfloat bool
}

func NewInt(i int64) Number {
	ret := new(big.Rat)
	ret.SetInt64(i)
	return Number{Rat: ret, isfloat: false}
}

func (n Number) ToInt() int64 {
	return n.Num().Int64()
}

func GetInt(v reflect.Value) (int64, bool) {
	if !v.IsValid() {
		return 0, false
	}
	switch t := v.Interface().(type) {
	case Number:
		return t.ToInt(), true
	}
	return 0, false
}

func (n Number) String() string {
	if !n.isfloat {
		return n.Num().String()
	} else {
		return n.Rat.String()
	}
}

func SscanNumber(str string) (Number, bool) {
	r := new(big.Rat)
	_, err := fmt.Sscan(str, r)
	if err != nil {
		return Number{}, false
	} else {
		return Number{Rat: r, isfloat: strings.Contains(str, ".")}, true
	}
}

func AddV(a, b reflect.Value) (reflect.Value, error) {
	switch l := a.Interface().(type) {
	case Number:
		switch r := b.Interface().(type) {
		case Number:
			ret := new(big.Rat)
			ret.Add(l.Rat, r.Rat)
			return reflect.ValueOf(Number{Rat: ret, isfloat: l.isfloat || r.isfloat}), nil
		}
	}
	return NIL, fmt.Errorf("Error in Add")
}

func SubV(a, b reflect.Value) (reflect.Value, error) {
	switch l := a.Interface().(type) {
	case Number:
		switch r := b.Interface().(type) {
		case Number:
			ret := new(big.Rat)
			ret.Sub(l.Rat, r.Rat)
			return reflect.ValueOf(Number{Rat: ret, isfloat: l.isfloat || r.isfloat}), nil
		}
	}
	return NIL, fmt.Errorf("Error in Sub")
}

func MulV(a, b reflect.Value) (reflect.Value, error) {
	switch l := a.Interface().(type) {
	case Number:
		switch r := b.Interface().(type) {
		case Number:
			ret := new(big.Rat)
			ret.Mul(l.Rat, r.Rat)
			return reflect.ValueOf(Number{Rat: ret, isfloat: l.isfloat || r.isfloat}), nil
		}
	}
	return NIL, fmt.Errorf("Error in Mul")
}

func DivV(a, b reflect.Value) (reflect.Value, error) {
	switch l := a.Interface().(type) {
	case Number:
		switch r := b.Interface().(type) {
		case Number:
			if !l.isfloat && !r.isfloat {
				i := new(big.Int)
				i.Div(l.Num(), r.Num())
				ret := new(big.Rat)
				ret.SetInt(i)
				return reflect.ValueOf(Number{Rat: ret, isfloat: false}), nil
			}
			ret := new(big.Rat)
			ret.Quo(l.Rat, r.Rat)
			return reflect.ValueOf(Number{Rat: ret, isfloat: true}), nil
		}
	}
	return NIL, fmt.Errorf("Error in Div")
}

func ModV(a, b reflect.Value) (reflect.Value, error) {
	switch l := a.Interface().(type) {
	case Number:
		switch r := b.Interface().(type) {
		case Number:
			i1 := l.Num()
			i2 := r.Num()
			res := new(big.Int)
			res.Mod(i1, i2)
			ret := new(big.Rat)
			ret.SetInt(res)
			return reflect.ValueOf(Number{Rat: ret, isfloat: false}), nil
		}
	}
	return NIL, fmt.Errorf("Error in Mod")
}

func CmpV(a, b reflect.Value) (int, error) {
	switch l := a.Interface().(type) {
	case Number:
		switch r := b.Interface().(type) {
		case Number:
			return l.Cmp(r.Rat), nil
		}
	}
	return 0, fmt.Errorf("Error in Cmp")
}
