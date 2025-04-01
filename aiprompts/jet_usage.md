
# Rendering Templates

> This section covers the Go side of things: preparing and executing your templates. See [Jet template syntax](https://github.com/CloudyKit/jet/wiki/3.-Jet-template-syntax) for help on writing your template files.

In the [Getting Started](https://github.com/CloudyKit/jet/wiki/1.-Getting-Started) section, we had this piece of code as the last step to execute a template:

```go
templateName := "home.jet"
t, err := set.GetTemplate(templateName)
if err != nil {
    // template could not be loaded
}
var w bytes.Buffer // needs to conform to io.Writer interface (like gin's context.Writer for example)
vars := make(jet.VarMap)
if err = t.Execute(&w, vars, nil); err != nil {
    // error when executing template
}
```

What's the `vars` map there as the second parameter? And why did we pass `nil` as the third parameter? How are templates located and loaded? Let's start there.

## Loading a Template

When you instantiate a `Set` and give it the directories for template lookup, it will not search them right away. Templates are located and loaded on-demand.

Imagine this tree of templates in your project folder:

```
├── main.go
├── README.md
└── views
    ├── common
    │   ├── _footer.jet
    │   └── _menu.jet
    ├── auth
    │   ├── _logo.jet
    │   └── login.jet
    ├── home.jet
    └── layouts
        └── application.jet
```

The `Set` might have been initialized in the `main.go` like this:

```go
var viewSet = jet.NewHTMLSet("./views")
```

Jet loads templates relative to the lookup directory; to load the `login.jet` template, you'd do:

```go
t, err := viewSet.GetTemplate("auth/login.jet")
```

Loading a template parses it and all included, imported, or extended templates – and caches the result so parsing only happens once.

## Reloading a Template in Development

While developing a website or web app in Go, it'd be nice to not cache the result after loading a template so you can leave your Go app running and still make incremental changes to the template(s). For this, Jet includes a development mode which disables caching the templates:

```go
viewSet.SetDevelopmentMode(true)
```

Be advised to disable the development mode on staging and in production to achieve maximum performance.

## Passing Variables When Executing a Template

When executing a template, you are passing the `io.Writer` object as well as the variable map and a context. Both of these will be explained next.

The variable map is a `jet.VarMap` for variables you want to access by name in your templates. Use the convenience method `Set(key, value)` to add variables:

```go
vars := make(jet.VarMap)
vars.Set("user", &User{})
```

You usually build up the variable map in one of your controller functions in response to an HTTP request by the user. One thing to be aware of is that the `jet.VarMap` is not safe to use across multiple goroutines concurrently because the backing type is a regular `map[string]reflect.Value`. If you're using wait groups to coordinate multiple concurrent fetches of data in your controllers or a similar construct, you may need to use a mutex to guard against data races. The decision was made to not do this in the core `jet.VarMap` implementation for ease of use and also because it's not a common usage scenario.

> The Appendix has a basic implementation of a mutex-protected variable map that you can use if the need arises.

Lastly, the context: the context is passed as the third parameter to the `t.Execute` template execution function and is accessed in the template using the dot. Anything can be used as a context, but if you are rendering a user edit form, it'd be best to pass the user as the context.

```html
<form action="/user" method="post">
  <input name="firstname" value="{{ .Firstname }}" />
</form>
```

Using a context can also be helpful when making blocks more reusable because the context can change while the template stays the same: `{{ .Text }}`.


# Built-in Functions

Some functions are available to you in every template. They may be invoked as regular functions:

```jet
{{ lower("TEST") }} <!-- outputs "test" -->
```

Or, which may be preferred, they can be invoked as pipelines, which also allows chaining:

```jet
{{ "test"|upper|trimSpace }} <!-- outputs "TEST" -->
```

For documentation on how to add your own (global) functions, see [Jet Template Syntax](https://github.com/CloudyKit/jet/wiki/3.-Jet-template-syntax).

## `isset()`

Can be used to check against truthy or falsy expressions:

```jet
{{ isset(.Title) ? .Title : "Title not set" }}
```

It can also be used to check for map key existence:

```jet
{{ isset(.["non_existent_key"]) ? "key does exist" : "key does not exist" }}
<!-- will print "key does not exist" -->
```

The example above uses the context, but of course, this also works with maps registered on the variable map.

## `len()`

Counts the elements in arrays, channels, slices, maps, and strings. When used on a struct argument, it returns the number of fields.

## From Go's `strings` package

- `lower` (`strings.ToLower`)
- `upper` (`strings.ToUpper`)
- `hasPrefix` (`strings.HasPrefix`)
- `hasSuffix` (`strings.HasSuffix`)
- `repeat` (`strings.Repeat`)
- `replace` (`strings.Replace`)
- `split` (`strings.Split`)
- `trimSpace` (`strings.TrimSpace`)

## Escape Helpers

- `html` (`html.EscapeString`)
- `url` (`url.QueryEscape`)
- `safeHtml` (escape HTML)
- `safeJs` (escape JavaScript)
- `raw`, `unsafe` (no escaping)
- `writeJson`, `json` to dump variables as JSON strings

## On-the-fly Map Creation

- `map`: `map(key1, value1, key2, value2)` 
    – use with caution because accessing these is slow when used with lots of elements and checked/read in loops.
