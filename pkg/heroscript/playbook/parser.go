package playbook

import (
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/heroscript/paramsparser"
	"github.com/freeflowuniverse/herolauncher/pkg/tools"
)

// State represents the parser state
type State int

const (
	StateStart State = iota
	StateCommentForActionMaybe
	StateAction
	StateOtherText
)

// PlayBookOptions contains options for creating a new PlayBook
type PlayBookOptions struct {
	Text      string
	Path      string
	GitURL    string
	GitPull   bool
	GitBranch string
	GitReset  bool
	Priority  int
}

// AddText adds heroscript text to the playbook
func (p *PlayBook) AddText(text string, priority int) error {
	// Normalize text
	text = strings.ReplaceAll(text, "\t", "    ")

	var state State = StateStart
	var action *Action
	var comments []string
	var paramsData []string

	// Process each line
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		lineStrip := strings.TrimSpace(line)

		if lineStrip == "" {
			continue
		}

		// Handle action state
		if state == StateAction {
			if !strings.HasPrefix(line, "  ") || lineStrip == "" || strings.HasPrefix(lineStrip, "!") {
				state = StateStart
				// End of action, parse params
				if len(paramsData) > 0 {
					params := strings.Join(paramsData, "\n")
					err := action.Params.Parse(params)
					if err != nil {
						return err
					}
					// Remove ID from params if present
					delete(action.Params.GetAll(), "id")
				}
				comments = []string{}
				paramsData = []string{}
				action = nil
			} else {
				paramsData = append(paramsData, line)
			}
		}

		// Handle comment state
		if state == StateCommentForActionMaybe {
			if strings.HasPrefix(lineStrip, "//") {
				comments = append(comments, strings.TrimLeft(lineStrip, "/ "))
			} else {
				if strings.HasPrefix(lineStrip, "!") {
					state = StateStart
				} else {
					state = StateStart
					p.OtherText += strings.Join(comments, "\n")
					if !strings.HasSuffix(p.OtherText, "\n") {
						p.OtherText += "\n"
					}
					comments = []string{}
				}
			}
		}

		// Handle start state
		if state == StateStart {
			if strings.HasPrefix(lineStrip, "!") && !strings.HasPrefix(lineStrip, "![") {
				// Start a new action
				state = StateAction

				// Create new action
				action = &Action{
					ID:       p.NrActions + 1,
					Priority: priority,
					Params:   paramsparser.New(),
					Result:   paramsparser.New(),
				}
				p.NrActions++

				// Set comments
				action.Comments = strings.Join(comments, "\n")
				comments = []string{}
				paramsData = []string{}

				// Parse action name
				actionName := lineStrip
				if strings.Contains(lineStrip, " ") {
					actionName = strings.TrimSpace(strings.Split(lineStrip, " ")[0])
					params := strings.TrimSpace(strings.Join(strings.Split(lineStrip, " ")[1:], " "))
					if params != "" {
						paramsData = append(paramsData, params)
					}
				}

				// Determine action type
				if strings.HasPrefix(actionName, "!!!!!") {
					return ErrInvalidActionPrefix
				} else if strings.HasPrefix(actionName, "!!!!") {
					action.ActionType = ActionTypeWAL
				} else if strings.HasPrefix(actionName, "!!!") {
					action.ActionType = ActionTypeMacro
				} else if strings.HasPrefix(actionName, "!!") {
					action.ActionType = ActionTypeSAL
				} else if strings.HasPrefix(actionName, "!") {
					action.ActionType = ActionTypeDAL
				}

				// Remove prefix
				actionName = strings.TrimLeft(actionName, "!")

				// Split into actor and action name
				parts := strings.Split(actionName, ".")
				if len(parts) == 1 {
					action.Actor = "core"
					action.Name = tools.NameFix(parts[0])
				} else if len(parts) == 2 {
					action.Actor = tools.NameFix(parts[0])
					action.Name = tools.NameFix(parts[1])
				} else {
					return ErrInvalidActionName
				}

				// Add action to playbook
				p.Actions = append(p.Actions, action)

				continue
			} else if strings.HasPrefix(lineStrip, "//") {
				state = StateCommentForActionMaybe
				comments = append(comments, strings.TrimLeft(lineStrip, "/ "))
			}
		}
	}

	// Process the last action if needed
	if state == StateAction && action != nil && action.ID != 0 {
		if len(paramsData) > 0 {
			params := strings.Join(paramsData, "\n")
			err := action.Params.Parse(params)
			if err != nil {
				return err
			}
			// Remove ID from params if present
			delete(action.Params.GetAll(), "id")
		}
	}

	// Process the last comment if needed
	if state == StateCommentForActionMaybe && len(comments) > 0 {
		p.OtherText += strings.Join(comments, "\n")
	}

	return nil
}

// NewFromFile creates a new PlayBook from a file
func NewFromFile(path string, priority int) (*PlayBook, error) {
	// This is a simplified version - in a real implementation, you'd read the file
	// and handle different file types (md, hero, etc.)

	// For now, we'll just create an empty playbook
	pb := New()

	// TODO: Implement file reading and parsing

	return pb, nil
}

// Errors
var (
	ErrInvalidActionPrefix = NewError("invalid action prefix")
	ErrInvalidActionName   = NewError("invalid action name")
)

// NewError creates a new error
func NewError(msg string) error {
	return &PlayBookError{msg}
}

// PlayBookError represents a playbook error
type PlayBookError struct {
	Msg string
}

// Error returns the error message
func (e *PlayBookError) Error() string {
	return e.Msg
}
