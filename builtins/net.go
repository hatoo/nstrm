package builtins

import (
	"fmt"
	"io"
	"log"
	"net"
	"reflect"

	"../pipe"
	"../vm"
)

//LoadNet defines net function
func LoadNet(env *vm.Env) {
	env.DefineBuiltin("tcp_server", reflect.ValueOf(vm.NewBuiltinFunction(func(args ...reflect.Value) (reflect.Value, error) {
		port, ok := vm.GetInt(args[0])
		if !ok {
			return vm.NIL, fmt.Errorf("%v is not number", port)
		}
		out := pipe.NewValve()
		if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err == nil {
			go func() {
				for {
					connection, err := ln.Accept()
					if err == nil {
						producer := pipe.NewValve()
						go func(conn net.Conn, v pipe.Valve) {
							buf := make([]byte, 4096)
							for {
								if n, err := conn.Read(buf); err == nil {
									for i := 0; i < n; i++ {
										v.Send(reflect.ValueOf(buf[i]))
									}
								} else if err == io.EOF {
									conn.Close()
									producer.Close()
									return
								}
							}
						}(connection, producer)

						consumer := func(conn net.Conn) func(<-chan reflect.Value) reflect.Value {
							return func(r <-chan reflect.Value) reflect.Value {
								for v := range r {
									switch t := v.Interface().(type) {
									case byte:
										if _, err := conn.Write([]byte{t}); err != nil {
											conn.Close()
											break
										}
									default:
										log.Println("unimplemented")
									}
								}
								return vm.NIL
							}
						}(connection)
						i := pipe.NewConsumer(consumer)
						o := pipe.NewProducer(producer)
						io := pipe.InOut(i, o)
						i.Decref()
						o.Decref()
						if !out.Send(reflect.ValueOf(io)) {
							io.Decref()
							out.Close()
							return
						}
					}
				}
			}()
			return reflect.ValueOf(pipe.NewProducer(out)), nil
		} else {
			return vm.NIL, err
		}
	})))
}
