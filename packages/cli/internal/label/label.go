package label

import (
	"evo/internal/goutils/maputils"
	"fmt"
	"strings"
)

type Label struct {
	Target string
	Scope  string
}

const Sep = "::"

func (l *Label) String() string {
	if l.Scope == "*" {
		return fmt.Sprintf("%s%s", Sep, l.Target)
	}

	return fmt.Sprintf("%s%s%s", l.Scope, Sep, l.Target)
}

func StringifyLabels(labels *[]Label) string {
	var res = []string{}

	for _, lb := range *labels {
		res = append(res, lb.String())
	}

	return strings.Join(res, ", ")
}

func GetScopeFromLabels(labels *[]Label) []string {
	var scopesMap = map[string]bool{}

	for _, lb := range *labels {
		if lb.Scope == "*" {
			continue
		}

		scopesMap[lb.Scope] = true
	}

	return maputils.GetKeys(scopesMap)
}
