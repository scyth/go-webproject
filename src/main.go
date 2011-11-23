package main

import (
	"../goconf/goconf"
	"../gorilla/mux/mux"
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
}

type AppConfig struct {
	ListenAddr   string
	ProjectRoot  string
	TempDir      string
	TemplatePath string
	Templates    map[string]*template.Template // keys = relative file path, vals = parsed template objects
}

func NewAppConfig() *AppConfig {
	ac := new(AppConfig)
	ac.Templates = make(map[string]*template.Template)
	return ac
}

func main() {
	// we read the config file
	loadConfig(ac)

	// we setup our url handlers - see handlers.go
	initHandlers(router)
	http.Handle("/", router)

	// serve the world
	err := http.ListenAndServe(ac.ListenAddr, nil)
	if err != nil {
		fmt.Printf("Failed to create listener: %s \n", err.Error())
		os.Exit(1)
	}
}

func loadTemplate(name string) (tpl *template.Template, err error) {
	if ac.Templates[name] != nil {
		return ac.Templates[name], nil
	}

	tpl, err = template.ParseFile(ac.TemplatePath + name)
	if err != nil {
		fmt.Println("Error loading template", err.Error())
		return nil, err
	}
	ac.Templates[name] = tpl
	return tpl, nil
}

func loadConfig(ac *AppConfig) {
	c, err := goconf.ReadConfigFile(configPath)
	checkConfigError(err)

	conf_root, err := c.GetString("default", "projectRoot")
	checkConfigError(err)

	conf_addr, err := c.GetString("default", "listen")
	checkConfigError(err)

	conf_template_path, err := c.GetString("default", "templatePath")
	checkConfigError(err)

	conf_tmpdir, err := c.GetString("default", "tmpDir")
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

}

func checkConfigError(e error) {
	if e != nil {
		fmt.Println("Config error: ", e.Error())
		os.Exit(1)
	}
}
