package cp

import (
	"strings"
	"testing"
)

func TestAll(t *testing.T) {
	// provide / require integer
	c := New()
	c.Provide("foo", 42)
	var i int
	c.Require("foo", &i)
	c.Compose()
	if i != 42 {
		t.Fatal()
	}

	// provide / require func
	ok := false
	c.Provide("foo", func() {
		ok = true
	})
	var fn func()
	c.Require("foo", &fn)
	c.Compose()
	fn()
	if !ok {
		t.Fatal()
	}

	// define / implment func
	fn = nil
	n := 0
	c.Define("foo", &fn)
	c.Implement("foo", func() {
		n++
	})
	c.Implement("foo", func() {
		n += 2
	})
	c.Compose()
	fn()
	if n != 3 {
		t.Fatal()
	}

	// define & provide func
	fn = nil
	n = 0
	c.DefineProvide("foo", &fn)
	c.Implement("foo", func() {
		n++
	})
	c.Impl("foo", func() {
		n += 3
	})
	var fn2 func()
	c.Require("foo", &fn2)
	c.Compose()
	fn2()
	if n != 4 {
		t.Fatal()
	}

	// func(int)
	var f1 func(int)
	c.Define("foo", &f1)
	c.Impl("foo", func(int) {})
	c.Compose()
	f1(42)
	var f2 func(string)
	c.Define("foo", &f2)
	c.Impl("foo", func(string) {})
	c.Compose()
	f2("foo")
	var f3 func(bool)
	c.Define("foo", &f3)
	c.Impl("foo", func(bool) {})
	c.Compose()
	f3(true)
	var f4 func(int, int)
	c.Define("foo", &f4)
	c.Impl("foo", func(int, int) {})
	c.Compose()
	f4(42, 42)
}

func TestPanic(t *testing.T) {
	checkErr := func(s string) {
		if err := recover(); err == nil || err.(string) != s {
			t.Fatal()
		}
	}

	func() {
		defer checkErr("multiple provides of foo")
		c := New()
		c.Provide("foo", 42)
		c.Provide("foo", 42)
		c.Compose()
	}()

	func() {
		defer checkErr("required foo is not a pointer")
		c := New()
		c.Require("foo", 42)
		c.Compose()
	}()

	func() {
		defer checkErr("defined foo must be a pointer to function")
		c := New()
		c.Define("foo", 42)
		c.Compose()
	}()

	func() {
		defer checkErr("multiple defines of foo")
		c := New()
		var foo func()
		c.Define("foo", &foo)
		c.Define("foo", &foo)
		c.Compose()
	}()

	func() {
		defer checkErr("implementation of foo must be function")
		c := New()
		c.Impl("foo", 42)
		c.Compose()
	}()
}

func TestComposePanic(t *testing.T) {
	checkErr := func(s string) {
		if err := recover(); err == nil || !strings.HasPrefix(err.(string), s) {
			t.Fatal()
		}
	}

	func() {
		defer checkErr("no implementation for foo")
		c := New()
		var foo func()
		c.Define("foo", &foo)
		c.Compose()
	}()

	func() {
		defer checkErr("defined func(), implemented func(int)")
		c := New()
		var foo func()
		c.Define("foo", &foo)
		c.Impl("foo", func(int) {})
		c.Compose()
	}()

	func() {
		defer checkErr("int provided, bool required")
		c := New()
		c.Provide("foo", 42)
		var b bool
		c.Require("foo", &b)
		c.Compose()
	}()

	func() {
		defer checkErr("foo not provided")
		c := New()
		var i int
		c.Require("foo", &i)
		c.Compose()
	}()
}
