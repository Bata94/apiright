package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	ui_pages "github.com/bata94/apiright/example/ui/pages"
	"github.com/bata94/apiright/example/uirouter"
	ar "github.com/bata94/apiright/pkg/core"
	ar_templ "github.com/bata94/apiright/pkg/templ"
	"github.com/bata94/apiright/pkg/auth/jwt"
)

//go:generate /Users/bata/Projects/personal/apiright/bin/apiright generate

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

	app.Use(ar.FavIcon("example/assets/favicon.ico"))
	app.Use(ar.PanicMiddleware())
	app.Use(ar.LogMiddleware(app.Logger))
	app.Use(ar.TimeoutMiddleware(ar.TimeoutConfigFromApp(app)))
	app.Use(ar.CORSMiddleware(corsConfig))

	// Need to check if "docs/" is already generated as it won't be on first run, and app will painc
	if _, err := os.Stat("docs/"); err == nil {
		app.ServeStaticDir("/static", "docs/")
	}
	app.ServeStaticDir("/assets", "example/assets/")
	app.ServeStaticDir("/assets_dynamic", "example/assets/", ar.WithoutPreLoad())
	app.ServeStaticFile("/file/not_found", "./tmp/not_existing.txt", ar.WithoutPreLoad())
	app.ServeStaticFile("/file/hallo", "example/test.txt", ar.WithoutPreLoad())
	app.ServeStaticFile("/file/pre_hallo", "example/test.txt", ar.WithPreCache())
	// app.ServeStaticDir("/assets_not_loaded", "example/assets/", ar.WithoutPreLoad())

	// app.ServeStaticFile("/iso", "/home/bata/Downloads/nixos-minimal-25.05.806273.650e572363c0-x86_64-linux.iso")

	uiRouter := app.NewRouter("")
	uirouter.RegisterUIRoutes(uiRouter)

	jwt.DefaultJWTConfig()
	jwt.SetLogger(app.Logger)

	app.GET("/jwt", func(c *ar.Ctx) error {
		app.Logger.Info("JWT")
		c.Session["userID"] = 123
		tokenPair, err := jwt.NewTokenPair(c)
		if err != nil {
			return err
		} else if (tokenPair == jwt.TokenPair{}) {
			return errors.New("tokenPair is nil")
		}

		msgStr := fmt.Sprintf("AccessToken: %s\n", tokenPair.AccessToken)
		msgStr += fmt.Sprintf("RefreshToken: %s\n", tokenPair.RefreshToken)

		// c.ObjOut = tokenPair
		c.Response.StatusCode = 200
		c.Response.SetMessage(msgStr)

		return nil
	},
	// ar.WithObjOut(&jwt.TokenPair{}),
	)

	app.GET(ar_templ.SimpleRenderer("/simpleRenderer", ui_pages.Index()))
	app.GET(ar_templ.SimpleRenderer("/upload", ui_pages.Upload()))

	app.Redirect("/redirect", "/test", 302)
	app.Redirect("/favicon.ico", "/assets/favicon.ico", 301)

	app.GET(
		"/params/{id}",
		func(c *ar.Ctx) error {
			fmt.Println(c.PathParams)
			fmt.Println(c.QueryParams)
			c.Response.SetMessagef("PathParams: %s\nQueryParams: %s\n", c.PathParams, c.QueryParams)
			return nil
		},
		ar.WithQueryParam("name", "Test Name Description", false, reflect.TypeOf("")),
		// ar.WithQueryParam("perPage", "Items per Page", true, reflect.TypeOf(123)),
	)

	app.GET("/test", func(c *ar.Ctx) error {
		dst := "/test/upload/file.txt"

		fmt.Println(dst)
		fmt.Println(filepath.Base(dst))
		fmt.Println(filepath.Dir(dst))

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

	app.POST("/upload", upload_handler)

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

func upload_handler(c *ar.Ctx) error {
	if err := c.SaveFile("file", "./example/uploads/upload.txt"); err != nil {
		return err
	}

	c.Response.SetMessage("File uploaded successfully")
	return nil
}
