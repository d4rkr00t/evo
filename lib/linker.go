package lib

import (
	"fmt"
	"os"
	"path"
	"strings"
)

func LinkWorkspaces(workspaces *map[string]string, project *Project) {
	for ws_name := range *workspaces {
		var ws = project.GetWs(ws_name)
		var node_modules = path.Join(ws.Path, "node_modules")

		os.RemoveAll(node_modules)

		for dep := range ws.Deps {
			if dep_ws, ok := project.Workspaces[dep]; ok {
				link_local_ws(&ws, &dep_ws)
			} else {
				link_external(project.Cwd, &ws, dep)
			}
		}
	}
}

func link_local_ws(ws *Workspace, dep_ws *Workspace) {
	var target = path.Join(ws.Path, "node_modules", foldername_from_packagename(dep_ws.Name))
	var dir_name = path.Dir(target)
	os.MkdirAll(dir_name, 0700)
	var err = os.Symlink(dep_ws.Path, target)
	if err != nil {
		fmt.Println(err)
	}
}

func link_external(cwd string, ws *Workspace, name string) {
	var target = path.Join(ws.Path, "node_modules", foldername_from_packagename(name))
	var source = path.Join(cwd, "node_modules", foldername_from_packagename(name))
	var dir_name = path.Dir(target)
	os.MkdirAll(dir_name, 0700)
	var err = os.Symlink(source, target)
	if err != nil {
		fmt.Println(err)
	}
}

func foldername_from_packagename(name string) string {
	if name[0] == '@' {
		var name = strings.Split(name, "/")
		return path.Join(name...)
	}
	return name
}
