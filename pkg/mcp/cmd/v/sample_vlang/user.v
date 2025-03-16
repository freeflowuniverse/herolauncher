// User represents a system user
pub struct User {
	id string
	name string
	email string
	is_active bool
}

// CreateUser creates a new user in the system
pub fn (u &User) CreateUser() {}

pub enum Role {
	admin
	user
	guest
}
