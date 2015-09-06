package pipe

import (
	"log"
	"sync"

	"../gc"
)

type producerChan struct {
	origin     Valve
	ws         []Valve
	newW       chan Valve
	exitnotify chan bool
	wg         sync.WaitGroup
	runonce    sync.Once
	exitonce   sync.Once
	wsmutex    sync.Mutex
	gc.Ref
}

//NewProducer creates new Producer
func NewProducer(Out Valve) Producer {
	p := &producerChan{
		origin:     Out,
		ws:         []Valve{},
		newW:       make(chan Valve),
		exitnotify: make(chan bool, 1),
	}
	p.Incref()
	go func() {
		p.Wait()
		p.NotifyExit()
	}()
	return p
}

func (p *producerChan) NotifyExit() {
	p.exitonce.Do(func() {
		p.exitnotify <- true
	})
}

func (p *producerChan) Run(wg *sync.WaitGroup) {
	wg.Add(1)
	p.runonce.Do(func() {
		p.wg.Add(1)
		log.Println("producer run")
		go func() {
			defer func() {
				p.wsmutex.Lock()
				log.Println("producer end", p.ws)
				for _, valve := range p.ws {
					valve.Send(EOF)
				}
				p.ws = nil
				p.wsmutex.Unlock()
				p.wg.Done()
				log.Println("producer return")
			}()
			exitable := false
			rchan := p.origin.Rchan()
			for {
				select {
				case value, ok := <-rchan:
					if ok {
						if IsEOF(value) {
							panic("do not send EOF")
						} else {
							p.wsmutex.Lock()
							valids := []Valve{}
							for _, valve := range p.ws {
								if valve.Send(value) {
									valids = append(valids, valve)
								}
							}
							log.Println("Producer sent ", value.Interface(), len(valids))
							p.ws = valids
							if exitable && len(p.ws) == 0 {
								p.origin.Close()
								p.wsmutex.Unlock()
								return
							}
							p.wsmutex.Unlock()
						}
					} else {
						return
					}
				case <-p.exitnotify:
					log.Println("producer exit notify")
					exitable = true
					p.wsmutex.Lock()
					if len(p.ws) == 0 {
						p.origin.Close()
						p.wsmutex.Unlock()
						return
					}
					p.wsmutex.Unlock()
				}
			}
		}()
	})

	go func() {
		p.wg.Wait()
		wg.Done()
	}()
}

func (p *producerChan) AddW(newv Valve) {
	p.wsmutex.Lock()
	defer p.wsmutex.Unlock()
	p.ws = append(p.ws, newv)
}
