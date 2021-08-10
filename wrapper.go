package wrapper

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

var (
	// ErrBinaryNotFound is returned when docker-compose binary is not found
	ErrBinaryNotFound = errors.New("docker-compose binary not found")
)

// ComposeWrapper provide a type for managing docker compose commands
type ComposeWrapper struct {
	binaryPath string
}

// NewComposeWrapper initializes a new ComposeWrapper service with local docker-compose binary.
func NewComposeWrapper(binaryPath string) (*ComposeWrapper, error) {
	if !IsBinaryPresent(programPath(binaryPath, "docker-compose")) {
		return nil, ErrBinaryNotFound
	}

	return &ComposeWrapper{binaryPath: binaryPath}, nil
}

// Up create and start containers
func (wrapper *ComposeWrapper) Up(filePaths []string, projectDir, host, projectName, envFilePath, configPath string) ([]byte, error) {
	return wrapper.Command(newUpCommand(filePaths), projectDir, host, projectName, envFilePath, configPath)
}

// Down stop and remove containers
func (wrapper *ComposeWrapper) Down(filePaths []string, projectDir string, host, projectName string) ([]byte, error) {
	return wrapper.Command(newDownCommand(filePaths), projectDir, host, projectName, "", "")
}

// Command exectue a docker-compose commanåd
func (wrapper *ComposeWrapper) Command(command composeCommand, workingDir, host, projectName, envFilePath, configPath string) ([]byte, error) {
	program := programPath(wrapper.binaryPath, "docker-compose")

	if projectName != "" {
		command.WithProjectName(projectName)
	}

	if envFilePath != "" {
		command.WithEnvFilePath(envFilePath)
	}

	if host != "" {
		command.WithHost(host)
	}

	var stderr bytes.Buffer
	cmd := exec.Command(program, command.ToArgs()...)
	cmd.Dir = workingDir

	if configPath != "" {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("DOCKER_CONFIG=%s", configPath))
	}

	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, stderr.String())
	}

	return output, nil
}

type composeCommand struct {
	command []string
	args    []string
}

func newCommand(command []string, filePaths []string) composeCommand {
	var args []string
	for _, path := range filePaths {
		args = append(args, "-f")
		args = append(args, strings.TrimSpace(path))
	}
	return composeCommand{
		args:    args,
		command: command,
	}
}

func newUpCommand(filePaths []string) composeCommand {
	return newCommand([]string{"up", "-d"}, filePaths)
}

func newDownCommand(filePaths []string) composeCommand {
	return newCommand([]string{"down", "--remove-orphans"}, filePaths)
}

func (command *composeCommand) WithProjectName(projectName string) {
	command.args = append(command.args, "-p", projectName)
}

func (command *composeCommand) WithEnvFilePath(envFilePath string) {
	command.args = append(command.args, "--env-file", envFilePath)
}

func (command *composeCommand) WithHost(host string) {
	command.args = append(command.args, "-H", host)
}

func (command *composeCommand) ToArgs() []string {
	return append(command.args, command.command...)
}
