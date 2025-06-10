# Nice - Error Handling Library for Go

[![Go](https://github.com/antonyho/nice/actions/workflows/go.yml/badge.svg)](https://github.com/antonyho/nice/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/antonyho/nice)](https://goreportcard.com/report/github.com/antonyho/nice)
[![Go Reference](https://pkg.go.dev/badge/github.com/antonyho/nice)](https://pkg.go.dev/github.com/antonyho/nice)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Nice** is a Go library that provides an alternative error handling pattern using Go's built-in `panic`, `defer`, and `recover` mechanisms. It offers a more structured approach to error handling, similar to try-catch patterns found in other programming languages, while maintaining Go's philosophy and idioms.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
  - [Tackle](#tackle)
  - [Handler.With](#handlerwith)
- [Usage Examples](#usage-examples)
  - [Basic Error Handling](#basic-error-handling)
  - [Multiple Error Types](#multiple-error-types)
  - [Custom Error Types](#custom-error-types)
  - [Multiple Handlers](#multiple-handlers)
- [Best Practices](#best-practices)
- [Performance Considerations](#performance-considerations)
- [Design Philosophy](#design-philosophy)
- [Contributing](#contributing)
- [License](#license)

## Overview

Nice provides a fail-fast error handling pattern for Go applications, particularly useful in scenarios where you want to handle multiple errors in a centralized manner. Instead of checking errors after each function call, Nice allows you to register error handlers and use `panic` to propagate errors up the call stack.

### Why Nice?

Traditional Go error handling:
```go
func processData(a, b string) error {
    x, err := strconv.Atoi(a)
    if err != nil {
        return err
    }
    
    y, err := strconv.Atoi(b)
    if err != nil {
        return err
    }
    
    result, err := calculate(x, y)
    if err != nil {
        return err
    }
    
    return saveResult(result)
}
```

With Nice:
```go
func processData(a, b string) {
    defer nice.Tackle(
        errors.New("conversion error"),
        errors.New("calculation error"),
    ).With(func(err any) {
        log.Printf("Processing failed: %v", err)
    })
    
    x := mustAtoi(a)
    y := mustAtoi(b)
    result := mustCalculate(x, y)
    mustSaveResult(result)
}
```

## Features

- **Centralized Error Handling**: Handle multiple error types in one place
- **Type-Safe Error Matching**: Register specific error types or values to catch
- **Multiple Handler Support**: Chain multiple handlers for different error scenarios
- **Custom Error Types**: Full support for custom error types and interfaces
- **Fail-Fast Pattern**: Stop execution immediately when an error occurs
- **Clean API**: Simple and intuitive API design

## Installation

```bash
go get github.com/antonyho/nice
```

## Quick Start

```go
package main

import (
    "errors"
    "log"
    "github.com/antonyho/nice"
)

var ErrDivideByZero = errors.New("divide by zero")

func main() {
    defer nice.Tackle(ErrDivideByZero).With(func(err any) {
        log.Printf("Caught error: %v", err)
    })
    
    result := divide(10, 0)
    log.Printf("Result: %d", result)
}

func divide(a, b int) int {
    if b == 0 {
        panic(ErrDivideByZero)
    }
    return a / b
}
```

## API Reference

### Tackle

`Tackle` is the primary function for registering error handlers. It accepts one or more error values or types to catch.

```go
func Tackle(artefacts ...any) Handler
```

#### Parameters
- `artefacts`: One or more error values, error types, or `reflect.Type` values to catch.

#### Returns
- `Handler`: A handler instance to attach callback functions.

#### Example
```go
defer nice.Tackle(
    io.EOF,
    reflect.TypeFor[os.PathError](),
    reflect.TypeFor[*MyCustomError](),
).With(errorHandler)
```

### Handler.With

`With` attaches a handler function to be called when a matching error is caught.

```go
func (h Handler) With(handle func(any))
```

#### Parameters
- `handler`: Function to call when a matching error is caught. Receives the error or artefact value.

#### Returns
- `*Handler`: The same handler instance for chaining.

## Usage Examples

### Basic Error Handling

```go
func readFile(filename string) []byte {
    defer nice.Tackle(
        reflect.TypeFor[os.PathError](),
        io.EOF,
    ).With(func(err any) {
        log.Printf("File operation failed: %v", err)
    })
    
    file := mustOpen(filename)
    defer file.Close()
    
    data := mustReadAll(file)
    return data
}

func mustOpen(filename string) *os.File {
    file, err := os.Open(filename)
    if err != nil {
        panic(err)
    }
    return file
}
```

### Multiple Error Types

```go
func processRequest(req *Request) *Response {
    defer nice.Tackle(
        ErrInvalidInput,
        ErrUnauthorized,
        ErrDatabaseConnection,
    ).With(func(err any) {
        switch err {
        case ErrInvalidInput:
            respondWithError(400, "Invalid input")
        case ErrUnauthorized:
            respondWithError(401, "Unauthorized")
        case ErrDatabaseConnection:
            respondWithError(500, "Database error")
        }
    })
    
    validateInput(req)
    authenticateUser(req)
    return processData(req)
}
```

### Custom Error Types

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

func validateForm(data map[string]string) {
    defer nice.Tackle(
        reflect.TypeFor[ValidationError](),
    ).With(func(err any) {
        if ve, ok := err.(ValidationError); ok {
            log.Printf("Validation failed: field=%s, msg=%s", ve.Field, ve.Message)
        }
    })
    
    if data["email"] == "" {
        panic(&ValidationError{Field: "email", Message: "required"})
    }
    
    if !isValidEmail(data["email"]) {
        panic(&ValidationError{Field: "email", Message: "invalid format"})
    }
}
```

### Multiple Handlers

```go
func complexOperation() {
    // First handler for database errors
    defer nice.Tackle(
        reflect.TypeFor[DBError]()),
        ErrConnectionLost,
    ).With(func(err any) {
        log.Error("Database error:", err)
        notifyOps(err)
    })
    
    // Second handler for business logic errors
    defer nice.Tackle(
        ErrInsufficientFunds,
        ErrAccountLocked,
    ).With(func(err any) {
        log.Warn("Business error:", err)
        auditLog(err)
    })
    
    // Operations that might panic with various errors
    performDatabaseOperation()
    performBusinessLogic()
}
```

## Best Practices

### 1. Use Specific Error Types

Register specific error types rather than catching all panics:

```go
// Use case
defer nice.Tackle(ErrSpecificError).With(handler)

// Replace this use case
defer func() {
    if r := recover(); r != nil {
        // Catches everything
    }
}()
```

### 2. Define Clear Error Variables

Create well-named error variables for different failure scenarios:

```go
var (
    ErrInvalidConfig   = errors.New("invalid configuration")
    ErrServiceUnavailable = errors.New("service unavailable")
    ErrRateLimitExceeded  = errors.New("rate limit exceeded")
)
```

### 3. Place Handlers at Appropriate Levels

Put error handlers at logical boundaries in your application:

```go
func httpHandler(w http.ResponseWriter, r *http.Request) {
    defer nice.Tackle(
        ErrBadRequest,
        ErrUnauthorized,
        ErrServerError,
    ).With(func(err any) {
        respondWithAppropriateError(w, err)
    })
    
    // Request processing logic
}
```

### 4. Use for Fail-Fast Scenarios

Nice is ideal for scenarios where you want to stop execution immediately on error:

```go
func initializeApp() {
    defer nice.Tackle(ErrConfigError).With(func(err any) {
        log.Fatal("Failed to initialize:", err)
    })
    
    loadConfig()      // panic on error
    connectDB()       // panic on error
    startServices()   // panic on error
}
```

## Performance Considerations

Nice uses Go's `panic`, `recover`, and reflection mechanisms, which have performance implications:

- **Reflection Cost**: Type checking uses reflection, which adds overhead
- **Panic/Recover Cost**: These operations are more expensive than regular error returns
- **Best Use Cases**: 
  - Server-side request handlers
  - Initialization code
  - Batch processing
  - Any scenario where code clarity outweighs microsecond-level performance

For performance-critical code paths (e.g., tight loops, real-time systems), consider using traditional error handling.

## Design Philosophy

Nice embraces the idea that `panic` and `recover` are legitimate Go features that can be used effectively when applied appropriately. The library aims to:

1. **Reduce Boilerplate**: Minimize repetitive error checking code
2. **Improve Readability**: Make the happy path more apparent
3. **Centralize Handling**: Handle related errors in one place
4. **Maintain Go Idioms**: Work within Go's design principles

Nice is NOT trying to turn Go into Java or Python. It's providing an alternative pattern that can coexist with traditional Go error handling.

## When to Use Nice

✅ **Good Use Cases:**
- Web request handlers
- CLI applications
- Service initialization
- Batch processing
- Prototyping
- Any code where fail-fast behavior is desired

❌ **Avoid Using Nice For:**
- Library code (return errors instead)
- Performance-critical paths
- Goroutines (unless carefully managed)
- Code that needs fine-grained error handling

## Contributing

We welcome contributions! Please submit pull request for any contribution and suggestion.

To contribute:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure your code:
- Includes tests
- Follows Go conventions
- Updates documentation as needed

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by exception handling patterns in other languages
- Built for the Go community as an experiment in alternative error handling approaches
- Thanks to all contributors who help improve this library
