// Example usage of the paramsparser package
package main

import (
	"fmt"

	"github.com/freeflowuniverse/herolauncher/pkg/heroscript/paramsparser"
)

func main() {
	// Create a new parser
	parser := paramsparser.New()

	// Set some default values
	parser.SetDefaults(map[string]string{
		"host":     "localhost",
		"port":     "8080",
		"debug":    "false",
		"timeout":  "30",
		"greeting": "Hello, World!",
	})

	// Parse a string in the specified format
	inputStr := `
		name: 'myapp' 
		host: 'example.com'
		port: 25
		secure: 1
		reset: 1 
		description: '
			This is a multiline description
			for my application.

			It can span multiple lines.
		'
	`

	err := parser.Parse(inputStr)
	if err != nil {
		fmt.Printf("Error parsing input: %v\n", err)
		return
	}

	// Access parameters with type conversion
	name := parser.Get("name")
	host := parser.Get("host")
	port := parser.GetIntDefault("port", 8080)
	secure := parser.GetBool("secure")
	reset := parser.GetBool("reset")
	description := parser.Get("description")

	fmt.Println("Configuration:")
	fmt.Printf("  Name:        %s\n", name)
	fmt.Printf("  Host:        %s\n", host)
	fmt.Printf("  Port:        %d\n", port)
	fmt.Printf("  Secure:      %t\n", secure)
	fmt.Printf("  Reset:       %t\n", reset)
	fmt.Printf("  Description: %s\n", description)

	// Get all parameters
	fmt.Println("\nAll parameters:")
	for key, value := range parser.GetAll() {
		if key == "description" {
			// Truncate long values for display
			if len(value) > 30 {
				value = value[:30] + "..."
			}
		}
		fmt.Printf("  %s = %s\n", key, value)
	}

	// Example of using MustGet for required parameters
	if parser.Has("name") {
		fmt.Printf("\nRequired parameter: %s\n", parser.MustGet("name"))
	}

	// Example of a simpler one-line parse
	simpleParser := paramsparser.New()
	simpleParser.ParseString("name: 'simple' version: '1.0' active: 1")
	fmt.Println("\nSimple parser results:")
	fmt.Printf("  Name:    %s\n", simpleParser.Get("name"))
	fmt.Printf("  Version: %s\n", simpleParser.Get("version"))
	fmt.Printf("  Active:  %t\n", simpleParser.GetBool("active"))
}
