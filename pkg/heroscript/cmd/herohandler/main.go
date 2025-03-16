package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/heroscript/cmd/herohandler/internal"
)

func main() {
	// Create a new example handler
	handler := internal.NewExampleHandler()

	// Check if input is coming from stdin (piped input)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Reading from stdin (pipe or redirect)
		processStdin(handler)
		return
	}

	// Check if there are command-line arguments
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	// Get the command from arguments
	command := strings.Join(os.Args[1:], " ")
	
	// Format as proper HeroScript with !! prefix if not already prefixed
	script := command
	if !strings.HasPrefix(script, "!!") {
		script = "!!" + script
	}

	// Process the script
	result, err := handler.Play(script, handler)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print the result
	fmt.Println(result)
}

func printUsage() {
	fmt.Println("Usage: herohandler <action>")
	fmt.Println("       cat script.hero | herohandler")
	fmt.Println("\nExample commands:")
	fmt.Println("  example.set key:mykey value:myvalue")
	fmt.Println("  example.get key:mykey")
	fmt.Println("  example.list")
	fmt.Println("  example.delete key:mykey")
	fmt.Println("\nNote: The command will be automatically formatted as HeroScript with !! prefix.")
	fmt.Println("      You can also pipe a multi-line HeroScript file to process multiple commands.")
}

// processStdin reads and processes HeroScript from stdin
func processStdin(handler *internal.ExampleHandler) {
	reader := bufio.NewReader(os.Stdin)
	var scriptBuilder strings.Builder

	// Read all lines from stdin
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Printf("Error reading from stdin: %v\n", err)
			return
		}

		// Add the line to our script
		scriptBuilder.WriteString(line)

		// If we've reached EOF, break
		if err == io.EOF {
			break
		}
	}

	// Process the complete script
	script := scriptBuilder.String()
	if script == "" {
		fmt.Println("Error: Empty script")
		return
	}

	// Process the script
	result, err := handler.Play(script, handler)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print the result
	fmt.Println(result)
}
