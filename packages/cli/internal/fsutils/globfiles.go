package fsutils

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

func GlobFiles(dirPath string, includePattens *[]string, excludePattens *[]string) []string {
	var include []string
	var exclude []string
	var result []string

	for _, p := range *includePattens {
		include = append(include, path.Join(dirPath, p))
	}

	for _, p := range *excludePattens {
		exclude = append(exclude, path.Join(dirPath, p))
	}

	var includePattern = "{" + strings.Join(include, ",") + "}"
	var excludePattern = "{" + strings.Join(exclude, ",") + "}"
	var _ = filepath.Walk(dirPath, func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", p, err)
			return err
		}

		if val, _ := doublestar.Match(excludePattern, p); val {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if val, _ := doublestar.Match(includePattern, p); val || len(*includePattens) == 0 {
			result = append(result, p)
		}

		return nil
	})

	return result
}
