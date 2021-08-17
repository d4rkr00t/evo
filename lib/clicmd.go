package lib

import (
	"os"
	"strings"

	"github.com/ionrock/procs"
)

type Cmd struct {
	name   string
	dir    string
	cmd    string
	stdout func(msg string)
}

func NewCmd(name string, dir string, cmd string, stdout func(msg string)) Cmd {
	return Cmd{
		name, dir, cmd, stdout,
	}
}

func (c Cmd) Run() (string, error) {
	var cmd = procs.NewProcess(c.cmd)
	var out = []string{}
	cmd.Dir = c.dir
	cmd.Env = procs.ParseEnv(os.Environ())

	cmd.OutputHandler = func(line string) string {
		if len(line) > 0 {
			c.stdout(line)
			out = append(out, line)
		}
		return line
	}

	var err = cmd.Run()
	return strings.Join(out, "\n"), err
}
