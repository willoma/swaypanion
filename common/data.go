package common

type Data[T any] interface {
	Equal(other T) bool
}

type Int struct {
	Disabled bool
	Value    int
}

func (i Int) Equal(other Int) bool {
	return i.Disabled == other.Disabled && i.Value == other.Value
}

func IntPoller(fn func() (value int, disabled, ok bool)) func() (Int, bool) {
	return func() (Int, bool) {
		value, disabled, ok := fn()
		return Int{Disabled: disabled, Value: value}, ok
	}
}
