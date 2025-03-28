# About MD Viewer

<div style="text-align: center;">
  <img src="/images/screenshot.svg" alt="MD Viewer Screenshot" width="350" style="border: 1px solid #ddd; border-radius: 8px; margin: 20px auto;">
</div>

## Project Overview

This project demonstrates how to create a simple yet powerful Markdown viewer using modern web technologies.

## Technologies Used

<img src="/images/diagram.svg" alt="Architecture Diagram" style="float: right; width: 200px; margin-left: 15px; margin-bottom: 10px;">

### PicoCSS
A minimal CSS framework that provides beautiful defaults without any classes. It makes creating clean, responsive interfaces incredibly simple.

### Pug Templates
A template engine that offers a clean, whitespace-sensitive syntax for writing HTML, integrated with Go's Fiber framework.

### Unpoly
A JavaScript library that enhances HTML applications with the benefits of a single-page application without requiring you to write any JavaScript.

### Go Fiber
A fast, flexible, and minimalist web framework for Go, inspired by Express.js.

## How It Works

The application serves Markdown files statically while using Unpoly to dynamically fetch and render content without full page reloads. This creates a smooth, app-like experience while maintaining the simplicity of a static site.

<div style="display: flex; justify-content: space-between; margin-top: 30px;">
  <img src="/images/chart.svg" alt="Performance Chart" width="45%">
  <img src="/images/diagram.svg" alt="Flow Diagram" width="45%">
</div>

*The charts above illustrate the application's performance and data flow.*
