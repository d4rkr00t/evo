package workspace

import (
	"path"
	"scu/main/fileutils"
	"sort"
)

type Workspace struct {
	path  string
	hash  string
	files []string
}

func NewWorkspace(ws_path string) Workspace {
	var files []string
	files = append(files, path.Join(ws_path, "package.json"))

	if IsTypeScriptWS(ws_path) {
		files = append(files, GetTSConfigPath(ws_path))
		files = append(files, GetFilesFromTSConfig(ws_path)...)
	}

	sort.Strings(files)

	return Workspace{
		path:  ws_path,
		hash:  fileutils.GetFileListHash(files),
		files: files,
	}
}
