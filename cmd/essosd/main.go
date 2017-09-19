package essosd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"plugin"
	"regexp"

	"github.com/johanliu/Vidar/constant"
	"github.com/johanliu/essos"
	"github.com/johanliu/mlog"
	"github.com/johanliu/vidar"
	"github.com/johanliu/vidar/middlewares"
)

const (
	pluginDir  = "./components"
	entryPoint = "Entry"
)

type compon struct {
	operations map[string]essos.Operation
}

type essosd struct {
	log        *mlog.Logger
	components map[string]*compon
}

func (e *essosd) loadPlugins() error {
	if _, err := os.Stat(pluginDir); err != nil {
		return err
	}

	ps, err := listFiles(pluginDir, `*.so`)
	if err != nil {
		return err
	}

	for _, p := range ps {
		//Open library in PLUGIN_DIR read from configuration file
		lib, err := plugin.Open(path.Join(pluginDir, p.Name()))
		if err != nil {
			fmt.Printf("failed to open plugin %s: %v\n", p.Name(), err)
			continue
		}

		//Lookup Component variable in plugin which is entry point
		symbol, err := lib.Lookup(entryPoint)
		if err != nil {
			fmt.Printf("plugin %s does not export symbol \"%s\"\n",
				p.Name(), entryPoint)
			continue
		}

		//Validate the symbol loaded from plugin
		component, ok := symbol.(essos.Component)
		if !ok {
			fmt.Printf("Symbol %s (from %s) does not implement Component interface\n",
				entryPoint, p.Name())
			continue
		}

		//Get the operations supported from plugin
		ops := component.Discover()
		if ops == nil {
			fmt.Printf("No operations in %s\n", p.Name())
			continue
		}

		c := new(compon)
		for name, cmd := range ops {
			c.operations[name] = cmd
		}
		e.components[p.Name()] = c
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

type response struct {
	Message string `json:message`
	Code    int    `json:code`
}

type compFunc func(context.Context, []string) (context.Context, error)

func handlerWrapper(cf compFunc) vidar.ContextUserFunc {
	// res = cf()
	res := new(response)

	return func(c *vidar.Context) {
		c.JSON(200, res)
	}
}

func (e *essosd) Run(args ...string) {
	commonHandler := vidar.NewChain()
	v := vidar.New()

	commonHandler.Append(middlewares.LoggingHandler)
	commonHandler.Append(middlewares.RecoverHandler)

	v.Router.Add("GET", "/", commonHandler.Use(indexHandler))
	// v.Router.Add("POST", "/dns/create", commonHandler.Use(dnsCreateHandler))
	v.Router.NotFound = commonHandler.Use(notFoundHandler)

	if err := v.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func main() {

	ParseConfig()

	e := NewEssosd()

	e.loadPlugins()

	if err := e.Run(os.Args[1:]...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func NewEssosd() *essosd {
	return &essosd{
		log: mlog.NewLogger(),
	}
}
