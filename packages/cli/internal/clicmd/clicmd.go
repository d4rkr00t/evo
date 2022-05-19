package clicmd

import (
	"os"
	"path"
	"strings"

	"github.com/ionrock/procs"
	"github.com/mattn/go-shellwords"
)

type Cmd struct {
	name   string
	dir    string
	cmd    string
	stdout func(msg string)
	stderr func(msg string)
}

func NewCmd(name string, dir string, cmd string, stdout func(msg string), stderr func(msg string)) Cmd {
	return Cmd{
		name, dir, cmd, stdout, stderr,
	}
}

func (c *Cmd) Run() (string, error) {
	if len(c.cmd) == 0 {
		return "", nil
	}

	var envs, args, _ = shellwords.ParseWithEnvs(c.cmd)
	var cmd = procs.NewProcess(strings.Join(args, " "))
	var out = []string{}
	var oldPath = os.Getenv("PATH")

	// TODO: somehow make this generic, maybe
	os.Setenv("PATH", path.Join(c.dir, "node_modules", ".bin")+":"+os.ExpandEnv("$PATH"))

	cmd.Dir = c.dir
	cmd.Env = procs.ParseEnv(os.Environ())
	cmd.Env["CWD"] = c.dir
	for _, envString := range envs {
		var splitEnvString = strings.Split(envString, "=")
		if len(splitEnvString) == 2 {
			cmd.Env[splitEnvString[0]] = splitEnvString[1]
		}
	}

	cmd.OutputHandler = func(line string) string {
		if len(line) > 0 {
			c.stdout(line)
			out = append(out, line)
		}
		return line
	}

	cmd.ErrHandler = func(line string) string {
		if len(line) > 0 {
			c.stderr(line)
		}
		return line
	}

	var err = cmd.Run()
	os.Setenv("PATH", oldPath)
	return strings.Join(out, "\n"), err
}
