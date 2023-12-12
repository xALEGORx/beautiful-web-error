package main

import (
	"fmt"
	"net/http"

	bwError "github.com/xALEGORx/beautiful-web-error"
)

var berror bwError.BeautifulError = bwError.BeautifulError{
	Page: true,
}

func main() {
	if err := berror.Init(); err != nil {
		panic(err)
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if err := makeRequest(r); berror.Catch(err, w, r) {
		return
	}

	fmt.Fprint(w, "Hello, World!")
}

func makeRequest(r *http.Request) error {
	return fmt.Errorf("failed request")
}
