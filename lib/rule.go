package lib

import (
	"fmt"
	"strings"
)

type Rule struct {
	Cmd         string
	Deps        []string
	CacheOutput bool
}

func (r Rule) String() string {
	return fmt.Sprintf("%s:%s:%v", r.Cmd, strings.Join(r.Deps, ","), r.CacheOutput)
}
