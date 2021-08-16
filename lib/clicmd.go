package lib

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Cmd struct {
	name   string
	dir    string
	cmd    string
	params []string
	stdout func(msg string)
}

func NewCmd(name string, dir string, cmd string, params []string, stdout func(msg string)) Cmd {
	return Cmd{
		name, dir, cmd, params, stdout,
	}
}

func (c Cmd) Run() (string, error) {
	var params = []string{}

	for _, param := range c.params {
		params = append(params, os.ExpandEnv(param))
	}
	var cmd = exec.Command(c.cmd, params...)
	cmd.Dir = c.dir

	var stdout, _ = cmd.StdoutPipe()
	var stderr, _ = cmd.StderrPipe()

	cmd.Start()

	var combined = io.MultiReader(stdout, stderr)
	var scanner = bufio.NewScanner(combined)
	var out = []string{}
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var m = scanner.Text()
		if len(m) > 0 {
			out = append(out, m)
			c.stdout(m)
		}
	}
	var err = cmd.Wait()

	return strings.Join(out, "\n"), err
}
