package transfer

import (
	"fmt"
	"os"
	"os/exec"
)

// FileTransfer represents a file transfer client
type FileTransfer struct {
	Host    string
	User    string
	LogFile *os.File
}

// NewFileTransfer creates a new file transfer client
func NewFileTransfer(host string, user string) *FileTransfer {
	return &FileTransfer{
		Host: host,
		User: user,
	}
}

// WithLogFile sets a log file for the file transfer client
func (t *FileTransfer) WithLogFile(logFile *os.File) *FileTransfer {
	t.LogFile = logFile
	return t
}

// TransferFile transfers a file to the remote server using rsync
func (t *FileTransfer) TransferFile(localPath, remotePath string) error {
	// Build the rsync command
	rsyncCmd := fmt.Sprintf("rsync -avz --progress %s %s@%s:%s", localPath, t.User, t.Host, remotePath)

	cmd := exec.Command("bash", "-c", rsyncCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// If we have a log file, we need to capture the output to write to it
	if t.LogFile != nil {
		output, err := cmd.CombinedOutput()

		// Write to log file
		_, logErr := t.LogFile.WriteString(string(output))
		if logErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to write to log file: %v\n", logErr)
		}

		if err != nil {
			return fmt.Errorf("failed to transfer file %s to %s: %w", localPath, remotePath, err)
		}

		return nil
	}

	// Otherwise just run the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to transfer file %s to %s: %w", localPath, remotePath, err)
	}

	return nil
}
