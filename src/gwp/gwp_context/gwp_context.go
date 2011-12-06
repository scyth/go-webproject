package gwp_context

import (
	"html/template"
	"gwp/libs/gorilla/mux"
)

var (
        TypeInt     uint8 = 0x01 
        TypeBool    uint8 = 0x02 
        TypeStr     uint8 = 0x03 
        TypeFloat64 uint8 = 0x04 
)

// Context is used to store all runtime app data (modules, templates, configs...)
type Context struct {
	ConfigFile string
	Router     *mux.Router
	LiveTplMsg chan *ParsedTemplate
	ErrorMsg   chan error
	App        *AppConfig
	Templates  map[string]*template.Template // keys = relative file path, vals = parsed template objects
}

// NewContext creates new instance of Context, and returns pointer to it
func NewContext() *Context {
	c := new(Context)
	c.App = NewAppConfig()
	c.LiveTplMsg = make(chan *ParsedTemplate)
	c.ErrorMsg = make(chan error)
	c.Templates = make(map[string]*template.Template)
	return c
}

// AppConfig holds data parsed from configuration file, [default] and [project] sections only
type AppConfig struct {
	ListenAddr    string
	Mux           string
	ProjectRoot   string
	TempDir       string
	TemplatePath  string
	LiveTemplates bool
}

// NewAppConfig creates new instance of AppConfig, and returns pointer to it
func NewAppConfig() *AppConfig {
	ac := new(AppConfig)
	return ac
}

// ParsedTemplate is a wrapper type around template.Template
type ParsedTemplate struct {
	Name string
	Tpl  *template.Template
}


// Param is generic declaration of individual custom config file parameter, defined by modules
type ModParam struct {   
        Name    string
        Value   interface{}
        Default interface{}
        Type    uint8
        Must    bool
}

type ModParams []*ModParam
