package deployment

import (
	"fmt"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/system/builders/lib/logging"
	"github.com/freeflowuniverse/herolauncher/pkg/system/builders/lib/ssh"
	"github.com/freeflowuniverse/herolauncher/pkg/system/builders/lib/transfer"
)

// RemoteDeployment represents a remote deployment
type RemoteDeployment struct {
	Host           string
	User           string
	RemotePath     string
	Logger         *logging.Logger
	SSHClient      *ssh.SSHClient
	TransferClient *transfer.FileTransfer
}

// NewRemoteDeployment creates a new remote deployment
func NewRemoteDeployment(host, user, remotePath string, logger *logging.Logger) *RemoteDeployment {
	sshClient := ssh.NewSSHClient(host, user)
	if logger != nil {
		sshClient.WithLogFile(logger.LogFile)
	}

	transferClient := transfer.NewFileTransfer(host, user)
	if logger != nil {
		transferClient.WithLogFile(logger.LogFile)
	}

	return &RemoteDeployment{
		Host:           host,
		User:           user,
		RemotePath:     remotePath,
		Logger:         logger,
		SSHClient:      sshClient,
		TransferClient: transferClient,
	}
}

// CreateRemoteDirectory creates the remote directory
func (d *RemoteDeployment) CreateRemoteDirectory() error {
	d.Logger.Log(fmt.Sprintf("Creating deployment directory on server: %s", d.RemotePath))
	return d.SSHClient.CreateDirectory(d.RemotePath)
}

// InstallDependencies installs dependencies on the remote server
func (d *RemoteDeployment) InstallDependencies(packages ...string) error {
	if len(packages) == 0 {
		return nil
	}

	packageList := ""
	for i, pkg := range packages {
		if i > 0 {
			packageList += ", "
		}
		packageList += pkg
	}

	d.Logger.Log(fmt.Sprintf("Installing dependencies (%s) on server...", packageList))
	err := d.SSHClient.InstallPackages(packages...)
	if err != nil {
		d.Logger.LogFailure(fmt.Sprintf("Failed to install dependencies on the server: %v", err))
		return err
	}

	d.Logger.LogSuccess("Dependencies installed successfully")
	return nil
}

// TransferFile transfers a file to the remote server
func (d *RemoteDeployment) TransferFile(localPath string, remoteFileName string) error {
	remotePath := filepath.Join(d.RemotePath, remoteFileName)
	d.Logger.Log(fmt.Sprintf("Transferring file %s to server at %s...", localPath, remotePath))

	err := d.TransferClient.TransferFile(localPath, d.RemotePath)
	if err != nil {
		d.Logger.LogFailure(fmt.Sprintf("Failed to transfer file to server: %v", err))
		return err
	}

	d.Logger.LogSuccess("File transferred successfully")
	return nil
}

// SetExecutablePermissions sets executable permissions on a file
func (d *RemoteDeployment) SetExecutablePermissions(fileName string) error {
	filePath := filepath.Join(d.RemotePath, fileName)
	d.Logger.Log(fmt.Sprintf("Setting executable permissions on %s...", filePath))

	err := d.SSHClient.SetPermissions(filePath, "+x")
	if err != nil {
		d.Logger.LogFailure(fmt.Sprintf("Failed to set permissions: %v", err))
		return err
	}

	d.Logger.LogSuccess("Permissions set successfully")
	return nil
}

// ExecuteRemoteCommand executes a command on the remote server
func (d *RemoteDeployment) ExecuteRemoteCommand(command string) (string, error) {
	d.Logger.Log(fmt.Sprintf("Executing command on server: %s", command))

	output, err := d.SSHClient.Execute(command)
	if err != nil {
		d.Logger.LogFailure(fmt.Sprintf("Command execution failed: %v", err))
		return output, err
	}

	d.Logger.LogSuccess("Command executed successfully")
	return output, nil
}

// ExecuteRemoteBuilder executes a builder binary on the remote server
func (d *RemoteDeployment) ExecuteRemoteBuilder(builderName string) (string, error) {
	d.Logger.Log(fmt.Sprintf("Running %s on server...", builderName))

	command := fmt.Sprintf("cd %s && ./%s", d.RemotePath, builderName)
	output, err := d.SSHClient.Execute(command)

	if err != nil {
		d.Logger.LogFailure(fmt.Sprintf("Builder execution failed: %v", err))
		return output, err
	}

	d.Logger.LogSuccess(fmt.Sprintf("%s completed successfully", builderName))
	return output, nil
}

// VerifyInstallation verifies the installation on the remote server
func (d *RemoteDeployment) VerifyInstallation(checkPaths []string) (bool, map[string]bool, error) {
	d.Logger.Log("Verifying installation on server...")

	results := make(map[string]bool)
	allFound := true

	for _, path := range checkPaths {
		command := fmt.Sprintf("ls -la %s 2>/dev/null || echo 'Not found'", path)
		output, err := d.SSHClient.Execute(command)

		if err != nil {
			d.Logger.LogWarning(fmt.Sprintf("Error checking path %s: %v", path, err))
			results[path] = false
			allFound = false
			continue
		}

		found := output != "" && output != "Not found"
		results[path] = found

		if found {
			d.Logger.LogSuccess(fmt.Sprintf("Found %s", path))
		} else {
			d.Logger.LogWarning(fmt.Sprintf("%s not found", path))
			allFound = false
		}
	}

	if allFound {
		d.Logger.LogSuccess("All installation paths verified successfully")
	} else {
		d.Logger.LogWarning("Some installation paths were not found")
	}

	return allFound, results, nil
}

// Deploy performs a complete deployment
func (d *RemoteDeployment) Deploy(localBinaryPath string, binaryName string, dependencies []string) error {

	// Create remote directory
	if err := d.CreateRemoteDirectory(); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Install dependencies
	if err := d.InstallDependencies(dependencies...); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Transfer binary
	if err := d.TransferFile(localBinaryPath, binaryName); err != nil {
		return fmt.Errorf("failed to transfer binary: %w", err)
	}

	// Set executable permissions
	if err := d.SetExecutablePermissions(binaryName); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	// Execute binary
	_, err := d.ExecuteRemoteBuilder(binaryName)
	if err != nil {
		return fmt.Errorf("failed to execute remote builder: %w", err)
	}

	return nil
}
