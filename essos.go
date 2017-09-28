package essos

import (
	"context"
)

type Operation interface {
	Description() string
	Do(context.Context, []string) (context.Context, error)
}

type Component interface {
	Discover() map[string]Operation
}

type Rpc interface {
	InitConnection(string, string) (error)
}

type RpcComponent interface {
	NewConnection(string, int)					// ip:port
	Discover() map[string]Operation
}

type Response struct {
	Message []byte
	Code    int
}
