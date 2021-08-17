package lib

import (
	"os"
	"strings"

	"github.com/ionrock/procs"
	"github.com/mattn/go-shellwords"
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
	var envs, args, _ = shellwords.ParseWithEnvs(c.cmd)
	var cmd = procs.NewProcess(strings.Join(args, " "))
	var out = []string{}
	cmd.Dir = c.dir

	cmd.Env = procs.ParseEnv(os.Environ())
	cmd.Env["CWD"] = c.dir
	for _, env_string := range envs {
		var split_env_string = strings.Split(env_string, "=")
		if len(split_env_string) == 2 {
			cmd.Env[split_env_string[0]] = split_env_string[1]
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
			c.stdout(line)
		}
		return line
	}

	var err = cmd.Run()
	return strings.Join(out, "\n"), err
}
