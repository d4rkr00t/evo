package runner

import (
	"scu/main/project"
)

type Runner struct {
	cwd     string
	project project.Project
}

func NewRunner(cwd string) Runner {
	var proj = project.NewProject(cwd)
	return Runner{cwd: cwd, project: proj}
}

func (r Runner) GetCwd() string {
	return r.cwd
}
