package main

import (
	"fmt"
	"net/http"

	ar "github.com/bata94/apiright/v0/core"
)

func infoHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func main() {
	fmt.Println("Hello, World!")

	app := ar.InitApp()

	app.Get("/info", infoHandler)

	app.Run()

	fmt.Println("Bye, World!")
}
