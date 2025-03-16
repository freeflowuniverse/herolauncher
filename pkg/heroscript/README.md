# HeroScript

A Go package for parsing and executing HeroScript, a small scripting language for defining actions.

## HeroScript Format

HeroScript is a simple scripting language with the following structure:

```heroscript
!!actor.action
    name: 'value'
    key: 'value'
    numeric_value: 25
    boolean_value: 1
    description: '
        A multiline description
        can be added here.
        
        It supports multiple paragraphs.
    '
```

Key features:
- Every action starts with `!!` (for SAL actions)
  - The first part is the actor (e.g., `mailclient`)
  - The second part is the action name (e.g., `configure`)
- Parameters follow in an indented block
- Multiline strings are supported using single quotes
- Comments start with `//`

## Usage

### Basic Usage

```go
import (
    "fmt"
    "github.com/freeflowuniverse/herolauncher/pkg/heroscript/playbook"
)

// Create a new playbook from HeroScript text
script := `
!!mailclient.configure
    name: 'myname'
    host: 'localhost'
    port: 25
    secure: 1
`

pb, err := playbook.NewFromText(script)
if err != nil {
    // Handle error
}

// Access actions
for _, action := range pb.Actions {
    fmt.Printf("Action: %s.%s\n", action.Actor, action.Name)
    
    // Access parameters
    name := action.Params.Get("name")
    host := action.Params.Get("host")
    port := action.Params.GetInt("port")
    secure := action.Params.GetBool("secure")
    
    // Do something with the action...
}
```

### Finding Actions

```go
// Find all actions for a specific actor
mailActions, err := pb.FindActions(0, "mailclient", "", playbook.ActionTypeUnknown)
if err != nil {
    // Handle error
}

// Find a specific action
configAction, err := pb.GetAction(0, "mailclient", "configure")
if err != nil {
    // Handle error
}
```

### Generating HeroScript

```go
// Generate HeroScript from the playbook
script := pb.HeroScript(true)  // true to include done actions
fmt.Println(script)

// Generate HeroScript excluding done actions
script = pb.HeroScript(false)
fmt.Println(script)
```

## Action Types

HeroScript supports different action types:

- `!action` - DAL (Data Access Layer)
- `!!action` - SAL (Service Access Layer)
- `!!!action` - Macro

## Integration with HandlerFactory

The HeroScript package can be used with the HandlerFactory to process commands. Each handler is associated with an actor and implements methods for each action it supports.

### Handler Implementation

To create a handler that works with the HandlerFactory and HeroScript:

```go
// MyHandler handles actions for the "myactor" actor
type MyHandler struct {
	handlerfactory.BaseHandler
}

// NewMyHandler creates a new MyHandler
func NewMyHandler() *MyHandler {
	return &MyHandler{
		BaseHandler: handlerfactory.BaseHandler{
			ActorName: "myactor",
		},
	}
}

// Play processes all actions for this handler's actor
func (h *MyHandler) Play(script string, handler interface{}) (string, error) {
	return h.BaseHandler.Play(script, handler)
}

// DoSomething handles the myactor.do_something action
func (h *MyHandler) DoSomething(script string) string {
	log.Printf("MyActor.DoSomething called with: %s", script)
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}
	
	// Process the action...
	return "Action completed successfully"
}
```

### Using with HandlerFactory

```go
// Create a new handler factory
factory := handlerfactory.NewHandlerFactory()

// Create and register a handler
myHandler := NewMyHandler()
err := factory.RegisterHandler(myHandler)
if err != nil {
	log.Fatalf("Failed to register handler: %v", err)
}

// Process a HeroScript command
result, err := factory.ProcessHeroscript(`
!!myactor.do_something
    param1: 'value1'
    param2: 'value2'
`)
if err != nil {
	log.Fatalf("Error processing heroscript: %v", err)
}

fmt.Println(result)
```

## Example

See the [example](./example/main.go) for a complete demonstration of how to use this package.

## Running Tests

```bash
go test -v ./pkg/heroscript/playbook
```
