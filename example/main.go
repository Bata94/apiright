package main

import (
	"errors"
	"fmt"

	ar "github.com/bata94/apiright"
)

func main() {
	fmt.Println("Starting...")

	app := ar.NewApp()

	// app.GET("*", func(c *ar.Ctx) error {
	// 	_, err := c.Writer.Write([]byte("Catch All"))
	// 	return err
	// })

	app.GET("/", func(c *ar.Ctx) error {
		c.Response.Message = "Hello"
		return nil
	})

	app.GET("/err", func(c *ar.Ctx) error {
		err := errors.New("Test Error")
		return err
	})

	app.GET("/err_inside", func(c *ar.Ctx) error {
		err := errors.New("Test Error")
		c.Response.StatusCode = 503
		return err
	})

	app.GET("/panic", func(c *ar.Ctx) error {
		panic("Test Panic")
	})

	group := app.NewRouter("/group")
	group.GET("/", func(c *ar.Ctx) error {
		c.Response.Message = "Hello from Group Index"
		return nil
	})

	group.GET("/hello", func(c *ar.Ctx) error {
		c.Response.Message = "Hello from Group"
		return nil
	})

	err := app.Run()
	if err != nil {
		fmt.Println("Exited with err: ", err)
		return
	}

	fmt.Println("Exited without err")
}
