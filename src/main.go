package main

import (
	"../goconf/goconf"
	"../gorilla/mux/mux"
	"../gorilla/sessions/sessions"
	"exp/inotify"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"
)

var (
	configPath string
	ac         *AppConfig
	router     *mux.Router
	WatchList  map[string]bool
)

func init() {
	flag.StringVar(&configPath, "config", "config/server.conf", "path to configuration file")
	flag.Parse()
	_, err := os.Stat(configPath)
	if err != nil {
		fmt.Printf("Error, config file does not exist: %s\n", configPath)
		os.Exit(1)
	}
	ac = NewAppConfig()
	router = new(mux.Router)

	WatchList = make(map[string]bool)

	sessions.SetStoreKeys("cookie", []byte("my-simple-key-hmac"))

}

type AppConfig struct {
	ListenAddr    string
	Gorilla       bool
	ProjectRoot   string
	TempDir       string
	TemplatePath  string
	LiveTemplates bool
	LiveMsg	      chan *ParsedTemplate
	Templates     map[string]*template.Template // keys = relative file path, vals = parsed template objects
}

func NewAppConfig() *AppConfig {
	ac := new(AppConfig)
	ac.LiveMsg = make(chan *ParsedTemplate)
	ac.Templates = make(map[string]*template.Template)
	return ac
}

func main() {
	// we read the config file
	loadConfig(ac)

	if ac.Gorilla == true {
		initHandlers(router)
		http.Handle("/", router)
	} else {
		initHandlers(nil)
	}

	// run the watcher for templates
	go WatchTemplates()

	// serve the world
	err := http.ListenAndServe(ac.ListenAddr, nil)
	if err != nil {
		fmt.Printf("Failed to create listener: %s \n", err.Error())
		os.Exit(1)
	}
}


type ParsedTemplate struct {
	Name string
	Tpl  *template.Template
}


func WatchTemplates() {
	// we're tracking live changes to template files
	if ac.LiveTemplates == true {
		watcher, err := inotify.NewWatcher()
		if err != nil {
			fmt.Println("Could not create inotify watcher: ", err.Error())
			return
		}
		defer watcher.Close()

		for {
			select {
			case ev := <-watcher.Event:
				// cached file was modified
				if ac.Templates[ev.Name] != nil {
					delete(ac.Templates, ev.Name)
				}
				if WatchList[ev.Name] == true {
					watcher.RemoveWatch(ev.Name)
					WatchList[ev.Name] = false
				}

			case ev := <-watcher.Error:
				fmt.Println("Error from inotify.Watcher: ", ev)
				os.Exit(1)
			case ev := <-ac.LiveMsg:
				ac.Templates[ev.Name] = ev.Tpl

				// check if we're already watching this file name
				if WatchList[ev.Name] == true {
					watcher.RemoveWatch(ev.Name)
					watcher.AddWatch(ev.Name, inotify.IN_MODIFY)
				} else {
					watcher.AddWatch(ev.Name, inotify.IN_MODIFY)
					WatchList[ev.Name] = true
				}
			}
		}
	} else {

		for {
			ev := <-ac.LiveMsg
			ac.Templates[ev.Name] = ev.Tpl
		}
	}


}


func loadTemplate(name string) (tpl *template.Template, err error) {
	if ac.Templates[ac.TemplatePath + name] != nil {
		return ac.Templates[ac.TemplatePath + name], nil
	}

	tpl, err = template.ParseFile(ac.TemplatePath + name)
	if err != nil {
		fmt.Println("Error loading template", err.Error())
		return nil, err
	}
	pt := &ParsedTemplate{ac.TemplatePath + name, tpl}

	ac.LiveMsg <- pt
	return tpl, nil
}

func loadConfig(ac *AppConfig) {
	c, err := goconf.ReadConfigFile(configPath)
	checkConfigError(err)

	conf_root, err := c.GetString("project", "root")
	checkConfigError(err)

	conf_addr, err := c.GetString("default", "listen")
	checkConfigError(err)

	conf_template_path, err := c.GetString("project", "templatePath")
	checkConfigError(err)

	conf_mux, err := c.GetBool("default", "gorilla-mux")
	checkConfigError(err)

	conf_tmpdir, err := c.GetString("project", "tmpDir")
	checkConfigError(err)

	conf_livetpl, err := c.GetBool("project", "live-templates")
	checkConfigError(err)

	// check if we have write access to temp dir
	if strings.HasSuffix(conf_tmpdir, "/") {
		ac.TempDir = conf_tmpdir
	} else {
		ac.TempDir = conf_tmpdir + "/"
	}

	testpath := ac.TempDir + "go-webproject_tmptest"
	if err := os.Mkdir(testpath, 0755); err != nil {
		fmt.Println("Error with tmp dir configuration:", err.Error())
		os.Exit(1)
	} else {
		os.Remove(testpath)
	}

	p := strings.TrimSpace(conf_template_path)
	// check if path exists
	if _, err := os.Stat(p); err != nil {
		fmt.Println("Configuration error, template directory does not exist")
		os.Exit(1)
	}
	if strings.HasSuffix(p, "/") {
		ac.TemplatePath = p
	} else {
		ac.TemplatePath = p + "/"
	}

	ac.ListenAddr = conf_addr
	ac.TempDir = conf_tmpdir
	ac.ProjectRoot = conf_root
	ac.Gorilla = conf_mux
	ac.LiveTemplates = conf_livetpl

}

func checkConfigError(e error) {
	if e != nil {
		fmt.Println("Config error: ", e.Error())
		os.Exit(1)
	}
}
