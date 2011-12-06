package main

import (
	"gwp/gwp_context"
	"gwp/gwp_template"
	"gwp/gwp_module"
	"gwp/libs/gorilla/mux"
	"gwp/modules/mod_sessions"
	"gwp/modules/mod_example"
	"bytes"
	"net/http"
	"fmt"
)

type Example struct {
	ID string
	Name string
	LoggedIn bool
	ErrorMsg string
}


// initHandlres defines all the routes for our web application
func initHandlers(r *mux.Router) {
	// If gorilla-mux is enabled, r is pointer to mux.Router and we have to use it.
	// Once first match is found, appropriate handler is called, so make sure you 
	// order patterns appropriately. Pattern is regexp.
	// usage would be: r.HandleFunc("/some-regexp-pattern", handlerFn)
	//
	// If gorilla-mux is disabled, you will use default http's way of defining patterns.
	// Patterns are not regexp. Instead, longest match will win, so you don't have to worry
	// about ordering in this case.
	// usage: http.HandleFunc("/pattern", handlerFn)
	//
	// this code is based on sample config provided, where gorilla-mux is enabled, so
	r.HandleFunc("/", indexPage) // otherwise, we would use http.HandleFunc("/", indexPage)
	r.HandleFunc("/login", loginPage)
		
}

func initModules(ctx *gwp_context.Context) {
	// load example module
	example := mod_example.LoadModule()
	gwp_module.RegisterModule(ctx, example)

	// load sessions module
	sess := mod_sessions.LoadModule()
	gwp_module.RegisterModule(ctx, sess)
	
	secretKey := mod_sessions.ReadParamStr("secret-key")
	encKey := mod_sessions.ReadParamStr("encryption-key")
	
        // setup session management. We use filestore as default backend
        mod_sessions.SetStore("filestore", new(mod_sessions.FileSessionStore))
	if len(encKey) == 0 {
        	mod_sessions.SetStoreKeys("filestore", []byte(secretKey))
	} else {
		mod_sessions.SetStoreKeys("filestore", []byte(secretKey), []byte(encKey))
	}
	
}
// checkSession initializes the session, and can also check for specified session parameter
// returns session data and bool if match is found, or just session data
func checkSession(req *http.Request, writer http.ResponseWriter, param ...string) (mod_sessions.SessionData, bool) {
	sess, err := mod_sessions.Session(req, "sf", "filestore")
	
	if err != nil {
		fmt.Println("Session error: ", err.Error())
		return mod_sessions.SessionData{}, false
	}
	mod_sessions.Init(req, writer)
	if len(param) > 0 {
		if _,ok := sess[param[0]]; ok {
			return sess, true
		}
	}
	return sess, false
	
}


// indexPage() is a handler which will load some template and send the result back to the client
func indexPage(writer http.ResponseWriter, req *http.Request) {
	tpl, err := gwp_template.Load(ctx, "index.html")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	var displayContent bool

	// if session parameter "session_id" is present, we already have session set and we can show the content
	// otherwise, we will display login form
	sess,ok := checkSession(req, writer, "session_id")
	if ok {
		displayContent = true
	} else { 
		displayContent = false 
	}
		
	var s_id string
	if sid,ok := sess["session_id"]; ok {
		s_id = sid.(string)
	} else {
		s_id = sess.GetId()
	}

	errmsg := req.FormValue("error")
	var msg string
	if errmsg == "login" { msg = "Invalid login" } 
		
	mydata := Example{ID: s_id, Name: "Joe", LoggedIn: displayContent, ErrorMsg: msg}
	buff := new(bytes.Buffer)

	tpl.Execute(buff, mydata)
	writer.Write(buff.Bytes())
}


// loginPage authenticates users
func loginPage(writer http.ResponseWriter, req *http.Request) {
	// here we define some static record for comparison
	valid_user := "testu"
	valid_pass := "testp"

	if req.FormValue("user") == valid_user && req.FormValue("pass") == valid_pass {
		sess,_ := checkSession(req, writer)
		sess["session_id"] = sess.GetId() // we set this to indicate we're logged in.
		mod_sessions.Save(req, writer) 
		http.Redirect(writer, req, "/", http.StatusFound)
		return
	}
	http.Redirect(writer, req, "/?error=login", http.StatusFound)
}

