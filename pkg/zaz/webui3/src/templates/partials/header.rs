use maud::{html, Markup};

pub fn header() -> Markup {
    html! {
        // Header component for the Freezone Manager using PicoCSS native approach
        header.container {
            nav {
                // Logo on the left
                ul {
                    li {
                        a href="/" {
                            img src="/img/freezone-icon.svg" alt="Freezone Logo" width="24" height="24";
                            strong { " Freezone Manager" }
                        }
                    }
                }
                
                // Navigation items in the center
                ul {
                    li { a.secondary href="/" { "Home" } }
                    li { a.secondary href="/companies" { "Companies" } }
                    li { a.secondary href="/shareholders" { "Shareholders" } }
                    li { a.secondary href="/boardmeetings" { "Meetings" } }
                    li { a.secondary href="/votes" { "Votes" } }
                    li { a.secondary href="/sales" { "Sales" } }
                }
                
                // Profile dropdown on the right
                ul {
                    li {
                        details.dropdown {
                            summary { "Account" }
                            ul dir="rtl" {
                                li { a href="/profile" { "Profile" } }
                                li { a href="/settings" { "Settings" } }
                                li { a href="/logout" { "Logout" } }
                            }
                        }
                    }
                }
            }
        }
    }
}
