package lib

import (
	"errors"
	"fmt"
	"strings"
)

func ValidateExternalDeps(workspaces *WorkspacesMap, root_pkg_json PackageJson) error {
	var err = []string{}

	for _, ws := range *workspaces {
		for dep_name, dep_ver := range ws.Deps {
			if _, ok := (*workspaces)[dep_name]; ok {
				continue
			}

			if ver, ok := root_pkg_json.Dependencies[dep_name]; ok {
				if ver != dep_ver {
					err = append(
						err,
						fmt.Sprintf("Dependency '%s' of a package '%s' doesn't match root package.json version '%s' != '%s'", dep_name, ws.Name, dep_ver, ver),
					)
				}
			} else if ver, ok := root_pkg_json.DevDependencies[dep_name]; ok {
				if ver != dep_ver {
					err = append(
						err,
						fmt.Sprintf("Dependency '%s' of a package '%s' doesn't match root package.json version '%s' != '%s'", dep_name, ws.Name, dep_ver, ver),
					)
				}
			} else {
				err = append(
					err,
					fmt.Sprintf("Unknown dependency '%s' of a package '%s'", dep_name, ws.Name),
				)
			}
		}
	}

	if len(err) > 0 {
		return errors.New(strings.Join(err, "\n"))
	}

	return nil
}

func ValidateDepsGraph(dg *DepGraph) error {
	var cycles, path = dg.HasCycles()

	if cycles {
		return fmt.Errorf("cycle in the dependecy graph [ %s ]", path)
	}

	return nil
}
