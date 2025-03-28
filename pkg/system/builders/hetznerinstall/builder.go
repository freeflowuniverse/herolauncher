package hetznerinstall

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

// Struct to parse lsblk JSON output
type lsblkOutput struct {
	BlockDevices []lsblkDevice `json:"blockdevices"`
}

type lsblkDevice struct {
	Name string `json:"name"`
	Rota bool   `json:"rota"` // Rotational device (false for SSD/NVMe)
	Type string `json:"type"` // disk, part, lvm, etc.
}

const installImageConfigPath = "/root/.installimage" // Standard path in Rescue System

// DefaultImage is the default OS image to install.
const DefaultImage = "Ubuntu-2404"

// Partition represents a partition definition in the installimage config.
type Partition struct {
	MountPoint string // e.g., "/", "/boot", "swap"
	FileSystem string // e.g., "ext4", "swap"
	Size       string // e.g., "512M", "all", "8G"
}

// HetznerInstallBuilder configures and runs the Hetzner installimage process.
type HetznerInstallBuilder struct {
	// Drives are now auto-detected
	Hostname    string      // Target hostname
	Image       string      // OS Image name, e.g., "Ubuntu-2404"
	Partitions  []Partition // Partition layout
	Swraid      bool        // Enable software RAID
	SwraidLevel int         // RAID level (0, 1, 5, 6, 10)
	ClearPart   bool        // Wipe disks before partitioning
	// Add PostInstallScript path later if needed
	detectedDrives []string // Stores drives detected by detectSSDDevicePaths
}

// NewBuilder creates a new HetznerInstallBuilder with default settings.
func NewBuilder() *HetznerInstallBuilder {
	return &HetznerInstallBuilder{
		Image:       DefaultImage,
		ClearPart:   true, // Default to wiping disks
		Swraid:      false,
		SwraidLevel: 0,
		Partitions: []Partition{ // Default simple layout
			{MountPoint: "/boot", FileSystem: "ext4", Size: "512M"},
			{MountPoint: "/", FileSystem: "ext4", Size: "all"},
		},
	}
}

// WithHostname sets the target hostname.
func (b *HetznerInstallBuilder) WithHostname(hostname string) *HetznerInstallBuilder {
	b.Hostname = hostname
	return b
}

// WithImage sets the OS image to install.
func (b *HetznerInstallBuilder) WithImage(image string) *HetznerInstallBuilder {
	b.Image = image
	return b
}

// WithPartitions sets the partition layout. Replaces the default.
func (b *HetznerInstallBuilder) WithPartitions(partitions ...Partition) *HetznerInstallBuilder {
	if len(partitions) > 0 {
		b.Partitions = partitions
	}
	return b
}

// WithSoftwareRAID enables and configures software RAID.
func (b *HetznerInstallBuilder) WithSoftwareRAID(enable bool, level int) *HetznerInstallBuilder {
	b.Swraid = enable
	if enable {
		b.SwraidLevel = level
	} else {
		b.SwraidLevel = 0 // Ensure level is 0 if RAID is disabled
	}
	return b
}

// WithClearPart enables or disables wiping disks.
func (b *HetznerInstallBuilder) WithClearPart(clear bool) *HetznerInstallBuilder {
	b.ClearPart = clear
	return b
}

// Validate checks if the builder configuration is valid *before* running install.
// Note: Drive validation happens in RunInstall after auto-detection.
func (b *HetznerInstallBuilder) Validate() error {
	if b.Hostname == "" {
		return fmt.Errorf("hostname must be specified using WithHostname()")
	}
	if b.Image == "" {
		return fmt.Errorf("OS image must be specified using WithImage()")
	}
	if len(b.Partitions) == 0 {
		return fmt.Errorf("at least one partition must be specified using WithPartitions()")
	}
	// Add more validation as needed (e.g., valid RAID levels, partition sizes)
	return nil
}

