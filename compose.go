package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/thamaji/devcontainer-docker/devcontainer"
	"github.com/thamaji/devcontainer-docker/parser"
	"gopkg.in/yaml.v2"
)

func convertComposeOptions(environment *devcontainer.Environment, options parser.Options) (parser.Options, func(), error) {
	var file string
	var projectDirectory string

	newOptions := make(parser.Options, 0, len(options))

	for _, option := range options {
		switch option.Name {
		case "-f", "--file":
			file = option.Value

		case "--project-directory":
			projectDirectory = option.Value

		default:
			newOptions.Add(option.Name, option.Value)
		}
	}

	if file != "" && projectDirectory == "" {
		projectDirectory = filepath.Dir(file)
	}

	if file == "" {
		file = "docker-compose.yml"
		if _, err := os.Stat(file); err != nil {
			file = "docker-compose.yaml"
		}
	}

	if projectDirectory == "" {
		projectDirectory = "."
	}

	path, err := createFakeComposeYaml(environment, projectDirectory, file)
	if err != nil {
		return nil, nil, err
	}
	if path == "" {
		return options, nil, nil
	}

	newOptions.Add("--file", path)
	newOptions.Add("--project-directory", projectDirectory)
	onExit := func() { os.Remove(path) }
	return newOptions, onExit, nil
}

func createFakeComposeYaml(environment *devcontainer.Environment, projectDirectory string, file string) (dst string, err error) {
	defer func() {
		if recover() != nil {
			dst = ""
		}
	}()

	text, err := ioutil.ReadFile(file)
	if err != nil {
		return "", nil
	}

	var data interface{}

	if err := yaml.Unmarshal(text, &data); err != nil {
		return "", nil
	}

	compose := reflect.ValueOf(data)
	services := compose.MapIndex(reflect.ValueOf("services")).Elem()

	for _, key := range services.MapKeys() {
		service := services.MapIndex(key).Elem()
		volumes := service.MapIndex(reflect.ValueOf("volumes")).Elem()

		for i := 0; i < volumes.Len(); i++ {
			volume := volumes.Index(i).Elem()

			switch volume.Kind() {
			case reflect.String:
				params := strings.Split(volume.String(), ":")
				if len(params) <= 0 {
					// error
					break
				}

				if t := compose.MapIndex(reflect.ValueOf("volumes")); t.IsValid() {
					if t.Elem().Kind() != reflect.Map || t.Elem().MapIndex(reflect.ValueOf(params[0])).IsValid() {
						break
					}
				}

				path := params[0]
				if !filepath.IsAbs(path) {
					path = filepath.Join(projectDirectory, path)
				}

				hostPath, err := environment.GetHostPath(path)
				if err != nil {
					return "", err
				}
				params[0] = hostPath

				volumes.Index(i).Set(reflect.ValueOf(strings.Join(params, ":")))

			case reflect.Map:
				vtype := volume.MapIndex(reflect.ValueOf("type")).Elem()
				if vtype.String() == "bind" {
					path := volume.MapIndex(reflect.ValueOf("source")).Elem().String()
					if !filepath.IsAbs(path) {
						path = filepath.Join(projectDirectory, path)
					}

					hostPath, err := environment.GetHostPath(path)
					if err != nil {
						return "", err
					}
					volume.SetMapIndex(reflect.ValueOf("source"), reflect.ValueOf(hostPath))
				}
			}
		}
	}

	f, err := os.CreateTemp(os.TempDir(), "docker-compose-*.yml")
	if err != nil {
		return "", err
	}
	err = yaml.NewEncoder(f).Encode(data)
	f.Close()
	if err != nil {
		os.Remove(f.Name())
		return "", err
	}

	return f.Name(), nil
}
