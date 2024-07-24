package common

type Data[T any] interface {
	Equal(other T) bool
}

type Int struct {
	Value int
}

func (i Int) Equal(other Int) bool {
	return i.Value == other.Value
}

func IntPoller(fn func() (int, bool)) func() (Int, bool) {
	return func() (Int, bool) {
		value, ok := fn()
		return Int{Value: value}, ok
	}
}
