# Apiright

Apiright (Name is WIP), is yet another go Webframework, mainly for APIs. Why another one, aren't there enough?
Probably yes... But I usually write my APIs with the stdlib and use often the same boilerplate.
So this Framework will be a wrapper for the stdlib net/http server, with some quality of life additions.

Therefore I will try to add the best features I have seen in other Frameworks into it.

## Features:
    - [ ] "ExpressJS" like Router, with a global Error Handler and panic recovery
    - [ ] Catch-all/default Route
    - [ ] Multi domain support
    - [ ] Routergroups
    - [ ] Middlewares
    - [ ] "Fastapi" like, streamlined simple CRUD Operations
    - [ ] Automatic OpenAPI Documention
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
