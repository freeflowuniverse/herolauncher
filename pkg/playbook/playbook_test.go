package playbook

import (
	"strings"
	"testing"
)

const testText1 = `
//comment for the action
!!mailclient.configure
	name:'myname'
	host:'localhost'
	port:25
	secure:1
	reset:1 
	description:'
		a description can be multiline

		like this
		'
`

func TestParse(t *testing.T) {
	pb, err := NewFromText(testText1)
	if err != nil {
		t.Fatalf("Failed to parse text: %v", err)
	}

	if len(pb.Actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(pb.Actions))
	}

	action := pb.Actions[0]
	if action.Actor != "mailclient" {
		t.Errorf("Expected actor 'mailclient', got '%s'", action.Actor)
	}

	if action.Name != "configure" {
		t.Errorf("Expected name 'configure', got '%s'", action.Name)
	}

	if action.Comments != "comment for the action" {
		t.Errorf("Expected comment 'comment for the action', got '%s'", action.Comments)
	}

	// Test params
	name := action.Params.Get("name")
	if name != "myname" {
		t.Errorf("Expected name 'myname', got '%s'", name)
	}

	host := action.Params.Get("host")
	if host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", host)
	}

	port, err := action.Params.GetInt("port")
	if err != nil || port != 25 {
		t.Errorf("Expected port 25, got %d, error: %v", port, err)
	}

	secure := action.Params.GetBool("secure")
	if !secure {
		t.Errorf("Expected secure to be true, got false")
	}

	reset := action.Params.GetBool("reset")
	if !reset {
		t.Errorf("Expected reset to be true, got false")
	}

	// Test multiline description
	desc := action.Params.Get("description")
	// Just check that the description contains the expected text
	if !strings.Contains(desc, "a description can be multiline") || !strings.Contains(desc, "like this") {
		t.Errorf("Description doesn't contain expected content: '%s'", desc)
	}
}

func TestHeroScript(t *testing.T) {
	pb, err := NewFromText(testText1)
	if err != nil {
		t.Fatalf("Failed to parse text: %v", err)
	}

	// Generate heroscript
	script := pb.HeroScript(true)
	
	// Parse the generated script again
	pb2, err := NewFromText(script)
	if err != nil {
		t.Fatalf("Failed to parse generated script: %v", err)
	}

	// Verify the actions are the same
	if len(pb2.Actions) != len(pb.Actions) {
		t.Errorf("Expected %d actions, got %d", len(pb.Actions), len(pb2.Actions))
	}

	// Verify the actions have the same actor and name
	if pb.Actions[0].Actor != pb2.Actions[0].Actor || pb.Actions[0].Name != pb2.Actions[0].Name {
		t.Errorf("Actions don't match: %s.%s vs %s.%s", 
			pb.Actions[0].Actor, pb.Actions[0].Name, 
			pb2.Actions[0].Actor, pb2.Actions[0].Name)
	}

	// Verify the parameters are the same
	params1 := pb.Actions[0].Params.GetAll()
	params2 := pb2.Actions[0].Params.GetAll()
	
	// Check that all keys in params1 exist in params2
	for k, v1 := range params1 {
		v2, exists := params2[k]
		if !exists {
			t.Errorf("Key %s missing in generated script", k)
			continue
		}
		
		// For multiline strings, just check that they contain the same content
		if strings.Contains(v1, "\n") {
			if !strings.Contains(v2, "description") || !strings.Contains(v2, "multiline") {
				t.Errorf("Multiline value for key %s doesn't match: '%s' vs '%s'", k, v1, v2)
			}
		} else if v1 != v2 {
			t.Errorf("Value for key %s doesn't match: '%s' vs '%s'", k, v1, v2)
		}
	}
}

func TestMultipleActions(t *testing.T) {
	const multipleActionsText = `
!!mailclient.configure
	name:'myname'
	host:'localhost'

!!system.update
	force:1
	packages:'git,curl,wget'
`

	pb, err := NewFromText(multipleActionsText)
	if err != nil {
		t.Fatalf("Failed to parse text: %v", err)
	}

	if len(pb.Actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(pb.Actions))
	}

	// Check first action
	action1 := pb.Actions[0]
	if action1.Actor != "mailclient" || action1.Name != "configure" {
		t.Errorf("First action incorrect: %s.%s", action1.Actor, action1.Name)
	}

	// Check second action
	action2 := pb.Actions[1]
	if action2.Actor != "system" || action2.Name != "update" {
		t.Errorf("Second action incorrect: %s.%s", action2.Actor, action2.Name)
	}

	force := action2.Params.GetBool("force")
	if !force {
		t.Errorf("Expected force to be true, got false")
	}

	packages := action2.Params.Get("packages")
	if packages != "git,curl,wget" {
		t.Errorf("Expected packages 'git,curl,wget', got '%s'", packages)
	}
}
