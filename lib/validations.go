package lib

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

func ValidateExternalDeps(wm *WorkspacesMap, root_pkg_json PackageJson) error {
	var err = []string{}

	for _, ws := range wm.workspaces {
		for dep_name, dep_ver := range ws.Deps {
			if _, ok := wm.workspaces[dep_name]; ok {
				continue
			}

			if ver, ok := root_pkg_json.Dependencies[dep_name]; ok {
				if ver != dep_ver {
					err = append(
						err,
						fmt.Sprintf("Dependency '%s' of a package '%s' doesn't match %s version '%s' != '%s'", color.CyanString(dep_name), color.GreenString(ws.Name), color.YellowString("root package.json"), color.GreenString(dep_ver), color.YellowString(ver)),
					)
				}
			} else if ver, ok := root_pkg_json.DevDependencies[dep_name]; ok {
				if ver != dep_ver {
					err = append(
						err,
						fmt.Sprintf("Dependency '%s' of a package '%s' doesn't match %s version '%s' != '%s'", color.CyanString(dep_name), color.GreenString(ws.Name), color.YellowString("root package.json"), color.GreenString(dep_ver), color.YellowString(ver)),
					)
				}
			} else {
				err = append(
					err,
					fmt.Sprintf("Unknown dependency '%s' of a package '%s'", color.CyanString(dep_name), color.GreenString(ws.Name)),
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
