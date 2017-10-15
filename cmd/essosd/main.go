package main

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"plugin"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/johanliu/essos"
	"github.com/johanliu/essos/cmd"
	"github.com/johanliu/essos/components"
	"github.com/johanliu/mlog"
	"github.com/johanliu/vidar"
	"github.com/johanliu/vidar/plugins"
)

const configPath = "/etc/essos.conf"

type essosd struct {
	log        *mlog.Logger
	components map[string]essos.Component
	server     *vidar.Vidar
	chain      *vidar.Plugin
}

func (e *essosd) loadPlugins(pluginDir string, li cmd.LibraryInfo) error {
	if _, err := os.Stat(pluginDir); err != nil {
		return err
	}

	ps, err := listFiles(pluginDir, `.so`)
	if err != nil {
		return err
	}

	// for all plugins
	for _, p := range ps {
		name := strings.Split(p.Name(), ".")[0]

		conf := reflect.ValueOf(li).FieldByName(strings.Title(name))

		if !conf.IsValid() {
			e.log.Warning("Component %s is not found in configs file", name)
			continue
		}

		if !conf.FieldByName("Enabled").Bool() {
			e.log.Warning("Component %s is not enabled", name)
			continue
		}

		// Open library in PLUGIN_DIR read from configuration file
		_, err := plugin.Open(path.Join(pluginDir, p.Name()))
		if err != nil {
			e.log.Warning("Failed to open plugin %s: %v\n", name, err)
			continue
		}

		// Validate the object loaded from plugin
		object, ok := components.ComponentSets[name]
		if !ok {
			e.log.Warning("Faile to start plugin %s", name)
			continue
		}

		component, ok := object.(essos.Component)
		if !ok {
			e.log.Warning("Object %s (from %s) does not implement Component interface %v\n",
				object, name, ok)
			continue
		}

		if err := component.Start(conf.Interface()); err != nil {
			e.log.Error(err)
		}

		//Get the operations supported from plugin
		ops := component.Discover()
		if ops == nil {
			e.log.Warning("No operations in %s\n", name)
			continue
		}

		e.components[name] = object
	}

	return nil
}

func (e *essosd) stopPlugins() {
	wg := sync.WaitGroup{}

	for _, c := range e.components {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			c.Stop()
			wg.Done()
		}(&wg)
	}

	wg.Wait()
}

func staticResource(root string) vidar.ContextFunc {
	return func(ctx *vidar.Context) {
		ctx.File(root)
	}
}

func (e *essosd) renderPortal(prefix, root string) error {
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

func notFoundHandler(c *vidar.Context) {
	c.Error(vidar.NotFoundError)
}

type compFunc func(context.Context, []string) (context.Context, error)

func (e *essosd) handlerWrapper(cf compFunc) vidar.ContextFunc {
	return func(c *vidar.Context) {
		input := c.Body()
		ctxArgs := context.WithValue(context.Background(), "input", input)

		ctxReturn, err := cf(ctxArgs, nil)
		if err != nil {
			e.log.Error(err)
			c.Error(err)
		} else {
			if ctxReturn.Value("result") != nil {
				result := ctxReturn.Value("result")
				e.log.Info("Result return by caller: %+v", result)
				c.JSON(200, result)
			} else {
				e.log.Info("No result is returned")
				c.Error(vidar.NotImplementedError)
			}
		}
	}
}

func (e *essosd) addHandler() error {
	for componentName, component := range e.components {
		operations := component.Discover()
		e.log.Info("componentName: %s\n", componentName)

		for methodName, method := range operations {
			e.log.Info("methodName: %s\n", methodName)
			e.server.Router.POST(
				strings.Join([]string{"", componentName, methodName, ""}, "/"),
				e.chain.Apply(e.handlerWrapper(method.Do)),
			)
		}
	}

	return nil
}

func (e *essosd) runServer(args ...string) error {
	if err := e.server.Run(); err != nil {
		e.log.Error(err)
		return err
	}

	return nil
}

func main() {
	//TODO: To be refactor
	e := NewEssosd()

	tc, err := cmd.ParseConfig(configPath)
	if err != nil {
		e.log.Error(err)
	}

	if err := e.loadPlugins(tc.LibraryPath, tc.Library); err != nil {
		e.log.Error(err)
	}

	defer e.stopPlugins()

	e.chain.Append(plugins.SlashHandler)
	e.chain.Append(plugins.LoggingHandler)
	e.chain.Append(plugins.RecoverHandler)

	if err := e.addHandler(); err != nil {
		e.log.Error(err)
	}

	/*
		if err := e.renderPortal("/portal", "portal"); err != nil {
			e.log.Error(err)
		}
	*/

	e.server.Router.NotFound = e.chain.Apply(notFoundHandler)

	// fmt.Println(e.server.Router.ShowHandler())

	if err := e.runServer(os.Args[1:]...); err != nil {
		e.log.Error(err)
		os.Exit(1)
	}
}

func NewEssosd() *essosd {
	return &essosd{
		log:        mlog.NewLogger(),
		components: map[string]essos.Component{},
		server:     vidar.New(),
		chain:      vidar.NewPlugin(),
	}
}
