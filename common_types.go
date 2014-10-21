package cp

func init() {
	AddFuncType((*func(int))(nil), func(impls []interface{}) interface{} {
		return func(i int) {
			for _, impl := range impls {
				impl.(func(int))(i)
			}
		}
	})
	AddFuncType((*func(string))(nil), func(impls []interface{}) interface{} {
		return func(s string) {
			for _, impl := range impls {
				impl.(func(string))(s)
			}
		}
	})
	AddFuncType((*func(bool))(nil), func(impls []interface{}) interface{} {
		return func(b bool) {
			for _, impl := range impls {
				impl.(func(bool))(b)
			}
		}
	})
	AddFuncType((*func())(nil), func(impls []interface{}) interface{} {
		return func() {
			for _, impl := range impls {
				impl.(func())()
			}
		}
	})
}
