package pipe

import (
	//"github.com/k0kubun/pp"

	"log"
	"reflect"
	"sync"

	"../gc"
)

type filterChan struct {
	filter     func(<-chan reflect.Value, Valve)
	runed      bool
	reader     Valve
	numsources int
	ws         []Valve
	newW       chan Valve
	exportnewR chan Valve
	exitnotify chan bool
	exitonce   sync.Once
	runonce    sync.Once
	runedmutex sync.Mutex
	wg         sync.WaitGroup
	gc.Ref
}

//NewFilter creates new filter
func NewFilter(f func(<-chan reflect.Value, Valve)) Filter {
	ret := &filterChan{
		filter:     f,
		runed:      false,
		reader:     NewValve(),
		numsources: 0,
		ws:         []Valve{},
		newW:       make(chan Valve),
		exportnewR: make(chan Valve),
		exitnotify: make(chan bool, 1),
	}
	ret.Incref()
	go func() {
		ret.Wait()
		ret.NotifyExit()
	}()
	return ret
}

func (f *filterChan) NotifyExit() {
	f.exitonce.Do(func() {
		f.exitnotify <- true
	})
}

func (f *filterChan) AddW(v Valve) {
	f.runedmutex.Lock()
	defer f.runedmutex.Unlock()
	if f.runed {
		f.newW <- v
	} else {
		f.ws = append(f.ws, v)
	}
}

func (f *filterChan) NewR() Valve {
	f.runedmutex.Lock()
	defer f.runedmutex.Unlock()
	if f.runed {
		return <-f.exportnewR
	}
	f.numsources++
	return f.reader
}

func (f *filterChan) Run(wg *sync.WaitGroup) {
	wg.Add(1)
	f.runonce.Do(func() {
		f.runedmutex.Lock()
		f.runed = true
		f.runedmutex.Unlock()
		f.wg.Add(3)
		log.Println("filter run", wg)
		r := make(chan reflect.Value)
		w := NewValve()

		enR := make(chan bool, 1)
		enW := make(chan bool, 1)

		funend := make(chan bool, 2)

		go func() {
			<-f.exitnotify
			go func() { enR <- true }()
			go func() { enW <- true }()
		}()

		go func() {
			f.filter(r, w)
			w.Close()
			f.wg.Done()
			funend <- true
			log.Println("filter func end", wg)
		}()

		//read part
		go func() {
			defer func() {
				close(r)
				f.reader.Close()
				f.wg.Done()
			}()
			rchan := f.reader.Rchan()
			exitable := false
			for {
				select {
				case v := <-rchan:
					if IsEOF(v) {
						f.numsources--
						if f.numsources == 0 && exitable {
							return
						}
					} else {
						select {
						case r <- v:
						case <-funend:
							return
						}
					}
				case f.exportnewR <- f.reader:
					f.numsources++
				case <-enR:
					exitable = true
					if f.numsources == 0 && exitable {
						return
					}
				}
			}
		}()

		//write part
		go func() {
			defer func() {
				funend <- true
				for _, valve := range f.ws {
					valve.Send(EOF)
				}
				f.ws = nil
				w.Close()
				f.wg.Done()
			}()
			rchan := w.Rchan()
			exitable := false
			for {
				if len(f.ws) == 0 {
					select {
					case nw := <-f.newW:
						f.ws = append(f.ws, nw)
					case <-enW:
						exitable = true
						return
					}
				} else {
					select {
					case v, ok := <-rchan:
						if ok {
							valids := []Valve{}
							for _, valve := range f.ws {
								if valve.Send(v) {
									valids = append(valids, valve)
								}
							}
							f.ws = valids
							if exitable && len(f.ws) == 0 {
								return
							}
						} else {
							if exitable {
								return
							}
							select {
							case nw := <-f.newW:
								f.ws = append(f.ws, nw)
							case <-enW:
								exitable = true
								return
							}
						}
					case nw := <-f.newW:
						f.ws = append(f.ws, nw)
					case <-enW:
						exitable = true
						if len(f.ws) == 0 {
							return
						}
					}
				}
			}
		}()
	})

	go func() {
		f.wg.Wait()
		wg.Done()
	}()
}
