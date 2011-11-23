package main

import (
	"../gorilla/mux/mux"
	"bytes"
	"net/http"
)

type Example struct {
	Name string
}


// in this function, we define all our patterns and their 
// handler functions
func initHandlers(r *mux.Router) {
	// 1. If gorilla-mux is enabled, r is pointer to mux.Router and we have to use it.
	// Once first match is found, appropriate handler is called, so make sure you 
	// order patterns appropriately. Pattern is regexp.
	// usage would be: r.HandleFunc("/some-regexp-pattern", handlerFn)
	//
	// 2. If gorilla-mux is disabled, you will use default http's way of defining patterns.
	// Patterns are not regexp. Instead, longest match will win, so you don't have to worry
	// about ordering in this case.
	// usage: http.HandleFunc("/pattern", handlerFn)
	//
	// this code is based on sample config provided, where gorilla-mux is disabled, so
	http.HandleFunc("/", indexPage)
	
}


// indexPage() is a handler which will load some template and send the result back to the client
func indexPage(writer http.ResponseWriter, req *http.Request) {
	tpl, err := loadTemplate("index.html")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	mydata := Example{Name: "Joe"}
	buff := new(bytes.Buffer)

	tpl.Execute(buff, mydata)
	writer.Write(buff.Bytes())
}

