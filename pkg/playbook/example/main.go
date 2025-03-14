package main

import (
	"fmt"
	"log"

	"github.com/freeflowuniverse/herolauncher/pkg/playbook"
)

const exampleScript = `
//This is a mail client configuration
!!mailclient.configure
	name: 'mymail'
	host: 'smtp.example.com'
	port: 25
	secure: 1
	reset: 1 
	description: '
		This is a multiline description
		for my mail client configuration.

		It supports multiple paragraphs.
		'

//System update action
!!system.update
	force: 1
	packages: 'git,curl,wget'
`

func main() {
	// Parse heroscript
	pb, err := playbook.NewFromText(exampleScript)
	if err != nil {
		log.Fatalf("Failed to parse heroscript: %v", err)
	}

	// Print the playbook
	fmt.Println("Playbook contains:")
	fmt.Printf("- %d actions\n", len(pb.Actions))
	fmt.Println("- Hash: " + pb.HashKey())
	fmt.Println()

	// Print each action
	for i, action := range pb.Actions {
		fmt.Printf("Action %d: %s.%s\n", i+1, action.Actor, action.Name)
		fmt.Printf("  Comments: %s\n", action.Comments)
		fmt.Printf("  Parameters:\n")
		
		for key, value := range action.Params.GetAll() {
			fmt.Printf("    %s: %s\n", key, value)
		}
		fmt.Println()
	}

	// Generate heroscript
	fmt.Println("Generated HeroScript:")
	fmt.Println("---------------------")
	fmt.Println(pb.HeroScript(true))
	fmt.Println("---------------------")

	// Demonstrate finding actions
	mailActions, err := pb.FindActions(0, "mailclient", "", playbook.ActionTypeUnknown)
	if err != nil {
		log.Fatalf("Error finding actions: %v", err)
	}
	fmt.Printf("\nFound %d mail client actions\n", len(mailActions))

	// Mark an action as done
	if len(pb.Actions) > 0 {
		pb.Actions[0].Done = true
		fmt.Println("\nAfter marking first action as done:")
		fmt.Println(pb.HeroScript(false)) // Don't show done actions
	}
}
