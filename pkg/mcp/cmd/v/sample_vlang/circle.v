// Circle represents a geometric shape
pub struct Circle {
	radius f64
	center Point
}

// Point represents a 2D coordinate
pub struct Point {
	x f64
	y f64
}

// Area calculates the area of the circle
pub fn (c &Circle) Area() f64 {}

// Circumference calculates the circumference of the circle
pub fn (c &Circle) Circumference() f64 {}

// ShapeType represents different geometric shapes
pub enum ShapeType {
	circle
	rectangle
	triangle
}
