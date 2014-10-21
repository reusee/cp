package cp

import "testing"

func TestAll(t *testing.T) {
	c := New()
	c.Provide("foo", 42)
}
