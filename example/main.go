package main

import (
	"errors"
	"fmt"
	"os"

	ar "github.com/bata94/apiright/pkg/core"
)

type PostStruct struct {
	Name string `json:"name"`
}

// @title My Go Web Framework API
// @description This is a sample API for my Go web framework.
// @version 1.0
// @host localhost:8080
// @BasePath /api/v1
// @schemes http
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	fmt.Println("Starting...")

	app := ar.NewApp()

	// app.GET("*", func(c *ar.Ctx) error {
	// 	_, err := c.Writer.Write([]byte("Catch All"))
	// 	return err
	// })

	app.GET("/", func(c *ar.Ctx) error {
		c.Response.AddHeader("Content-Type", "text/html; charset=utf-8")

		filePath := "./index.html"
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

	app.GET("/err", err_handler)

	app.GET("/err_inside", func(c *ar.Ctx) error {
		err := errors.New("Test Error")
		c.Response.StatusCode = 503
		return err
	})

	app.GET("/panic", func(c *ar.Ctx) error {
		panic("Test Panic")
	})

	app.POST(
		"/post",
		post_test,
		ar.WithObjIn(&PostStruct{}),
		ar.WithObjOut(&PostStruct{}),
	)

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

func post_test(c *ar.Ctx) error {
	fmt.Printf("Post Test ObjectIn: %v ObjInType: %v\n", c.ObjIn, c.ObjInType)
	test, ok := c.ObjIn.(*PostStruct)
	if !ok {
		return errors.New("Object is not correctly parsed")
	}

	fmt.Println("Sucess!")
	c.Response.SetMessage("Test ObjectIn: " + test.Name)
	c.ObjOut = test

	return nil
}

// GetUsers godoc
// @Summary Get all users
// @Description Retrieve a list of all users
// @Tags users
// @Accept json
// @Produce json
// @Success 200
// @Router /err [get]
func err_handler(c *ar.Ctx) error {
	err := errors.New("Test Error")
	return err
}
