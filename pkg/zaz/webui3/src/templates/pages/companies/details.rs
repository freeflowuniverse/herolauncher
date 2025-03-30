use maud::{html, Markup};
use chrono::NaiveDate;

/// Company details data structure
#[derive(Debug, Clone)]
pub struct CompanyDetails {
    pub id: u32,
    pub name: String,
    pub registration_number: String,
    pub incorporation_date: NaiveDate,
    pub status: String,
    pub business_type: String,
    pub industry: String,
    pub email: String,
    pub phone: String,
    pub website: Option<String>,
    pub description: Option<String>,
    pub address: Option<String>,
    pub shareholders_count: usize,
}

/// Shareholder data structure for company details
#[derive(Debug, Clone)]
pub struct ShareholderSummary {
    pub id: u32,
    pub name: String,
    pub shares_count: u32,
    pub ownership_percentage: f32,
}

/// Company details page
pub fn company_details(
    company: &CompanyDetails,
    shareholders: &[ShareholderSummary],
) -> Markup {
    html! {
        article.compact {
            header.slim-header {
                h2 { "Company Details" }
                div.header-actions.button-group {
                    a href={"/companies/" (company.id) "/edit"} role="button" class="secondary small" { "Edit Company" }
                    a href="/companies" role="button" class="outline small" { "Back to List" }
                }
            }
            
            div.company-info.grid {
                div.company-profile {
                    h3 { (company.name) }
                    
                    p.registration {
                        strong { "Registration #: " }
                        span { (company.registration_number) }
                    }
                    p.incorporation {
                        strong { "Incorporated: " }
                        span { (company.incorporation_date.format("%Y-%m-%d")) }
                    }
                    p.status {
                        strong { "Status: " }
                        span.status { (company.status) }
                    }
                    p.business-type {
                        strong { "Business Type: " }
                        span { (company.business_type) }
                    }
                    p.industry {
                        strong { "Industry: " }
                        span { (company.industry) }
                    }
                }
                
                div.company-contact {
                    h3 { "Contact Information" }
                    p {
                        strong { "Email: " }
                        a href={"mailto:" (company.email)} { (company.email) }
                    }
                    p {
                        strong { "Phone: " }
                        span { (company.phone) }
                    }
                    p {
                        strong { "Website: " }
                        @if let Some(website) = &company.website {
                            a href=(website) target="_blank" { (website) }
                        } @else {
                            span { "Not provided" }
                        }
                    }
                }
                
                div.company-description {
                    h3 { "Description" }
                    @if let Some(description) = &company.description {
                        p { (description) }
                    } @else {
                        p.no-data { "No description provided" }
                    }
                }
            }
            
            section.shareholders {
                header {
                    h3 { "Shareholders" }
                    span.count { "Total: " (company.shareholders_count) }
                }
                
                @if !shareholders.is_empty() {
                    table {
                        thead {
                            tr {
                                th { "Name" }
                                th { "Shares" }
                                th { "Ownership %" }
                                th { "Actions" }
                            }
                        }
                        tbody {
                            @for shareholder in shareholders {
                                tr {
                                    td { (shareholder.name) }
                                    td { (shareholder.shares_count) }
                                    td { (format!("{:.2}%", shareholder.ownership_percentage)) }
                                    td.actions {
                                        a href={"/shareholders/" (shareholder.id)} title="View Shareholder" { "Details" }
                                    }
                                }
                            }
                        }
                    }
                } @else {
                    p.no-data { "No shareholders registered for this company" }
                    a href={"/shareholders/create?company_id=" (company.id)} role="button" class="secondary" { "Add Shareholder" }
                }
            }
            
            div.action-buttons {
                a href={"/boardmeetings/create?company_id=" (company.id)} role="button" { "Schedule Meeting" }
                a href={"/votes/create?company_id=" (company.id)} role="button" class="secondary" { "Create Vote" }
                a href={"/sales/create?company_id=" (company.id)} role="button" class="contrast" { "Record Sale" }
            }
        }
    }
}
