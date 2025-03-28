package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// SSHClient represents a simple SSH client for executing commands on remote servers
type SSHClient struct {
	Host     string
	User     string
	LogFile  *os.File
	LogToStd bool // Whether to also log to stdout/stderr
}

// NewSSHClient creates a new SSH client
func NewSSHClient(host string, user string) *SSHClient {
	return &SSHClient{
		Host:     host,
		User:     user,
		LogToStd: true,
	}
}

// WithLogFile sets a log file for the SSH client
func (c *SSHClient) WithLogFile(logFile *os.File) *SSHClient {
	c.LogFile = logFile
	return c
}

// WithLogToStd sets whether to log to stdout/stderr
func (c *SSHClient) WithLogToStd(logToStd bool) *SSHClient {
	c.LogToStd = logToStd
	return c
}

// Execute executes a command on the remote server
func (c *SSHClient) Execute(command string) (string, error) {
	sshCommand := fmt.Sprintf("ssh %s@%s \"%s\"", c.User, c.Host, command)

	cmd := exec.Command("bash", "-c", sshCommand)

	// Configure stdout/stderr handling
	if c.LogToStd {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// If we have a log file, we need to capture the output to write to it
	if c.LogFile != nil {
		// Execute and capture output
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Write to log file
		_, logErr := c.LogFile.WriteString(outputStr)
		if logErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to write to log file: %v\n", logErr)
		}

		return outputStr, err
	} else if !c.LogToStd {
		// If we're not logging to stdout/stderr or a file, just capture the output
		output, err := cmd.CombinedOutput()
		return string(output), err
	} else {
		// Just run the command with stdout/stderr connected
		err := cmd.Run()
		return "", err
	}
}

// InstallPackages installs packages on the remote server
func (c *SSHClient) InstallPackages(packages ...string) error {
	if len(packages) == 0 {
		return nil
	}

	// Join packages with spaces
	packageList := strings.Join(packages, " ")

	// Create the apt-get command
	command := fmt.Sprintf("apt-get update && apt-get install -y %s", packageList)

	// Execute the command
	output, err := c.Execute(command)

	if err != nil {
		return fmt.Errorf("failed to install packages %s: %w\nOutput: %s", packageList, err, output)
	}

	return nil
}

// CreateDirectory creates a directory on the remote server
func (c *SSHClient) CreateDirectory(path string) error {
	command := fmt.Sprintf("mkdir -p %s", path)

	output, err := c.Execute(command)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w\nOutput: %s", path, err, output)
	}

	return nil
}

// SetPermissions sets permissions on a file or directory on the remote server
func (c *SSHClient) SetPermissions(path string, permissions string) error {
	command := fmt.Sprintf("chmod %s %s", permissions, path)

	output, err := c.Execute(command)
	if err != nil {
		return fmt.Errorf("failed to set permissions %s on %s: %w\nOutput: %s", permissions, path, err, output)
	}

	return nil
}
