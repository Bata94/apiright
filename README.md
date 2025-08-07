# Apiright

[![Go Report Card](https://goreportcard.com/badge/github.com/bata94/apiright)](https://goreportcard.com/report/github.com/bata94/apiright)

* [Description](#description)
* [Features](#features)
* [Limitation](#limitation)
* [Usage](#usage)
* [How to use](#how-to-use)

## Description

Apiright (Name is WIP), is yet another go Webframework, mainly for APIs. Why another one, aren't there enough?
Probably yes... But I usually write my APIs with the stdlib and use often the same boilerplate.
So this Framework will be a wrapper for the stdlib net/http server, with some quality of life additions.

Therefore I will try to add the best features I have seen in other Frameworks into it.

Apiright won't be the fastest Go Webframework out there (but it will still be way fast enough for most use cases), the main goal is Development experience and Development speed.
The goal is an easy to use and optionated "batteries included" Framework. But in a modular way, so you can exchange modules to fit your needs and only use what you need.

## Features:
These are the features I have in mind, they are roughly ordered by importance.

    - [X] "ExpressJS" like Router
    - [X] simple GET Filebased Routing, with full Router support
    - [ ] more advanced Filebased Routing
    - [X] SubRouter support for grouping
    - [X] Path- and QueryParams Support
    - [X] Catch-all/default Route
    - [X] Global ErrorHandling
    - [~] App Logger, compatible/exchangeable with a slog Logger (Default logger supporting colored and formatted logging)
    - [ ] Custom global ErrorHandling
    - [ ] Multi domain support
    - [X] Routergroups
    - [X] Static File serving
    - [X] Middleware support
    - [X] Buildin Middleware (Logging, CORS, CSRF, PanicRecovery, Timeout)
    - [X] Addable custom Middlewares
    - [ ] More Buildin Middlewares (RateLimit, Cache, Compress, SecureCookies, Session)
    - [ ] A good Auth library like "BetterAuth" (light)
    - [ ] Builtin Auth Middlewares (Basic-Auth, JWT-Auth, OAuth, ApiKey-Auth)
    - [ ] Function defined skip condition for Middlewares
    - [X] "Fastapi" like, streamlined simple CRUD Operations, meaning auto conversion to defined struct for RequestInput and ResponseOutput
    - [ ] Simple CRUD Endpoint support for fast v0 or prototyping
    - [X] Automatic MIMEType parsing based on Headers (JSON, YAML and XML)
    - [ ] More MIMETypes settable and exchangeable parsers for build in MIMETypes
    - [X] Automatic OpenAPI Documention
    - [ ] Embedded SQLc and Goose implementation (like in Flask and SQLAlchemy)
    - [ ] Embedded HTMX Support
    - [ ] HtmGo Support (?!)
    - [ ] CLI setup/blueprint tool
    - [ ] Dokerfile template
    - [ ] Metrics
    - [ ] Simple ReverseProxy (not recommended for production, but for hobby projects maybe nice to have)
    - [ ] Embedded "cron-jobs", like included Microservices (run once, run every hour etc.)
    - [ ] Multiworker scaling
    - [ ] CLI-Mode
    - [ ] Extensive Testsuite
    - [X] HTTP/2.0 Support (Handled automatically in net/http)
    - [ ] HTTP/3.0 Support (I think in Go net/http it's not prod ready)

## Limitation

As long I haven't tagged or released a Version 1.X.X, you shouldn't use this in production! Breaking changes will happen on a regular basis.

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
