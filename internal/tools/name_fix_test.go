package tools

import "testing"

func TestNameFix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello_world"},
		{"Hello-World", "hello_world"},
		{"Hello,World", "hello_world"},
		{"UPPER case", "upper_case"},
		{"Mixed-Case,String", "mixed_case_string"},
		{"Non-ASCII: éñçödéd", "non_ascii__dd"},
		{"Multiple   spaces", "multiple___spaces"},
		{"Symbols: !@#$%^&*()", "symbols____________"},
		{"", ""},
	}

	for _, test := range tests {
		result := NameFix(test.input)
		if result != test.expected {
			t.Errorf("NameFix(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}
