package target

import (
	"strings"
)

func GetScopeAndTargetFromString(str string) (string, string, bool) {
	var splits = strings.Split(str, "#")
	if len(splits) < 2 {
		return "", "", false
	}

	return splits[0], splits[1], true
}
