package npm

import (
	"evo/internal/errors"
	"evo/internal/fsutils"
	"evo/internal/project"
	"evo/internal/workspace"
	"fmt"
	"path"
	"strings"

	"github.com/fatih/color"
)

func AddNpmDependencies(proj *project.Project, wsName string) error {
	if !fsutils.Exist(path.Join(proj.Path, "package.json")) {
		return nil
	}

	var rootPkgJson, rootPkgJsonError = NewPackageJson(path.Join(proj.Path, "package.json"))
	if rootPkgJsonError != nil {
		return rootPkgJsonError
	}

	var ws, _ = proj.Load(wsName)
	var pkgJson, pkgJsonErr = NewPackageJson(path.Join(ws.Path, "package.json"))

	if pkgJsonErr != nil {
		return pkgJsonErr
	}

	var rootDeps = rootPkgJson.GetAllDependencies()
	var npmDepsErrors = []string{}

	for name, version := range pkgJson.GetAllDependencies() {
		var _, exist = proj.Load(name)
		if exist {
			ws.Deps[name] = workspace.WorkspaceDependency{
				Name:     name,
				Version:  version,
				Provider: "npm",
				Type:     "local",
			}
		} else {
			if rootDepVersion, ok := rootDeps[name]; ok {
				if rootDepVersion != version {
					npmDepsErrors = append(
						npmDepsErrors,
						fmt.Sprintf("Dependency '%s' of a package '%s' doesn't match %s version '%s' != '%s'", color.CyanString(name), color.GreenString(ws.Name), color.YellowString("root package.json"), color.GreenString(version), color.YellowString(rootDepVersion)),
					)
				} else {
					ws.Deps[name] = workspace.WorkspaceDependency{
						Name:     name,
						Version:  version,
						Provider: "npm",
						Type:     "external",
					}
				}
			} else {
				npmDepsErrors = append(
					npmDepsErrors,
					fmt.Sprintf("Unknown dependency '%s' of a package '%s'", color.CyanString(name), color.GreenString(ws.Name)),
				)
			}
		}
	}

	proj.Store(ws)

	if len(npmDepsErrors) > 1 {
		return errors.New(
			errors.ErrorNPMIntegrationInvalidDeps,
			strings.Join(append([]string{"Invalid 'npm' dependencies."}, npmDepsErrors...), "\n"),
		)
	}

	return nil
}
