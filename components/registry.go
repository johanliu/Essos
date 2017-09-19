package components

import "github.com/johanliu/essos"

type Init func() essos.Component

var Components = map[string]Init{}

func Add(name string, init Init) {
	Components[name] = init
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
