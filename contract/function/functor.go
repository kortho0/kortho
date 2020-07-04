package functor

type MapFunc (func(interface{}) interface{})

type FoldFunc (func(interface{}, interface{}) interface{})

type Functor interface {
	Map(f MapFunc) Functor
	Foldl(f FoldFunc, a interface{}) interface{}
	Foldr(f FoldFunc, a interface{}) interface{}
}

func Map(f MapFunc, a interface{}) interface{} {
	return f(a)
}

func Foldl(f FoldFunc, a interface{}, b interface{}) interface{} {
	return f(a, b)
}

func Foldr(f FoldFunc, a interface{}, b interface{}) interface{} {
	return f(a, b)
}
