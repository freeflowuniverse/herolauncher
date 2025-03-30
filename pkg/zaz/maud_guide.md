# Maud: A Macro for Writing HTML in Rust

Maud is an HTML template engine for Rust implemented as a macro (`html!`), which compiles your markup to specialized Rust code. This approach makes Maud templates fast, type-safe, and easy to deploy.

## Introduction

```rust
html! {
    h1 { "Hello, world!" }
    p.intro {
        "This is an example of the "
        a href="https://github.com/lambda-fairy/maud" { "Maud" }
        " template language."
    }
}
```

### Key Features

- **Tight integration with Rust**: Pattern matching and for loops work as they do in Rust. There is no need to derive JSON conversions, as your templates can work with Rust values directly.
- **Type safety**: Templates are checked by the compiler, just like the code around them. Any typos will be caught at compile time.
- **Minimal runtime**: Since most of the work happens at compile time, the runtime footprint is small.
- **Simple deployment**: No need to track separate template files, since all relevant code is linked into the final executable.

## Getting Started

### Add Maud to your project

Once Rust is set up, create a new project with Cargo:

```bash
cargo new --bin pony-greeter
cd pony-greeter
```

Add maud to your Cargo.toml:

```toml
[dependencies]
maud = "*"
```

Then save the following to src/main.rs:

```rust
use maud::html;

fn main() {
    let name = "Lyra";
    let markup = html! {
        p { "Hi, " (name) "!" }
    };
    println!("{}", markup.into_string());
}
```

The `html!` macro takes a single argument: a template using Maud's custom syntax. This call expands to an expression of type `Markup`, which can then be converted to a `String` using `.into_string()`.

Run this program with `cargo run`, and you should get the following:

```html
<p>Hi, Lyra!</p>
```

## Text and Escaping

### Text

Literal strings use the same syntax as Rust. Wrap them in double quotes, and use a backslash for escapes.

```rust
html! {
    "Oatmeal, are you crazy?"
}
```

### Raw Strings

If the string is long, or contains many special characters, then it may be worth using raw strings instead:

```rust
html! {
    pre {
        r#"
            Rocks, these are my rocks.
            Sediments make me sedimental.
            Smooth and round,
            Asleep in the ground.
            Shades of brown
            And gray.
        "#
    }
}
```

### Escaping and PreEscaped

By default, HTML special characters are escaped automatically. Wrap the string in `(PreEscaped())` to disable this escaping.

```rust
use maud::PreEscaped;
html! {
    "<script>alert(\"XSS\")</script>"                // &lt;script&gt;...
    (PreEscaped("<script>alert(\"XSS\")</script>"))  // <script>...
}
```

### The DOCTYPE constant

If you want to add a `<!DOCTYPE html>` declaration to your page, you may use the `maud::DOCTYPE` constant instead of writing it out by hand:

```rust
use maud::DOCTYPE;
html! {
    (DOCTYPE)  // <!DOCTYPE html>
}
```

## Elements and Attributes

### Elements with contents: p {}

Write an element using curly braces:

```rust
html! {
    h1 { "Poem" }
    p {
        strong { "Rock," }
        " you are a rock."
    }
}
```

### Void elements: br;

Terminate a void element using a semicolon:

```rust
html! {
    link rel="stylesheet" href="poetry.css";
    p {
        "Rock, you are a rock."
        br;
        "Gray, you are gray,"
        br;
        "Like a rock, which you are."
        br;
        "Rock."
    }
}
```

The result will be rendered with HTML syntax – `<br>` not `<br />`.

### Custom elements and data attributes

Maud also supports elements and attributes with hyphens in them. This includes custom elements, data attributes, and ARIA annotations.

```rust
html! {
    article data-index="12345" {
        h1 { "My blog" }
        tag-cloud { "pinkie pie pony cute" }
    }
}
```

### Non-empty attributes: title="yay"

Add attributes using the syntax: `attr="value"`. You can attach any number of attributes to an element. The values must be quoted: they are parsed as string literals.

