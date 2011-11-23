package main

import (
	"../gorilla/mux/mux"
	"bytes"
	"net/http"
)

func initHandlers(r *mux.Router) {
	http.HandleFunc("/", myhandler)
}

func myhandler(writer http.ResponseWriter, req *http.Request) {
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

type Example struct {
	Name string
}
