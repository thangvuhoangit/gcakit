package gcakit

import "context"

type AppContext struct {
	context.Context
	*App
}
