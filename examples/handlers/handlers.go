package main

import (
	"../gorilla/mux/mux"
	"../gorilla/sessions/sessions"
	"bytes"
	"fmt"
	"net/http"
)

type Example struct {
	Name     string
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

// checkSession checks for specified session parameter
// returns true (eg. logged in) if present, false if not
func checkSession(req *http.Request, param string) bool {
	sess, err := sessions.Session(req)
	if err != nil {
		fmt.Println("Session error: ", err.Error())
		return false
	}

	if sess[param] != nil {
		return true
	}
	return false

}

// indexPage() is a handler which will load some template and send the result back to the client
func indexPage(writer http.ResponseWriter, req *http.Request) {
	tpl, err := LoadTemplate("index.html")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	var displayContent bool

	// if session parameter "one" is present, we already have session set and we can show the content
	// otherwise, we will display login form
	if checkSession(req, "one") == true {
		displayContent = true
	} else {
		displayContent = false
	}

	errmsg := req.FormValue("error")
	var msg string

	if errmsg == "login" {
		msg = "Invalid login"
	} else {
		if errmsg == "session" {
			msg = "Session error"
		}
	}

	mydata := Example{Name: "Joe", LoggedIn: displayContent, ErrorMsg: msg}
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
		sess, err := sessions.Session(req)
		if err != nil {
			// something went wrong
			http.Redirect(writer, req, "/?error=session", http.StatusFound)
			return
		}
		sess["one"] = "two" // we set this to indicate we're logged in.
		sessions.Save(req, writer)
		http.Redirect(writer, req, "/", http.StatusFound)
		return
	}
	http.Redirect(writer, req, "/?error=login", http.StatusFound)
}
