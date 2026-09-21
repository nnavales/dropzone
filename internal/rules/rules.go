// Package rules defines runtime matching for files.
package rules

import (
	"fmt"

	"github.com/nnavales/dropzone/internal/actions"
)

// Rule is a runtime rule.
type Rule struct {
	Name   string
	Match  Match
	Action actions.Action
}

// NewRule builds a Rule.
func NewRule(name string, match Match, action actions.Action) (Rule, error) {
	if action == nil {
		return Rule{}, fmt.Errorf("action must not be nil")
	}
	return Rule{Name: name, Match: match, Action: action}, nil
}

// Matches reports whether file satisfies the rule's match.
func (r Rule) Matches(file File) bool {
	return r.Match.match(file)
}