```rust
html! {
    ul {
        li {
            a href="about:blank" { "Apple Bloom" }
        }
        li class="lower-middle" {
            "Sweetie Belle"
        }
        li dir="rtl" {
            "Scootaloo "
            small { "(also a chicken)" }
        }
    }
}
```

### Empty attributes: checked

Declare an empty attribute by omitting the value.

```rust
html! {
    form {
        input type="checkbox" name="cupcakes" checked;
        " "
        label for="cupcakes" { "Do you like cupcakes?" }
    }
}
```

Before version 0.22.2, Maud required a `?` suffix on empty attributes: `checked?`. This is no longer necessary, but still supported for backward compatibility.

### Classes and IDs: .foo #bar

Add classes and IDs to an element using `.foo` and `#bar` syntax. You can chain multiple classes and IDs together, and mix and match them with other attributes:

```rust
html! {
    input #cannon .big.scary.bright-red type="button" value="Launch Party Cannon";
}
```

In Rust 2021, the `#` symbol must be preceded by a space, to avoid conflicts with reserved syntax:

```rust
html! {
    // Works on all Rust editions
    input #pinkie;

    // Works on Rust 2018 and older only
    input#pinkie;
}
```

The classes and IDs can be quoted. This is useful for names with numbers or symbols which otherwise wouldn't parse:

```rust
html! {
    div."col-sm-2" { "Bootstrap column!" }
}
```

### Implicit div elements

If the element name is omitted, but there is a class or ID, then it is assumed to be a div.

```rust
html! {
    #main {
        "Main content!"
        .tip { "Storing food in a refrigerator can make it 20% cooler." }
    }
}
```

## Splices and Toggles

### Splices: (foo)

Use `(foo)` syntax to insert the value of foo at runtime. Any HTML special characters are escaped by default.

```rust
let best_pony = "Pinkie Pie";
let numbers = [1, 2, 3, 4];
html! {
    p { "Hi, " (best_pony) "!" }
    p {
        "I have " (numbers.len()) " numbers, "
        "and the first one is " (numbers[0])
    }
}
```

Arbitrary Rust code can be included in a splice by using a block:

```rust
html! {
    p {
        ({
            let f: Foo = something_convertible_to_foo()?;
            f.time().format("%H%Mh")
        })
    }
}
```

#### Splices in attributes

Splices work in attributes as well:

```rust
let secret_message = "Surprise!";
html! {
    p title=(secret_message) {
        "Nothing to see here, move along."
    }
}
```

To concatenate multiple values within an attribute, wrap the whole thing in braces:

```rust
const GITHUB: &'static str = "https://github.com";
html! {
    a href={ (GITHUB) "/lambda-fairy/maud" } {
        "Fork me on GitHub"
    }
}
```

#### Splices in classes and IDs

Splices can also be used in classes and IDs:

```rust
let name = "rarity";
let severity = "critical";
html! {
    aside #(name) {
        p.{ "color-" (severity) } { "This is the worst! Possible! Thing!" }
    }
}
```

#### What can be spliced?

You can splice any value that implements `Render`. Most primitive types (such as `str` and `i32`) implement this trait, so they should work out of the box.

To get this behavior for a custom type, you can implement the `Render` trait by hand. The `PreEscaped` wrapper type, which outputs its argument without escaping, works this way.

```rust
use maud::PreEscaped;
let post = "<p>Pre-escaped</p>";
html! {
    h1 { "My super duper blog post" }
    (PreEscaped(post))
}
```

### Toggles: [foo]

Use `[foo]` syntax to show or hide classes and boolean attributes on a HTML element based on a boolean expression `foo`.

Toggle boolean attributes:

```rust
let allow_editing = true;
html! {
    p contenteditable[allow_editing] {
        "Edit me, I "
        em { "dare" }
        " you."
    }
}
```

And classes:

```rust
let cuteness = 95;
html! {
    p.cute[cuteness > 50] { "Squee!" }
}
```

#### Optional attributes with values: title=[Some("value")]

Add optional attributes to an element using `attr=[value]` syntax, with square brackets. These are only rendered if the value is `Some<T>`, and entirely omitted if the value is `None`.

