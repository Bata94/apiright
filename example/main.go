package main

import (
	"errors"
	"fmt"
	"os"

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
		c.Response.AddHeader("Content-Type", "text/html; charset=utf-8")

		filePath := "./example/index.html"
		content, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				c.Response.SetStatus(404)
				c.Response.SetMessage("File not found")
				return errors.New("File not found")
			} else {
				c.Response.SetStatus(500)
				c.Response.SetMessage("File not readable")
				return errors.New("File not readable")
			}
		}

		c.Response.SetData(content)

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
