package spec

import (
	"bufio"
	"os/exec"
	"strings"
	"unicode"

	"github.com/thamaji/devcontainer-docker/parser"
)

type Spec struct {
	ComposeSupported bool
	GlobalOptions    *parser.OptionSpec
	RunOptions       *parser.OptionSpec
	ComposeOptions   *parser.OptionSpec
}

func GetSpec(path string) (*Spec, error) {
	spec := Spec{}
	var err error

	spec.GlobalOptions, err = parseOptions(path)
	if err != nil {
		return nil, err
	}

	spec.RunOptions, err = parseOptions(path, "run")
	if err != nil {
		return nil, err
	}

	spec.ComposeSupported, err = hasComposeSubCommand(path)
	if err != nil {
		return nil, err
	}

	if spec.ComposeSupported {
		spec.ComposeOptions, err = parseOptions(path, "compose")
		if err != nil {
			return nil, err
		}
	}

	return &spec, nil
}

func parseOptions(path string, args ...string) (*parser.OptionSpec, error) {
	spec := parser.OptionSpec{
		ShortOptions: map[string]parser.OptionType{},
		LongOptions:  map[string]parser.OptionType{},
	}

	cmd := exec.Command(path, append(args, "--help")...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	skip := true
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		if skip {
			if strings.HasPrefix(line, "Options:") {
				skip = false
			}
			continue
		}

		if strings.TrimSpace(line) == "" || len(line) == 0 {
			continue
		}

		if !unicode.IsSpace(rune(line[0])) {
			break
		}

		var optionNames []*strings.Builder
		var optionArg *strings.Builder = &strings.Builder{}

		mode := 0
		for _, r := range line {
			switch mode {
			case 0:
				switch {
				case r == '-': // option
					mode = 1
					optionNames = append(optionNames, &strings.Builder{})
					optionNames[len(optionNames)-1].WriteRune(r)
				case !unicode.IsSpace(r): // description
					mode = 99
				}
			case 1:
				switch {
				case r == ',': // aliace
					mode = 2
				case r == ' ': // end or arg
					mode = 3
				default:
					optionNames[len(optionNames)-1].WriteRune(r)
				}
			case 2:
				switch {
				case r == '-':
					mode = 1
					optionNames = append(optionNames, &strings.Builder{})
					optionNames[len(optionNames)-1].WriteRune(r)
				case r == ' ':
				default:
					mode = 0
				}
			case 3:
				switch {
				case unicode.IsSpace(r): // description
					mode = 99
				default:
					mode = 4
					optionArg.WriteRune(r)
				}
			case 4:
				switch {
				case unicode.IsSpace(r): // description
					mode = 99
				default: // arg
					optionArg.WriteRune(r)
				}
			}
		}

		for _, optionName := range optionNames {
			name := optionName.String()

			if strings.HasPrefix(name, "--") {
				spec.LongOptions[strings.TrimPrefix(name, "--")] = parser.OptionType{
					IsBool: optionArg.Len() == 0,
				}
			} else {
				spec.ShortOptions[strings.TrimPrefix(name, "-")] = parser.OptionType{
					IsBool: optionArg.Len() == 0,
				}
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return &spec, nil
}

func hasComposeSubCommand(path string) (bool, error) {
	cmd := exec.Command(path, "compose")
	if err := cmd.Start(); err != nil {
		return false, err
	}
	_ = cmd.Wait()
	return 0 == cmd.ProcessState.ExitCode(), nil
}
