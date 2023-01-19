package devcontainer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func NewEnvironment(cliPath string) *Environment {
	return &Environment{cliPath: cliPath}
}

type Environment struct {
	cliPath     string
	containerID string
	mounts      []mountInfo
}

type mountInfo struct {
	Type        string
	Source      string
	Destination string
}

var patterns = []*regexp.Regexp{
	regexp.MustCompile("/docker/([\\w+-.]{64}) "),
	regexp.MustCompile("/docker/containers/([\\w+-.]{64})/"),
}

func (environment *Environment) GetContainerID() (string, error) {
	if environment.containerID != "" {
		return environment.containerID, nil
	}

	var cid string

	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
SCAN:
	for scanner.Scan() {
		line := scanner.Text()
		for _, pattern := range patterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) < 2 {
				continue
			}
			cid = matches[1]
			break SCAN
		}
	}
	if cid == "" {
		return "", errors.New("failed to get container id")
	}

	environment.containerID = cid

	return environment.containerID, nil
}

func (environment *Environment) GetHostPath(path string) (string, error) {
	if environment.mounts == nil {
		containerID, err := environment.GetContainerID()
		if err != nil {
			return "", err
		}

		mounts, err := listMountInfos(environment.cliPath, containerID)
		if err != nil {
			return "", err
		}

		environment.mounts = mounts
	}

	for _, mount := range environment.mounts {
		out, err := exec.Command("realpath", "-m", "--relative-base="+mount.Destination, path).Output()
		if err != nil {
			return "", err
		}

		path = strings.TrimSpace(string(out))

		if !strings.HasPrefix(path, "/") {
			return filepath.Join(mount.Source, path), nil
		}
	}

	return "", errors.New("path is not in host filesystem: " + path)
}

func listMountInfos(cliPath string, containerID string) ([]mountInfo, error) {
	cmd := exec.Command(cliPath, "container", "inspect", containerID)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr := bytes.NewBuffer(nil)
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	inspect := []struct {
		Mounts []mountInfo
	}{}
	if err := json.NewDecoder(stdout).Decode(&inspect); err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, errors.New(string(bytes.TrimSpace(stderr.Bytes())))
	}

	return inspect[0].Mounts, nil
}
