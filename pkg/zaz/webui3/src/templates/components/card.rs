use maud::{html, Markup};

/// Card component with optional title and footer
pub fn card(title: Option<&str>, content: Markup, footer: Option<Markup>) -> Markup {
    html! {
        article.card {
            @if let Some(title_text) = title {
                header {
                    h3 { (title_text) }
                }
            }
            
            div.card-content {
                (content)
            }
            
            @if let Some(footer_content) = footer {
                footer {
                    (footer_content)
                }
            }
        }
    }
}

/// Stat card for displaying a statistic with label
pub fn stat_card(label: &str, value: impl std::fmt::Display, unit: Option<&str>) -> Markup {
    html! {
        div.stat-card {
            h4 { (label) }
            p {
                strong { (value) }
                @if let Some(unit_text) = unit {
                    span { " " (unit_text) }
                }
            }
        }
    }
}

/// Activity card for displaying a list of activities
pub fn activity_card(title: &str, activities: &[(String, String)], empty_message: &str) -> Markup {
    html! {
        article {
            header {
                h3 { (title) }
            }
            
            @if !activities.is_empty() {
                ul.activity-list {
                    @for (description, time_ago) in activities {
                        li { (description) " (" (time_ago) ")" }
                    }
                }
            } @else {
                p.no-data { (empty_message) }
            }
        }
    }
}
