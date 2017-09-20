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

	"github.com/johanliu/Vidar/constant"
	"github.com/johanliu/essos"
	"github.com/johanliu/essos/cmd"
	"github.com/johanliu/essos/components"
	"github.com/johanliu/mlog"
	"github.com/johanliu/vidar"
	"github.com/johanliu/vidar/middlewares"
)

const configPath = "/etc/essos.conf"

type essosd struct {
	log        *mlog.Logger
	components map[string]essos.Component
	server     *vidar.Vidar
}

func (e *essosd) loadPlugins(pluginDir string, li cmd.LibraryInfo) error {
	if _, err := os.Stat(pluginDir); err != nil {
		return err
	}

	ps, err := listFiles(pluginDir, `.so`)
	if err != nil {
		return err
	}

	for _, p := range ps {
		name := strings.Split(p.Name(), ".")[0]

		condition := reflect.ValueOf(li).FieldByName(strings.Title(name))

		if !condition.IsValid() {
			e.log.Warning("Component %s is not included in configs file", name)
			continue
		}

		if !condition.FieldByName("Enabled").Bool() {
			e.log.Warning("Component %s is not enabled", name)
			continue
		}

		//Open library in PLUGIN_DIR read from configuration file
		_, err := plugin.Open(path.Join(pluginDir, p.Name()))
		if err != nil {
			e.log.Warning("Failed to open plugin %s: %v\n", name, err)
			continue
		}

		//Validate the object loaded from plugin
		object := components.ComponentSets[name]
		component, ok := object.(essos.Component)
		if !ok {
			e.log.Warning("Object %s (from %s) does not implement Component interface\n",
				object, name)
			continue
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

func (e *essosd) loadRPC(cmd.RPCInfo) error {
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
	c.Error(constant.NotFoundError)
}

type compFunc func(context.Context, []string) (context.Context, error)

func (e *essosd) handlerWrapper(cf compFunc) vidar.ContextUserFunc {
	return func(c *vidar.Context) {
		formValues, err := c.FormParams()
		if err != nil {
			e.log.Error(err)
		}
		ctxArgs := context.WithValue(context.Background(), "Form", formValues)

		// TODO: to be implemented
		args := []string{"hello", "world"}

		ctxReturn, err := cf(ctxArgs, args)
		if err != nil {
			e.log.Error(err)
		}

		result := ctxReturn.Value("result").(essos.Response)

		c.JSON(result.Code, result.Message)
	}
}

func (e *essosd) addHandler() error {
	commonHandler := vidar.NewChain()
	commonHandler.Append(middlewares.LoggingHandler)
	commonHandler.Append(middlewares.RecoverHandler)

	for componentName, component := range e.components {
		operations := component.Discover()
		e.log.Info("componentName: %s\n", componentName)

		for methodName, method := range operations {
			e.log.Info("methodName: %s\n", methodName)
			e.server.Router.Add(
				"POST",
				strings.Join([]string{"", componentName, methodName}, "/"),
				e.handlerWrapper(method.Do),
			)
		}
	}

	e.server.Router.NotFound = commonHandler.Use(notFoundHandler)

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

	if err := e.loadRPC(tc.RPC); err != nil {
		e.log.Error(err)
	}

	if err := e.addHandler(); err != nil {
		e.log.Error(err)
	}

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
	}
}
