package cmd

import (
	"evo/main/lib"
	"path"
)

func DetectScopeFromCWD(root_path string, cwd string) []string {
	if root_path != cwd {
		var maybepkgjson_path = path.Join(cwd, "package.json")
		var maybepkgjson, err = lib.NewPackageJson(maybepkgjson_path)

		if err != nil {
			return []string{}
		}

		return []string{maybepkgjson.Name}
	}

	return []string{}
}
