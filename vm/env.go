package vm

import (
	"log"
	"reflect"
	"sync"

	"../gc"
	"../pipe"
)

//Env is a state for run vm
type Env struct {
	parent          *Env
	namespace       map[string]reflect.Value
	out             pipe.Valve
	runnotify       map[pipe.Pipe]bool
	decreflist      []gc.GcThing
	task            sync.WaitGroup
	namespacemutex  sync.RWMutex
	decreflistmutex sync.Mutex
	runnotifymutex  sync.Mutex
	outmutex        sync.RWMutex
}

//Incref is implements for gc.GcThing
func (env *Env) Incref() {
	env.task.Add(1)
}

//Decref is implements for gc.GcThing
func (env *Env) Decref() {
	env.task.Done()
}

func (env *Env) decrefAll() {
	env.namespacemutex.RLock()
	defer env.namespacemutex.RUnlock()
	log.Println("decrefAll")
	for _, v := range env.namespace {
		gc.Decif(v)
	}
}

//Wait is implements for gc.GcThing
func (env *Env) Wait() {
	env.task.Wait()
}

//RunLater regist pipe to run when called Env.Run
func (env *Env) RunLater(p pipe.Pipe) {
	env.runnotifymutex.Lock()
	defer env.runnotifymutex.Unlock()
	env.runnotify[p] = true
}

//DecrefLater adds to DecrefList. Decrease its refcount when this score is out.
func (env *Env) DecrefLater(v gc.GcThing) {
	env.decreflistmutex.Lock()
	defer env.decreflistmutex.Unlock()
	env.decreflist = append(env.decreflist, v)
}

//DecrefLaterV provides DecrefLater for reflect.Value
func (env *Env) DecrefLaterV(v reflect.Value) {
	if v.IsValid() {
		switch t := v.Interface().(type) {
		case gc.GcThing:
			env.DecrefLater(t)
		}
	}
}

//SetOut set out pipe
func (env *Env) SetOut(out pipe.Valve) {
	env.outmutex.Lock()
	defer env.outmutex.Unlock()
	env.out = out
}

//Send send value to current scope's out pipe. it's called 'Emit' expression
func (env *Env) Send(v reflect.Value) bool {
	env.outmutex.RLock()
	defer env.outmutex.RUnlock()
	return env.out.Send(v)
}

//NewEnv creates new Env. use sync.WaitGroup to wait until root environment's refcount is zero.
func NewEnv(wg *sync.WaitGroup) *Env {
	log.Println("newenv")
	e := &Env{
		parent:     nil,
		namespace:  make(map[string]reflect.Value),
		out:        pipe.NewValve(),
		runnotify:  make(map[pipe.Pipe]bool),
		decreflist: []gc.GcThing{},
	}
	wg.Add(1)
	e.Incref()
	go func() {
		defer wg.Done()
		e.Wait()
		e.namespacemutex.Lock()
		e.namespace = nil
		e.namespacemutex.Unlock()
		log.Println("root env end")
	}()
	return e
}

//ChildEnv creates child Env for Function Call
func (parent *Env) ChildEnv() *Env {
	parent.outmutex.RLock()
	defer parent.outmutex.RUnlock()
	parent.Incref()
	e := &Env{
		parent:     parent,
		namespace:  make(map[string]reflect.Value),
		out:        parent.out,
		runnotify:  make(map[pipe.Pipe]bool),
		decreflist: []gc.GcThing{},
	}
	e.Incref()
	go func() {
		e.Wait()
		e.namespacemutex.Lock()
		e.namespace = nil
		e.namespacemutex.Unlock()
		parent.Decref()
		e.parent = nil
		log.Println("child env end")
	}()
	return e
}

//RunWait runs pipe connection and leave current scope. block until all connection is end
func (env *Env) RunWait(retvalue reflect.Value) {
	env.decreflistmutex.Lock()
	env.runnotifymutex.Lock()
	defer env.decreflistmutex.Unlock()
	defer env.runnotifymutex.Unlock()
	var wg sync.WaitGroup
	env.Incref()
	for p, b := range env.runnotify {
		if b {
			log.Println("run", p)
			p.Run(&wg)
		}
	}
	for _, t := range env.decreflist {
		t.Decref()
	}
	env.runnotify = nil
	env.decreflist = nil
	wg.Wait()
	log.Println("wg end root")
	gc.Waitif(retvalue)
	log.Println("decref all root")
	env.decrefAll()
	env.Decref()
}

//Run runs pipe connection and leave current scope
func (env *Env) Run(retvalue reflect.Value) {
	env.decreflistmutex.Lock()
	env.runnotifymutex.Lock()
	defer env.decreflistmutex.Unlock()
	defer env.runnotifymutex.Unlock()
	var wg sync.WaitGroup
	env.Incref()
	for p, b := range env.runnotify {
		if b {
			log.Println("child run", p)
			p.Run(&wg)
		}
	}
	for _, t := range env.decreflist {
		t.Decref()
	}
	env.runnotify = nil
	env.decreflist = nil
	go func() {
		wg.Wait()
		gc.Waitif(retvalue)
		log.Println("decref all")
		env.decrefAll()
		env.Decref()
	}()
}

//Lookup lookup variable
func (env *Env) Lookup(key string) (reflect.Value, bool) {
	env.namespacemutex.RLock()
	defer env.namespacemutex.RUnlock()
	if env == nil {
		return reflect.ValueOf(nil), false
	}

	if v, ok := env.namespace[key]; ok {
		//gc.Incif(v)
		return v, ok
	}
	if env.parent == nil {
		return reflect.ValueOf(nil), false
	}
	return env.parent.Lookup(key)
}

//DefineBuiltin
func (env *Env) DefineBuiltin(key string, v reflect.Value) {
	env.namespacemutex.Lock()
	defer env.namespacemutex.Unlock()
	env.namespace[key] = v
}

//Define defines variable to environment
func (env *Env) Define(key string, v reflect.Value) {
	s := env
	for {
		s.namespacemutex.Lock()
		_, ok := s.namespace[key]
		if ok {
			break
		} else {
			s.namespacemutex.Unlock()
		}
		s = s.parent
		if s == nil {
			break
		}
	}
	if s == nil {
		gc.Incif(v)
		env.namespacemutex.Lock()
		env.namespace[key] = v
		env.namespacemutex.Unlock()
	} else {
		gc.Incif(v)
		gc.Decif(s.namespace[key])
		s.namespace[key] = v
		s.namespacemutex.Unlock()
	}
}
