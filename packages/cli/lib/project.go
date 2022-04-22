package lib

import (
	"errors"
	"evo/main/lib/fileutils"
	"path"
)

func FindProject(cwd string) (PackageJson, Config, error) {
	var pkg_json PackageJson
	cfg, err := FindRootConfig(cwd)
	if err == nil {
		pkg_json, err = NewPackageJson(path.Join(path.Dir(cfg.Path), "package.json"))
	}

	if err != nil {
		return pkg_json, cfg, err
	}

	return pkg_json, cfg, nil
}

func FindRootConfig(cwd string) (Config, error) {
	var cfg Config

	for {
		var maybecfg_path = path.Join(cwd, CONFIG_FILE_NAME)

		if fileutils.Exist(maybecfg_path) {
			var maybecfg, _ = NewConfig(maybecfg_path)
			if len(maybecfg.Workspaces) > 0 {
				return maybecfg, nil
			}
		}

		if cwd == path.Dir(cwd) {
			break
		}

		cwd = path.Dir(cwd)
	}

	return cfg, errors.New("couldn't locate main evo config")
}
