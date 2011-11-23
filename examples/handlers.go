package main

import (
	"net/http"
	"bytes"
	"../gorilla/mux/mux"
)


func initHandlers(r *mux.Router) {
	r.HandleFunc("/", myhandler)
}


func myhandler(writer http.ResponseWriter, req *http.Request) {
        tpl, err := loadTemplate("example.html")
        if err != nil {
                http.Error(writer, err.Error(), http.StatusInternalServerError)
                return
        }

        mydir := Example{Name: "Joe"}
        buff := new(bytes.Buffer) 

        tpl.Execute(buff, mydir)
        writer.Write(buff.Bytes())
}
 

type Example struct {
        Name    string
}
