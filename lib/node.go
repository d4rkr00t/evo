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

func InstallNodeDeps(root string, lg *LoggerGroup) error {
	var cmd = NewCmd("pnpm install", root, "pnpm install", func(msg string) {
		lg.Badge("pnpm").Info(msg)
	}, func(msg string) {
		lg.Badge("pnpm").Error(msg)
	})
	var _, err = cmd.Run()
	return err
}
