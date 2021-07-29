package typescript

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

type TSConfig struct {
	Include []string
	Exclude []string
}

func IsTypeScriptWS(ws_path string) bool {
	var _, err = ioutil.ReadFile(GetTSConfigPath(ws_path))
	return err == nil
}

func GetFilesFromTSConfig(ws_path string) []string {
	var tscfg = read_tsconfig(ws_path)
	var include []string
	var result []string

	for _, p := range tscfg.Include {
		include = append(include, path.Join(ws_path, p))
	}

	var g = glob.MustCompile("{" + strings.Join(include, ",") + "}")
	var _ = filepath.Walk(ws_path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() && info.Name() == "node_modules" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		if g.Match(path) {
			result = append(result, path)
		}

		return nil
	})

	return result
}

func GetTSConfigPath(ws_path string) string {
	return filepath.Join(ws_path, "tsconfig.json")
}

func read_tsconfig(ws_path string) TSConfig {
	var tscfg TSConfig
	var dat, _ = ioutil.ReadFile(GetTSConfigPath(ws_path))
	json.Unmarshal(dat, &tscfg)
	return tscfg
}
