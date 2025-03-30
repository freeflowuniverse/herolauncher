use maud::{html, Markup};

/// Table component for displaying tabular data
pub struct Table<'a> {
    headers: Vec<&'a str>,
    rows: Vec<Vec<Markup>>,
    caption: Option<&'a str>,
    striped: bool,
    hoverable: bool,
    bordered: bool,
    compact: bool,
}

impl<'a> Table<'a> {
    /// Create a new table with headers
    pub fn new(headers: Vec<&'a str>) -> Self {
        Self {
            headers,
            rows: Vec::new(),
            caption: None,
            striped: false,
            hoverable: false,
            bordered: false,
            compact: false,
        }
    }
    
    /// Add a caption to the table
    pub fn with_caption(mut self, caption: &'a str) -> Self {
        self.caption = Some(caption);
        self
    }
    
    /// Add striped rows to the table
    pub fn striped(mut self) -> Self {
        self.striped = true;
        self
    }
    
    /// Add hover effect to rows
    pub fn hoverable(mut self) -> Self {
        self.hoverable = true;
        self
    }
    
    /// Add borders to the table
    pub fn bordered(mut self) -> Self {
        self.bordered = true;
        self
    }
    
    /// Make the table more compact
    pub fn compact(mut self) -> Self {
        self.compact = true;
        self
    }
    
    /// Add a row to the table
    pub fn add_row(mut self, cells: Vec<Markup>) -> Self {
        if cells.len() != self.headers.len() {
            panic!("Row length must match header length");
        }
        self.rows.push(cells);
        self
    }
    
    /// Render the table as Markup
    pub fn render(self) -> Markup {
        let mut classes = Vec::new();
        
        if self.striped {
            classes.push("striped");
        }
        
        if self.hoverable {
            classes.push("hoverable");
        }
        
        if self.bordered {
            classes.push("bordered");
        }
        
        if self.compact {
            classes.push("compact");
        }
        
        let class_str = classes.join(" ");
        
        html! {
            table class=(class_str) {
                @if let Some(caption_text) = self.caption {
                    caption { (caption_text) }
                }
                
                thead {
                    tr {
                        @for header in self.headers {
                            th { (header) }
                        }
                    }
                }
                
                tbody {
                    @for row in self.rows {
                        tr {
                            @for cell in row {
                                td { (cell) }
                            }
                        }
                    }
                }
            }
        }
    }
}

/// Simple function to create a basic table with headers and rows
pub fn simple_table(headers: &[&str], rows: &[Vec<Markup>]) -> Markup {
    html! {
        table {
            thead {
                tr {
                    @for header in headers {
                        th { (header) }
                    }
                }
            }
            
            tbody {
                @for row in rows {
                    tr {
                        @for cell in row {
                            td { (cell) }
                        }
                    }
                }
            }
        }
    }
}
