package lib

import (
	"path"
	"scu/main/lib/fileutils"
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

func InstallNodeDeps(root string, lg *LoggerGroup) {
	var cmd = NewCmd("pnpm install", root, "pnpm", []string{"install"}, func(msg string) {
		lg.LogWithBadge("pnpm", msg)
	})
	cmd.Run()
}
