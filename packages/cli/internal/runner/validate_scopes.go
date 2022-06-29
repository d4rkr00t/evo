package runner

import (
	"evo/internal/errors"
	"evo/internal/project"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

func ValidateScopes(proj *project.Project, scopes *[]string) error {
	var missing = []string{}

	for _, scope := range *scopes {
		if _, ok := proj.Load(scope); !ok {
			missing = append(missing, scope)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return errors.New(errors.ErrorScopesNotFound, fmt.Sprintf("Trying to scope the run to non-existing workspaces: %s", color.YellowString(strings.Join(missing, ", "))))
}
