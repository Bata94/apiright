package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/bata94/apiright/example/ui-router/gen"
	pages "github.com/bata94/apiright/example/ui/pages"
	ar "github.com/bata94/apiright/pkg/core"
	ar_templ "github.com/bata94/apiright/pkg/templ"
)

//go:generate /Users/bata/Projects/personal/apiright/bin/apiright-cli -i ./ui/pages -o ./ui-router/gen/routes_gen.go -p gen

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

	app.ServeStaticDir("/static", "docs/")
	app.ServeStaticDir("/assets", "example/assets/")
	app.ServeStaticFile("/file/not_found", "./tmp/not_existing.txt", ar.WithoutPreLoad())
	app.ServeStaticFile("/file/hallo", "example/test.txt", ar.WithoutPreLoad())
	// app.ServeStaticDir("/assets_not_loaded", "example/assets/", ar.WithoutPreLoad())

	uiRouter := app.NewRouter("")
	gen.RegisterUIRoutes(uiRouter)

	app.GET(ar_templ.SimpleRenderer("/simpleRenderer", pages.Index()))

	app.Redirect("/redirect", "/test", 302)
	app.Redirect("/favicon.ico", "/assets/favicon.ico", 301)

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
