package playbook

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/paramsparser"
)

// ActionType represents the type of action
type ActionType int

const (
	ActionTypeUnknown ActionType = iota
	ActionTypeDAL
	ActionTypeSAL
	ActionTypeWAL
	ActionTypeMacro
)

// Action represents a single action in a heroscript
type Action struct {
	ID         int
	CID        string
	Name       string
	Actor      string
	Priority   int
	Params     *paramsparser.ParamsParser
	Result     *paramsparser.ParamsParser
	ActionType ActionType
	Comments   string
	Done       bool
}

// PlayBook represents a collection of actions
type PlayBook struct {
	Actions    []*Action
	Priorities map[int][]int // key is priority, value is list of action indices
	OtherText  string        // text outside of actions
	Result     string
	NrActions  int
	Done       []int
}

// NewAction creates a new action and adds it to the playbook
func (p *PlayBook) NewAction(cid, name, actor string, priority int, actionType ActionType) *Action {
	p.NrActions++
	action := &Action{
		ID:         p.NrActions,
		CID:        cid,
		Name:       name,
		Actor:      actor,
		Priority:   priority,
		ActionType: actionType,
		Params:     paramsparser.New(),
		Result:     paramsparser.New(),
	}
	p.Actions = append(p.Actions, action)
	return action
}

// New creates a new PlayBook
func New() *PlayBook {
	return &PlayBook{
		Actions:    make([]*Action, 0),
		Priorities: make(map[int][]int),
		NrActions:  0,
		Done:       make([]int, 0),
	}
}

// NewFromText creates a new PlayBook from heroscript text
func NewFromText(text string) (*PlayBook, error) {
	pb := New()
	err := pb.AddText(text, 10) // Default priority 10
	if err != nil {
		return nil, err
	}
	return pb, nil
}

// String returns the heroscript representation of the action
func (a *Action) String() string {
	out := a.HeroScript()
	if a.Result != nil && len(a.Result.GetAll()) > 0 {
		out += "\n\nResult:\n"
		// Indent the result
		resultParams := a.Result.GetAll()
		for k, v := range resultParams {
			out += "    " + k + ": '" + v + "'\n"
		}
	}
	return out
}

// HeroScript returns the heroscript representation of the action
func (a *Action) HeroScript() string {
	var out strings.Builder

	// Add comments if any
	if a.Comments != "" {
		lines := strings.Split(a.Comments, "\n")
		for _, line := range lines {
			out.WriteString("// " + line + "\n")
		}
	}

	// Add action type prefix
	switch a.ActionType {
	case ActionTypeDAL:
		out.WriteString("!")
	case ActionTypeSAL:
		out.WriteString("!!")
	case ActionTypeMacro:
		out.WriteString("!!!")
	default:
		out.WriteString("!!") // Default to SAL
	}

	// Add actor and name
	if a.Actor != "" {
		out.WriteString(a.Actor + ".")
	}
	out.WriteString(a.Name + " ")

	// Add ID if present
	if a.ID > 0 {
		out.WriteString(fmt.Sprintf("id:%d ", a.ID))
	}

	// Add parameters
	if a.Params != nil && len(a.Params.GetAll()) > 0 {
		params := a.Params.GetAll()
		firstLine := true
		for k, v := range params {
			if firstLine {
				out.WriteString(k + ": '" + v + "'\n")
				firstLine = false
			} else {
				out.WriteString("    " + k + ": '" + v + "'\n")
			}
		}
	}

	return out.String()
}

// HashKey returns a unique hash for the action
func (a *Action) HashKey() string {
	h := sha1.New()
	h.Write([]byte(a.HeroScript()))
	return hex.EncodeToString(h.Sum(nil))
}

// HashKey returns a unique hash for the playbook
func (p *PlayBook) HashKey() string {
	h := sha1.New()
	for _, action := range p.Actions {
		h.Write([]byte(action.HashKey()))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// HeroScript returns the heroscript representation of the playbook
func (p *PlayBook) HeroScript(showDone bool) string {
	var out strings.Builder
	
	actions, _ := p.ActionsSorted(false)
	for _, action := range actions {
		if !showDone && action.Done {
			continue
		}
		out.WriteString(action.HeroScript() + "\n")
	}
	
	if p.OtherText != "" {
		out.WriteString(p.OtherText)
	}
	
	return out.String()
}

// ActionsSorted returns the actions sorted by priority
func (p *PlayBook) ActionsSorted(prioOnly bool) ([]*Action, error) {
	var result []*Action
	
	// If no priorities are set, return all actions
	if len(p.Priorities) == 0 {
		return p.Actions, nil
	}
	
	// Get all priority numbers and sort them
	var priorities []int
	for prio := range p.Priorities {
		priorities = append(priorities, prio)
	}
	sort.Ints(priorities)
	
	// Add actions in priority order
	for _, prio := range priorities {
		if prioOnly && prio > 49 {
			continue
		}
		
		actionIDs := p.Priorities[prio]
		for _, id := range actionIDs {
			action, err := p.GetAction(id, "", "")
			if err != nil {
				return nil, err
			}
			result = append(result, action)
		}
	}
	
	return result, nil
}

// GetAction finds an action by ID, actor, or name
func (p *PlayBook) GetAction(id int, actor, name string) (*Action, error) {
	actions, err := p.FindActions(id, actor, name, ActionTypeUnknown)
	if err != nil {
		return nil, err
	}
	
	if len(actions) == 1 {
		return actions[0], nil
	} else if len(actions) == 0 {
		return nil, fmt.Errorf("couldn't find action with id: %d, actor: %s, name: %s", id, actor, name)
	} else {
		return nil, fmt.Errorf("multiple actions found with id: %d, actor: %s, name: %s", id, actor, name)
	}
}

// FindActions finds actions based on criteria
func (p *PlayBook) FindActions(id int, actor, name string, actionType ActionType) ([]*Action, error) {
	var result []*Action
	
	for _, a := range p.Actions {
		// If ID is specified, return only the action with that ID
		if id != 0 {
			if a.ID == id {
				return []*Action{a}, nil
			}
			continue
		}
		
		// Filter by actor if specified
		if actor != "" && a.Actor != actor {
			continue
		}
		
		// Filter by name if specified
		if name != "" && a.Name != name {
			continue
		}
		
		// Filter by actionType if specified
		if actionType != ActionTypeUnknown && a.ActionType != actionType {
			continue
		}
		
		// If the action passes all filters, add it to the result
		result = append(result, a)
	}
	
	return result, nil
}

// ActionExists checks if an action exists
func (p *PlayBook) ActionExists(id int, actor, name string) bool {
	actions, err := p.FindActions(id, actor, name, ActionTypeUnknown)
	if err != nil || len(actions) == 0 {
		return false
	}
	return true
}

// String returns a string representation of the playbook
func (p *PlayBook) String() string {
	return p.HeroScript(true)
}

// EmptyCheck checks if there are any actions left to execute
func (p *PlayBook) EmptyCheck() error {
	var undoneActions []*Action
	
	for _, a := range p.Actions {
		if !a.Done {
			undoneActions = append(undoneActions, a)
		}
	}
	
	if len(undoneActions) > 0 {
		return fmt.Errorf("there are actions left to execute: %d", len(undoneActions))
	}
	
	return nil
}
