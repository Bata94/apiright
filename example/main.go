package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	ar "github.com/bata94/apiright/pkg/core"
)

type PostStruct struct {
	Name  string `json:"name" xml:"name" yml:"name" example:"John Doe"`
	Email string `json:"email" xml:"email" yml:"email" example:"jdoe@me.com"`
	Age   int    `json:"age" xml:"age" yml:"age" example:"30"`
}

func main() {
	fmt.Println("Starting...")

	app := ar.NewApp(
		ar.AppAddr("0.0.0.0", "5500"),
		ar.AppTimeout(time.Duration(10)*time.Second),
	)

	// Create CORS config with permissive settings for quick integration
	// corsConfig := ar.DefaultCORSConfig()
	corsConfig := ar.ExposeAllCORSConfig()

	app.Use(ar.PanicMiddleware())
	app.Use(ar.LogMiddleware(app.Logger))
	app.Use(ar.TimeoutMiddleware(ar.TimeoutConfigFromApp(app)))
	app.Use(ar.CORSMiddleware(corsConfig))

	app.ServeStaticFile("/index", "./example/index.html", ar.WithPreCache())
	app.ServeStaticDir("/static", "docs/")

	app.GET("/", func(c *ar.Ctx) error {
		c.Response.AddHeader("Content-Type", "text/html; charset=utf-8")

		filePath := "./example/index.html"
		content, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				c.Response.SetStatus(404)
				c.Response.SetMessage("File not found")
				return errors.New("file not found")
			} else {
				c.Response.SetStatus(500)
				c.Response.SetMessage("File not readable")
				return errors.New("file not readable")
			}
		}

		c.Response.SetData(content)

		return nil
	})

	app.GET("/params/{id}", func(c *ar.Ctx) error {
		fmt.Println(c.PathParams)
		fmt.Println(c.QueryParams)
		c.Response.SetMessagef("PathParams: %s\nQueryParams: %s\n", c.PathParams, c.QueryParams)
		return nil
	})

	app.GET("/test", func(c *ar.Ctx) error {
		c.Response.SetMessage("Test")
		return nil
	})

	app.GET("/timeout", func(c *ar.Ctx) error {
		fmt.Println("Waiting 30 seconds")

		// This and time.Sleep(...) block and don't get canceled by the timeout Middleware
		// Need to test with "real" http calls or DB calls
		<-time.After(30 * time.Second)
		fmt.Println("Done waiting")

		c.Response.SetStatus(200)
		c.Response.SetMessage("Test Timeout")
		return nil
	})

	app.GET("/err", err_handler)

	app.GET("/err_inside", func(c *ar.Ctx) error {
		err := errors.New("test error")
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
		ar.WithOpenApiInfos("Post Test", "A simple Route to test Posting with ObjectIn and ObjectOut Structs"),
		ar.WithOpenApiTags("Test", "post"),
	)

	group := app.NewRouter("/group")
	group.GET(
		"/",
		func(c *ar.Ctx) error {
			fmt.Println("Group Index")
			c.Response.SetStatus(200)
			c.Response.SetMessage("Hello from Group Index")
			return nil
		},
		ar.WithOpenApiInfos("Group Index", "A simple Route in a Group"),
		ar.WithOpenApiDeprecated(),
	)

	group.GET("/hello", func(c *ar.Ctx) error {
		c.Response.SetMessage("Hello from Group")
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
		return errors.New("object is not correctly parsed")
	}

	fmt.Println("Success!")
	c.Response.SetMessage("Test ObjectIn: " + test.Name)
	c.ObjOut = test

	return nil
}

func err_handler(c *ar.Ctx) error {
	err := errors.New("test Error")
	return err
}
