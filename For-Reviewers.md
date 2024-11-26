# For Reviewers

This project is built to demonstrate production-ready code, with some tradeoffs or exclusions to keep things reasonable for time and complexity. Otherwise, its structure and design are consistent with how I build HTTP servers in Go.

Some specific considerations:

#### Logging

Logging is done at a rudimentary level; enough to debug this simple stack. A real world project would implement more robust structured logging to be used for debugging and observability, analytics, alterting, etc.

#### Pagination

Leaving a GET list endpoint to retrieve all of a given resource would be bad news in a real world scenario. Pagination with some sensible maximum defaults would be provided for any list endpoints.

#### Tracing

An HTTP server should be tracing its call stack using something like OpenTracing to enable observability into performance (local and third-party/network calls) and some debugging scenarios. Being a very simple API exercize, it is not necessary for this.

#### Server Router

I typically will use a library like [go-chi](https://github.com/go-chi/chi) for routing requests. It eliminates a lot of boilerplate, and makes things like creating subroutes and applying middleware pretty easy. For this project, a few routes on the server would make that overkill.

## Testing

Testing is done by running integration tests against the server as an outside HTTP client. The table of tests will look a lot like unit tests, and the goal is to test each happy path and possible error that could be returned by the server from the perspective of an HTTP client, our intended user. By testing by making calls 'over the wire', we're not excluding any part of our stack and can test in an environment that is nearly identical to a real world production environment.

A few tests are provided to show how it is done, the test cases are not exhaustive.

If any given package within a project could benefit from direct unit tests, they can also be created.
