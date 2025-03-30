use maud::{html, Markup};

// Example of how to create a new page in the Maud-based webui3
pub fn example_page(title: &str, items: &[String]) -> Markup {
    html! {
        article {
            header {
                h2 { (title) }
                p { "This is an example page created with Maud" }
            }
            
            section {
                h3 { "Dynamic Content Example" }
                
                @if !items.is_empty() {
                    ul {
                        @for item in items {
                            li { (item) }
                        }
                    }
                } @else {
                    p { "No items available" }
                }
            }
            
            section {
                h3 { "Form Example" }
                
                form method="post" action="/submit" {
                    div.grid {
                        label for="name" { "Name:" }
                        input type="text" id="name" name="name" required;
                    }
                    
                    div.grid {
                        label for="email" { "Email:" }
                        input type="email" id="email" name="email" required;
                    }
                    
                    div.grid {
                        label for="message" { "Message:" }
                        textarea id="message" name="message" rows="4" required {}
                    }
                    
                    button type="submit" { "Submit" }
                }
            }
        }
    }
}
