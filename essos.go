package essos

import (
	"context"
)

type Operation interface {
	Description() string
	Do(context.Context, []string) (context.Context, error)
}

type Component interface {
	Start(interface{}) error //Used to initialize the component
	Discover() map[string]Operation
	Stop() error //Used to clean the component
}
