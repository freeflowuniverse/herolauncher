package paramsparser

import (
	"testing"
)

func TestParamsParserBasic(t *testing.T) {
	input := "name: 'myname' host: 'localhost' port: 25 secure: 1 reset: 1"
	parser := New()
	err := parser.ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"string value", "name", "myname"},
		{"another string value", "host", "localhost"},
		{"numeric value", "port", "25"},
		{"boolean-like value", "secure", "1"},
		{"another boolean-like value", "reset", "1"},
		{"non-existent key", "nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.Get(tt.key); got != tt.expected {
				t.Errorf("ParamsParser.Get(%q) = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}

func TestParamsParserMultiline(t *testing.T) {
	input := "name: 'myname' description: '\n\t\ta description can be multiline\n\n\t\tlike this\n'"
	parser := New()
	err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}

	// Check the name parameter
	if got := parser.Get("name"); got != "myname" {
		t.Errorf("ParamsParser.Get(\"name\") = %q, want %q", got, "myname")
	}

	// Check the multiline description
	expectedDesc := "\n\t\ta description can be multiline\n\n\t\tlike this\n"
	if got := parser.Get("description"); got != expectedDesc {
		t.Errorf("ParamsParser.Get(\"description\") = %q, want %q", got, expectedDesc)
	}
}

func TestParamsParserDefaults(t *testing.T) {
	parser := New()
	parser.SetDefault("key1", "default1")
	parser.SetDefaults(map[string]string{
		"key2": "default2",
		"key3": "default3",
	})

	// Override one default
	parser.Set("key2", "override")

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"default value", "key1", "default1"},
		{"overridden value", "key2", "override"},
		{"another default", "key3", "default3"},
		{"non-existent key", "key4", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.Get(tt.key); got != tt.expected {
				t.Errorf("ParamsParser.Get(%q) = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}

func TestParamsParserTypes(t *testing.T) {
	parser := New()
	parser.Set("int", "123")
	parser.Set("float", "3.14")
	parser.Set("bool_true", "true")
	parser.Set("bool_yes", "yes")
	parser.Set("bool_1", "1")
	parser.Set("bool_false", "false")
	parser.Set("invalid_int", "not_an_int")
	parser.Set("invalid_float", "not_a_float")

	t.Run("GetInt", func(t *testing.T) {
		if val, err := parser.GetInt("int"); err != nil || val != 123 {
			t.Errorf("GetInt(\"int\") = %d, %v, want 123, nil", val, err)
		}
		if val, err := parser.GetInt("invalid_int"); err == nil {
			t.Errorf("GetInt(\"invalid_int\") = %d, %v, want error", val, err)
		}
		if val, err := parser.GetInt("nonexistent"); err == nil {
			t.Errorf("GetInt(\"nonexistent\") = %d, %v, want error", val, err)
		}
	})

	t.Run("GetIntDefault", func(t *testing.T) {
		if val := parser.GetIntDefault("int", 0); val != 123 {
			t.Errorf("GetIntDefault(\"int\", 0) = %d, want 123", val)
		}
		if val := parser.GetIntDefault("invalid_int", 42); val != 42 {
			t.Errorf("GetIntDefault(\"invalid_int\", 42) = %d, want 42", val)
		}
		if val := parser.GetIntDefault("nonexistent", 42); val != 42 {
			t.Errorf("GetIntDefault(\"nonexistent\", 42) = %d, want 42", val)
		}
	})

	t.Run("GetFloat", func(t *testing.T) {
		if val, err := parser.GetFloat("float"); err != nil || val != 3.14 {
			t.Errorf("GetFloat(\"float\") = %f, %v, want 3.14, nil", val, err)
		}
		if val, err := parser.GetFloat("invalid_float"); err == nil {
			t.Errorf("GetFloat(\"invalid_float\") = %f, %v, want error", val, err)
		}
		if val, err := parser.GetFloat("nonexistent"); err == nil {
			t.Errorf("GetFloat(\"nonexistent\") = %f, %v, want error", val, err)
		}
	})

	t.Run("GetFloatDefault", func(t *testing.T) {
		if val := parser.GetFloatDefault("float", 0.0); val != 3.14 {
			t.Errorf("GetFloatDefault(\"float\", 0.0) = %f, want 3.14", val)
		}
		if val := parser.GetFloatDefault("invalid_float", 2.71); val != 2.71 {
			t.Errorf("GetFloatDefault(\"invalid_float\", 2.71) = %f, want 2.71", val)
		}
		if val := parser.GetFloatDefault("nonexistent", 2.71); val != 2.71 {
			t.Errorf("GetFloatDefault(\"nonexistent\", 2.71) = %f, want 2.71", val)
		}
	})

	t.Run("GetBool", func(t *testing.T) {
		if val := parser.GetBool("bool_true"); !val {
			t.Errorf("GetBool(\"bool_true\") = %v, want true", val)
		}
		if val := parser.GetBool("bool_yes"); !val {
			t.Errorf("GetBool(\"bool_yes\") = %v, want true", val)
		}
		if val := parser.GetBool("bool_1"); !val {
			t.Errorf("GetBool(\"bool_1\") = %v, want true", val)
		}
		if val := parser.GetBool("bool_false"); val {
			t.Errorf("GetBool(\"bool_false\") = %v, want false", val)
		}
		if val := parser.GetBool("nonexistent"); val {
			t.Errorf("GetBool(\"nonexistent\") = %v, want false", val)
		}
	})

	t.Run("GetBoolDefault", func(t *testing.T) {
		if val := parser.GetBoolDefault("bool_true", false); !val {
			t.Errorf("GetBoolDefault(\"bool_true\", false) = %v, want true", val)
		}
		if val := parser.GetBoolDefault("nonexistent", true); !val {
			t.Errorf("GetBoolDefault(\"nonexistent\", true) = %v, want true", val)
		}
	})
}

func TestParamsParserGetAll(t *testing.T) {
	parser := New()
	parser.SetDefault("key1", "default1")
	parser.SetDefault("key2", "default2")
	parser.Set("key2", "override")
	parser.Set("key3", "value3")

	all := parser.GetAll()
	
	expected := map[string]string{
		"key1": "default1",
		"key2": "override",
		"key3": "value3",
	}

	if len(all) != len(expected) {
		t.Errorf("GetAll() returned map with %d entries, want %d", len(all), len(expected))
	}

	for k, v := range expected {
		if all[k] != v {
			t.Errorf("GetAll()[%q] = %q, want %q", k, all[k], v)
		}
	}
}

func TestParamsParserMustGet(t *testing.T) {
	parser := New()
	parser.Set("key", "value")
	parser.Set("int", "123")
	parser.Set("float", "3.14")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustGet on non-existent key did not panic")
		}
	}()

	// These should not panic
	if val := parser.MustGet("key"); val != "value" {
		t.Errorf("MustGet(\"key\") = %q, want \"value\"", val)
	}
	if val := parser.MustGetInt("int"); val != 123 {
		t.Errorf("MustGetInt(\"int\") = %d, want 123", val)
	}
	if val := parser.MustGetFloat("float"); val != 3.14 {
		t.Errorf("MustGetFloat(\"float\") = %f, want 3.14", val)
	}

	// This should panic
	parser.MustGet("nonexistent")
}
