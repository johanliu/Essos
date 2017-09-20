package components

import "github.com/johanliu/essos"

var ComponentSets = map[string]essos.Component{}

func Add(name string, c essos.Component) {
	ComponentSets[name] = c
}

type DNS struct {
	Enabled bool
	Path    string `toml:"library_path"`
	Api     string `toml:"api_location"`
	Etcd    string `toml:"etcd_address"`
	Domain  string
}

type ConfigManagement struct {
	Enabled bool
	Path    string `toml:"library_path"`
	Api     string `toml:"api_location"`
}

type Pipeline struct {
	Enabled bool
	IP      string
	Port    string
}
