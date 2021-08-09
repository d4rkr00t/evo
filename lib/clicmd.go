package lib

import (
	"bufio"
	"fmt"
	"os/exec"
)

type Cmd struct {
	name   string
	dir    string
	cmd    string
	params []string
}

func NewCmd(name string, dir string, cmd string, params []string) Cmd {
	return Cmd{
		name, dir, cmd, params,
	}
}

func (c Cmd) Run() error {
	var cmd = exec.Command(c.cmd, c.params...)
	cmd.Dir = c.dir

	var stdout, _ = cmd.StdoutPipe()

	cmd.Start()

	var scanner = bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var m = scanner.Text()
		if len(m) > 0 {
			fmt.Println("["+c.name+"] → ", m)
		}
	}
	var err = cmd.Wait()

	return err
}
