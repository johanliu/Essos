package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"

	"github.com/johanliu/Vidar/constant"
	"github.com/johanliu/mlog"
	"github.com/johanliu/vidar"
	"github.com/johanliu/vidar/middlewares"
)

const (
	PLUGIN_DIR  = "./components"
	ENTRY_POINT = "Entry"
)

type component struct {
	operations map[string]essos.Operation
}

type essos struct {
	log        *mlog.Logger
	components map[string]*component
}

func (e *essos) loadPlugins() error {
	if _, err := os.Stat(PLUGIN_DIR); err != nil {
		return err
	}

	plugins, err := listFiles(PLUGIN_DIR, `*.so`)
	if err != nil {
		return err
	}

	for _, plugin := range plugins {
		//Open library in PLUGIN_DIR read from configuration file
		lib, err := plugin.Open(path.Join(PLUGIN_DIR, plugin.Name()))
		if err != nil {
			fmt.Printf("failed to open plugin %s: %v\n", plugin.Name(), err)
			continue
		}

		//Lookup Component variable in plugin which is entry point
		symbol, err := lib.Lookup(ENTRY_POINT)
		if err != nil {
			fmt.Printf("plugin %s does not export symbol \"%s\"\n",
				plugin.Name(), ENTRY_POINT)
			continue
		}

		//Validate the symbol loaded from plugin
		component, ok := symbol.(essos.Component)
		if !ok {
			fmt.Printf("Symbol %s (from %s) does not implement Component interface\n",
				ENTRY_POINT, plugin.Name())
			continue
		}

		//Get the operations supported from plugin
		ops := component.Discover()
		if ops == nil {
			fmt.Printf("No operations in %s\n", plugin.Name())
			continue
		}

		c := new(component)
		for name, cmd := range ops {
			c.operations[name] = cmd
		}
		e.components[plugin.Name()] = c
	}

	return nil
}

func listFiles(dir, pattern string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	filteredFiles := []os.FileInfo{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		matched, err := regexp.MatchString(pattern, file.Name())
		if err != nil {
			return nil, err
		}
		if matched {
			filteredFiles = append(filteredFiles, file)
		}
	}
	return filteredFiles, nil
}

func indexHandler(c *vidar.Context) {
	c.Text(200, "HELLO")
}

func notFoundHandler(c *vidar.Context) {
	c.Error(constant.NotFoundError)
}

func dnsHandler(c *vidar.Context) {
	loadPlugins()
}

func main() {
	commonHandler := vidar.NewChain()
	v := vidar.New()

	commonHandler.Append(middlewares.LoggingHandler)
	commonHandler.Append(middlewares.RecoverHandler)

	v.Router.Add("GET", "/", commonHandler.Use(indexHandler))
	v.Router.Add("POST", "/dns/create", commonHandler.Use(dnsHandler))
	v.Router.Add("POST", "/dns/read", commonHandler.Use(dnsHandler))
	v.Router.Add("POST", "/dns/update", commonHandler.Use(dnsHandler))
	v.Router.Add("POST", "/dns/delete", commonHandler.Use(dnsHandler))
	v.Router.NotFound = commonHandler.Use(notFoundHandler)

	v.Run()
}

func NewEssos() *essos {
	return &essos{
		log: mlog.NewLogger(),
	}
}
