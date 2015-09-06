package pipe

import (
	"log"
	"reflect"
	"sync"

	"../gc"
)

//EOF. send EOF instead of close channel
var EOF = reflect.ValueOf(nil)

//IsEOF check is it EOF
func IsEOF(v reflect.Value) bool {
	return v == reflect.ValueOf(nil)
}

//Valve
type Valve interface {
	Send(reflect.Value) bool
	Receive() (reflect.Value, bool)
	Close()
	Rchan() chan reflect.Value
}

type valveimpl struct {
	ch        chan reflect.Value
	closeonce sync.Once
	done      chan bool
}

type nilvalve struct{}

//NewValve creates new Valve
func NewValve() Valve {
	ret := &valveimpl{
		ch:   make(chan reflect.Value),
		done: make(chan bool),
	}
	return ret
}

func NilValve() Valve {
	return &nilvalve{}
}

func (v *valveimpl) Close() {
	v.closeonce.Do(func() {
		close(v.done)
	})
}

func (v *nilvalve) Close() {}

func (valve *valveimpl) Send(v reflect.Value) bool {
	select {
	case valve.ch <- v:
		return true
	case <-valve.done:
		return false
	}
}

func (valve *nilvalve) Send(v reflect.Value) bool {
	return false
}
func (valve *nilvalve) Receive() (reflect.Value, bool) {
	return EOF, false
}
func (valve *valveimpl) Receive() (reflect.Value, bool) {
	select {
	case v := <-valve.ch:
		return v, true
	case <-valve.done:
		return EOF, false
	}
}

func (*nilvalve) Rchan() chan reflect.Value {
	ch := make(chan reflect.Value)
	close(ch)
	return ch
}

func (valve *valveimpl) Rchan() chan reflect.Value {
	r := make(chan reflect.Value, 1)
	go func() {
		defer close(r)
		for {
			select {
			case v := <-valve.ch:
				r <- v

				/*
					select {
					case r <- v:
					case <-valve.done:
						return
					}
				*/
			case <-valve.done:
				return
			}
		}
	}()
	return r
}

//Pipe
type Pipe interface {
	Run(*sync.WaitGroup)
	NotifyExit()
	gc.GcThing
}

type Producer interface {
	Pipe
	AddW(Valve)
}

type Consumer interface {
	Pipe
	NewR() Valve
	Result() reflect.Value
}

type Filter interface {
	Pipe
	AddW(Valve)
	NewR() Valve
}

type Terminal interface {
	Pipe
	terminal()
	Result() reflect.Value
}

type terminalimpl struct{}

func (*terminalimpl) terminal() {}

type Handle interface {
	Pipe
	AddW(Valve)
	NewR() Valve
	Result() reflect.Value
}

type port struct {
	P Producer
	C Consumer
	gc.Ref
}

type connectedPC struct {
	P Producer
	C Consumer
	gc.Ref
	terminalimpl
}

type connectedPF struct {
	P Producer
	F Filter
	gc.Ref
}

type connectedFC struct {
	F Filter
	C Consumer
	gc.Ref
}

type connectedFF struct {
	F1 Filter
	F2 Filter
	gc.Ref
}

func ConnectPC(p Producer, c Consumer) Terminal {
	log.Println("ConnectPC")
	p.AddW(c.NewR())
	log.Println("ConnectPC ok")
	//p.Incref()
	//c.Incref()
	ret := &connectedPC{P: p, C: c}
	ret.Incref()
	/*
		go func() {
			ret.Wait()
			//p.Decref()
			//c.Decref()
		}()
	*/
	return ret
}

func ConnectPF(p Producer, f Filter) Producer {
	log.Println("ConnectPF")
	p.AddW(f.NewR())
	//p.Incref()
	//f.Incref()
	ret := &connectedPF{P: p, F: f}
	ret.Incref()
	/*
		go func() {
			ret.Wait()
			//p.Decref()
			//f.Decref()
		}()
	*/
	return ret
}

func ConnectFC(f Filter, c Consumer) Consumer {
	log.Println("ConnectFC")
	f.AddW(c.NewR())
	//f.Incref()
	//c.Incref()
	ret := &connectedFC{F: f, C: c}
	ret.Incref()
	/*
		go func() {
			ret.Wait()
			log.Println("FC end")
			//f.Decref()
			//c.Decref()
		}()
	*/
	return ret
}

func ConnectFF(f1 Filter, f2 Filter) Filter {
	log.Println("ConnectFF")
	f1.AddW(f2.NewR())
	//f1.Incref()
	//f2.Incref()
	ret := &connectedFF{F1: f1, F2: f2}
	ret.Incref()
	/*
		go func() {
			ret.Wait()
			//f1.Decref()
			//f2.Decref()
		}()
	*/
	return ret
}

func InOut(c Consumer, p Producer) Handle {
	//p.Incref()
	//c.Incref()
	ret := &port{P: p, C: c}
	ret.Incref()
	/*
		go func() {
			ret.Wait()
			//p.Decref()
			//c.Decref()
		}()
	*/
	return ret
}

func (this *connectedPC) Run(wg *sync.WaitGroup) {
	this.P.Run(wg)
	this.C.Run(wg)
}

func (this *connectedPF) Run(wg *sync.WaitGroup) {
	this.P.Run(wg)
	this.F.Run(wg)
}

func (this *connectedFC) Run(wg *sync.WaitGroup) {
	this.F.Run(wg)
	this.C.Run(wg)
}

func (this *connectedFF) Run(wg *sync.WaitGroup) {
	this.F1.Run(wg)
	this.F2.Run(wg)
}

func (this *port) Run(wg *sync.WaitGroup) {
	this.P.Run(wg)
	this.C.Run(wg)
}

func (this *connectedPC) NotifyExit() {
	this.P.NotifyExit()
	this.C.NotifyExit()
}

func (this *connectedPF) NotifyExit() {
	this.P.NotifyExit()
	this.F.NotifyExit()
}

func (this *connectedFC) NotifyExit() {
	this.F.NotifyExit()
	this.C.NotifyExit()
}

func (this *connectedFF) NotifyExit() {
	this.F1.NotifyExit()
	this.F2.NotifyExit()
}

func (this *port) NotifyExit() {
	this.P.NotifyExit()
	this.C.NotifyExit()
}

func (this *connectedFC) NewR() Valve {
	return this.F.NewR()
}

func (this *connectedFF) NewR() Valve {
	return this.F1.NewR()
}

func (this *port) NewR() Valve {
	return this.C.NewR()
}

func (this *connectedFC) Result() reflect.Value {
	return this.C.Result()
}

func (this *connectedPC) Result() reflect.Value {
	return this.C.Result()
}

func (this *port) Result() reflect.Value {
	return this.C.Result()
}

func (this *connectedPF) AddW(valve Valve) {
	this.F.AddW(valve)
}

func (this *connectedFF) AddW(valve Valve) {
	this.F2.AddW(valve)
}

func (this *port) AddW(valve Valve) {
	this.P.AddW(valve)
}
