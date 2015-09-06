package pipe

import (
	"log"
	"reflect"
	"sync"

	"../gc"
)

type consumerFunction struct {
	consumer   func(<-chan reflect.Value) reflect.Value
	runed      bool
	numsources int
	reader     Valve
	exportnewR chan Valve
	result     reflect.Value
	exitnotify chan bool
	exitonce   sync.Once
	runonce    sync.Once
	exitmutex  sync.RWMutex
	runedmutex sync.Mutex
	wg         sync.WaitGroup
	gc.Ref
}

//NewConsumer creates new Consumer
func NewConsumer(f func(<-chan reflect.Value) reflect.Value) Consumer {
	c := &consumerFunction{
		consumer:   f,
		runed:      false,
		numsources: 0,
		reader:     NewValve(),
		exportnewR: make(chan Valve),
		exitnotify: make(chan bool, 1),
	}
	c.exitmutex.Lock()
	c.Incref()
	go func() {
		c.Wait()
		log.Println("consumer end")
		c.NotifyExit()
	}()
	return c
}

func (c *consumerFunction) NotifyExit() {
	c.exitonce.Do(func() {
		c.exitnotify <- true
	})
}

func (c *consumerFunction) NewR() Valve {
	c.runedmutex.Lock()
	defer c.runedmutex.Unlock()
	if c.runed {
		return <-c.exportnewR
	}
	c.numsources++
	return c.reader
}

func (c *consumerFunction) Result() reflect.Value {
	c.exitmutex.RLock()
	defer c.exitmutex.RUnlock()
	return c.result
}

func (c *consumerFunction) Run(wg *sync.WaitGroup) {
	wg.Add(1)

	c.runonce.Do(func() {
		c.wg.Add(1)
		c.runedmutex.Lock()
		c.runed = true
		c.runedmutex.Unlock()
		log.Println("consumer run", c.wg)
		r := make(chan reflect.Value)
		w := make(chan reflect.Value)
		go func() {
			w <- c.consumer(r)
		}()

		go func() {
			defer func() {
				log.Println("consumer end", c.wg)
				c.reader.Close()
				c.exitmutex.Unlock()
				c.wg.Done()
				log.Println("consumer return", c.result, c.wg)
			}()
			exitable := false
			rchan := c.reader.Rchan()
			for {
				select {
				case v := <-rchan:
					if IsEOF(v) {
						c.numsources--
						log.Println("consumer got eof", c.wg)
						if exitable && c.numsources == 0 {
							log.Println("close r")
							close(r)
						}
					} else {
						log.Println("consumer got", v, c)
						select {
						case r <- v:
						case res := <-w:
							log.Println("consumer got w", c.wg)
							c.result = res
							return
						}
					}
				case res := <-w:
					log.Println("consumer got w", c.wg)
					c.result = res
					return
				case c.exportnewR <- c.reader:
					c.numsources++
				case <-c.exitnotify:
					log.Println("consumer exit notify", c.numsources, c.wg)
					exitable = true
					if c.numsources == 0 {
						log.Println("close r")
						close(r)
					}
				}
			}
		}()
	})

	go func() {
		c.wg.Wait()
		wg.Done()
	}()
}
