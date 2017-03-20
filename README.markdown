Fill struct values from http.Request form values or attachment files
====================================================================

## Description

Use go-structer to fill struct values from http.Request form values or attachment files.

## Usage

``` golang
package main

import (
	"log"
	"net/http"

	"github.com/mash/go-structer"
)

type HelloParameters struct {
	Nickname string
	Icon     structer.Attachment
}

func hello(w http.ResponseWriter, req *http.Request) {
	parameters := HelloParameters{}
	err := structer.ToStruct(req, &parameters)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	log.Printf("parameters: %#v, file: %#v, header: %#v", parameters, parameters.Icon.File, parameters.Icon.Header)
	w.Write([]byte("Hello, " + parameters.Nickname))
}

func main() {
	handler := http.HandlerFunc(hello)
	port := ":8080"
	log.Println("Going to listen on " + port)
	log.Println("Try: `curl -v \"http://localhost:8080/\" -F \"nickname=Doe\" -F \"icon=@./README.markdown\"`")
	log.Fatal(http.ListenAndServe(port, handler))
}
```
