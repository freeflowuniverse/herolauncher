use maud::{html, Markup};

/// Button style variants
pub enum ButtonStyle {
    Primary,
    Secondary,
    Contrast,
    Outline,
}

impl ButtonStyle {
    fn class_name(&self) -> &'static str {
        match self {
            ButtonStyle::Primary => "",
            ButtonStyle::Secondary => "secondary",
            ButtonStyle::Contrast => "contrast",
            ButtonStyle::Outline => "outline",
        }
    }
}

/// Button component with configurable style and attributes
pub fn button(
    text: &str,
    style: ButtonStyle,
    href: Option<&str>,
    icon: Option<&str>,
    attributes: Option<&[(&str, &str)]>,
) -> Markup {
    let class_name = style.class_name();
    
    html! {
        @if let Some(link) = href {
            a href=(link) class=(class_name) role="button" {
                @if let Some(icon_name) = icon {
                    span.icon { (icon_name) }
                }
                (text)
            }
        } @else {
            button class=(class_name) {
                @if let Some(icon_name) = icon {
                    span.icon { (icon_name) }
                }
                (text)
                
                @if let Some(attrs) = attributes {
                    @for (name, value) in attrs {
                        @if *value == "" {
                            (name)
                        } @else {
                            (format!("{0}=\"{1}\"", name, value))
                        }
                    }
                }
            }
        }
    }
}

/// Action button for tables and lists
pub fn action_button(text: &str, href: &str, title: Option<&str>) -> Markup {
    html! {
        a.action-button href=(href) title=(title.unwrap_or(text)) { (text) }
    }
}

/// Button group component for grouping related buttons
pub fn button_group(buttons: Vec<Markup>) -> Markup {
    html! {
        div.button-group {
            @for button in buttons {
                (button)
            }
        }
    }
}
