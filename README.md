# Apiright

Apiright (Name is WIP), is yet another go Webframework, mainly for APIs. Why another one, aren't there enough?
Probably yes... But I usually write my APIs with the stdlib and use often the same boilerplate.
So this Framework will be a wrapper for the stdlib net/http server, with some quality of life additions.

Therefore I will try to add the best features I have seen in other Frameworks into it.

Apiright won't be the fastest Go Webframework out there (but it will still be way fast enough for most use cases), the main goal is Development experience and Development speed.

## Features:
    - [X] "ExpressJS" like Router, with a global Error Handler and panic recovery
    - [X] Catch-all/default Route
    - [ ] Multi domain support
    - [X] Routergroups
    - [~] Middlewares
    - [~] "Fastapi" like, streamlined simple CRUD Operations
    - [X] Automatic OpenAPI Documention
    - [ ] Embedded SQLc and Goose implementation (like in Flask and SQLAlchemy)
    - [ ] Embedded HTMX Support
    - [ ] Static File serving
    - [ ] Metrics
    - [ ] Simple ReverseProxy (not recommended for production, but for hobby projects maybe nice to have)
    - [ ] Embedded "cron-jobs", like included Microservices (run once, run every hour etc.)
    - [ ] Multiworker scaling
    - [ ] CLI-Mode
    - [ ] Extensive Testsuite

## Limitation

As long I haven't tagged or released a Version 1.X.X, you shouldn't use this in production!

## Usage

Install via go pkg and hope for the best :)

## How to use

As it is more or less a stdlib wrapper, most of the syntax is the same.

### Adding Routes

A difference to the stdlib pkg is, that the "/" path isn't a Catch-all route. I think this choice by go is rather strange, so in Apiright "/" is a valid route that only serves "/".
You can implement a Catch-all route, if you want. It should mainly be used for a 404 or some redirecting logic.

``` go
app := apiright.NewApp()

app.SetDefaultRoute(func(c *Ctx) error {
	c.Writer.Write([]byte("Custom not found!"))
	return nil
})

```
