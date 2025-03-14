package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/webdavserver"
)

func main() {
	// Default filesystem directory
	defaultFs := filepath.Join(os.TempDir(), "herolauncher")

	// Parse command line flags
	port := flag.Int("port", 9999, "WebDAV server port")
	host := flag.String("host", "0.0.0.0", "WebDAV server host")
	basePath := flag.String("base-path", "/", "Base path for WebDAV server")
	fileSystem := flag.String("fs", defaultFs, "File system path for WebDAV server (also used for certificates)")
	debugMode := flag.Bool("debug", true, "Enable debug mode with verbose logging")

	// Authentication options
	useAuth := flag.Bool("auth", false, "Enable basic authentication")
	username := flag.String("username", "admin", "Username for basic authentication")
	password := flag.String("password", "1234", "Password for basic authentication")

	// HTTPS options
	useHTTPS := flag.Bool("https", false, "Enable HTTPS")
	certFile := flag.String("cert", "", "Path to TLS certificate file")
	keyFile := flag.String("key", "", "Path to TLS key file")
	autoGenCerts := flag.Bool("auto-gen-certs", true, "Auto-generate certificates if they don't exist")
	certValidityDays := flag.Int("cert-validity", 365, "Validity period in days for auto-generated certificates")
	certOrg := flag.String("cert-org", "HeroLauncher WebDAV Server", "Organization name for auto-generated certificates")

	flag.Parse()

	// Create WebDAV server configuration
	config := webdavserver.DefaultConfig()
	config.Host = *host
	config.Port = *port
	config.BasePath = *basePath
	config.FileSystem = *fileSystem
	config.DebugMode = *debugMode

	// Configure authentication if enabled
	config.UseAuth = *useAuth
	config.Username = *username
	config.Password = *password

	// Configure HTTPS if enabled
	config.UseHTTPS = *useHTTPS
	config.CertFile = *certFile
	config.KeyFile = *keyFile
	config.AutoGenerateCerts = *autoGenCerts
	config.CertValidityDays = *certValidityDays
	config.CertOrganization = *certOrg

	// Log configuration details
	if *debugMode {
		log.Printf("Debug mode enabled")
	}

	if *useAuth {
		log.Printf("Authentication enabled for user: %s", *username)
	}

	if *useHTTPS {
		if *autoGenCerts {
			log.Printf("HTTPS enabled with auto-certificate generation (if needed)")
			if *certFile != "" && *keyFile != "" {
				log.Printf("Using certificate files if they exist: cert=%s, key=%s", *certFile, *keyFile)
			} else {
				certsDir := filepath.Join(filepath.Dir(*fileSystem), "certificates")
				log.Printf("Certificate files not specified, will auto-generate in %s directory", certsDir)
			}
		} else {
			log.Printf("HTTPS enabled with cert: %s, key: %s", *certFile, *keyFile)
		}
	}

	// Create WebDAV server
	server, err := webdavserver.NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create WebDAV server: %v", err)
	}

	// Start the server in a goroutine
	go func() {
		protocol := "HTTP"
		if config.UseHTTPS {
			protocol = "HTTPS"
		}

		log.Printf("Starting WebDAV server (%s) on %s:%d with file system %s",
			protocol, config.Host, config.Port, config.FileSystem)

		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start WebDAV server: %v", err)
		}
	}()

	// Wait for termination signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	log.Println("Shutting down WebDAV server...")
	if err := server.Stop(); err != nil {
		log.Printf("Error stopping WebDAV server: %v", err)
	}
}
