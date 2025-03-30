use axum::{
    routing::get,
    Router,
    response::{Html, IntoResponse},
    extract::State,
};
use maud::{html, Markup};
use std::net::SocketAddr;
use std::path::PathBuf;
use tokio::net::TcpListener;
use std::sync::{Arc, Mutex};
use tower_http::services::ServeDir;
use chrono::{Local, Datelike};
use serde::{Serialize, Deserialize};

mod templates;
use templates::layout;
use templates::components::{button, card, form, table};
use templates::pages::companies::{list, create, details, edit};
use axum::extract::Path;

// Sample data structures to match the existing templates
#[derive(Debug, Clone, Serialize, Deserialize)]
struct Company {
    id: u32,
    name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Meeting {
    id: u32,
    date: String,
    title: String,
    company: Company,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Activity {
    description: String,
    time_ago: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct FlashMessage {
    message_type: String,
    text: String,
}

// Application state
struct AppState {
    companies_count: usize,
    active_companies_count: usize,
    shareholders_count: usize,
    recent_activities: Vec<Activity>,
    upcoming_meetings: Vec<Meeting>,
}

#[tokio::main]
async fn main() {
    // Initialize sample data
    let state = Arc::new(Mutex::new(AppState {
        companies_count: 15,
        active_companies_count: 12,
        shareholders_count: 48,
        recent_activities: vec![
            Activity {
                description: "New company registered: TechCorp".to_string(),
                time_ago: "2 hours ago".to_string(),
            },
            Activity {
                description: "Board meeting scheduled for Alpha Inc".to_string(),
                time_ago: "5 hours ago".to_string(),
            },
            Activity {
                description: "Vote completed for Beta Corp".to_string(),
                time_ago: "1 day ago".to_string(),
            },
        ],
        upcoming_meetings: vec![
            Meeting {
                id: 1,
                date: "2025-04-15".to_string(),
                title: "Quarterly Review".to_string(),
                company: Company { id: 1, name: "Alpha Inc".to_string() },
            },
            Meeting {
                id: 2,
                date: "2025-04-20".to_string(),
                title: "Strategic Planning".to_string(),
                company: Company { id: 2, name: "Beta Corp".to_string() },
            },
        ],
    }));

    // Get the current directory for static file serving
    let static_dir = std::env::current_dir().unwrap().join("static");
    println!("Serving static files from: {}", static_dir.display());
    
    // Build our application with routes
    let app = Router::new()
        .route("/", get(index_handler))
        .route("/components", get(components_demo_handler))
        .route("/companies", get(companies_handler))
        .route("/companies/create", get(companies_create_handler))
        .route("/companies/:id", get(company_details_handler))
        .route("/companies/:id/edit", get(company_edit_handler))
        .nest_service("/static", ServeDir::new(static_dir))
        .with_state(state);

    // Run the server
    let addr = SocketAddr::from(([127, 0, 0, 1], 3000));
    let listener = TcpListener::bind(addr).await.unwrap();
    println!("Listening on http://{}", addr);
    axum::serve(listener, app).await.unwrap();
}

// Route handlers
async fn index_handler(State(state): State<Arc<Mutex<AppState>>>) -> impl IntoResponse {
    let state = state.lock().unwrap();
    let current_year = Local::now().year();
    
    let content = html! {
        (index_content(&state, current_year))
    };
    
    let page = layout::layout(
        Some("Dashboard"),
        None,
        None,
        None,
        content,
        current_year,
    );
    
    Html(page.into_string())
}

// Components demo handler to showcase all Maud components
async fn components_demo_handler() -> impl IntoResponse {
    let current_year = Local::now().year();
    
    let content = html! {
        article {
            header {
                h2 { "Maud Components Demo" }
                p { "This page showcases the reusable Maud components" }
            }
            
            // Card components
            section {
                h3 { "Card Components" }
                
                div.grid {
                    // Basic card
                    div {
                        (card::card(
                            Some("Basic Card"),
                            html! { p { "This is a basic card with a title." } },
                            None
                        ))
                    }
                    
                    // Card with footer
                    div {
                        (card::card(
                            Some("Card with Footer"),
                            html! { p { "This card has both a title and a footer." } },
                            Some(html! { p { "Card footer" } })
                        ))
                    }
                    
                    // Stat card
                    div {
                        (card::stat_card("Total Users", 1234, Some("registered")))
                    }
                }
                
                // Activity card
                (card::activity_card(
                    "Recent Activities",
                    &[
                        ("User login".to_string(), "5 minutes ago".to_string()),
                        ("New post created".to_string(), "1 hour ago".to_string()),
                        ("System update".to_string(), "1 day ago".to_string()),
                    ],
                    "No recent activities"
                ))
            }
            
            // Button components
            section {
                h3 { "Button Components" }
                
                div.grid {
                    div {
                        (button::button(
                            "Primary Button",
                            button::ButtonStyle::Primary,
                            None,
                            None,
                            None
                        ))
                    }
                    
                    div {
                        (button::button(
                            "Secondary Button",
                            button::ButtonStyle::Secondary,
                            None,
                            None,
                            None
                        ))
                    }
                    
                    div {
                        (button::button(
                            "Contrast Button",
                            button::ButtonStyle::Contrast,
                            None,
                            None,
                            None
                        ))
                    }
                    
                    div {
                        (button::button(
                            "Outline Button",
                            button::ButtonStyle::Outline,
                            None,
                            None,
                            None
                        ))
                    }
                }
                
                div.grid {
                    div {
                        (button::button(
                            "Link Button",
                            button::ButtonStyle::Primary,
                            Some("#"),
                            None,
                            None
                        ))
                    }
                    
                    div {
                        (button::action_button("View", "#", Some("View details")))
                    }
                }
                
                (button::button_group(vec![
                    button::button("Left", button::ButtonStyle::Primary, None, None, None),
                    button::button("Middle", button::ButtonStyle::Secondary, None, None, None),
                    button::button("Right", button::ButtonStyle::Contrast, None, None, None),
                ]))
            }
            
            // Table components
            section {
                h3 { "Table Components" }
                
                // Simple table
                (table::simple_table(
                    &["Name", "Email", "Role"],
                    &[
                        vec![
                            html! { "John Doe" },
                            html! { "john@example.com" },
                            html! { "Admin" },
                        ],
                        vec![
                            html! { "Jane Smith" },
                            html! { "jane@example.com" },
                            html! { "User" },
                        ],
                    ]
                ))
                
                // Builder pattern table
                (
                    table::Table::new(vec!["ID", "Product", "Price", "Actions"])
                        .with_caption("Products Table")
                        .striped()
                        .hoverable()
                        .add_row(vec![
                            html! { "1" },
                            html! { "Widget" },
                            html! { "$10.00" },
                            html! { (button::action_button("Edit", "#", None)) },
                        ])
                        .add_row(vec![
                            html! { "2" },
                            html! { "Gadget" },
                            html! { "$15.00" },
                            html! { (button::action_button("Edit", "#", None)) },
                        ])
                        .render()
                )
            }
            
            // Form components
            section {
                h3 { "Form Components" }
                
                (form::form("post", "/submit", html! {
                    (form::form_grid(vec![
                        form::input_field(
                            form::InputType::Text,
                            "name",
                            Some("Name"),
                            None,
                            Some("Enter your name"),
                            true,
                            None
                        ),
                        form::input_field(
                            form::InputType::Email,
                            "email",
                            Some("Email"),
                            None,
                            Some("Enter your email"),
                            true,
                            None
                        ),
                    ]))
                    
                    (form::textarea(
                        "message",
                        Some("Message"),
                        None,
                        Some("Enter your message"),
                        4,
                        true
                    ))
                    
                    (form::select(
                        "role",
                        Some("Role"),
                        &[
                            ("user".to_string(), "User".to_string()),
                            ("admin".to_string(), "Administrator".to_string()),
                            ("editor".to_string(), "Editor".to_string()),
                        ],
                        None,
                        true
                    ))
                    
                    (button::button("Submit", button::ButtonStyle::Primary, None, None, None))
                }))
            }
        }
    };
    
    let page = layout::layout(
        Some("Components Demo"),
        None,
        None,
        None,
        content,
        current_year,
    );
    
    Html(page.into_string())
}

// Companies handlers
async fn companies_handler(State(state): State<Arc<Mutex<AppState>>>) -> impl IntoResponse {
    let current_year = Local::now().year();
    
    // Sample company data
    let companies = vec![
        list::Company {
            id: 1,
            name: "Alpha Inc".to_string(),
            status: "Active".to_string(),
            industry: "Technology".to_string(),
        },
        list::Company {
            id: 2,
            name: "Beta Corp".to_string(),
            status: "Active".to_string(),
            industry: "Finance".to_string(),
        },
        list::Company {
            id: 3,
            name: "Gamma LLC".to_string(),
            status: "Pending".to_string(),
            industry: "Manufacturing".to_string(),
        },
    ];
    
    let content = list::companies_list(&companies, companies.len(), None);
    
    let page = layout::layout(
        Some("Companies"),
        None,
        None,
        None,
        content,
        current_year,
    );
    
    Html(page.into_string())
}

async fn companies_create_handler(State(state): State<Arc<Mutex<AppState>>>) -> impl IntoResponse {
    let current_year = Local::now().year();
    
    // Sample form data
    let form_data = create::CompanyForm::default();
    let form_errors: Vec<String> = Vec::new();
    
    // Sample select options
    let months = vec![
        create::SelectOption { value: "1".to_string(), name: "January".to_string() },
        create::SelectOption { value: "2".to_string(), name: "February".to_string() },
        create::SelectOption { value: "3".to_string(), name: "March".to_string() },
        // ... other months
    ];
    
    let business_types = vec![
        create::SelectOption { value: "corporation".to_string(), name: "Corporation".to_string() },
        create::SelectOption { value: "llc".to_string(), name: "LLC".to_string() },
        create::SelectOption { value: "partnership".to_string(), name: "Partnership".to_string() },
    ];
    
    let industries = vec![
        create::SelectOption { value: "technology".to_string(), name: "Technology".to_string() },
        create::SelectOption { value: "finance".to_string(), name: "Finance".to_string() },
        create::SelectOption { value: "manufacturing".to_string(), name: "Manufacturing".to_string() },
    ];
    
    let content = create::company_create(
        &form_data,
        &form_errors,
        &months,
        &business_types,
        &industries,
        Some("csrf_token_placeholder"),
    );
    
    let page = layout::layout(
        Some("Create Company"),
        None,
        None,
        None,
        content,
        current_year,
    );
    
    Html(page.into_string())
}

async fn company_details_handler(Path(id): Path<u32>, State(state): State<Arc<Mutex<AppState>>>) -> impl IntoResponse {
    let current_year = Local::now().year();
    
    // Sample company data
    let company = details::CompanyDetails {
        id,
        name: format!("Company {}", id),
        registration_number: format!("REG{:06}", id),
        incorporation_date: chrono::NaiveDate::from_ymd_opt(2020, 1, 15).unwrap(),
        status: "Active".to_string(),
        business_type: "Corporation".to_string(),
        industry: "Technology".to_string(),
        email: format!("contact@company{}.com", id),
        phone: "+1 234 567 8900".to_string(),
        website: Some(format!("https://company{}.com", id)),
        description: Some("This is a sample company description.".to_string()),
        address: Some("123 Business St, Tech City, TC 12345".to_string()),
        shareholders_count: 3,
    };
    
    // Sample shareholders data
    let shareholders = vec![
        details::ShareholderSummary {
            id: 1,
            name: "John Doe".to_string(),
            shares_count: 500,
            ownership_percentage: 50.0,
        },
        details::ShareholderSummary {
            id: 2,
            name: "Jane Smith".to_string(),
            shares_count: 300,
            ownership_percentage: 30.0,
        },
        details::ShareholderSummary {
            id: 3,
            name: "Bob Johnson".to_string(),
            shares_count: 200,
            ownership_percentage: 20.0,
        },
    ];
    
    let content = details::company_details(&company, &shareholders);
    
    let page = layout::layout(
        Some(&format!("Company Details - {}", company.name)),
        None,
        None,
        None,
        content,
        current_year,
    );
    
    Html(page.into_string())
}

async fn company_edit_handler(Path(id): Path<u32>, State(state): State<Arc<Mutex<AppState>>>) -> impl IntoResponse {
    let current_year = Local::now().year();
    
    // Sample company data for editing
    let company = edit::CompanyEditForm {
        id,
        name: format!("Company {}", id),
        registration_number: format!("REG{:06}", id),
        business_type: "Corporation".to_string(),
        industry: "Technology".to_string(),
        status: "Active".to_string(),
        email: format!("contact@company{}.com", id),
        phone: "+1 234 567 8900".to_string(),
        website: format!("https://company{}.com", id),
        description: "This is a sample company description.".to_string(),
        address: "123 Business St, Tech City, TC 12345".to_string(),
    };
    
    let content = edit::company_edit(&company, Some("csrf_token_placeholder"));
    
    let page = layout::layout(
        Some(&format!("Edit Company - {}", company.name)),
        None,
        None,
        None,
        content,
        current_year,
    );
    
    Html(page.into_string())
}

// Index page content
fn index_content(state: &AppState, _current_year: i32) -> Markup {
    html! {
        article {
            header {
                h2 { "Freezone Manager Dashboard" }
                p { "Welcome to the Freezone Management System" }
            }
            
            div.grid {
                div {
                    article {
                        header {
                            h3 { "Company Statistics" }
                        }
                        
                        div.grid {
                            div {
                                h4 { "Total Companies" }
                                p {
                                    strong { (state.companies_count) }
                                    span { " registered" }
                                }
                            }
                            
                            div {
                                h4 { "Active Companies" }
                                p {
                                    strong { (state.active_companies_count) }
                                    span { " companies" }
                                }
                            }
                            
                            div {
                                h4 { "Total Shareholders" }
                                p {
                                    strong { (state.shareholders_count) }
                                    span { " shareholders" }
                                }
                            }
                        }
                    }
                }
                
                div {
                    article {
                        header {
                            h3 { "Recent Activity" }
                        }
                        
                        @if !state.recent_activities.is_empty() {
                            ul {
                                @for activity in &state.recent_activities {
                                    li { (activity.description) " (" (activity.time_ago) ")" }
                                }
                            }
                        } @else {
                            p.no-data { "No recent activities to display." }
                        }
                    }
                }
            }
            
            article {
                header {
                    h3 { "Quick Actions" }
                }
                
                div.grid {
                    div {
                        a href="/companies/create" role="button" { "Register Company" }
                    }
                    
                    div {
                        a href="/boardmeetings/create" role="button" class="secondary" { "Schedule Meeting" }
                    }
                    
                    div {
                        a href="/votes/create" role="button" class="contrast" { "Create Vote" }
                    }
                }
            }
            
            @if !state.upcoming_meetings.is_empty() {
                article {
                    header {
                        h3 { "Upcoming Meetings" }
                    }
                    
                    table {
                        thead {
                            tr {
                                th { "Date" }
                                th { "Company" }
                                th { "Title" }
                                th { "Actions" }
                            }
                        }
                        tbody {
                            @for meeting in &state.upcoming_meetings {
                                tr {
                                    td { (meeting.date) }
                                    td { (meeting.company.name) }
                                    td { (meeting.title) }
                                    td {
                                        a.action-button href={"/boardmeetings/" (meeting.id)} title="View Details" { "Details" }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
