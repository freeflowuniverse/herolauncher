package tools

import "testing"

func TestNameFix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello_world"},
		{"Hello_World", "hello_world"},
		{"Hello__World", "hello_world"},
		{"Hello-World", "hello_world"},
		{"Hello,World", "hello_world"},
		{"Hello,World.MD", "hello_world.md"},
		{"Hello,World.MD", "hello_world.md"},
		{"UPPER case", "upper_case"},
		{"Mixed-Case,String", "mixed_case_string"},
		{"Non-ASCII: éñçödéd", "non_ascii_dd"},
		{"Multiple   spaces", "multiple_spaces"},
		{"Symbols: !@#$%^&*()", "symbols_"},
		{"Getting- Started", "getting_started"},
	}

	for _, test := range tests {
		result := NameFix(test.input)
		if result != test.expected {
			t.Errorf("NameFix(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}