```rust
html! {
    p title=[Some("Good password")] { "Correct horse" }

    @let value = Some(42);
    input value=[value];

    @let title: Option<&str> = None;
    p title=[title] { "Battery staple" }
}
```

## Control Structures

### Branching with @if and @else

Use `@if` and `@else` to branch on a boolean expression. As with Rust, braces are mandatory and the `@else` clause is optional.

```rust
#[derive(PartialEq)]
enum Princess { Celestia, Luna, Cadance, TwilightSparkle }

let user = Princess::Celestia;

html! {
    @if user == Princess::Luna {
        h1 { "Super secret woona to-do list" }
        ul {
            li { "Nuke the Crystal Empire" }
            li { "Kick a puppy" }
            li { "Evil laugh" }
        }
    } @else if user == Princess::Celestia {
        p { "Sister, please stop reading my diary. It's rude." }
    } @else {
        p { "Nothing to see here; move along." }
    }
}
```

`@if let` is supported as well:

```rust
let user = Some("Pinkie Pie");
html! {
    p {
        "Hello, "
        @if let Some(name) = user {
            (name)
        } @else {
            "stranger"
        }
        "!"
    }
}
```

### Looping with @for

Use `@for .. in ..` to loop over the elements of an iterator:

```rust
let names = ["Applejack", "Rarity", "Fluttershy"];
html! {
    p { "My favorite ponies are:" }
    ol {
        @for name in &names {
            li { (name) }
        }
    }
}
```

### Declaring variables with @let

Declare a new variable within a template using `@let`:

```rust
let names = ["Applejack", "Rarity", "Fluttershy"];
html! {
    @for name in &names {
        @let first_letter = name.chars().next().unwrap();
        p {
            "The first letter of "
            b { (name) }
            " is "
            b { (first_letter) }
            "."
        }
    }
}
```

### Matching with @match

Pattern matching is supported with `@match`:

```rust
enum Princess { Celestia, Luna, Cadance, TwilightSparkle }

let user = Princess::Celestia;

html! {
    @match user {
        Princess::Luna => {
            h1 { "Super secret woona to-do list" }
            ul {
                li { "Nuke the Crystal Empire" }
                li { "Kick a puppy" }
                li { "Evil laugh" }
            }
        },
        Princess::Celestia => {
            p { "Sister, please stop reading my diary. It's rude." }
        },
        _ => p { "Nothing to see here; move along." }
    }
}
```

## Partials

Maud does not have a built-in concept of partials or sub-templates. Instead, you can compose your markup with any function that returns `Markup`.

The following example defines a header and footer function. These functions are combined to form the final page:

```rust
use maud::{DOCTYPE, html, Markup};

/// A basic header with a dynamic `page_title`.
fn header(page_title: &str) -> Markup {
    html! {
        (DOCTYPE)
        meta charset="utf-8";
        title { (page_title) }
    }
}

/// A static footer.
fn footer() -> Markup {
    html! {
        footer {
            a href="rss.atom" { "RSS Feed" }
        }
    }
}

/// The final Markup, including `header` and `footer`.
///
/// Additionally takes a `greeting_box` that's `Markup`, not `&str`.
pub fn page(title: &str, greeting_box: Markup) -> Markup {
    html! {
        // Add the header markup to the page
        (header(title))
        h1 { (title) }
        (greeting_box)
        (footer())
    }
}
```

Using the page function will return the markup for the whole page:

```rust
page("Hello!", html! {
    div { "Greetings, Maud." }
});
```

## The Render Trait

For most types, Maud will use the `std::fmt::Display` trait to convert (spliced) values to HTML. (The result will be escaped automatically.) If you'd like to override this behavior for your own type, then you can implement the `Render` trait instead.

Below are some examples of implementing `Render`:

### Example: A shorthand for including CSS stylesheets

```rust
use maud::{html, Markup, Render};

/// Links to a CSS stylesheet at the given path.
struct Css(&'static str);

impl Render for Css {
    fn render(&self) -> Markup {
        html! {
            link rel="stylesheet" type="text/css" href=(self.0);
        }
    }
}
```