// GenerateConfig generates the content for the installimage config file.
func (b *HetznerInstallBuilder) GenerateConfig() (string, error) {
	if err := b.Validate(); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	// Use detectedDrives for the template
	if len(b.detectedDrives) == 0 {
		// This should ideally be caught earlier in RunInstall, but double-check
		return "", fmt.Errorf("internal error: GenerateConfig called with no detected drives")
	}

	tmplData := struct {
		*HetznerInstallBuilder          // Embed original builder fields
		Drives                 []string // Override Drives field for the template
	}{
		HetznerInstallBuilder: b,
		Drives:                b.detectedDrives,
	}

	tmpl := `{{range $i, $drive := .Drives}}DRIVE{{add $i 1}} {{$drive}}
{{end}}
SWRAID {{if .Swraid}}1{{else}}0{{end}}
SWRAIDLEVEL {{.SwraidLevel}}

HOSTNAME {{.Hostname}}
BOOTLOADER grub
IMAGE {{.Image}}

{{range .Partitions}}PART {{.MountPoint}} {{.FileSystem}} {{.Size}}
{{end}}
# Wipe disks
CLEARPART {{if .ClearPart}}yes{{else}}no{{end}}
`
	// Using text/template requires a function map for simple arithmetic like add
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	t, err := template.New("installimageConfig").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse config template: %w", err)
	}

	var configContent bytes.Buffer
	// Execute template with the overridden Drives data
	if err := t.Execute(&configContent, tmplData); err != nil {
		return "", fmt.Errorf("failed to execute config template: %w", err)
	}

	return configContent.String(), nil
}

// detectSSDDevicePaths finds non-rotational block devices (SSDs, NVMe).
// Assumes lsblk is available and supports JSON output.
func detectSSDDevicePaths() ([]string, error) {
	fmt.Println("Attempting to detect SSD/NVMe devices using lsblk...")
	cmd := exec.Command("lsblk", "-J", "-o", "NAME,ROTA,TYPE")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute lsblk: %w. Output: %s", err, string(output))
	}

	var data lsblkOutput
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("failed to parse lsblk JSON output: %w", err)
	}

	var ssdPaths []string
	for _, device := range data.BlockDevices {
		// We only care about top-level disks, not partitions
		if device.Type == "disk" && !device.Rota {
			fullPath := "/dev/" + device.Name
			fmt.Printf("Detected potential SSD/NVMe device: %s\n", fullPath)
			ssdPaths = append(ssdPaths, fullPath)
		}
	}

	if len(ssdPaths) == 0 {
		fmt.Println("Warning: No SSD/NVMe devices detected via lsblk.")
		// Don't return an error here, let RunInstall decide if it's fatal
	} else {
		fmt.Printf("Detected SSD/NVMe devices: %v\n", ssdPaths)
	}

	return ssdPaths, nil
}

