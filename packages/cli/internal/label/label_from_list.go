package label

import (
	"evo/internal/errors"
	"fmt"
	"strings"
)

func GetLablesFromList(maybeLabels []string, defaultScope string) ([]Label, error) {
	var lables = []Label{}

	for _, tgt := range maybeLabels {
		var s, t = getScopeAndTargetFromString(tgt)
		if len(s) > 0 && len(t) > 0 {
			lables = append(lables, Label{Target: t, Scope: s})
		} else if len(t) > 0 {
			lables = append(lables, Label{Target: t, Scope: defaultScope})
		} else {
			return lables, errors.New(
				errors.ErrorExtractingScopeAndTargetsFromInput,
				fmt.Sprintf("Ambiguous input, couldn't extract scope and target from `%s`. Try `evo run %s%starget` to run `target` for `%s` or `evo run %s%s` if `%s` is a target", tgt, tgt, Sep, tgt, Sep, tgt, tgt),
			)
		}
	}

	return lables, nil
}

func getScopeAndTargetFromString(str string) (string, string) {
	if strings.HasPrefix(str, Sep) {
		return "", strings.Replace(str, Sep, "", -1)
	}

	var splits = strings.Split(str, Sep)
	if len(splits) < 2 {
		return str, ""
	}

	return splits[0], splits[1]
}
