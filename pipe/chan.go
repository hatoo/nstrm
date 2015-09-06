package pipe

import (
	//"github.com/k0kubun/pp"

	"log"
	"reflect"
	"sync"

	"../gc"
)

type pipechan struct {
	runed      bool
	reader     Valve
	numsources int
	done       chan bool
	ws         []Valve
	newW       chan Valve
	exportnewR chan Valve
	runonce    sync.Once
	runedmutex sync.RWMutex
	gc.Ref
}

//NewChan creates new pipechan
func NewChan() Handle {
	ret := &pipechan{
		runed:      false,
		reader:     NewValve(),
		numsources: 0,
		done:       make(chan bool),
		ws:         []Valve{},
		newW:       make(chan Valve),
		exportnewR: make(chan Valve),
	}
	ret.Incref()
	go func() {
		ret.Wait()
		log.Println("chan exit notify")
	}()
	return ret
}

func (f *pipechan) NotifyExit() {
}

func (f *pipechan) AddW(v Valve) {
	f.runedmutex.RLock()
	defer f.runedmutex.RUnlock()

	if f.runed {
		f.newW <- v
	} else {
		f.ws = append(f.ws, v)
	}
}

func (f *pipechan) NewR() Valve {
	f.runedmutex.RLock()
	defer f.runedmutex.RUnlock()

	if f.runed {
		return <-f.exportnewR
	}
	f.numsources++
	return f.reader
}

func (f *pipechan) Result() reflect.Value {
	f.Wait()
	return reflect.ValueOf(nil)
}

func (f *pipechan) Run(*sync.WaitGroup) {
	f.runonce.Do(func() {
		log.Println("pipechan run")
		f.runedmutex.Lock()
		f.runed = true
		f.runedmutex.Unlock()

		r := make(chan reflect.Value)
		w := make(chan reflect.Value)

		go func() {
			for v := range r {
				w <- v
			}
		}()

		go func() {
			buf := []reflect.Value{}
			rchan := f.reader.Rchan()
			for {
				if len(buf) != 0 {
					select {
					case r <- buf[0]:
						buf = buf[1:]
					case f.exportnewR <- f.reader:
						f.numsources++
					}
				} else {
					select {
					case v := <-rchan:
						if IsEOF(v) {
							f.numsources--
						} else {
							buf = append(buf, v)
						}
					case f.exportnewR <- f.reader:
						f.numsources++
					}
				}
			}
		}()

		go func() {
			buf := []reflect.Value{}
			for {
				if len(f.ws) != 0 {
					if len(buf) != 0 {
						select {
						case valve := <-f.newW:
							f.ws = append(f.ws, valve)
						default:
							valids := []Valve{}
							for _, valve := range f.ws {
								if valve.Send(buf[0]) {
									valids = append(valids, valve)
								}
							}
							f.ws = valids
							if len(valids) != 0 {
								buf = buf[1:]
							}
						}
					} else {
						select {
						case value := <-w:
							if IsEOF(value) {
								panic("should not send EOF")
							}
							buf = append(buf, value)
						case valve := <-f.newW:
							f.ws = append(f.ws, valve)
						}
					}
				} else {
					select {
					case valve := <-f.newW:
						f.ws = append(f.ws, valve)
					}
				}
			}
		}()
	})
}
