package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/johanliu/essos/components"
	"github.com/labstack/gommon/log"
)

type tomlConfig struct {
	Hostname string
	Override bool `toml:"config-override"`
	Server   struct {
		IP    string
		Port  string
		HTTPS bool `toml:"https-enabled"`
	}
	Logging struct {
		LogPath string `toml:"log_path"`
		Level   string
	}
	Library struct {
		DNS              components.DNS
		ConfigManagement components.ConfigManagement `toml:"cm"`
	}
	RPC struct {
		Pipeline components.Pipeline
	}
}

var tc tomlConfig

func ParseConfig(configFile string) (*tomlConfig, error) {
	f, err := os.Open(configFile)
	if err != nil {
		log.Error(err)
	}

	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Error(err)
	}

	if err := toml.Unmarshal(buf, &tc); err != nil {
		log.Error(err)
	}

	return &tc, nil
}

func main() {
	tc, err := ParseConfig("/etc/essos.conf")
	if err != nil {
		log.Error(err)
	}
	fmt.Printf("%+v", tc)
}
