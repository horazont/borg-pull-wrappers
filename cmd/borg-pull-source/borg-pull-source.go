package main

import (
	"fmt"
	"log"
	"os"

	"github.com/syndtr/gocapability/capability"

	"github.com/horazont/borg-pull-wrappers/internal/config"
	"github.com/horazont/borg-pull-wrappers/internal/executil"
)

func composeBorgCommand(argv []string, clientName string, runtime config.RuntimeConfig) (*executil.Command, error) {
	env := []string{
		fmt.Sprintf("BORG_RSH=%s", runtime.SocketWrapper),
		fmt.Sprintf("BORGSOCKETWRAP_SOCKET_PATH=%s/%s.sock", runtime.SocketDir, clientName),
		"PATH=/usr/bin:/bin",
		fmt.Sprintf("HOME=%s", runtime.Home),
	}

	argv = append([]string{"/usr/bin/borg"}, argv...)

	cmd := executil.NewCommand(
		argv[0],
		argv,
		env,
		[]capability.Cap{capability.CAP_DAC_READ_SEARCH},
	)
	return cmd, nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s CONFIG_FILE CLIENT_NAME\n", os.Args[0])
		os.Exit(1)
	}

	env, err := config.LoadEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load environment: %s\n", err)
		os.Exit(2)
	}
	if env.SSHOriginalCommand == nil {
		fmt.Fprintf(os.Stderr, "SSH_ORIGINAL_COMMAND environment must be set and well-formed\n")
		os.Exit(2)
	}

	arg_config := os.Args[1]
	arg_client_name := os.Args[2]

	cfg, err := config.LoadGlobalConfig(arg_config)
	if err != nil {
		log.Printf("failed to load configuration: %s\n", err)
		os.Exit(3)
	}

	client_cfg, client_exists := cfg.Clients[arg_client_name]
	if !client_exists {
		log.Printf("invalid client: %s\n", arg_client_name)
		os.Exit(2)
	}

	restrictions, err := config.RestrictionsFromConfig(
		&client_cfg,
	)
	if err != nil {
		log.Printf("failed to process client config: %s\n", err)
		os.Exit(3)
	}

	command, err := config.ParseBorgCommand(
		env.SSHOriginalCommand,
	)
	if err != nil {
		log.Printf("failed to process command: %s\n", err)
		os.Exit(2)
	}

	err = command.CheckRestrictions(restrictions)
	if err != nil {
		log.Printf("command forbidden: %s\n", err)
		os.Exit(4)
	}

	command_argv := command.ToArgv(arg_client_name)

	log.Printf("argv = %s\n", command_argv)

	cmd, err := composeBorgCommand(command_argv, arg_client_name, cfg.Runtime)
	if err != nil {
		log.Printf("failed to prepare execution: %s\n", err)
		os.Exit(4)
	}

	err = cmd.Exec()
	if err != nil {
		log.Printf("borg failed: %s\n", err)
		os.Exit(126)
	}
}
