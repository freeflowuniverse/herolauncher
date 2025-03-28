package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/freeflowuniverse/herolauncher/pkg/system/builders/hetznerinstall"
)

func main() {
	// Define command-line flags
	hostname := flag.String("hostname", "", "Target hostname for the server (required)")
	image := flag.String("image", hetznerinstall.DefaultImage, "OS image to install (e.g., Ubuntu-2404)")

	flag.Parse()

	// Validate required flags
	if *hostname == "" {
		fmt.Fprintln(os.Stderr, "Error: -hostname flag is required.")
		flag.Usage()
		os.Exit(1)
	}
	// Drives are now always auto-detected by the builder

	// Create a new HetznerInstall builder
	builder := hetznerinstall.NewBuilder().
		WithHostname(*hostname).
		WithImage(*image)

	// Example: Add custom partitions (optional, overrides default)
	// builder.WithPartitions(
	// 	hetznerinstall.Partition{MountPoint: "/boot", FileSystem: "ext4", Size: "1G"},
	// 	hetznerinstall.Partition{MountPoint: "swap", FileSystem: "swap", Size: "4G"},
	// 	hetznerinstall.Partition{MountPoint: "/", FileSystem: "ext4", Size: "all"},
	// )

	// Example: Enable Software RAID 1 (optional)
	// builder.WithSoftwareRAID(true, 1)

	// Run the Hetzner installation process
	// The builder will handle drive detection/validation internally if drives were not set
	fmt.Printf("Starting Hetzner installation for hostname %s using image %s...\n",
		*hostname, *image)
	if err := builder.RunInstall(); err != nil {
		fmt.Fprintf(os.Stderr, "Error during Hetzner installation: %v\n", err)
		os.Exit(1) // Ensure we exit with non-zero status on error
	}

	// Note: If RunInstall succeeds, the system typically reboots,
	// so this message might not always be seen.
	fmt.Println("Hetzner installation process initiated successfully!")
}
