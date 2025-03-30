use maud::{html, Markup};
use crate::templates::components::form;

/// Month option for select fields
#[derive(Debug, Clone)]
pub struct SelectOption {
    pub value: String,
    pub name: String,
}

/// Company creation form data
#[derive(Debug, Clone, Default)]
pub struct CompanyForm {
    pub name: String,
    pub registration_number: String,
    pub incorporation_date: String,
    pub fiscal_year_end: String,
    pub email: String,
    pub phone: String,
    pub website: String,
    pub address: String,
    pub business_type: String,
    pub industry: String,
    pub description: String,
}

/// Company creation page
pub fn company_create(
    form_data: &CompanyForm,
    form_errors: &[String],
    months: &[SelectOption],
    business_types: &[SelectOption],
    industries: &[SelectOption],
    csrf_token: Option<&str>,
) -> Markup {
    html! {
        article {
            header {
                h2 { "Create New Company" }
                p { "Register a new company in the Freezone" }
            }
            
            @if !form_errors.is_empty() {
                div.form-errors {
                    h4 { "Please correct the following errors:" }
                    ul {
                        @for error in form_errors {
                            li { (error) }
                        }
                    }
                }
            }
            
            form action="/companies/create" method="post" {
                section {
                    h3 { "Company Information" }
                    
                    @if let Some(token) = csrf_token {
                        input type="hidden" name="csrf_token" value=(token) {}
                    }
                    
                    div.grid {
                        div {
                            label for="name" { "Company Name" }
                            input type="text" id="name" name="name" placeholder="Enter company name" value=(form_data.name) {}
                        }
                        
                        div {
                            label for="registration_number" { "Registration Number" }
                            input type="text" id="registration_number" name="registration_number" placeholder="e.g. BRN12345678" value=(form_data.registration_number) {}
                        }
                    }
                    
                    div.grid {
                        div {
                            label for="incorporation_date" { "Incorporation Date" }
                            input type="date" id="incorporation_date" name="incorporation_date" value=(form_data.incorporation_date) {}
                        }
                        
                        div {
                            label for="fiscal_year_end" { "Fiscal Year End" }
                            select id="fiscal_year_end" name="fiscal_year_end" {
                                option value="" { "Select month" }
                                @for month in months {
                                    @if month.value == form_data.fiscal_year_end {
                                        option value=(month.value) selected { (month.name) }
                                    } @else {
                                        option value=(month.value) { (month.name) }
                                    }
                                }
                            }
                        }
                    }
                }
                
                hr {}
                
                section {
                    h3 { "Contact Information" }
                    div.grid {
                        div {
                            label for="email" { "Email" }
                            input type="email" id="email" name="email" placeholder="company@example.com" value=(form_data.email) {}
                        }
                        
                        div {
                            label for="phone" { "Phone" }
                            input type="tel" id="phone" name="phone" placeholder="+1 234 567 8900" value=(form_data.phone) {}
                        }
                    }
                    
                    div.grid {
                        div {
                            label for="website" { "Website" }
                            input type="url" id="website" name="website" placeholder="https://example.com" value=(form_data.website) {}
                        }
                        
                        div {
                            label for="address" { "Address" }
                            textarea id="address" name="address" placeholder="Company Address" rows="3" {
                                (form_data.address)
                            }
                        }
                    }
                }
                
                fieldset {
                    legend { "Company Details" }
                    div.grid {
                        div {
                            label for="business_type" { "Business Type" }
                            select id="business_type" name="business_type" {
                                option value="" { "Select type" }
                                @for business_type in business_types {
                                    @if business_type.value == form_data.business_type {
                                        option value=(business_type.value) selected { (business_type.name) }
                                    } @else {
                                        option value=(business_type.value) { (business_type.name) }
                                    }
                                }
                            }
                        }
                        
                        div {
                            label for="industry" { "Industry" }
                            select id="industry" name="industry" {
                                option value="" { "Select industry" }
                                @for industry in industries {
                                    @if industry.value == form_data.industry {
                                        option value=(industry.value) selected { (industry.name) }
                                    } @else {
                                        option value=(industry.value) { (industry.name) }
                                    }
                                }
                            }
                        }
                    }
                    
                    div {
                        label for="description" { "Company Description" }
                        textarea id="description" name="description" placeholder="Brief description of the company" rows="4" {
                            (form_data.description)
                        }
                    }
                }
                
                div.form-actions {
                    button type="submit" { "Create Company" }
                    a href="/companies" role="button" class="secondary" { "Cancel" }
                }
            }
        }
    }
}
