package mod_sessions

import (
	"os"
	"fmt"
	"net/http"
	"github.com/scyth/go-webproject/gwp/gwp_context"
	"github.com/scyth/go-webproject/gwp/gwp_module"
	"github.com/scyth/go-webproject/gwp/libs/gorilla/sessions"
	"github.com/scyth/go-webproject/gwp/libs/gorilla/securecookie"
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
	M.Store = new(sessions.FilesystemStore)
	return M
}

// ModSessions is base struct for this module. It will implement Module interface.
type ModSessions struct {
	ModCtx *gwp_module.ModContext
	Store *sessions.FilesystemStore
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

// RegisterStore registers a session store. This module uses FilesystemStore
func RegisterStore(keyPairs ...[]byte) {
	store := sessions.NewFilesystemStore("", keyPairs...)
	M.Store = store
}


// GetSession returns a session
func GetSession(r *http.Request, session_name string) (*sessions.Session, error) {
	s, err := M.Store.Get(r, session_name)
	if s.ID == "" {
		k := securecookie.GenerateRandomKey(24)
		s.ID = fmt.Sprintf("%x", k)
	}
	return s, err
}

// Save calls sessions.Save
func Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	return M.Store.Save(r, w, s)
}


// checkSession initializes the session, and can also check for specified session parameter
// returns session data and bool if match is found, or just session data
func CheckSession(req *http.Request, writer http.ResponseWriter, param ...string) (*sessions.Session, bool) {
        sess, err := GetSession(req, "sf")
        
        if err != nil {
                fmt.Println("Session error: ", err.Error())
                return sess, false
        }
        if len(param) > 0 {
                if _,ok := sess.Values[param[0]]; ok {
                        return sess, true
                }
        }
        return sess, false
}