### Example: A wrapper that calls std::fmt::Debug

```rust
use maud::{Escaper, html, Render};
use std::fmt;
use std::fmt::Write as _;

/// Renders the given value using its `Debug` implementation.
struct Debug<T: fmt::Debug>(T);

impl<T: fmt::Debug> Render for Debug<T> {
    fn render_to(&self, output: &mut String) {
        let mut escaper = Escaper::new(output);
        write!(escaper, "{:?}", self.0).unwrap();
    }
}
```

### Example: Rendering Markdown using pulldown-cmark and ammonia

```rust
use ammonia;
use maud::{Markup, PreEscaped, Render};
use pulldown_cmark::{Parser, html};

/// Renders a block of Markdown using `pulldown-cmark`.
struct Markdown<T>(T);

impl<T: AsRef<str>> Render for Markdown<T> {
    fn render(&self) -> Markup {
        // Generate raw HTML
        let mut unsafe_html = String::new();
        let parser = Parser::new(self.0.as_ref());
        html::push_html(&mut unsafe_html, parser);
        // Sanitize it with ammonia
        let safe_html = ammonia::clean(&unsafe_html);
        PreEscaped(safe_html)
    }
}
```

## Web Framework Integration

Maud includes support for these web frameworks: Actix, Rocket, Rouille, Tide, Axum, Warp, Submillisecond, and Poem.

### Actix

Actix support is available with the "actix-web" feature:

```toml
[dependencies]
maud = { version = "*", features = ["actix-web"] }
```

Actix request handlers can use a `Markup` that implements the `actix_web::Responder` trait:

```rust
use actix_web::{get, App, HttpServer, Result as AwResult};
use maud::{html, Markup};
use std::io;

#[get("/")]
async fn index() -> AwResult<Markup> {
    Ok(html! {
        html {
            body {
                h1 { "Hello World!" }
            }
        }
    })
}

#[actix_web::main]
async fn main() -> io::Result<()> {
    HttpServer::new(|| App::new().service(index))
        .bind(("127.0.0.1", 8080))?
        .run()
        .await
}
```

### Rocket

Rocket works in a similar way, except using the rocket feature:

```toml
[dependencies]
maud = { version = "*", features = ["rocket"] }
```

This adds a `Responder` implementation for the `Markup` type, so you can return the result directly:

```rust
use maud::{html, Markup};
use rocket::{get, routes};
use std::borrow::Cow;

#[get("/<n>")]
fn hello(name: &str) -> Markup {
    html! {
        h1 { "Hello, " (name) "!" }
        p { "Nice to meet you!" }
    }
}

#[rocket::launch]
fn launch() -> _ {
    rocket::build().mount("/", routes![hello])
}
```

### Rouille

Unlike with the other frameworks, Rouille doesn't need any extra features at all! Calling `Response::html` on the rendered `Markup` will Just Work®:

```rust
use maud::html;
use rouille::{Response, router};

fn main() {
    rouille::start_server("localhost:8000", move |request| {
        router!(request,
            (GET) (/{name: String}) => {
                Response::html(html! {
                    h1 { "Hello, " (name) "!" }
                    p { "Nice to meet you!" }
                })
            },
            _ => Response::empty_404()
        )
    });
}
```

### Tide

Tide support is available with the "tide" feature:

```toml
[dependencies]
maud = { version = "*", features = ["tide"] }
```

This adds an implementation of `From<PreEscaped<String>>` for the `Response` struct:

```rust
use maud::html;
use tide::Request;
use tide::prelude::*;

#[async_std::main]
async fn main() -> tide::Result<()> {
    let mut app = tide::new();
    app.at("/hello/:name").get(|req: Request<()>| async move {
        let name: String = req.param("name")?.parse()?;
        Ok(html! {
            h1 { "Hello, " (name) "!" }
            p { "Nice to meet you!" }
        })
    });
    app.listen("127.0.0.1:8080").await?;
    Ok(())
}
```

### Axum

Axum support is available with the "axum" feature:

```toml
[dependencies]
maud = { version = "*", features = ["axum"] }
```

