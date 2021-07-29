package typescript

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/mattn/go-zglob"
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

	for _, str := range tscfg.Include {
		include = append(include, filepath.Join(ws_path, str))
	}
	var matches, _ = zglob.Glob(strings.Join(include, "|"))
	return matches
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
