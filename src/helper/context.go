package helper

import (
	c "context"
)

var context c.Context

func init() {
	context = c.Background()
}

func GetContext() c.Context {
	return context
}