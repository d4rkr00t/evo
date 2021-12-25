package fileutils

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

func GlobFiles(ws_path string, include_pattens *[]string, exclude_pattens *[]string) []string {
	var include []string
	var exclude []string
	var result []string

	for _, p := range *include_pattens {
		include = append(include, path.Join(ws_path, p))
	}

	for _, p := range *exclude_pattens {
		exclude = append(exclude, path.Join(ws_path, p))
	}

	var include_pattern = "{" + strings.Join(include, ",") + "}"
	var exclude_pattern = "{" + strings.Join(exclude, ",") + "}"
	var _ = filepath.Walk(ws_path, func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", p, err)
			return err
		}

		if val, _ := doublestar.Match(exclude_pattern, p); val {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if val, _ := doublestar.Match(include_pattern, p); val || len(*include_pattens) == 0 {
			result = append(result, p)
		}

		return nil
	})

	return result
}
