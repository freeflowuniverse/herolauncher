# ParamsParser

A Go package for parsing and manipulating parameters from text in a key-value format with support for multiline strings.

## Features

- Parse parameters in a natural format: `key: 'value' anotherKey: 'another value'`
- Support for multiline string values
- Support for numeric values without quotes: `port: 25`
- Support for boolean-like values: `secure: 1`
- Type conversion helpers (string, int, float, boolean)
- Default value support
- Required parameter validation with panic-on-missing options
- Simple and intuitive API

## Usage

### Basic Usage

```go
import (
    "github.com/freeflowuniverse/herolauncher/pkg/paramsparser"
)

// Create a new parser
parser := paramsparser.New()

// Parse a string with parameters
inputStr := `
    name: 'myapp' 
    host: 'localhost'
    port: 25
    secure: 1
    reset: 1 
    description: '
        A multiline description
        for my application.
    '
`

err := parser.Parse(inputStr)
if err != nil {
    // Handle error
}

// Or parse a simpler one-line string
parser.ParseString("name: 'myapp' version: '1.0' active: 1")

// Set default values
parser.SetDefault("host", "localhost")
parser.SetDefault("port", "8080")

// Or set multiple defaults at once
parser.SetDefaults(map[string]string{
    "debug": "false",
    "timeout": "30",
})

// Get values with type conversion
name := parser.Get("name")
port := parser.GetIntDefault("port", 8080)
secure := parser.GetBool("secure")
```

### Type Conversion

```go
// String value (with default if not found)
value := parser.Get("key")

// Integer value
intValue, err := parser.GetInt("key")
// Or with default
intValue := parser.GetIntDefault("key", 42)

// Float value
floatValue, err := parser.GetFloat("key")
// Or with default
floatValue := parser.GetFloatDefault("key", 3.14)

// Boolean value (true, yes, 1, on are considered true)
boolValue := parser.GetBool("key")
// Or with default
boolValue := parser.GetBoolDefault("key", false)
```

### Required Parameters

```go
// These will panic if the parameter is missing or invalid
value := parser.MustGet("required_param")
intValue := parser.MustGetInt("required_int_param")
floatValue := parser.MustGetFloat("required_float_param")
```

### Getting All Parameters

```go
// Get all parameters (including defaults)
allParams := parser.GetAll()
for key, value := range allParams {
    fmt.Printf("%s = %s\n", key, value)
}
```

## Example Input Format

The parser supports the following format:

```
name: 'myname' host: 'localhost'
port: 25
secure: 1
reset: 1 
description: '
    a description can be multiline

    like this
'
```

Key features of the format:
- Keys are alphanumeric (plus underscore)
- String values are enclosed in single quotes
- Numeric values don't need quotes
- Boolean values can be specified as 1/0
- Multiline strings start with a single quote and continue until a closing quote is found

## Example

See the [example](./example/main.go) for a complete demonstration of how to use this package.

## Running Tests

```bash
go test -v ./pkg/paramsparser
```
