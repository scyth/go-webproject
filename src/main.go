package main

import (
	"../goconf/goconf"
	"../gorilla/mux/mux"
	"../gorilla/sessions/sessions"
	"exp/inotify"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
)

var (
	configPath string
	ac         *AppConfig
	router     *mux.Router
	WatchList  map[string]bool
)

const (
	dflt_conf_addr    = "127.0.0.1:8000"
	dflt_conf_mux     = true
	dflt_conf_tmpdir  = "/tmp/"
	dflt_conf_livetpl = false
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
	sessions.SetStore("filestore", new(sessions.FileSessionStore))
	sessions.SetStoreKeys("filestore", []byte("my-simple-key-hmac"))

}

type AppConfig struct {
	ListenAddr    string
	Gorilla       bool
	ProjectRoot   string
	TempDir       string
	TemplatePath  string
	LiveTemplates bool
	LiveMsg       chan *ParsedTemplate
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
	go watchTemplates()

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

// watchTemplates is responsible for template caching
// and live reloading (if live-templates option is activated)
func watchTemplates() {
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
				// this probably means something has gone terribly wrong, so we exit
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
		// we're just preloading/caching templates. No runtime updates are possible.
	} else {

		for {
			ev := <-ac.LiveMsg
			ac.Templates[ev.Name] = ev.Tpl
		}
	}

}

// LoadTemplate is API call which will return parsed template object, and will do this fast.
// It is also thread safe
func LoadTemplate(name string) (tpl *template.Template, err error) {
	if ac.Templates[ac.TemplatePath+name] != nil {
		return ac.Templates[ac.TemplatePath+name], nil
	}

	tpl, err = template.ParseFiles(ac.TemplatePath + name)
	if err != nil {
		fmt.Println("Error loading template", err.Error())
		return nil, err
	}
	pt := &ParsedTemplate{ac.TemplatePath + name, tpl}

	ac.LiveMsg <- pt
	return tpl, nil
}

// loadConfig parses the configuration file and does meaningful checks on defined parameters
// Also sets default values if optional parameters are not set
func loadConfig(ac *AppConfig) {
	// config file must parse successfully
	c, err := goconf.ReadConfigFile(configPath)
	checkConfigError(err, true)

	// read params from [default] section
	conf_addr, err := c.GetString("default", "listen")
	checkConfigError(err, false)
	if err != nil {
		conf_addr = dflt_conf_addr
	}

	conf_mux, err := c.GetBool("default", "gorilla-mux")
	checkConfigError(err, false)
	if err != nil {
		conf_mux = dflt_conf_mux
	}

	// read params from [project] section
	conf_root, err := c.GetString("project", "root")
	checkConfigError(err, true)
	if !strings.HasSuffix(conf_root, "/") {
		conf_root += "/"
	}

	conf_tmpdir, err := c.GetString("project", "tmpDir")
	checkConfigError(err, false)
	if err != nil {
		conf_tmpdir = dflt_conf_tmpdir
	}
	if !strings.HasSuffix(conf_tmpdir, "/") {
		conf_tmpdir += "/"
	}

	conf_template_path, err := c.GetString("project", "templatePath")
	checkConfigError(err, false)
	if err != nil {
		conf_template_path = conf_root + "templates/"
	}
	if !strings.HasSuffix(conf_template_path, "/") {
		conf_template_path += "/"
	}

	conf_livetpl, err := c.GetBool("project", "live-templates")
	checkConfigError(err, false)
	if err != nil {
		conf_livetpl = dflt_conf_livetpl
	}

	testpath := conf_tmpdir + "go-webproject_tmptest"
	if err := os.Mkdir(testpath, 0755); err != nil {
		fmt.Println("Error with tmp dir configuration:", err.Error())
		os.Exit(1)
	} else {
		os.Remove(testpath)
	}

	p := strings.TrimSpace(conf_template_path)
	// check if path exists
	if _, err := os.Stat(p); err != nil {
		fmt.Println("Configuration error, template directory does not exist: ", conf_template_path)
		os.Exit(1)
	}

	ac.ListenAddr = conf_addr
	ac.Gorilla = conf_mux
	ac.ProjectRoot = conf_root
	ac.TempDir = conf_tmpdir
	ac.TemplatePath = conf_template_path
	ac.LiveTemplates = conf_livetpl

}

// checkConfigError checks for configuration errors. If strict parameter is true, 
// we can't proceed with execution
func checkConfigError(e error, strict bool) {
	if e != nil && strict == true {
		fmt.Println("Config error: ", e.Error())
		fmt.Println("See examples/config/server.conf for all the options")
		os.Exit(1)
	}
}
