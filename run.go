package main

import (
	"strings"

	"github.com/thamaji/devcontainer-docker/devcontainer"
	"github.com/thamaji/devcontainer-docker/parser"
)

func convertRunOptions(environment *devcontainer.Environment, options parser.Options) (parser.Options, error) {
	for _, option := range options {
		switch option.Name {
		case "-v", "--volume":
			params := strings.Split(option.Value, ":")
			if len(params) <= 0 {
				break
			}

			hostPath, err := environment.GetHostPath(params[0])
			if err != nil {
				return nil, err
			}
			params[0] = hostPath

			option.Value = strings.Join(params, ":")
		}
	}

	return options, nil
}
