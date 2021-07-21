package uidGenerator

type FuncCaller func(int64) []int64

func (f FuncCaller) provide(momentInSecond int64) []int64 {
	return f(momentInSecond)
}
