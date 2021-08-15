package lib

import (
	"bufio"
	"io"
	"os/exec"
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

func (c Cmd) Run() error {
	var cmd = exec.Command(c.cmd, c.params...)
	cmd.Dir = c.dir

	var stdout, _ = cmd.StdoutPipe()
	var stderr, _ = cmd.StderrPipe()

	cmd.Start()

	var combined = io.MultiReader(stdout, stderr)
	var scanner = bufio.NewScanner(combined)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var m = scanner.Text()
		if len(m) > 0 {
			c.stdout(m)
		}
	}
	var err = cmd.Wait()

	return err
}
