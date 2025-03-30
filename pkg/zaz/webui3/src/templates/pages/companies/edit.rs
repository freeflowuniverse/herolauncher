use maud::{html, Markup};

/// Company edit form data
#[derive(Debug, Clone)]
pub struct CompanyEditForm {
    pub id: u32,
    pub name: String,
    pub registration_number: String,
    pub business_type: String,
    pub industry: String,
    pub status: String,
    pub email: String,
    pub phone: String,
    pub website: String,
    pub description: String,
    pub address: String,
}

/// Company edit page
pub fn company_edit(
    company: &CompanyEditForm,
    csrf_token: Option<&str>,
) -> Markup {
    html! {
        article.compact {
            header.slim-header {
                h2 { "Edit Company" }
                div.header-actions.button-group {
                    a href={"/companies/" (company.id)} role="button" class="outline small" { "Cancel" }
                    a href="/companies" role="button" class="outline small" { "Back to List" }
                }
            }
            
            form action={"/companies/" (company.id) "/edit"} method="POST" {
                @if let Some(token) = csrf_token {
                    input type="hidden" name="csrf_token" value=(token) {}
                }
                
                div.grid {
                    div.form-group {
                        label for="name" { "Company Name" }
                        input type="text" id="name" name="name" value=(company.name) {}
                    }
                    
                    div.form-group {
                        label for="registration_number" { "Registration Number" }
                        input type="text" id="registration_number" name="registration_number" value=(company.registration_number) {}
                    }
                    
                    div.form-group {
                        label for="business_type" { "Business Type" }
                        select id="business_type" name="business_type" {
                            @for business_type in &["Corporation", "LLC", "Partnership", "Sole Proprietorship", "Non-Profit"] {
                                @if *business_type == company.business_type {
                                    option value=(business_type) selected { (business_type) }
                                } @else {
                                    option value=(business_type) { (business_type) }
                                }
                            }
                        }
                    }
                    
                    div.form-group {
                        label for="industry" { "Industry" }
                        input type="text" id="industry" name="industry" value=(company.industry) {}
                    }
                    
                    div.form-group {
                        label for="status" { "Status" }
                        select id="status" name="status" {
                            @for status in &["Active", "Inactive", "Pending", "Dissolved"] {
                                @if *status == company.status {
                                    option value=(status) selected { (status) }
                                } @else {
                                    option value=(status) { (status) }
                                }
                            }
                        }
                    }
                    
                    div.form-group {
                        label for="email" { "Email" }
                        input type="email" id="email" name="email" value=(company.email) {}
                    }
                    
                    div.form-group {
                        label for="phone" { "Phone" }
                        input type="tel" id="phone" name="phone" value=(company.phone) {}
                    }
                    
                    div.form-group {
                        label for="website" { "Website" }
                        input type="url" id="website" name="website" value=(company.website) {}
                    }
                    
                    div.form-group.full-width {
                        label for="address" { "Address" }
                        textarea id="address" name="address" rows="3" {
                            (company.address)
                        }
                    }
                    
                    div.form-group.full-width {
                        label for="description" { "Description" }
                        textarea id="description" name="description" rows="4" {
                            (company.description)
                        }
                    }
                }
                
                div.form-actions {
                    button type="submit" { "Save Changes" }
                    a href={"/companies/" (company.id)} role="button" class="secondary" { "Cancel" }
                }
            }
        }
    }
}
