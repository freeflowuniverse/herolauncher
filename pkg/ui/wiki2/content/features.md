# Key Features

![Feature Performance Chart](/images/chart.svg)

## Dynamic Content Loading

This application uses Unpoly to load content dynamically without refreshing the entire page. When you click on a link in the sidebar, only the content area is updated, providing a smooth, app-like experience.

## Markdown Rendering

All content is written in Markdown and rendered as HTML on the server side using Goldmark, a fast Markdown parser for Go.

## Responsive Design

The application is built with PicoCSS, a minimal CSS framework that provides beautiful defaults without requiring any classes. This ensures the application looks great on all devices, from mobile phones to desktop computers.

## Go Fiber Backend

The backend is powered by Go Fiber, a fast and lightweight web framework inspired by Express.js. This provides excellent performance and a clean API for handling HTTP requests.

## Pug Templates

The user interface is built using Pug templates, which offer a clean, whitespace-sensitive syntax for writing HTML. This makes the templates more readable and maintainable.

## Image Support

![Application Flow Diagram](/images/diagram.svg)

The application fully supports embedding images in markdown files, as demonstrated above. Images are rendered with proper scaling and alignment.
