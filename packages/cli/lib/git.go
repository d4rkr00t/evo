package lib

import (
	"fmt"
	"strings"
)

func GetChangedSince(root string, since string) []string {
	var changed, _ = getChangedFilesSince(root, since)
	var untracked_changed, _ = getUntrackedChangedFiles(root)
	var result = []string{}
	for _, line := range changed {
		if len(line) > 0 {
			result = append(result, line)
		}
	}
	for _, line := range untracked_changed {
		if len(line) > 0 {
			result = append(result, line)
		}
	}
	return result
}

func getChangedFilesSince(root string, since string) ([]string, error) {
	var cmd = NewCmd("Get changed files", root, fmt.Sprintf("git diff --name-only %s", since), func(msg string) {}, func(msg string) {})
	var changed, err = cmd.Run()
	return strings.Split(changed, "\n"), err
}

func getUntrackedChangedFiles(root string) ([]string, error) {
	var cmd = NewCmd("Get untracked changed files", root, "git ls-files --others --exclude-standard", func(msg string) {}, func(msg string) {})
	var changed, err = cmd.Run()
	return strings.Split(changed, "\n"), err
}
