use maud::{html, Markup};

/// Input field types
pub enum InputType {
    Text,
    Email,
    Password,
    Number,
    Date,
    Checkbox,
    Radio,
    Hidden,
}

impl InputType {
    fn as_str(&self) -> &'static str {
        match self {
            InputType::Text => "text",
            InputType::Email => "email",
            InputType::Password => "password",
            InputType::Number => "number",
            InputType::Date => "date",
            InputType::Checkbox => "checkbox",
            InputType::Radio => "radio",
            InputType::Hidden => "hidden",
        }
    }
}

/// Form input field component
pub fn input_field(
    input_type: InputType,
    name: &str,
    label: Option<&str>,
    value: Option<&str>,
    placeholder: Option<&str>,
    required: bool,
    attributes: Option<&[(&str, &str)]>,
) -> Markup {
    let id = format!("field-{}", name);
    let input_type_str = input_type.as_str();
    
    html! {
        div.form-group {
            @if let Some(label_text) = label {
                label for=(id) { (label_text) }
            }
            
            input type=(input_type_str) id=(id) name=(name) {
                @if let Some(val) = value {
                    (format!("value=\"{}\"", val))
                }
                @if let Some(ph) = placeholder {
                    (format!("placeholder=\"{}\"", ph))
                }
                @if required {
                    ("required")
                }
                
                @if let Some(attrs) = attributes {
                    @for (attr_name, attr_value) in attrs {
                        @if *attr_value == "" {
                            (attr_name)
                        } @else {
                            (format!("{0}=\"{1}\"", attr_name, attr_value))
                        }
                    }
                }
            }
        }
    }
}

/// Textarea component
pub fn textarea(
    name: &str,
    label: Option<&str>,
    value: Option<&str>,
    placeholder: Option<&str>,
    rows: usize,
    required: bool,
) -> Markup {
    let id = format!("field-{}", name);
    
    html! {
        div.form-group {
            @if let Some(label_text) = label {
                label for=(id) { (label_text) }
            }
            
            textarea id=(id) name=(name) rows=(rows) {
                @if let Some(ph) = placeholder {
                    (format!("placeholder=\"{}\"", ph))
                }
                @if required {
                    ("required")
                }
                @if let Some(content) = value {
                    (content)
                }
            }
        }
    }
}

/// Select dropdown component
pub fn select(
    name: &str,
    label: Option<&str>,
    options: &[(String, String)],
    selected: Option<&str>,
    required: bool,
) -> Markup {
    let id = format!("field-{}", name);
    
    html! {
        div.form-group {
            @if let Some(label_text) = label {
                label for=(id) { (label_text) }
            }
            
            select id=(id) name=(name) {
                @if required {
                    ("required")
                }
                @for (value, text) in options {
                    @if let Some(sel) = selected {
                        @if sel == value {
                            option value=(value) selected { (text) }
                        } @else {
                            option value=(value) { (text) }
                        }
                    } @else {
                        option value=(value) { (text) }
                    }
                }
            }
        }
    }
}

/// Form component with method and action
pub fn form(method: &str, action: &str, content: Markup) -> Markup {
    html! {
        form method=(method) action=(action) {
            (content)
        }
    }
}

/// Form grid layout for side-by-side fields
pub fn form_grid(fields: Vec<Markup>) -> Markup {
    html! {
        div.grid {
            @for field in fields {
                (field)
            }
        }
    }
}
