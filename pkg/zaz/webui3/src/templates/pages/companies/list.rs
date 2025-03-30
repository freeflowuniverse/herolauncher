use maud::{html, Markup};
use crate::templates::components::card;

/// Company data structure
#[derive(Debug, Clone)]
pub struct Company {
    pub id: u32,
    pub name: String,
    pub status: String,
    pub industry: String,
}

/// Companies list page
pub fn companies_list(
    companies: &[Company], 
    count: usize, 
    search: Option<&str>
) -> Markup {
    html! {
        article {
            header.action-header {
                hgroup {
                    h2 { "Companies" }
                    p { "Manage all companies registered in the Freezone" }
                }
                a href="/companies/create" role="button" { "Create Company" }
            }

            div.action-bar {
                form.search-form role="search" action="/companies" method="get" {
                    input type="search" name="search" placeholder="Search companies..." value=[search] {}
                    button type="submit" { "Search" }
                }
            }

            // Debug info
            p { "Found " (count) " companies in the database" }
            
            @if !companies.is_empty() {
                table {
                    thead {
                        tr {
                            th { "ID" }
                            th { "Name" }
                            th { "Status" }
                            th { "Industry" }
                            th { "Actions" }
                        }
                    }
                    tbody {
                        @for company in companies {
                            tr {
                                td { (company.id) }
                                td { (company.name) }
                                td { (company.status) }
                                td { (company.industry) }
                                td.actions {
                                    a href={"/companies/" (company.id)} title="View Details" { "Details" }
                                    a href={"/companies/" (company.id) "/edit"} title="Edit Company" { "Edit" }
                                }
                            }
                        }
                    }
                }
            } @else {
                div.empty-state {
                    p { "No companies found" }
                    
                    @if let Some(search_term) = search {
                        p { "No results for \"" (search_term) "\"" }
                        a href="/companies" { "Clear search" }
                    } @else {
                        a href="/companies/create" role="button" { "Register your first company" }
                    }
                }
            }
        }
    }
}
