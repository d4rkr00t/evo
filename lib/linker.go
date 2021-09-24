package lib

import (
	"fmt"
	"os"
	"path"
	"strings"
)

func LinkWorkspaces(root string, wm *WorkspacesMap) {
	for ws_name := range wm.updated {
		var ws = wm.workspaces[ws_name]
		var node_modules = GetNodeModulesPath(ws.Path)

		os.RemoveAll(node_modules)

		for dep := range ws.Deps {
			if dep_ws, ok := wm.workspaces[dep]; ok {
				link_local_ws(&ws, &dep_ws)
			} else {
				link_external(root, &ws, dep)
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

	var bin_dir = path.Join(ws.Path, "node_modules", ".bin")
	os.MkdirAll(bin_dir, 0700)
	for bin_name, bin_target := range dep_ws.PkgJson.Bin {
		var bin_link_src = path.Join(ws.Path, "node_modules", foldername_from_packagename(dep_ws.Name), bin_target)
		var bin_link_target = path.Join(bin_dir, bin_name)
		var data = fmt.Sprintf("#!/usr/bin/env node\nrequire(\"%s\")", bin_link_src)
		var err = os.WriteFile(bin_link_target, []byte(data), 0744)
		if err != nil {
			fmt.Println(err)
		}
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
