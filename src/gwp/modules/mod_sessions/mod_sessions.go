package mod_sessions

import (
	"gwp/gwp_context"
	"gwp/gwp_module"
	"os"
	"fmt"
)

// myname represents 'official' module name
var myname = "mod_sessions"

// myparams is an example of how custom attributes can be exposed to server.conf.
var myparams = &gwp_context.ModParams{
        &gwp_context.ModParam{Name: "secret-key", Value: "", Default: "", Type: gwp_context.TypeStr, Must: true},
	&gwp_context.ModParam{Name: "encryption-key", Value: "", Default: "", Type: gwp_context.TypeStr, Must: false},
}

var M *ModSessions

// LoadModule is a MUST for every module. It returns Module interface.
func LoadModule() gwp_module.Module {
	M = new(ModSessions)
	return M
}

// ModSessions is base struct for this module. It will implement Module interface.
type ModSessions struct {
	ModCtx *gwp_module.ModContext
}


// ModInit sets the runtime ModContext for this module
func (ms *ModSessions) ModInit(modCtx *gwp_module.ModContext, err error) {
	if err != nil {
		fmt.Println("Error initializing module:", myname, "-", err.Error())
		os.Exit(1)
	}
	ms.ModCtx = modCtx
}

// GetParams returns *ModParams or nil if we don't want custom parameters in server.conf.
func (ms *ModSessions) GetParams() *gwp_context.ModParams {
        return myparams
}

// SaveParams updates current ModContext for this module.
func (ms *ModSessions) SaveParams(params gwp_context.ModParams) {
	ms.ModCtx.Params = &params
}

// GetName returns name of the module.
func (ms *ModSessions) GetName() string {
	return myname
}

// ReadParamStr returns named parameter value from ModContext.
func ReadParamStr(name string) string {
	for _,v := range *M.ModCtx.Params {
		if v.Name == name {
			return v.Value.(string)
		}
	}
	return ""
}
