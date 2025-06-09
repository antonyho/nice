# nice
[![Go](https://github.com/antonyho/nice/actions/workflows/go.yml/badge.svg)](https://github.com/antonyho/nice/actions/workflows/go.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/antonyho/nice)](https://goreportcard.com/report/github.com/antonyho/nice) [![PkgGoDev](https://pkg.go.dev/badge/github.com/antonyho/nice)](https://pkg.go.dev/github.com/antonyho/nice)

Nice way to handle error in Go


## Objective
Provides a fail fast and catch anticipated error, as design pattern for enterprise service usage on Go.
Using [`panic`](https://go.dev/ref/spec#Handling_panics), [`defer`](https://go.dev/ref/spec#Defer_statements), and `recover`. A different control flow is provided.


## Background
In response to the Go team's [decision](https://go.dev/blog/error-syntax) on the call for error handling by the community. This package is created to work as an experiment and proof of concept. I would welcome any input and adoption from the community if this project could develop into a wider adoption.


## TL;DR
The idiomatic way of Go error handling is like:
```
func printSum(a, b string) error {
    x, err := strconv.Atoi(a)
    if err != nil {
        return err
    }
    y, err := strconv.Atoi(b)
    if err != nil {
        return err
    }
    fmt.Println("result:", x + y)
    return nil
}
```
The verbosity annoys many developers especially those who have extensive Java, C#, Python software development background.


## Usage
Add package to your project's Go module.
```
$ go get github.com/antonyho/nice
```

### Implementation
`defer` a handle to specific error or artefact. Then `panic` with an error or artefact.
```
import (
    "errors"
    "github.com/antonyho/nice"
)

var err2Handle = errors.New("error to handle")

func main() {
    defer nice.Tackle(err2Handle).With(
        func(err any) {
            log.Printf("An error has happened: %v", err)
        }
    )

    panic(err2Handle)
}
```

Use [`panic`](https://go.dev/ref/spec#Handling_panics) to raise the problem. With an artefact to describe the problem.
Then register the artefact which you want to `Tackle` from the panic. Provide a handler function to handle it `With`.

### Example
```
type CustomError struct {
    CustomMessage string
}
func (e *CustomError) Error() string {
    return e.CustomMessage
}

func ExampleTackle() {
    var customError = &CustomError{
        CustomMessage: "error: custom message",
    }
    customErrA := errors.New("error: a")
    customErrB := errors.New("error: b")
    
    
    errHandleFunc := func(artefact any) {
        fmt.Printf("It panicked. Error: %+v", artefact)
    }
    // Must be deferred in order to handle any panic at function return.
    defer nice.Tackle(
        customErrA,
        customErrB,
        // Custom error type should be registered as type.
        reflect.TypeFor[customError](),
    ).With(errHandleFunc)
    
    
    strHandleFunc := func(artefact any) {
        fmt.Printf("It panicked. With message: %s", artefact)
    }
    // Register multiple handlers to handle different causes.
    defer nice.Ticket(
        reflect.TypeFor[string](),
    ).With(strHandleFunc)
    
    
    
    // Business logic...
    func() {
        panic(customErrB) // This will fail first.
    }()

    func() {
        panic(customErrA) // This will be skipped due to previous panic.
    }
}
```


## Disagreement
### Panic
Many developers panic to use `panic`. And feel that it's a counter-pattern for Go. I welcome the dicussion with an open-mind. I like the simplicity and readability on Go as well. Please don't get me wrong. I left Java because of this. However, I don't mind using `panic` and `recover`. Go has them from the beginning. That means it's there for us to use by design. I don't think `panic` should only be used for run-time panic. But developer's preferences change over time.

### Readability
Yes, it's more clear to unfold the error handlings from each function calls. It's easy to follow by reading the follow up of each `if err {}`. But sometimes we just want to handle all errors in the same function by logging and incrementing the metric to the instrument. Use appropriate method by use case.

### Performance
The performance difference is significant. The cost of `reflect`, `panic` and `recover` are high. However, these are neglectable to most enterprise applications. For example, a server-side gRPC function.


## Contribution
Please submit pull request for any contribution and suggestion.
