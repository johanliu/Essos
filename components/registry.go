package components

import "github.com/johanliu/essos"

type Init func() essos.Component

var Components = map[string]Init{}

func Add(name string, init Init) {
	Components[name] = init
}
