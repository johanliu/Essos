package components

import "gitlab.mzsvn.com/SRE/essos"

var ComponentSets = map[string]essos.Component{}

func Add(name string, c essos.Component) {
	ComponentSets[name] = c
}
