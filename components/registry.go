package components

import "github.com/johanliu/essos"

var ComponentSets = map[string]essos.Component{}

func Add(name string, c essos.Component) {
	ComponentSets[name] = c
}
