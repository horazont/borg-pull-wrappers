package config

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

type ClientRestrictions struct {
	AllowedPaths []string
}

func RestrictionsFromConfig(c *ClientConfig) (*ClientRestrictions, error) {
	return &ClientRestrictions{
		AllowedPaths: c.Paths,
	}, nil
}

type BorgCommand interface {
	ToArgv(clientName string) (argv []string)
	CheckRestrictions(r *ClientRestrictions) (err error)
}

type BorgCreateCommand struct {
	RepositoryPath string
	ArchiveName    string
	SourcePath     string
	Compression    string
	SameFileSystem bool
	ShowProgress   bool
}

func ParseBorgCommand(argv []string) (cmd BorgCommand, err error) {
	flagSet := flag.NewFlagSet("SSH_ORIGINAL_COMMAND", flag.ContinueOnError)

	compression := flagSet.String("C", "", "borg -C")
	progress := flagSet.Bool("p", false, "borg -p")
	same_file_system := flagSet.Bool("x", false, "borg -x")

	err = flagSet.Parse(argv)
	if err != nil {
		return nil, err
	}

	if flagSet.NArg() < 4 || flagSet.Arg(0) != "create" {
		return nil, errors.New("SSH usage: [FLAGS...] create REPO_PATH ARCHIVE_NAME SOURCE_PATH")
	}

	arg_repo_path := flagSet.Arg(1)
	arg_archive_name := flagSet.Arg(2)
	arg_local_path := flagSet.Arg(3)

	return &BorgCreateCommand{
		RepositoryPath: arg_repo_path,
		ArchiveName:    arg_archive_name,
		SourcePath:     arg_local_path,
		Compression:    *compression,
		SameFileSystem: *same_file_system,
		ShowProgress:   *progress,
	}, nil
}

func (c *BorgCreateCommand) ToArgv(clientName string) (argv []string) {
	argv = make([]string, 1)
	argv[0] = "create"

	if c.Compression != "" {
		argv = append(argv, "-C", c.Compression)
	}

	if c.ShowProgress {
		argv = append(argv, "-p")
	}

	if c.SameFileSystem {
		argv = append(argv, "-x")
	}

	stripped_path := strings.TrimLeft(c.RepositoryPath, "/")
	url := fmt.Sprintf("ssh://%s/%s::%s", clientName, stripped_path, c.ArchiveName)

	argv = append(argv, url, c.SourcePath)

	return argv
}

func (c *BorgCreateCommand) CheckRestrictions(r *ClientRestrictions) (err error) {
	path_allowed := false
	for _, allowed_path := range r.AllowedPaths {
		if c.SourcePath == allowed_path {
			path_allowed = true
			break
		}
	}

	if !path_allowed {
		return fmt.Errorf("given source path is restricted by server config")
	}

	return nil
}
