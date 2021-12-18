package lib

import (
	"evo/main/lib/fileutils"
	"path"
)

func IsNodeModulesExist(p string) bool {
	return fileutils.Exist(GetNodeModulesPath(p))
}

func GetNodeModulesPath(p string) string {
	return path.Join(p, "node_modules")
}

func GetNodeModulesBinPath(p string) string {
	return path.Join(GetNodeModulesPath(p), ".bin")
}

func InstallNodeDeps(root string, pkg_mgr string, lg *LoggerGroup) error {
	var cmd = NewCmd(pkg_mgr+" install", root, pkg_mgr+" install", func(msg string) {
		lg.Badge(pkg_mgr).Info(msg)
	}, func(msg string) {
		lg.Badge(pkg_mgr).Error(msg)
	})
	var _, err = cmd.Run()
	return err
}

func DetectPackageManager(root string) string {
	if fileutils.Exist(path.Join(root, "yarn.lock")) {
		return "yarn"
	}
	if fileutils.Exist(path.Join(root, "pnpm-lock.yaml")) {
		return "pnpm"
	}
	return "npm"
}
