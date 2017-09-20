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

type Response struct {
	Message []byte
	Code    int
}
