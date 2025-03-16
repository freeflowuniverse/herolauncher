package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/playbook"
)

func main() {
	// Define command line flags
	parseCmd := flag.NewFlagSet("parse", flag.ExitOnError)
	parseFile := parseCmd.String("file", "", "Path to heroscript file to parse")
	parseText := parseCmd.String("text", "", "Heroscript text to parse")
	parsePriority := parseCmd.Int("priority", 10, "Default priority for actions")

	executeCmd := flag.NewFlagSet("execute", flag.ExitOnError)
	executeFile := executeCmd.String("file", "", "Path to heroscript file to execute")
	executeText := executeCmd.String("text", "", "Heroscript text to execute")
	executePriority := executeCmd.Int("priority", 10, "Default priority for actions")
	executeActor := executeCmd.String("actor", "", "Only execute actions for this actor")
	executeAction := executeCmd.String("action", "", "Only execute actions with this name")

	// Check if a subcommand is provided
	if len(os.Args) < 2 {
		fmt.Println("Expected 'parse' or 'execute' subcommand")
		os.Exit(1)
	}

	// Parse the subcommand
	switch os.Args[1] {
	case "parse":
		parseCmd.Parse(os.Args[2:])
		handleParseCommand(*parseFile, *parseText, *parsePriority)
	case "execute":
		executeCmd.Parse(os.Args[2:])
		handleExecuteCommand(*executeFile, *executeText, *executePriority, *executeActor, *executeAction)
	default:
		fmt.Println("Expected 'parse' or 'execute' subcommand")
		os.Exit(1)
	}
}

func handleParseCommand(file, text string, priority int) {
	var pb *playbook.PlayBook
	var err error

	// Parse from file or text
	if file != "" {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		pb, err = playbook.NewFromText(string(content))
	} else if text != "" {
		pb, err = playbook.NewFromText(text)
	} else {
		log.Fatalf("Either -file or -text must be provided")
	}

	if err != nil {
		log.Fatalf("Failed to parse heroscript: %v", err)
	}

	// Print the parsed playbook
	fmt.Printf("Parsed %d actions:\n\n", len(pb.Actions))
	for i, action := range pb.Actions {
		fmt.Printf("Action %d: %s.%s (Priority: %d)\n", i+1, action.Actor, action.Name, action.Priority)
		if action.Comments != "" {
			fmt.Printf("  Comments: %s\n", action.Comments)
		}
		fmt.Printf("  Parameters:\n")
		for key, value := range action.Params.GetAll() {
			// Format multiline values
			if strings.Contains(value, "\n") {
				fmt.Printf("    %s: |\n", key)
				lines := strings.Split(value, "\n")
				for _, line := range lines {
					fmt.Printf("      %s\n", line)
				}
			} else {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}
		fmt.Println()
	}

	// Print the generated heroscript
	fmt.Println("Generated HeroScript:")
	fmt.Println("---------------------")
	fmt.Println(pb.HeroScript(true))
	fmt.Println("---------------------")
}

func handleExecuteCommand(file, text string, priority int, actor, action string) {
	var pb *playbook.PlayBook
	var err error

	// Parse from file or text
	if file != "" {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		pb, err = playbook.NewFromText(string(content))
	} else if text != "" {
		pb, err = playbook.NewFromText(text)
	} else {
		log.Fatalf("Either -file or -text must be provided")
	}

	if err != nil {
		log.Fatalf("Failed to parse heroscript: %v", err)
	}

	// Find actions to execute
	var actionsToExecute []*playbook.Action
	if actor != "" || action != "" {
		// Find specific actions
		actionsToExecute, err = pb.FindActions(0, actor, action, playbook.ActionTypeUnknown)
		if err != nil {
			log.Fatalf("Failed to find actions: %v", err)
		}
	} else {
		// Execute all actions in priority order
		actionsToExecute, err = pb.ActionsSorted(false)
		if err != nil {
			log.Fatalf("Failed to sort actions: %v", err)
		}
	}

	// Execute the actions
	fmt.Printf("Executing %d actions:\n\n", len(actionsToExecute))
	for i, action := range actionsToExecute {
		fmt.Printf("Executing action %d: %s.%s\n", i+1, action.Actor, action.Name)
		
		// In a real implementation, you would have handlers for different actors and actions
		// For this example, we'll just simulate execution
		fmt.Printf("  Parameters:\n")
		for key, value := range action.Params.GetAll() {
			fmt.Printf("    %s: %s\n", key, value)
		}
		
		// Mark the action as done
		action.Done = true
		
		// Set some result data
		action.Result.Set("status", "success")
		action.Result.Set("execution_time", "0.5s")
		
		fmt.Printf("  Result: success\n\n")
	}

	// Check if all actions are done
	err = pb.EmptyCheck()
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
	} else {
		fmt.Println("All actions executed successfully!")
	}
}
