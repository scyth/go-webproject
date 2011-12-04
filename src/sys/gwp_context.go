package gwp_context

import (
	"html/template"
)

// Context gets shared across the app and modules
type Context struct {
	ConfigFile string
	LiveTplMsg chan *ParsedTemplate
	ErrorMsg   chan error
	App        *AppConfig
}

func NewContext() *Context {
	c := new(Context)
	c.App = NewAppConfig()
	c.LiveTplMsg = make(chan *ParsedTemplate)
	c.ErrorMsg = make(chan error)
	return c
}

type AppConfig struct {
	ListenAddr    string
	Mux           string
	ProjectRoot   string
	TempDir       string
	TemplatePath  string
	LiveTemplates bool
	Templates     map[string]*template.Template // keys = relative file path, vals = parsed template objects
}

func NewAppConfig() *AppConfig {
	ac := new(AppConfig)
	ac.Templates = make(map[string]*template.Template)
	return ac
}

type ParsedTemplate struct {
	Name string
	Tpl  *template.Template
}