// findAndStopRaidArrays attempts to find and stop all active RAID arrays.
// Uses multiple methods to ensure arrays are properly stopped.
func findAndStopRaidArrays() error {
	fmt.Println("--- Attempting to find and stop active RAID arrays ---")
	var overallErr error

	// Method 1: Use lsblk to find md devices
	fmt.Println("Method 1: Finding md devices using lsblk...")
	cmdLsblk := exec.Command("lsblk", "-J", "-o", "NAME,TYPE")
	output, err := cmdLsblk.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to execute lsblk to find md devices: %v. Trying alternative methods.\n", err)
	} else {
		var data lsblkOutput
		if err := json.Unmarshal(output, &data); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to parse lsblk JSON for md devices: %v. Trying alternative methods.\n", err)
		} else {
			for _, device := range data.BlockDevices {
				// Check for various RAID types lsblk might report
				isRaid := strings.HasPrefix(device.Type, "raid") || device.Type == "md"
				if strings.HasPrefix(device.Name, "md") && isRaid {
					mdPath := "/dev/" + device.Name
					fmt.Printf("Attempting to stop md device: %s\n", mdPath)
					// Try executing via bash -c
					stopCmdStr := fmt.Sprintf("mdadm --stop %s", mdPath)
					cmdStop := exec.Command("bash", "-c", stopCmdStr)
					stopOutput, stopErr := cmdStop.CombinedOutput() // Capture both stdout and stderr
					if stopErr != nil {
						fmt.Fprintf(os.Stderr, "Warning: Failed to stop %s: %v. Output: %s\n", mdPath, stopErr, string(stopOutput))
						if overallErr == nil {
							overallErr = fmt.Errorf("failed to stop some md devices")
						}
					} else {
						fmt.Printf("Stopped %s successfully.\n", mdPath)
					}
				}
			}
		}
	}

	// Method 2: Use /proc/mdstat to find arrays
	fmt.Println("Method 2: Finding md devices using /proc/mdstat...")
	cmdCat := exec.Command("cat", "/proc/mdstat")
	mdstatOutput, mdstatErr := cmdCat.Output()
	if mdstatErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to read /proc/mdstat: %v\n", mdstatErr)
	} else {
		// Parse mdstat output to find active arrays
		// Example line: md0 : active raid1 sda1[0] sdb1[1]
		lines := strings.Split(string(mdstatOutput), "\n")
		for _, line := range lines {
			if strings.Contains(line, "active") {
				parts := strings.Fields(line)
				if len(parts) >= 1 && strings.HasPrefix(parts[0], "md") {
					mdPath := "/dev/" + parts[0]
					fmt.Printf("Found active array in mdstat: %s\n", mdPath)
					stopCmd := exec.Command("mdadm", "--stop", mdPath)
					stopOutput, stopErr := stopCmd.CombinedOutput()
					if stopErr != nil {
						fmt.Fprintf(os.Stderr, "Warning: Failed to stop %s: %v. Output: %s\n", mdPath, stopErr, string(stopOutput))
					} else {
						fmt.Printf("Stopped %s successfully.\n", mdPath)
					}
				}
			}
		}
	}

	// Method 3: Brute force attempt to stop common md devices
	fmt.Println("Method 3: Attempting to stop common md devices...")
	commonMdPaths := []string{"/dev/md0", "/dev/md1", "/dev/md2", "/dev/md3", "/dev/md127"}
	for _, mdPath := range commonMdPaths {
		fmt.Printf("Attempting to stop %s (brute force)...\n", mdPath)
		stopCmd := exec.Command("mdadm", "--stop", mdPath)
		stopOutput, _ := stopCmd.CombinedOutput() // Ignore errors, just try
		fmt.Printf("Output: %s\n", string(stopOutput))
	}

	// Sync to ensure changes are written
	syncCmd := exec.Command("sync")
	syncCmd.Run() // Ignore errors

	fmt.Println("--- Finished attempting to stop RAID arrays ---")
	return overallErr
}

// zeroSuperblocks attempts to zero mdadm superblocks on all given devices.
func zeroSuperblocks(physicalDevices []string) error {
	fmt.Println("--- Zeroing mdadm superblocks on physical devices ---")
	var overallErr error

	for _, devicePath := range physicalDevices {
		fmt.Printf("Executing: mdadm --zero-superblock %s\n", devicePath)
		// Try executing via bash -c
		zeroCmdStr := fmt.Sprintf("mdadm --zero-superblock %s", devicePath)
		cmdZero := exec.Command("bash", "-c", zeroCmdStr)
		zeroOutput, zeroErr := cmdZero.CombinedOutput() // Capture both stdout and stderr
		if zeroErr != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "Warning: Failed to zero superblock on %s: %v. Output: %s\n", devicePath, zeroErr, string(zeroOutput))
			if overallErr == nil {
				overallErr = fmt.Errorf("failed to zero superblock on some devices")
			}
		} else {
			fmt.Printf("Zeroed superblock on %s successfully.\n", devicePath)
		}
	}

	// Sync to ensure changes are written
	syncCmd := exec.Command("sync")
	syncCmd.Run() // Ignore errors

	fmt.Println("--- Finished zeroing superblocks ---")
	return overallErr
}

