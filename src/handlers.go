package main

import (
	"net/http"
	"bytes"
)


func initHandlers() {
	http.HandleFunc("/", myhandler)
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
