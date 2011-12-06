/*
Package mod_example shows how to write 3rd party modules which register their own handlers
*/
package mod_example

import (
        "gwp/gwp_context"
        "gwp/gwp_module"
	"gwp/gwp_template"
	"gwp/modules/mod_sessions"
	"net/http"
	"bytes"
        "os"
        "fmt"
)

// myname represents 'official' module name
var myname = "mod_example"

// myparams is an example of how custom attributes can be exposed to server.conf.
var myparams = &gwp_context.ModParams{
        &gwp_context.ModParam{Name: "test1", Value: "", Default: "", Type: gwp_context.TypeStr, Must: true},
        &gwp_context.ModParam{Name: "test2", Value: "", Default: "testvalue2", Type: gwp_context.TypeStr, Must: false},
}
// M is our global module var
var M *ModExample

// LoadModule is a MUST for every module. It returns Module interface.
func LoadModule() gwp_module.Module {
	M = new(ModExample)
	return M

}

// ModExample is base struct for this module. It will implement Module interface.
type ModExample struct {
        ModCtx *gwp_module.ModContext
}


// ModInit sets the runtime ModContext for this module
func (me *ModExample) ModInit(modCtx *gwp_module.ModContext, err error) {
        if err != nil {
                fmt.Println("Error initializing module:", myname, "-", err.Error())
                os.Exit(1)
        }
        me.ModCtx = modCtx
	
	// we register our handlers here
	gwp_module.RegisterHandler(me.ModCtx.Ctx, "/admin", adminHandler)
}

// GetParams returns *ModParams or nil if we don't want custom parameters in server.conf.
func (me *ModExample) GetParams() *gwp_context.ModParams {
        return myparams
}

// SaveParams updates current ModContext for this module.
func (me *ModExample) SaveParams(params gwp_context.ModParams) {
        me.ModCtx.Params = &params
}

// GetName returns name of the module.
func (me *ModExample) GetName() string {
        return myname
}

// Content type is merged with template
type Content struct {
	ExampleData string
}

// adminHandler function serves content.
func adminHandler(w http.ResponseWriter, r *http.Request) {
	sess,_ := mod_sessions.Session(r, "sf", "filestore")
        tpl, err := gwp_template.Load(M.ModCtx.Ctx, "admin.html")
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        mydata := Content{ExampleData: sess.GetId()}
        buff := new(bytes.Buffer)

        tpl.Execute(buff, mydata)
        w.Write(buff.Bytes())
}