// overwriteDiskStart uses dd to zero out the beginning of a disk.
// EXTREMELY DANGEROUS. Use only when absolutely necessary to destroy metadata.
func overwriteDiskStart(devicePath string) error {
	fmt.Printf("☢️☢️ EXTREME WARNING: Overwriting start of disk %s with zeros using dd!\n", devicePath)
	// Write 10MB of zeros. Should be enough to kill most metadata (MBR, GPT, RAID superblocks)
	// bs=1M count=10
	ddCmdStr := fmt.Sprintf("dd if=/dev/zero of=%s bs=1M count=10 oflag=direct", devicePath)
	fmt.Printf("Executing: %s\n", ddCmdStr)

	cmdDD := exec.Command("bash", "-c", ddCmdStr)
	ddOutput, ddErr := cmdDD.CombinedOutput()
	if ddErr != nil {
		// Log error but consider it potentially non-fatal if subsequent wipefs works
		fmt.Fprintf(os.Stderr, "Warning: dd command on %s failed: %v. Output: %s\n", devicePath, ddErr, string(ddOutput))
		// Return the error so the caller knows something went wrong
		return fmt.Errorf("dd command failed on %s: %w", devicePath, ddErr)
	}

	fmt.Printf("✅ Successfully overwrote start of %s with zeros.\n", devicePath)
	return nil
}

// wipeDevice erases partition table signatures from a given device path.
// USE WITH EXTREME CAUTION.
func wipeDevice(devicePath string) error {
	fmt.Printf("⚠️ WARNING: Preparing to wipe partition signatures from device %s\n", devicePath)
	fmt.Printf("Executing: wipefs --all --force %s\n", devicePath)

	cmd := exec.Command("wipefs", "--all", "--force", devicePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to wipe device %s: %w", devicePath, err)
	}

	fmt.Printf("✅ Successfully wiped partition signatures from %s\n", devicePath)
	return nil
}

