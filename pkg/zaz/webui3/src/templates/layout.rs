use maud::{html, DOCTYPE, Markup};
use super::partials;

pub fn layout(
    page_title: Option<&str>,
    body_class: Option<&str>,
    flash_message: Option<(String, String)>,
    custom_css: Option<&str>,
    content: Markup,
    current_year: i32,
) -> Markup {
    html! {
        (DOCTYPE)
        html lang="en" {
            head {
                meta charset="UTF-8";
                meta name="viewport" content="width=device-width, initial-scale=1.0";
                
                @if let Some(title) = page_title {
                    title { (title) " - Freezone Manager" }
                } @else {
                    title { "Freezone Manager" }
                }
                
                link rel="icon" href="/static/img/freezone-icon.svg" type="image/svg+xml";
                link rel="shortcut icon" href="/static/favicon.ico";
                link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap";
                link rel="stylesheet" href="/static/css/style.css";
                link rel="stylesheet" href="/static/css/pico.min.css";
                link rel="stylesheet" href="/static/css/freezone.css";
                link rel="stylesheet" href="/static/css/right-sidebar.css";
                link rel="stylesheet" href="/static/css/unpoly.min.css";
                
                @if let Some(css) = custom_css {
                    style { (css) }
                }
            }
            
            body class=(body_class.unwrap_or("")) {
                @if let Some((message_type, text)) = flash_message {
                    div class={"flash-message " (message_type)} {
                        div.container {
                            p { (text) }
                            button.close-flash { "×" }
                        }
                    }
                }
                
                (partials::header::header())
                
                div.container {
                    div.grid.grid-sidebar {
                        // Left sidebar
                        (partials::sidebar::sidebar())
                        
                        // Main content area
                        main.content-area {
                            (content)
                        }
                        
                        // Right sidebar
                        (partials::right_sidebar::right_sidebar())
                    }
                }
                
                footer {
                    div.container {
                        p { "© " (current_year) " Freezone Manager - A FreeFlow Universe Project" }
                    }
                }
                
                script src="/static/js/unpoly.min.js" {}
                script src="/static/js/echarts/echarts.min.js" {}
                script src="/static/js/freezone.js" {}
                script src="/static/js/sidebar.js" {}
            }
        }
    }
}
