package uidGenerator

type FuncCaller func(int64) []int64

func (f FuncCaller) provide(momentInSecond int64) []int64 {
	return f(momentInSecond)
}

type GenerateId func() int64

func (g GenerateId) assignWorkerId() int64 {
	return g()
}