// executeInstallImage attempts to execute the installimage command using multiple methods.
// Returns the first successful execution or the last error.
func executeInstallImage(configPath string) error {
	fmt.Println("--- Attempting to execute installimage using multiple methods ---")

	// Define all the methods we'll try
	methods := []struct {
		name    string
		cmdArgs []string
	}{
		{
			name:    "Method 1: Interactive bash shell",
			cmdArgs: []string{"bash", "-i", "-c", fmt.Sprintf("installimage -a -c %s", configPath)},
		},
		{
			name:    "Method 2: Login bash shell",
			cmdArgs: []string{"bash", "-l", "-c", fmt.Sprintf("installimage -a -c %s", configPath)},
		},
		{
			name:    "Method 3: Source profile first",
			cmdArgs: []string{"bash", "-c", fmt.Sprintf("source /etc/profile && installimage -a -c %s", configPath)},
		},
		{
			name:    "Method 4: Try absolute path /usr/sbin/installimage",
			cmdArgs: []string{"/usr/sbin/installimage", "-a", "-c", configPath},
		},
		{
			name:    "Method 5: Try absolute path /root/bin/installimage",
			cmdArgs: []string{"/root/bin/installimage", "-a", "-c", configPath},
		},
		{
			name:    "Method 6: Try absolute path /bin/installimage",
			cmdArgs: []string{"/bin/installimage", "-a", "-c", configPath},
		},
		{
			name:    "Method 7: Try absolute path /sbin/installimage",
			cmdArgs: []string{"/sbin/installimage", "-a", "-c", configPath},
		},
	}

	var lastErr error
	for _, method := range methods {
		fmt.Printf("Trying %s\n", method.name)
		fmt.Printf("Executing: %s\n", strings.Join(method.cmdArgs, " "))

		cmd := exec.Command(method.cmdArgs[0], method.cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err == nil {
			fmt.Printf("✅ Success with %s\n", method.name)
			return nil
		}

		fmt.Printf("❌ Failed with %s: %v\n", method.name, err)
		lastErr = err

		// Short pause between attempts
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("--- All installimage execution methods failed ---")
	return fmt.Errorf("all installimage execution methods failed, last error: %w", lastErr)
}

// RunInstall detects drives if needed, wipes them, generates config, and executes installimage.
// Assumes it's running within the Hetzner Rescue System.
func (b *HetznerInstallBuilder) RunInstall() error {
	// 1. Auto-Detect Drives
	fmt.Println("Attempting auto-detection of SSD/NVMe drives...")
	detected, err := detectSSDDevicePaths()
	if err != nil {
		// Make detection failure fatal if we rely solely on it
		return fmt.Errorf("failed to auto-detect SSD devices: %w. Cannot proceed without target drives.", err)
	}
	if len(detected) == 0 {
		return fmt.Errorf("auto-detection did not find any suitable SSD/NVMe drives. Cannot proceed.")
	}
	b.detectedDrives = detected // Store detected drives
	fmt.Printf("Using auto-detected drives for installation: %v\n", b.detectedDrives)

	// 2. Validate other parameters (Hostname, Image, Partitions)
	if err := b.Validate(); err != nil {
		return fmt.Errorf("pre-install validation failed: %w", err)
	}

	// 3. Find and stop all RAID arrays (using multiple methods)
	if err := findAndStopRaidArrays(); err != nil {
		// Log the warning but proceed, as zeroing might partially succeed
		fmt.Fprintf(os.Stderr, "Warning during RAID array stopping: %v. Proceeding with disk cleaning...\n", err)
	}

	// 4. Zero superblocks on all detected drives
	if err := zeroSuperblocks(b.detectedDrives); err != nil {
		// Log the warning but proceed to dd/wipefs, as zeroing might partially succeed
		fmt.Fprintf(os.Stderr, "Warning during superblock zeroing: %v. Proceeding with dd/wipefs...\n", err)
	}

	// 5. Overwrite start of disks using dd (Forceful metadata destruction)
	fmt.Println("--- Preparing to Overwrite Disk Starts (dd) ---")
	var ddFailed bool
	for _, drivePath := range b.detectedDrives {
		if err := overwriteDiskStart(drivePath); err != nil {
			// Log the error, mark as failed, but continue to try wipefs
			fmt.Fprintf(os.Stderr, "ERROR during dd on %s: %v. Will still attempt wipefs.\n", drivePath, err)
			ddFailed = true // If dd fails, we rely heavily on wipefs
		}
	}
	fmt.Println("--- Finished Overwriting Disk Starts (dd) ---")
	// Sync filesystem buffers to disk
	fmt.Println("Syncing after dd...")
	syncCmdDD := exec.Command("sync")
	if syncErr := syncCmdDD.Run(); syncErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: sync after dd failed: %v\n", syncErr)
	}

	// 6. Wipe Target Drives (Partition Signatures) using wipefs (as a fallback/cleanup)
	fmt.Println("--- Preparing to Wipe Target Devices (wipefs) ---")
	for _, drivePath := range b.detectedDrives { // Use detectedDrives
		if err := wipeDevice(drivePath); err != nil {
			// If dd also failed, this wipefs failure is critical. Otherwise, maybe okay.
			if ddFailed {
				return fmt.Errorf("CRITICAL: dd failed AND wipefs failed on %s: %w. Aborting installation.", drivePath, err)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: wipefs failed on %s after dd succeeded: %v. Proceeding cautiously.\n", drivePath, err)
				// Allow proceeding if dd succeeded, but log prominently.
			}
		}
	}
	fmt.Println("--- Finished Wiping Target Devices (wipefs) ---")
	// Sync filesystem buffers to disk again
	fmt.Println("Syncing after wipefs...")
	syncCmdWipe := exec.Command("sync")
	if syncErr := syncCmdWipe.Run(); syncErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: sync after wipefs failed: %v\n", syncErr)
	}

	// 7. Generate installimage Config (using detectedDrives)
	fmt.Println("Generating installimage configuration...")
	configContent, err := b.GenerateConfig()
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// 8. Write Config File
	fmt.Printf("Writing configuration to %s...\n", installImageConfigPath)
	fmt.Printf("--- Config Content ---\n%s\n----------------------\n", configContent) // Log the config
	err = os.WriteFile(installImageConfigPath, []byte(configContent), 0600)           // Secure permissions
	if err != nil {
		return fmt.Errorf("failed to write config file %s: %w", installImageConfigPath, err)
	}
	fmt.Printf("Successfully wrote configuration to %s\n", installImageConfigPath)

	// 9. Execute installimage using multiple methods
	err = executeInstallImage(installImageConfigPath)
	if err != nil {
		return fmt.Errorf("installimage execution failed: %w", err)
	}

	// If installimage succeeds, it usually triggers a reboot.
	// This part of the code might not be reached in a typical successful run.
	fmt.Println("installimage command finished. System should reboot shortly if successful.")
	return nil
}
