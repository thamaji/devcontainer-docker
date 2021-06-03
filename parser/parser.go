package parser

import (
	"errors"
	"strings"
)

func NewContext(args []string) *Context {
	return &Context{
		args:  args,
		index: 0,
	}
}

type Context struct {
	args  []string
	index int
}

func (ctx *Context) Args() []string {
	return ctx.args[ctx.index:]
}

func (ctx *Context) Next() (string, bool) {
	if ctx.index >= len(ctx.args) {
		return "", false
	}
	next := ctx.args[ctx.index]
	ctx.index++
	return next, true
}

type Options []*Option

type Option struct {
	Name  string
	Value string
}

func (options *Options) Add(name string, value string) {
	*options = append((*options), &Option{Name: name, Value: value})
}

func (options *Options) Remove(i int) {
	*options = append((*options)[:i], (*options)[i+1:]...)
}

func (options Options) Args() []string {
	args := []string{}
	for _, option := range options {
		args = append(args, option.Name+"="+option.Value)
	}
	return args
}

type OptionSpec struct {
	LongOptions  map[string]OptionType
	ShortOptions map[string]OptionType
}

type OptionType struct {
	IsBool bool
}

func ParseOptions(ctx *Context, spec *OptionSpec) (Options, error) {
	options := Options{}

	for ; ctx.index < len(ctx.args); ctx.index++ {
		if strings.HasPrefix(ctx.args[ctx.index], "--") {
			// long option
			arg := ctx.args[ctx.index][2:]
			tokens := strings.SplitN(arg, "=", 2)
			name := tokens[0]
			if option, ok := spec.LongOptions[name]; ok {
				switch {
				case option.IsBool && len(tokens) != 2:
					options = append(options, &Option{Name: "--" + name, Value: "true"})

				case len(tokens) == 2:
					options = append(options, &Option{Name: "--" + name, Value: tokens[1]})

				default:
					ctx.index++
					options = append(options, &Option{Name: "--" + name, Value: ctx.args[ctx.index]})
				}
				continue
			}

			return nil, errors.New("unknown option: --" + name)
		}

		if strings.HasPrefix(ctx.args[ctx.index], "-") {
			// short option
			arg := ctx.args[ctx.index][1:]
			tokens := strings.SplitN(arg, "=", 2)
			for j := 0; j < len(tokens[0])-1; j++ {
				name := string(tokens[0][j])

				if option, ok := spec.ShortOptions[name]; ok {
					switch {
					case option.IsBool:
						options = append(options, &Option{Name: "-" + name, Value: "true"})

					default:
						// error
						return nil, errors.New("option need some value: -" + name)
					}

					continue
				}

				return nil, errors.New("unknown option: -" + name)
			}

			name := string(tokens[0][len(tokens[0])-1])
			if option, ok := spec.ShortOptions[name]; ok {
				switch {
				case option.IsBool && len(tokens) != 2:
					options = append(options, &Option{Name: "-" + name, Value: "true"})

				case len(tokens) == 2:
					options = append(options, &Option{Name: "-" + name, Value: tokens[1]})

				default:
					ctx.index++
					options = append(options, &Option{Name: "-" + name, Value: ctx.args[ctx.index]})
				}
				continue
			}

			return nil, errors.New("unknown option: -" + name)
		}

		// subcommand or arguments
		break
	}

	return options, nil
}
