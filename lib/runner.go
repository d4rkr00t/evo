package lib

import (
	"fmt"
	"scu/main/lib/cache"
)

type Runner struct {
	cwd     string
	project Project
	cache   cache.Cache
}

func NewRunner(cwd string) Runner {
	var cc = cache.NewCache(cwd)
	var proj = NewProject(cwd)
	return Runner{cwd: cwd, project: proj, cache: cc}
}

func (r Runner) GetCwd() string {
	return r.cwd
}

func (r Runner) Build() {
	fmt.Println("\nBuild: " + r.GetCwd())
}
