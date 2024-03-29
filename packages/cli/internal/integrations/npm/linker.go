package npm

import (
	"evo/internal/project"
	"evo/internal/workspace"
	"fmt"
	"os"
	"path"
	"strings"
)

func LinkNpmDependencies(proj *project.Project, wsName string) {
	var ws, _ = proj.Load(wsName)
	var nodeModules = GetNodeModulesPath(ws.Path)
	os.RemoveAll(nodeModules)

	for _, dep := range ws.Deps {
		if dep.Provider != "npm" {
			continue
		}

		if dep.Type == "local" {
			var depWs, _ = proj.Load(dep.Name)
			linkLocal(proj.Path, ws, depWs)
		} else {
			linkExternal(proj.Path, ws, &dep)
		}
	}
}

func linkExternal(rootPath string, ws *workspace.Workspace, dep *workspace.WorkspaceDependency) {
	var tgt = path.Join(ws.Path, "node_modules", folderNameFromPackageName(dep.Name))
	var src = path.Join(rootPath, "node_modules", folderNameFromPackageName(dep.Name))
	var dirName = path.Dir(tgt)
	os.MkdirAll(dirName, 0700)
	var err = os.Symlink(src, tgt)
	if err != nil {
		panic(err)
	}

	linkBin(ws.Path, src)
}

func linkLocal(rootPath string, ws *workspace.Workspace, depWs *workspace.Workspace) {
	var tgt = path.Join(ws.Path, "node_modules", folderNameFromPackageName(depWs.Name))
	var dirName = path.Dir(tgt)
	os.MkdirAll(dirName, 0700)
	var err = os.Symlink(depWs.Path, tgt)
	if err != nil {
		panic(err)
	}

	linkBin(ws.Path, depWs.Path)
}

func linkBin(wsPath string, depWsPath string) {
	var depWsPkgJson, depWsPkgJsonErr = NewPackageJson(path.Join(depWsPath, "package.json"))
	if depWsPkgJsonErr != nil {
		return
	}

	var binDir = GetNodeModulesBinPath(wsPath)
	os.MkdirAll(binDir, 0700)
	for binName, binTgt := range depWsPkgJson.GetBin() {
		var binLinkSrc = path.Join(wsPath, "node_modules", folderNameFromPackageName(depWsPkgJson.Name), binTgt)
		if strings.HasPrefix(binName, "@") {
			binName = strings.Split(binName, "/")[1]
		}
		var binLinkTgt = path.Join(binDir, binName)
		var data = fmt.Sprintf("#!/bin/bash \nnode \"%s\" \"$@\"", binLinkSrc)
		var err = os.WriteFile(binLinkTgt, []byte(data), 0744)
		if err != nil {
			panic(err)
		}
	}
}

func folderNameFromPackageName(name string) string {
	if name[0] == '@' {
		var name = strings.Split(name, "/")
		return path.Join(name...)
	}
	return name
}
