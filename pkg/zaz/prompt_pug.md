# Pug Templates: A Comprehensive Guide

IMPORTANT

- pug in go uses {{ }} for delimiters for variables
- for if and else and range need to use | {{ }}

## Syntax Guide

### 1. Basic Structure

```pug
// Pug syntax for HTML structure
header
  h2 {{.title}}  // Go template syntax for dynamic content
```

### 2. Conditionals

```pug
// If statement
| {{if condition}}
div This appears when condition is true
| {{end}}

// If-else statement
| {{if condition}}
div True condition
| {{else}}
div False condition
| {{end}}

// Checking length
if len .mailPreviews
  // Content when collection has items
else
  // Content when collection is empty
```

### 3. Loops

```pug
// Looping through collections
| {{range .mailPreviews}}
li
  .item {{.Subject}}  // Access current item's field
| {{end}}
```

### 4. Data Access

**NOTE**: No `|` needed

```pug
// Accessing struct fields with dot notation
h3 {{.mail.Subject}}
p {{.mail.Body}}

// Nested structures
span {{.user.Profile.Name}}
```

### 5. Custom Functions
```pug
// Using registered template functions
h2 {{title .name}}  // Using a custom "title" function
div | .htmlContent
```

## Common Mistakes to Avoid

1. **Mixing delimiters**: Always use `{{ }}` for Go templating, not other delimiters
2. **Indentation issues**: Maintain proper Pug indentation while inserting Go template directives
3. **Missing end tags**: Every `| {{if}}` or `| {{range}}` must have a corresponding `| {{end}}`
4. **Incorrect field access**: Use dot notation (`.fieldName`) to access struct fields
5. **Function usage**: Custom functions must be registered with `engine.AddFunc()` before use

## Example from Your Codebase

```pug
| {{if len .mailPreviews}}
  ul.mail-preview-list
    | {{range .mailPreviews}}
    li
      a.mail-preview(
        href="/mail/{{.ID}}" 
        up-target=".mail-content" 
        class="{{if not .Read}}unread{{end}}"
      )
        hgroup
          h3.mail-preview-from {{.From}}
          small.mail-preview-date {{.Date}}
        .mail-preview-subject
          | {{if .Starred}}
          span.star ★
          | {{end}}
          | {{.Subject}}
    | {{end}}
else
  .empty-mailbox
    p No emails in this mailbox.
```

## Resources for Reference

1. [Go html/template Documentation](https://pkg.go.dev/html/template)
2. [Fiber Template Documentation](https://docs.gofiber.io/template/html)
3. [Pug Template Syntax](https://pugjs.org/language/attributes.html) - be careful use golang template
4. [Fiber Pug Template Engine](https://github.com/gofiber/template/tree/master/pug)
