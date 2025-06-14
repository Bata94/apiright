package main

import (
	"errors"
	"fmt"

	ar "github.com/bata94/apiright"
)

func main() {
	fmt.Print("Starting...")

	app := ar.NewApp()

	app.GET("/", func(c *ar.Ctx) error {
		_, err := c.Writer.Write([]byte("Hello"))
		return err
	})

	app.GET("/err", func(c *ar.Ctx) error {
		err := errors.New("Test Error")
		return err
	})

	app.GET("/panic", func(c *ar.Ctx) error {
		panic("Test Panic")
	})

	err := app.Run()
	if err != nil {
		fmt.Print("Exited with err: ", err)
		return
	}

	fmt.Print("Exited without err")
}
