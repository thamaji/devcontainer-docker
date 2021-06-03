package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/thamaji/devcontainer-docker/devcontainer"
	"github.com/thamaji/devcontainer-docker/parser"
	"github.com/thamaji/devcontainer-docker/spec"
)

const DockerPath = "/usr/bin/docker"

func main() {
	spec, err := spec.GetSpec(DockerPath)
	if err != nil {
		log.Fatalln(err)
	}

	if !spec.ComposeSupported && "on" == os.Getenv("DOCKERCLI_COMPOSE") {
		// for test
		spec.ComposeSupported = true
		spec.ComposeOptions = &parser.OptionSpec{
			LongOptions: map[string]parser.OptionType{
				"ansi":         {IsBool: false},
				"env-file":     {IsBool: false},
				"file":         {IsBool: false},
				"profile":      {IsBool: false},
				"project-name": {IsBool: false},
			},
			ShortOptions: map[string]parser.OptionType{
				"f": {IsBool: false},
				"p": {IsBool: false},
			},
		}
	}

	environment := devcontainer.NewEnvironment(DockerPath)
	command, err := convertArgs(os.Args[1:], spec, environment)
	if err != nil {
		log.Fatalln(err)
	}

	exitCode := command.execute(DockerPath)

	os.Exit(exitCode)
}

type command struct {
	args   []string
	onExit func()
}

func (command *command) execute(cliPath string) int {
	cmd := exec.Command(cliPath, command.args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	exitCode := 255
	if err := cmd.Start(); err == nil {
		_ = cmd.Wait()
		exitCode = cmd.ProcessState.ExitCode()
	}
	if command.onExit != nil {
		command.onExit()
	}
	return exitCode
}

func convertArgs(args []string, spec *spec.Spec, environment *devcontainer.Environment) (*command, error) {
	ctx := parser.NewContext(args)

	globalOptions, err := parser.ParseOptions(ctx, spec.GlobalOptions)
	if err != nil {
		return nil, err
	}

	if next, ok := ctx.Next(); ok {
		switch next {
		case "container":
			if next, ok := ctx.Next(); ok {
				switch {
				case strings.HasPrefix(next, "-"):
					return nil, errors.New("unknown option: " + next)

				case next == "run":
					options, err := parser.ParseOptions(ctx, spec.RunOptions)
					if err != nil {
						return nil, err
					}

					options, err = convertRunOptions(environment, options)
					if err != nil {
						return nil, err
					}

					command := &command{args: globalOptions.Args(), onExit: nil}
					command.args = append(command.args, "container", "run")
					command.args = append(command.args, options.Args()...)
					command.args = append(command.args, ctx.Args()...)
					return command, nil
				}
			}

		case "run":
			options, err := parser.ParseOptions(ctx, spec.RunOptions)
			if err != nil {
				return nil, err
			}

			options, err = convertRunOptions(environment, options)
			if err != nil {
				return nil, err
			}

			command := &command{args: globalOptions.Args(), onExit: nil}
			command.args = append(command.args, "run")
			command.args = append(command.args, options.Args()...)
			command.args = append(command.args, ctx.Args()...)
			return command, nil

		case "compose":
			if !spec.ComposeSupported {
				break
			}

			options, err := parser.ParseOptions(ctx, spec.ComposeOptions)
			if err != nil {
				return nil, err
			}

			options, onExit, err := convertComposeOptions(environment, options)
			if err != nil {
				return nil, err
			}

			command := &command{args: globalOptions.Args(), onExit: onExit}
			command.args = append(command.args, "compose")
			command.args = append(command.args, options.Args()...)
			command.args = append(command.args, ctx.Args()...)
			return command, nil
		}
	}

	command := &command{args: args, onExit: nil}
	return command, nil
}
