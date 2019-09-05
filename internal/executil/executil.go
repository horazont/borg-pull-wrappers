package executil

import (
	"syscall"

	"github.com/syndtr/gocapability/capability"
)

type Command struct {
	arg0 string
	argv []string
	env  []string
	caps []capability.Cap
}

func NewCommand(arg0 string, argv []string, env []string, caps []capability.Cap) (cmd *Command) {
	return &Command{
		arg0: arg0,
		argv: argv,
		env:  env,
		caps: caps,
	}
}

func (cmd *Command) Exec() (err error) {
	caps, err := capability.NewPid(0)
	if err != nil {
		return err
	}

	for _, c := range cmd.caps {
		caps.Set(capability.INHERITABLE, c)
		caps.Set(capability.AMBIENT, c)
	}

	err = caps.Apply(capability.CAPS | capability.AMBIENT)
	if err != nil {
		return err
	}

	return syscall.Exec(
		cmd.arg0,
		cmd.argv,
		cmd.env,
	)
}