This adds an implementation of `IntoResponse` for `Markup`/`PreEscaped<String>`:

```rust
use maud::{html, Markup};
use axum::{Router, routing::get};

async fn hello_world() -> Markup {
    html! {
        h1 { "Hello, World!" }
    }
}

#[tokio::main]
async fn main() {
    // build our application with a single route
    let app = Router::new().route("/", get(hello_world));

    // run it with hyper on localhost:3000
    let listener = tokio::net::TcpListener::bind("0.0.0.0:3000").await.unwrap();

    axum::serve(listener, app.into_make_service()).await.unwrap();
}
```

### Warp

Warp support is available with the "warp" feature:

```toml
[dependencies]
maud = { version = "*", features = ["warp"] }
```

This enables `Markup` to be of type `warp::Reply`:

```rust
use maud::html;
use warp::Filter;

#[tokio::main]
async fn main() {
    let hello = warp::any().map(|| html! { h1 { "Hello, world!" } });
    warp::serve(hello).run(([127, 0, 0, 1], 8000)).await;
}
```

### Submillisecond

Submillisecond support is available with the "submillisecond" feature:

```toml
[dependencies]
maud = { version = "*", features = ["submillisecond"] }
```

This adds an implementation of `IntoResponse` for `Markup`/`PreEscaped<String>`:

```rust
use maud::{html, Markup};
use std::io::Result;
use submillisecond::{router, Application};

fn main() -> Result<()> {
    Application::new(router! {
        GET "/hello" => helloworld
    })
    .serve("0.0.0.0:3000")
}

fn helloworld() -> Markup {
    html! {
        h1 { "Hello, World!" }
    }
}
```

### Poem

Poem support is available with the "poem" feature:

```toml
[dependencies]
maud = { version = "*", features = ["poem"] }
```

This adds an implementation of `poem::IntoResponse` for `Markup`/`PreEscaped<String>`:

```rust
use maud::{html, Markup};
use poem::{get, handler, listener::TcpListener, Route, Server};

#[handler]
fn hello_world() -> Markup {
    html! {
        h1 { "Hello, World!" }
    }
}

#[tokio::main]
async fn main() -> Result<(), std::io::Error> {
    let app = Route::new().at("/hello", get(hello_world));
    Server::new(TcpListener::bind("0.0.0.0:3000"))
        .name("hello-world")
        .run(app)
        .await
}
```

## Frequently Asked Questions

### What is the origin of the name "Maud"?

Maud is named after a character from My Little Pony: Friendship is Magic. It does not refer to the poem by Alfred Tennyson, though other people have brought that up in the past. Here are some reasons why I chose this name:

- "Maud" shares three letters with "markup"
- The library is efficient and austere, like the character
- Google used to maintain a site called "HTML5 Rocks", and Maud (the character) is a geologist

### Why does html! always allocate a String? Wouldn't it be more efficient if it wrote to a handle directly?

Good question! In fact, Maud did work this way in the past. But it's hard to support buffer reuse in an ergonomic way. The approaches I tried either involved too much boilerplate, or caused mysterious lifetime issues, or both. Moreover, Maud's allocation pattern—with small, short-lived buffers—follow the fast path in modern allocators. These reasons are why I changed `html!` to return a String in version 0.11.

That said, Rust has changed a lot since then, and some of those old assumptions might no longer hold today. So this decision could be revisited prior to the 1.0 release.

### Why is Maud written as a procedural macro? Can't it use macro_rules! instead?

This is certainly possible, and indeed the Horrorshow library works this way. I use procedural macros because they are more flexible. There are some syntax constructs in Maud that are hard to parse with `macro_rules!`; better diagnostics are a bonus as well.

### Maud has had a lot of releases so far. When will it reach 1.0?

I originally planned to cut a 1.0 after implementing stable support. But now that's happened, I've realized that there are a couple design questions that I'd like to resolve before marking that milestone.

### Why doesn't Maud implement context-aware escaping?

I agree that context-aware escaping is very important, especially for the kind of small-scale development that Maud is used for. But it's a complex feature, with security implications, so I want to take the time to get it right.
