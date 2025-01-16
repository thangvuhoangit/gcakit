package gcakit

import "context"

type ExecuteFunc func() (execute func(ctx context.Context) error, interrupt func(error))

type Executor struct {
	Name      string
	Execute   func(ctx context.Context) error
	Interrupt func(error)
}

func NewExecutor(name string, execute func(ctx context.Context) error, interrupt func(error)) *Executor {
	return &Executor{
		Name:      name,
		Execute:   execute,
		Interrupt: interrupt,
	}
}
