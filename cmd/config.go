package vidar

import (
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/labstack/gommon/log"
)

type tomlConfig struct {
	Components []string
}

const configFile string = "essos.conf"

var tc tomlConfig

func init() {
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
}
