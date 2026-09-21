package daemon

import (
	"fmt"

	"github.com/nnavales/dropzone/internal/actions"
	"github.com/nnavales/dropzone/internal/config"
	"github.com/nnavales/dropzone/internal/rules"
)

type Zone struct {
	name  string
	path  string
	rules []rules.Rule
}

func newZone(zone config.Zone, conflict actions.ConflictPolicy) (Zone, error) {
	path, err := resolvePath(zone.Path)
	if err != nil {
		return Zone{}, err
	}

	built := make([]rules.Rule, 0, len(zone.Rules))
	for i := range zone.Rules {
		cr := zone.Rules[i]

		m, err := rules.NewMatch(cr.Match)
		if err != nil {
			return Zone{}, fmt.Errorf("rules[%d] (%q): match: %w", i, cr.Name, err)
		}

		a, err := actions.New(cr.Action, conflict)
		if err != nil {
			return Zone{}, fmt.Errorf("rules[%d] (%q): action: %w", i, cr.Name, err)
		}

		r, err := rules.NewRule(cr.Name, m, a)
		if err != nil {
			return Zone{}, fmt.Errorf("rules[%d] (%q): %w", i, cr.Name, err)
		}
		built = append(built, r)
	}

	return Zone{
		name:  zone.Name,
		path:  path,
		rules: built,
	}, nil
}
