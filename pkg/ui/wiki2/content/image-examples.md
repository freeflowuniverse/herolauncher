# Image Examples in Markdown

This page demonstrates various ways to include and format images in your markdown files.

## Basic Image

The simplest way to add an image is using the standard markdown syntax:

![Basic Screenshot](/images/screenshot.svg)

## Image with HTML Attributes

For more control, you can use HTML to add specific styling:

<img src="/images/diagram.svg" alt="Flow Diagram" width="300" height="150" style="border: 1px solid #ddd; border-radius: 5px;">

## Aligned Images

### Center Alignment

<div style="text-align: center;">
  <img src="/images/chart.svg" alt="Centered Chart" width="350">
  <p><em>A centered chart with caption</em></p>
</div>

### Right Alignment

<img src="/images/diagram.svg" alt="Right-aligned Diagram" style="float: right; width: 200px; margin-left: 15px;">

This text will wrap around the right-aligned image. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla facilisi. Maecenas non magna eget nunc tincidunt faucibus. Duis vitae nisi vel nunc commodo lacinia.

<div style="clear: both;"></div>

## Side-by-Side Images

<div style="display: flex; justify-content: space-between; gap: 20px;">
  <div style="flex: 1;">
    <img src="/images/screenshot.svg" alt="Screenshot" style="width: 100%;">
    <p><em>Application Interface</em></p>
  </div>
  <div style="flex: 1;">
    <img src="/images/chart.svg" alt="Chart" style="width: 100%;">
    <p><em>Performance Metrics</em></p>
  </div>
</div>

## Responsive Images

This image will resize based on the screen width:

<img src="/images/diagram.svg" alt="Responsive Diagram" style="max-width: 100%; height: auto;">

## Image with Border and Shadow

<img src="/images/screenshot.svg" alt="Styled Screenshot" style="max-width: 80%; border: 2px solid #ddd; border-radius: 8px; box-shadow: 0 4px 8px rgba(0,0,0,0.1); margin: 20px auto; display: block;">

## Image Gallery

<div style="display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; margin: 20px 0;">
  <img src="/images/diagram.svg" alt="Diagram" style="width: 100%; border-radius: 5px;">
  <img src="/images/chart.svg" alt="Chart" style="width: 100%; border-radius: 5px;">
  <img src="/images/screenshot.svg" alt="Screenshot" style="width: 100%; border-radius: 5px;">
</div>

## Background Image (Using CSS)

<div style="background-image: url('/images/diagram.svg'); background-size: contain; background-repeat: no-repeat; background-position: center; height: 200px; display: flex; align-items: center; justify-content: center;">
  <div style="background-color: rgba(255,255,255,0.7); padding: 20px; border-radius: 5px;">
    <h3 style="margin: 0;">Text Over Image</h3>
    <p style="margin: 10px 0 0 0;">This text is displayed on top of a background image</p>
  </div>
</div>

These examples demonstrate the flexibility of image formatting in markdown when combined with HTML and CSS.
