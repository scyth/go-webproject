package main

import (
	"gwp/gwp_context"
	"gwp/gwp_core"
	"gwp/libs/gorilla/mux"
	"gwp/modules/gorilla/sessions"
	"flag"
	"fmt"
	"net/http"
	"os"
)

var (
	configPath string
	ctx        *gwp_context.Context
	router     *mux.Router
)

const (
	dflt_conf_addr    = "127.0.0.1:8000"
	dflt_conf_mux     = true
	dflt_conf_tmpdir  = "/tmp/"
	dflt_conf_livetpl = false
)

func init() {
	// set global context
	ctx = gwp_context.NewContext()

	// parse command line for config path
	flag.StringVar(&configPath, "config", "config/server.conf", "path to configuration file")
	flag.Parse()
	_, err := os.Stat(configPath)
	if err != nil {
		fmt.Printf("Error, config file does not exist: %s\n", configPath)
		os.Exit(1)
	}
	ctx.ConfigFile = configPath

	// setup session management. We use filestore as default backend
	sessions.SetStore("filestore", new(sessions.FileSessionStore))
	sessions.SetStoreKeys("filestore", []byte("my-simple-key-hmac"))

}

func main() {
	// parse the config file and set the context
	appconf, err := gwp_core.ParseConfig(ctx.ConfigFile)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("See examples/config/server.conf for all the options")
		os.Exit(1)
	}
	ctx.App = appconf

	// if gorilla-mux is not set, we will use default methods from http package
	if ctx.App.Mux == "gorilla" {
		router = new(mux.Router)
		initHandlers(router)
		http.Handle("/", router)
	} else {
		initHandlers(nil)
	}

	// run the watcher for templates
	go gwp_core.WatchTemplates(ctx)

	// serve the world
	err = http.ListenAndServe(ctx.App.ListenAddr, nil)
	if err != nil {
		fmt.Printf("Failed to create listener: %s \n", err.Error())
		os.Exit(1)
	}

	// run the watcher for templates
	go gwp_core.WatchTemplates(ctx)
	err = <-ctx.ErrorMsg
	fmt.Println("Aborting runtime. Got error:", err.Error())
}
