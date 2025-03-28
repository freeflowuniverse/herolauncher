# How to Use MD Viewer

## Navigation

Navigate between different markdown files using the sidebar navigation. Each link will load the content dynamically without refreshing the entire page, thanks to Unpoly.

## Adding New Content

To add new markdown files:

1. Place your `.md` files in the `content` directory
2. The application will automatically detect and list them in the navigation
3. Files will be accessible via their filename (without the `.md` extension)

## Markdown Support

This viewer supports standard Markdown syntax including:

- **Bold** and *italic* text
- # Headers of different levels
- Lists (ordered and unordered)
- [Links](https://example.com)
- Code blocks with syntax highlighting
- > Blockquotes
- Tables
- And more!

## Customization

The application uses PicoCSS for styling, which means you can easily customize the appearance by modifying the CSS variables in the custom stylesheet.
