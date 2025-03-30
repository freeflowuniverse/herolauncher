# Webui3 - Maud-based Web UI for Herolauncher

This is a Rust implementation of the Freezone Manager web interface using the Maud templating engine. It replaces the previous template-based approach with Maud's compile-time HTML generation.

## Features

- **Compile-time HTML generation**: Templates are checked at compile time, eliminating runtime template errors
- **Type safety**: Full Rust type checking for all templates
- **Performance**: Minimal runtime overhead compared to traditional template engines
- **Seamless Rust integration**: Templates work directly with Rust data structures

## Project Structure

```
webui3/
├── Cargo.toml          # Rust project dependencies
├── src/
│   ├── main.rs         # Application entry point and route handlers
│   └── templates/      # Maud templates
│       ├── layout.rs   # Base layout template
│       └── partials/   # Reusable UI components
│           ├── header.rs
│           ├── sidebar.rs
│           └── right_sidebar.rs
```

## Getting Started

1. Make sure you have Rust installed
2. Build the project:

```bash
cd webui3
cargo build
```

3. Run the server:

```bash
cargo run
```

4. Open your browser at http://localhost:3000

## How Maud Works

Maud is a macro-based HTML template engine for Rust. Instead of using a separate template language, you write HTML directly in Rust code using the `html!` macro:

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

This gets compiled to efficient Rust code that generates HTML at runtime.

## Advantages Over Traditional Templates

- No need to escape between languages - everything is Rust
- Compile-time checking catches errors early
- Full IDE support for code completion and refactoring
- Direct access to Rust data structures without serialization
- Better performance with minimal runtime overhead
