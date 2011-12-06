package gwp_module

import (
	"gwp/gwp_context"
	"gwp/gwp_core"
	"net/http"
)


// Module interface
type Module interface {
	ModInit(*ModContext, error)
	GetName() string
	GetParams() (*gwp_context.ModParams)
	SaveParams(gwp_context.ModParams) 
	
}


// ModContext is passed back to module after registration
type ModContext struct {
	Name    string                 // module name
	Ctx     *gwp_context.Context   // pointer to global context
	Params  *gwp_context.ModParams // parsed parameters
}


// RegisterModule takes Module interface and registers the module within global Context.
// It calls *Module.ModInit() passing the ModContext, or nil if there as an error.
func RegisterModule(ctx *gwp_context.Context, m Module) {
	modctx := new(ModContext)
	modctx.Name = m.GetName()
	modctx.Ctx = ctx
	modctx.Params = m.GetParams()
	if modctx.Params != nil {
		err := gwp_core.ParseConfigParams(ctx.ConfigFile, modctx.Name, m.GetParams())
		if err != nil {
			m.ModInit(nil, err)
		}
	}
	m.ModInit(modctx, nil)
}

// RegisterHandler can be called to register handlers directly from modules.
// It takes standard http's(or mux's) pattern and a HandlerFunc as arguments, 
// along with a pointer to the global Context.
func RegisterHandler(ctx *gwp_context.Context, pattern string, 
	handler func(http.ResponseWriter, *http.Request)) {
	
	if ctx.App.Mux == "gorilla" {
		ctx.Router.HandleFunc(pattern, handler)
	} else {
		http.HandleFunc(pattern, handler)
	}
}
