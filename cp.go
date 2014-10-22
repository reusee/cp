package cp

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
)

var (
	sp = fmt.Sprintf
	pt = fmt.Printf
)

type Cp struct {
	provides map[string]info
	requires map[string][]info
	defs     map[string]info
	impls    map[string][]info
}

type info struct {
	v        interface{}
	provides bool
	trace    string
}

func getTrace() string {
	buf := new(bytes.Buffer)
	for i := 1; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		var name string
		if fn := runtime.FuncForPC(pc); fn != nil {
			name = fn.Name()
		}
		fmt.Fprintf(buf, "%s:%d %s\n", file, line, name)
	}
	return string(buf.Bytes())
}

func New() *Cp {
	return &Cp{
		provides: make(map[string]info),
		requires: make(map[string][]info),
		defs:     make(map[string]info),
		impls:    make(map[string][]info),
	}
}

func (c *Cp) Provide(name string, v interface{}) {
	if _, in := c.provides[name]; in {
		panic(sp("multiple provides of %s", name))
	}
	c.provides[name] = info{
		v:     v,
		trace: getTrace(),
	}
}

func (c *Cp) Require(name string, ptr interface{}) {
	t := reflect.TypeOf(ptr)
	if t.Kind() != reflect.Ptr {
		panic(sp("required %s is not a pointer", name))
	}
	c.requires[name] = append(c.requires[name], info{
		v:     ptr,
		trace: getTrace(),
	})
}

func (c *Cp) define(name string, fnPtr interface{}, provides bool) {
	t := reflect.TypeOf(fnPtr)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Func {
		panic(sp("defined %s must be a pointer to function", name))
	}
	if _, in := c.defs[name]; in {
		panic(sp("multiple defines of %s", name))
	}
	c.defs[name] = info{
		v:        fnPtr,
		trace:    getTrace(),
		provides: provides,
	}
}

func (c *Cp) Define(name string, fnPtr interface{}) {
	c.define(name, fnPtr, false)
}

func (c *Cp) DefineProvide(name string, fnPtr interface{}) {
	c.define(name, fnPtr, true)
}

func (c *Cp) Implement(name string, fn interface{}) {
	t := reflect.TypeOf(fn)
	if t.Kind() != reflect.Func {
		panic(sp("implementation of %s must be function", name))
	}
	c.impls[name] = append(c.impls[name], info{
		v:     fn,
		trace: getTrace(),
	})
}

func (c *Cp) Impl(name string, fn interface{}) {
	c.Implement(name, fn)
}

var fnHandlers = make(map[reflect.Type]func(interface{}, []interface{}))

func AddFuncType(fnNilPtr interface{}, handler func(impls []interface{}) interface{}) {
	fnHandlers[reflect.TypeOf(fnNilPtr).Elem()] = func(fnPtr interface{}, impls []interface{}) {
		fn := handler(impls)
		reflect.ValueOf(fnPtr).Elem().Set(reflect.ValueOf(fn))
	}
}

func (c *Cp) Compose() {
	// handle defs / impls
	for name, defInfo := range c.defs {
		fnPtr := defInfo.v
		impls := c.impls[name]
		if len(impls) == 0 {
			panic(sp("no implementation for %s defined at\n%s", name, defInfo.trace))
		}
		fnType := reflect.TypeOf(fnPtr).Elem()
		for _, implInfo := range impls {
			if t := reflect.TypeOf(implInfo.v); t != fnType {
				panic(sp("defined %v, implemented %v. defined at\n%simplemented at\n%s", fnType, t, defInfo.trace, implInfo.trace))
			}
		}
		handler, ok := fnHandlers[fnType]
		if ok {
			implFns := make([]interface{}, 0, len(impls))
			for _, implInfo := range impls {
				implFns = append(implFns, implInfo.v)
			}
			handler(fnPtr, implFns)
		} else {
			implValues := make([]reflect.Value, 0, len(impls))
			for _, implInfo := range impls {
				implValues = append(implValues, reflect.ValueOf(implInfo.v))
			}
			reflect.ValueOf(fnPtr).Elem().Set(reflect.MakeFunc(fnType,
				func(args []reflect.Value) (ret []reflect.Value) {
					for _, impl := range implValues {
						impl.Call(args)
					}
					return
				}))
		}
		if defInfo.provides {
			c.provides[name] = info{
				v:     reflect.ValueOf(fnPtr).Elem().Interface(),
				trace: defInfo.trace,
			}
		}
		delete(c.defs, name)
		delete(c.impls, name)
	}

	// match provides and requires
	for name, provideInfo := range c.provides {
		requireInfos := c.requires[name]
		provideValue := reflect.ValueOf(provideInfo.v)
		for _, requireInfo := range requireInfos {
			requireValue := reflect.ValueOf(requireInfo.v).Elem()
			if provideValue.Type() != requireValue.Type() {
				panic(sp("%v provided, %v required. provided at\n%srequired at\n%s", provideValue.Type(), requireValue.Type(),
					provideInfo.trace, requireInfo.trace))
			}
			requireValue.Set(provideValue)
		}
		delete(c.requires, name)
		delete(c.provides, name)
	}
	for name, _ := range c.requires {
		panic(sp("%s not provided", name))
	}

}

func (c *Cp) Combine(c2 *Cp) {
	for name, info := range c2.provides {
		if _, ok := c.provides[name]; ok {
			panic(sp("multiple provides of %s", name))
		}
		c.provides[name] = info
	}
	for name, infos := range c2.requires {
		c.requires[name] = append(c.requires[name], infos...)
	}
	for name, info := range c2.defs {
		if _, ok := c.defs[name]; ok {
			panic(sp("multiple defines of %s", name))
		}
		c.defs[name] = info
	}
	for name, infos := range c2.impls {
		c.impls[name] = append(c.impls[name], infos...)
	}
}
