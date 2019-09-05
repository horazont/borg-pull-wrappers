package config

import (
	"errors"
	"os"
	"strings"
)

type Environ struct {
	SSHOriginalCommand []string
}

var Err_InvalidOriginalCommand = errors.New("command malformed: shell metacharacters ($\"' etc.) and spaces are not supported")

func parseSSHOriginalCommand(s string) (result []string, err error) {
	// we currently do not support spaces or other shenangians
	if strings.Count(s, "\\") > 0 {
		return nil, Err_InvalidOriginalCommand
	}
	if strings.Count(s, "\"") > 0 {
		return nil, Err_InvalidOriginalCommand
	}
	if strings.Count(s, "'") > 0 {
		return nil, Err_InvalidOriginalCommand
	}
	if strings.Count(s, "$") > 0 {
		return nil, Err_InvalidOriginalCommand
	}
	return strings.Split(s, " "), nil
}

func LoadEnv() (result Environ, err error) {
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) < 2 {
			continue
		}
		k := parts[0]
		v := parts[1]

		switch k {
		case "SSH_ORIGINAL_COMMAND":
			parsed, err := parseSSHOriginalCommand(v)
			if err != nil {
				return result, err
			}
			result.SSHOriginalCommand = parsed
		default:
		}
	}

	return result, nil
}
