package gwp_core

import (
	"errors"
	"exp/inotify"
	"gwp/goconf"
	"gwp/gwp_context"
	"os"
	"strings"
)

// ----------------------------------------
// Runtime config operations API
// ----------------------------------------

const (
	dflt_conf_addr    = "127.0.0.1:8000"
	dflt_conf_mux     = true
	dflt_conf_tmpdir  = "/tmp/"
	dflt_conf_livetpl = false
)

// ParseConfig parses the configuration file and does meaningful checks on defined parameters
// If optional parameters are not met, it sets default values
// It parses only [default] and [project] sections
func ParseConfig(configPath string) (*gwp_context.AppConfig, error) {
	ac := gwp_context.NewAppConfig()

	// config file must parse successfully
	c, err := goconf.ReadConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	// read params from [default] section
	conf_addr, err := c.GetString("default", "listen")
	if err != nil {
		conf_addr = dflt_conf_addr
	}

	conf_mux, err := c.GetBool("default", "gorilla-mux")
	if err != nil {
		conf_mux = dflt_conf_mux
	}

	// read params from [project] section
	conf_root, err := c.GetString("project", "root")
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(conf_root, "/") {
		conf_root += "/"
	}

	conf_tmpdir, err := c.GetString("project", "tmpDir")
	if err != nil {
		conf_tmpdir = dflt_conf_tmpdir
	}
	if !strings.HasSuffix(conf_tmpdir, "/") {
		conf_tmpdir += "/"
	}

	conf_template_path, err := c.GetString("project", "templatePath")
	if err != nil {
		conf_template_path = conf_root + "templates/"
	}
	if !strings.HasSuffix(conf_template_path, "/") {
		conf_template_path += "/"
	}

	conf_livetpl, err := c.GetBool("project", "live-templates")
	if err != nil {
		conf_livetpl = dflt_conf_livetpl
	}

	testpath := conf_tmpdir + "go-webproject_tmptest"
	if err := os.Mkdir(testpath, 0755); err != nil {
		return nil, errors.New("Error with tmp dir configuration: " + err.Error())
	} else {
		os.Remove(testpath)
	}

	p := strings.TrimSpace(conf_template_path)
	// check if path exists
	if _, err := os.Stat(p); err != nil {
		return nil, errors.New("Configuration error, template directory does not exist: " + conf_template_path)
	}

	ac.ListenAddr = conf_addr
	if conf_mux {
		ac.Mux = "gorilla"
	} else {
		ac.Mux = "default"
	}
	ac.ProjectRoot = conf_root
	ac.TempDir = conf_tmpdir
	ac.TemplatePath = conf_template_path
	ac.LiveTemplates = conf_livetpl
	return ac, nil
}

// ----------------------------------------
// Runtime template operations and API
// ----------------------------------------

var (
	WatchList map[string]bool
)

// WatchTemplates is responsible for template caching
// and live reloading (if live-templates option is activated)
func WatchTemplates(ctx *gwp_context.Context) {
	// we're tracking live changes to template files
	if ctx.App.LiveTemplates == true {
		watcher, err := inotify.NewWatcher()
		if err != nil {
			ctx.ErrorMsg <- errors.New("Could not create inotify watcher: " + err.Error())
			return
		}
		defer watcher.Close()

		WatchList = make(map[string]bool)

		for {
			select {
			case ev := <-watcher.Event:
				// cached file was modified
				if ctx.App.Templates[ev.Name] != nil {
					delete(ctx.App.Templates, ev.Name)
				}
				if WatchList[ev.Name] == true {
					watcher.RemoveWatch(ev.Name)
					WatchList[ev.Name] = false
				}

			case ev := <-watcher.Error:
				// this probably means something has gone terribly wrong, so we exit
				ctx.ErrorMsg <- ev
				return

			case ev := <-ctx.LiveTplMsg:
				ctx.App.Templates[ev.Name] = ev.Tpl

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
			ev := <-ctx.LiveTplMsg
			ctx.App.Templates[ev.Name] = ev.Tpl
		}
	}

}
