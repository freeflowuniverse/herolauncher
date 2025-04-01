# Jet Template Engine Syntax Reference

## Delimiters

Template delimiters are `{{` and `}}`.  
Delimiters can use `.` to output the execution context:

```jet
hello {{ . }} <!-- context = "world" => "hello world" -->
```

### Whitespace Trimming

Whitespace around delimiters can be trimmed using `{{-` and `-}}`:

```jet
foo {{- "bar" -}} baz <!-- outputs "foobarbaz" -->
```

Whitespace includes spaces, tabs, carriage returns, and newlines.

### Comments

Comments use `{* ... *}`:

```jet
{* this is a comment *}

{*
    Multiline
    {{ expressions }} are ignored
*}
```

---

## Variables

### Initialization

```jet
{{ foo := "bar" }}
```

### Assignment

```jet
{{ foo = "asd" }}
{{ foo = 4711 }}
```

Skip assignment but still evaluate:

```jet
{{ _ := stillRuns() }}
{{ _ = stillRuns() }}
```

---

## Expressions

### Identifiers

Identifiers resolve to values:

```jet
{{ len("hello") }}
{{ isset(foo, bar) }}
```

### Indexing

#### String

```jet
{{ s := "helloworld" }}
{{ s[1] }} <!-- 101 (ASCII of 'e') -->
```

#### Slice / Array

```jet
{{ s := slice("foo", "bar", "asd") }}
{{ s[0] }}
{{ s[2] }}
```

#### Map

```jet
{{ m := map("foo", 123, "bar", 456) }}
{{ m["foo"] }}
```

#### Struct

```jet
{{ user["Name"] }}
```

### Field Access

#### Map

```jet
{{ m.foo }}
{{ range s }}
    {{ .foo }}
{{ end }}
```

#### Struct

```jet
{{ user.Name }}
{{ range users }}
    {{ .Name }}
{{ end }}
```

### Slicing

```jet
{{ s := slice(6, 7, 8, 9, 10, 11) }}
{{ sevenEightNine := s[1:4] }}
```

### Arithmetic

```jet
{{ 1 + 2 * 3 - 4 }}
{{ (1 + 2) * 3 - 4.1 }}
```

### String Concatenation

```jet
{{ "HELLO" + " " + "WORLD!" }}
```

#### Logical Operators

- `&&`
- `||`
- `!`
- `==`, `!=`
- `<`, `>`, `<=`, `>=`

```jet
{{ item == true || !item2 && item3 != "test" }}
{{ item >= 12.5 || item < 6 }}
```

### Ternary Operator

```jet
<title>{{ .HasTitle ? .Title : "Title not set" }}</title>
```

### Method Calls

```jet
{{ user.Rename("Peter") }}
{{ range users }}
    {{ .FullName() }}
{{ end }}
```

### Function Calls

```jet
{{ len(s) }}
{{ isset(foo, bar) }}
```

#### Prefix Syntax

```jet
{{ len: s }}
{{ isset: foo, bar }}
```

#### Pipelining

```jet
{{ "123" | len }}
{{ "FOO" | lower | len }}
{{ "hello" | repeat: 2 | len }}
```

**Escapers must be last in a pipeline:**

```jet
{{ "hello" | upper | raw }} <!-- valid -->
{{ raw: "hello" }}          <!-- valid -->
{{ raw: "hello" | upper }}  <!-- invalid -->
```

#### Piped Argument Slot

```jet
{{ 2 | repeat("foo", _) }}
{{ 2 | repeat("foo", _) | repeat(_, 3) }}
```

---

## Control Structures

### if

```jet
{{ if foo == "asd" }}
    foo is 'asd'!
{{ end }}
```

#### if / else

```jet
{{ if foo == "asd" }}
    ...
{{ else }}
    ...
{{ end }}
```

#### if / else if

```jet
{{ if foo == "asd" }}
{{ else if foo == 4711 }}
{{ end }}
```

#### if / else if / else

```jet
{{ if foo == "asd" }}
{{ else if foo == 4711 }}
{{ else }}
{{ end }}
```

### range

#### Slices / Arrays

```jet
{{ range s }}
    {{ . }}
{{ end }}

{{ range i := s }}
    {{ i }}: {{ . }}
{{ end }}

{{ range i, v := s }}
    {{ i }}: {{ v }}
{{ end }}
```

#### Maps

```jet
{{ range k := m }}
    {{ k }}: {{ . }}
{{ end }}

{{ range k, v := m }}
    {{ k }}: {{ v }}
{{ end }}
```

#### Channels

```jet
{{ range v := c }}
    {{ v }}
{{ end }}
```

#### Custom Ranger

Any Go type implementing `Ranger` can be ranged over.

#### else

```jet
{{ range searchResults }}
    {{ . }}
{{ else }}
    No results found :(
{{ end }}
```

### try

```jet
{{ try }}
    {{ foo }}
{{ end }}
```

### try / catch

```jet
{{ try }}
    {{ foo }}
{{ catch }}
    Fallback content
{{ end }}

{{ try }}
    {{ foo }}
{{ catch err }}
    {{ log(err.Error()) }}
    Error: {{ err.Error() }}
{{ end }}
```

---

## Templates

### include

```jet
{{ include "./user.jet" }}

<!-- user.jet -->
<div class="user">
    {{ .["name"] }}: {{ .["email"] }}
</div>
```

### return

```jet
<!-- foo.jet -->
{{ return "foo" }}

<!-- bar.jet -->
{{ foo := exec("./foo.jet") }}
Hello, {{ foo }}!
```

---

## Blocks

### block

```jet
{{ block copyright() }}
    <div>© ACME, Inc. 2020</div>
{{ end }}

{{ block inputField(type="text", label, id, value="", required=false) }}
    <label for="{{ id }}">{{ label }}</label>
    <input type="{{ type }}" value="{{ value }}" id="{{ id }}" {{ required ? "required" : "" }} />
{{ end }}
```

### yield

```jet
{{ yield copyright() }}

{{ yield inputField(id="firstname", label="First name", required=true) }}

{{ block buff() }}
    <strong>{{ . }}</strong>
{{ end }}

{{ yield buff() "Batman" }}
```

### content

```jet
{{ block link(target) }}
    <a href="{{ target }}">{{ yield content }}</a>
{{ end }}

{{ yield link(target="https://example.com") content }}
    Example Inc.
{{ end }}
```

```jet
{{ block header() }}
    <div class="header">
    {{ yield content }}
    </div>
{{ content }}
    <h1>Hey {{ name }}!</h1>
{{ end }}
```

### Recursion

```jet
{{ block menu() }}
    <ul>
        {{ range . }}
            <li>{{ .Text }}{{ if len(.Children) }}{{ yield menu() .Children }}{{ end }}</li>
        {{ end }}
    </ul>
{{ end }}
```

### extends

```jet
<!-- content.jet -->
{{ extends "./layout.jet" }}
{{ block body() }}
<main>This content can be yielded anywhere.</main>
{{ end }}

<!-- layout.jet -->
<html>
<body>
    {{ yield body() }}
</body>
</html>
```

### import

```jet
<!-- my_blocks.jet -->
{{ block body() }}
<main>This content can be yielded anywhere.</main>
{{ end }}

<!-- index.jet -->
{{ import "./my_blocks.jet" }}
<html>
<body>
    {{ yield body() }}
</body>
</html>
```