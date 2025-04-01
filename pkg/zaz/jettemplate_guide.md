# Jet Template Syntax Reference for Go

This guide focuses on the core syntax elements of the Jet template engine relevant for usage within Go applications.

## how to use in golang

```go
package main

import (
	"log"
	
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/jet/v2"
)

func main() {
	// Create a new engine
	engine := jet.New("./views", ".jet")

	// Or from an embedded system
	// See github.com/gofiber/embed for examples
	// engine := jet.NewFileSystem(http.Dir("./views", ".jet"))

	// Pass the engine to the views
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		// Render index
		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		})
	})

	app.Get("/layout", func(c *fiber.Ctx) error {
		// Render index within layouts/main
		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		}, "layouts/main")
	})

	log.Fatal(app.Listen(":3000"))
}
```


## Delimiters

Template delimiters are `{{` and `}}`. Delimiters can use `.` to output the execution context.

    hello {{ . }} <!-- with context of "world", outputs "hello world" -->

### Whitespace Trimming

Use `{{- ` and ` -}}` to trim preceding and following whitespace (spaces, tabs, carriage returns, newlines).

    foo {{ "bar" }} baz <!-- outputs "foo bar baz" -->
    foo {{- "bar" -}} baz <!-- outputs "foobarbaz" -->

### Comments

Comments `{* comment *}` are ignored during parsing and can span multiple lines.

    {* this is a comment *}

## Variables

### Initialization

Variables must be initialized before use: `{{ varName := value }}`.

    {{ foo := "bar" }}

### Assignment

Assign new values to initialized variables: `{{ varName = newValue }}`. Variables are dynamically typed within the template.

    {{ foo = "asd" }}
    {{ foo = 4711 }}

Discarding assignment (`_`): Executes the right side but discards the result. Useful for calling functions without rendering their output.

    {{ _ := functionCall() }}
    {{ _ = functionCall() }} <!-- Equivalent -->

## Expressions

### Identifiers

Names of functions (e.g., `len`, `isset`, `split`) and variables. Resolve to values in scope.

    {{ foo := "bar" }}
    {{ len(foo) }} <!-- foo is an identifier resolving to "bar" -->

### Indexing (`[]`)

Access elements within strings (by byte index), slices/arrays, maps, or struct fields (by string name).

    {{ "hello"[1] }}           <!-- Byte ASCII value -->
    {{ sliceVar[0] }}        <!-- Slice/Array element -->
    {{ mapVar["key"] }}       <!-- Map value -->
    {{ structVar["FieldName"] }} <!-- Struct field -->

### Field Access (`.`)

Access map values or struct fields using dot notation. Omit the identifier before `.` to access the current context (`.`).

    {{ mapVar.keyName }}       <!-- Map value -->
    {{ structVar.FieldName }} <!-- Struct field -->
    {{ range users }}{{ .Name }}{{ end }} <!-- Field in context -->

### Slicing (`[start:end]`)

Re-slice slices or arrays (Go-like syntax). `start` index inclusive, `end` index exclusive.

    {{ s := slice(6, 7, 8, 9) }}
    {{ subSlice := s[1:3] }} <!-- contains 7, 8 -->

### Arithmetic

Operators: `+`, `-`, `*`, `/`, `%`. Standard precedence, use `()` to override.

    {{ 1 + 2 * 3 }} <!-- outputs 7 -->

### String Concatenation

Use the `+` operator.

    {{ "Hello" + " " + "World" }}

### Logical Operators

- `&&` (and), `||` (or), `!` (not)
- `==`, `!=`, `<`, `>`, `<=`, `>=`

    {{ isAdmin && user.Active }}
    {{ count > 0 }}

### Ternary Operator

Syntax: `{{ condition ? valueIfTrue : valueIfFalse }}`.

    {{ a > b ? a : b }}

### Method Calls

Call exported methods on Go types: `{{ value.MethodName(arguments...) }}`.

    {{ user.HasPermission("edit") }}

### Function Calls

Call built-in or user-defined functions: `{{ functionName(arguments...) }}`.

    {{ len(mySlice) }}
    {{ strings.ToUpper("hello") }} <!-- Assuming 'strings' is available -->

#### Pipelining (`|`)

Pass the result of the left expression as the *last* argument to the function on the right. Chainable.

    {{ "hello" | upper }} <!-- Equivalent to upper("hello") -->
    {{ "hello" | repeat(3) | upper }} <!-- Equivalent to upper(repeat("hello", 3)) -->

## Control Structures

### `if` / `else if` / `else`

Conditional execution.

    {{ if user.IsAdmin }}
        Admin Controls
    {{ else if user.IsEditor }}
        Edit Controls
    {{ else }}
        View Controls
    {{ end }}

### `range`

Iterate over slices, arrays, maps. The `else` block executes if the collection is empty or nil.

#### Slices / Arrays

    {{ range index, value := mySlice }}
        {{ index }}: {{ value }}
    {{ else }}
        Slice is empty.
    {{ end }}
    
    {{ range value := mySlice }} <!-- Index omitted -->
        Value: {{ value }}
    {{ end }}

#### Maps

    {{ range key, value := myMap }}
        {{ key }} = {{ value }}
    {{ else }}
        Map is empty.
    {{ end }}

### `try` / `catch`

Handle potential errors during expression evaluation. Output inside `try` is buffered and only rendered on success.

    {{ try }}
        {{ mightFail() }}
    {{ catch err }}
        Error occurred: {{ err.Error() }}
    {{ end }}

## Templates

### `include`

Renders another template file, passing the current context by default, or an optional explicit context.

    {{ include "path/to/partial.jet" }}
    {{ include "path/to/user_card.jet" userContext }}

## Blocks

Reusable, named template segments, primarily for layouts.

### `block`

Defines a block. Can accept parameters with optional defaults.

    {{ block header(title="Default Title") }}
        <h1>{{ title }}</h1>
    {{ end }}

### `yield`

Executes/renders a named block, passing arguments by name.

    {{ yield header(title="My Page") }}

### `content`

Inside a block definition, `{{ yield content }}` renders content passed during the block's invocation.

    {{ block card(title) }}
      <div class="card">
        <h2>{{ title }}</h2>
        {{ yield content }}
      </div>
    {{ end }}

    {{ yield card(title="Info") content }}
        <p>This is the card body.</p>
    {{ end }}

### `extends`

Inherit from a base/layout template. Must be the first statement. The current template overrides blocks in the layout.

    {{ extends "layouts/base.jet" }}

    {{ block content() }}
        <p>Page-specific content here.</p>
    {{ end }}

### `import`

Imports all blocks defined in another template file, making them available to `yield`.

    {{ import "components/forms.jet" }}

    {{ yield inputField(label="Name", id="user_name") }}