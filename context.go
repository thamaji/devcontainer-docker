package main

import (
	"os"

	"github.com/thamaji/devcontainer-docker/devcontainer"
	"github.com/thamaji/devcontainer-docker/spec"
)

type context struct {
	spec        *spec.Spec
	environment *devcontainer.Environment

	index int
	args  []string

	onExit func()
}

// get current argument
func (context *context) next() (string, bool) {
	if context.index >= len(os.Args) {
		return "", false
	}
	arg := os.Args[context.index]
	context.index++

	context.args = append(context.args, arg)

	return arg, true
}

// replace current argument
func (context *context) replace(arg string) {
	context.args[len(context.args)-1] = arg
}

// remove current argument
func (context *context) remove() {
	context.args = context.args[:len(context.args)-1]
}
