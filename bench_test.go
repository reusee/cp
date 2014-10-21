package cp

import "testing"

func BenchmarkImpl(b *testing.B) {
	c := New()
	var fn func()
	c.Define("foo", &fn)
	c.Impl("foo", func() {})
	c.Compose()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
}

func BenchmarkImplReflectionBased(b *testing.B) {
	c := New()
	var fn func(struct{})
	c.Define("foo", &fn)
	c.Impl("foo", func(struct{}) {})
	c.Compose()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn(struct{}{})
	}
}
