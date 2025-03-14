package webdavserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/webdav"
)

// Config holds the configuration for the WebDAV server
type Config struct {
	Host                string
	Port                int
	BasePath            string
	FileSystem          string
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	DebugMode           bool
	UseAuth             bool
	Username            string
	Password            string
	UseHTTPS            bool
	CertFile            string
	KeyFile             string
	AutoGenerateCerts   bool
	CertValidityDays    int
	CertOrganization    string
}

// Server represents the WebDAV server
type Server struct {
	config     Config
	httpServer *http.Server
	handler    *webdav.Handler
	debugLog   func(format string, v ...interface{})
}

// responseWrapper wraps http.ResponseWriter to capture the status code
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code and passes it to the wrapped ResponseWriter
func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures a 200 status code if WriteHeader hasn't been called yet
func (rw *responseWrapper) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

// NewServer creates a new WebDAV server
func NewServer(config Config) (*Server, error) {
	log.Printf("Creating new WebDAV server with config: host=%s, port=%d, basePath=%s, fileSystem=%s, debug=%v, auth=%v, https=%v",
		config.Host, config.Port, config.BasePath, config.FileSystem, config.DebugMode, config.UseAuth, config.UseHTTPS)

	// Ensure the file system directory exists
	if err := os.MkdirAll(config.FileSystem, 0755); err != nil {
		log.Printf("ERROR: Failed to create file system directory %s: %v", config.FileSystem, err)
		return nil, fmt.Errorf("failed to create file system directory: %w", err)
	}
	
	// Log the file system path
	log.Printf("Using file system path: %s", config.FileSystem)
	
	// Create debug logger function
	debugLog := func(format string, v ...interface{}) {
		if config.DebugMode {
			log.Printf("[WebDAV DEBUG] "+format, v...)
		}
	}

	// Create WebDAV handler
	handler := &webdav.Handler{
		FileSystem: webdav.Dir(config.FileSystem),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Printf("WebDAV error: %s %s - %v", r.Method, r.URL.Path, err)
			} else {
				log.Printf("WebDAV: %s %s", r.Method, r.URL.Path)
			}
			
			// Additional debug logging
			if config.DebugMode {
				log.Printf("[WebDAV DEBUG] Request Headers: %v", r.Header)
				log.Printf("[WebDAV DEBUG] Request RemoteAddr: %s", r.RemoteAddr)
				log.Printf("[WebDAV DEBUG] Request UserAgent: %s", r.UserAgent())
			}
		},
	}

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return &Server{
		config:     config,
		httpServer: httpServer,
		handler:    handler,
		debugLog:   debugLog,
	}, nil
}

// Start starts the WebDAV server
func (s *Server) Start() error {
	log.Printf("Starting WebDAV server at %s with file system %s", s.httpServer.Addr, s.config.FileSystem)

	// Create a mux to handle the WebDAV requests
	mux := http.NewServeMux()

	// Register the WebDAV handler at the base path
	mux.HandleFunc(s.config.BasePath, func(w http.ResponseWriter, r *http.Request) {
		// Enhanced debug logging
		s.debugLog("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		s.debugLog("Request Protocol: %s", r.Proto)
		s.debugLog("User-Agent: %s", r.UserAgent())
		
		// Log all request headers
		for name, values := range r.Header {
			s.debugLog("Header: %s = %s", name, values)
		}
		
		// Log request depth (important for WebDAV)
		s.debugLog("Depth header: %s", r.Header.Get("Depth"))
		
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, DELETE, OPTIONS, PROPFIND, PROPPATCH, MKCOL, COPY, MOVE")
		w.Header().Set("Access-Control-Allow-Headers", "Depth, Authorization, Content-Type, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle OPTIONS requests for CORS and WebDAV discovery
		if r.Method == "OPTIONS" {
			// Add WebDAV specific headers for OPTIONS responses
			w.Header().Set("DAV", "1, 2")
			w.Header().Set("MS-Author-Via", "DAV")
			w.Header().Set("Allow", "OPTIONS, GET, HEAD, POST, PUT, DELETE, PROPFIND, PROPPATCH, MKCOL, COPY, MOVE")
			
			// Check if this is a macOS WebDAV client
			isMacOSClient := strings.Contains(r.UserAgent(), "WebDAVFS") || 
				strings.Contains(r.UserAgent(), "WebDAVLib") || 
				strings.Contains(r.UserAgent(), "Darwin")
			
			if isMacOSClient {
				s.debugLog("Detected macOS WebDAV client OPTIONS request, adding macOS-specific headers")
				// These headers help macOS Finder with WebDAV compatibility
				w.Header().Set("X-Dav-Server", "HeroLauncher WebDAV Server")
			}
			
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Handle authentication if enabled
		if s.config.UseAuth {
			s.debugLog("Authentication required for request")
			auth := r.Header.Get("Authorization")
			
			// Check if this is a macOS WebDAV client
			isMacOSClient := strings.Contains(r.UserAgent(), "WebDAVFS") || 
				strings.Contains(r.UserAgent(), "WebDAVLib") || 
				strings.Contains(r.UserAgent(), "Darwin")
			
			// Special handling for OPTIONS requests from macOS clients
			if r.Method == "OPTIONS" && isMacOSClient {
				s.debugLog("Detected macOS WebDAV client OPTIONS request, allowing without auth")
				// macOS sends OPTIONS without auth first, we need to let this through
				// but still send the auth challenge
				w.Header().Set("WWW-Authenticate", "Basic realm=\"WebDAV Server\"")
				return
			}
			
			if auth == "" {
				s.debugLog("No Authorization header provided for non-OPTIONS request")
				w.Header().Set("WWW-Authenticate", "Basic realm=\"WebDAV Server\"")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			// Parse the authentication header
			if !strings.HasPrefix(auth, "Basic ") {
				s.debugLog("Invalid Authorization header format: %s", auth)
				http.Error(w, "Invalid authorization header", http.StatusBadRequest)
				return
			}
			
			payload, err := base64.StdEncoding.DecodeString(auth[6:])
			if err != nil {
				s.debugLog("Failed to decode Authorization header: %v, raw header: %s", err, auth)
				http.Error(w, "Invalid authorization header", http.StatusBadRequest)
				return
			}
			
			pair := strings.SplitN(string(payload), ":", 2)
			if len(pair) != 2 {
				s.debugLog("Invalid credential format: could not split into username:password")
				w.Header().Set("WWW-Authenticate", "Basic realm=\"WebDAV Server\"")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			// Log username for debugging (don't log password)
			s.debugLog("Received credentials for user: %s", pair[0])
			
			if pair[0] != s.config.Username || pair[1] != s.config.Password {
				s.debugLog("Invalid credentials provided, expected user: %s", s.config.Username)
				w.Header().Set("WWW-Authenticate", "Basic realm=\"WebDAV Server\"")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			s.debugLog("Authentication successful for user: %s", pair[0])
		}

		// Log request body for WebDAV methods
		if r.Method == "PROPFIND" || r.Method == "PROPPATCH" || r.Method == "REPORT" || r.Method == "PUT" {
			if r.Body != nil {
				bodyBytes, err := io.ReadAll(r.Body)
				if err == nil {
					s.debugLog("Request body: %s", string(bodyBytes))
					// Create a new reader with the same content
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}
		}

		// Add macOS-specific headers for better compatibility
		isMacOSClient := strings.Contains(r.UserAgent(), "WebDAVFS") || 
			strings.Contains(r.UserAgent(), "WebDAVLib") || 
			strings.Contains(r.UserAgent(), "Darwin")
		
		if isMacOSClient {
			s.debugLog("Adding macOS-specific headers for better compatibility")
			// These headers help macOS Finder with WebDAV compatibility
			w.Header().Set("MS-Author-Via", "DAV")
			w.Header().Set("X-Dav-Server", "HeroLauncher WebDAV Server")
			w.Header().Set("DAV", "1, 2")
			
			// Special handling for PROPFIND requests from macOS
			if r.Method == "PROPFIND" {
				s.debugLog("Handling macOS PROPFIND request with special compatibility")
				// Make sure Content-Type is set correctly for PROPFIND responses
				w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			}
		}

		// Create a response wrapper to capture the response
		responseWrapper := &responseWrapper{ResponseWriter: w}

		// Handle WebDAV requests
		s.debugLog("Handling WebDAV request: %s %s", r.Method, r.URL.Path)
		s.handler.ServeHTTP(responseWrapper, r)

		// Log response details
		s.debugLog("Response status: %d", responseWrapper.statusCode)
		s.debugLog("Response content type: %s", w.Header().Get("Content-Type"))
		
		// Log detailed information for debugging connection issues
		if responseWrapper.statusCode >= 400 {
			s.debugLog("ERROR: WebDAV request failed with status %d", responseWrapper.statusCode)
			s.debugLog("Request method: %s, path: %s", r.Method, r.URL.Path)
			s.debugLog("Response headers: %v", w.Header())
		} else {
			s.debugLog("WebDAV request succeeded with status %d", responseWrapper.statusCode)
		}
	})

	// Set the mux as the HTTP server handler
	s.httpServer.Handler = mux

	// Start the server with HTTPS if configured
	var err error
	if s.config.UseHTTPS {
		// Check if certificate files exist or need to be generated
		if (s.config.CertFile == "" || s.config.KeyFile == "") && !s.config.AutoGenerateCerts {
			log.Printf("ERROR: HTTPS enabled but certificate or key file not provided and auto-generation is disabled")
			return fmt.Errorf("HTTPS enabled but certificate or key file not provided and auto-generation is disabled")
		}
		
		// Auto-generate certificates if needed
		if (s.config.CertFile == "" || s.config.KeyFile == "" || 
			!fileExists(s.config.CertFile) || !fileExists(s.config.KeyFile)) && 
			s.config.AutoGenerateCerts {
			
			s.debugLog("Certificate files not found, auto-generating...")
			
			// Get base directory from the file system path
			baseDir := filepath.Dir(s.config.FileSystem)
			
			// Create certificates directory if it doesn't exist
			certsDir := filepath.Join(baseDir, "certificates")
			if err := os.MkdirAll(certsDir, 0755); err != nil {
				log.Printf("ERROR: Failed to create certificates directory: %v", err)
				return fmt.Errorf("failed to create certificates directory: %w", err)
			}
			
			// Set default certificate paths if not provided
			if s.config.CertFile == "" {
				s.config.CertFile = filepath.Join(certsDir, "webdav.crt")
			}
			if s.config.KeyFile == "" {
				s.config.KeyFile = filepath.Join(certsDir, "webdav.key")
			}
			
			// Generate certificates
			if err := generateCertificate(
				s.config.CertFile, 
				s.config.KeyFile, 
				s.config.CertOrganization, 
				s.config.CertValidityDays,
				s.debugLog,
			); err != nil {
				log.Printf("ERROR: Failed to generate certificates: %v", err)
				return fmt.Errorf("failed to generate certificates: %w", err)
			}
			
			log.Printf("Successfully generated self-signed certificates at %s and %s", 
				s.config.CertFile, s.config.KeyFile)
		}
		
		// Verify certificate files exist
		if !fileExists(s.config.CertFile) || !fileExists(s.config.KeyFile) {
			log.Printf("ERROR: Certificate files not found at %s and/or %s", 
				s.config.CertFile, s.config.KeyFile)
			return fmt.Errorf("certificate files not found")
		}
		
		// Configure TLS
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		s.httpServer.TLSConfig = tlsConfig
		
		log.Printf("Starting WebDAV server with HTTPS on %s using certificates: %s, %s", 
			s.httpServer.Addr, s.config.CertFile, s.config.KeyFile)
		err = s.httpServer.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
	} else {
		log.Printf("Starting WebDAV server with HTTP on %s", s.httpServer.Addr)
		err = s.httpServer.ListenAndServe()
	}
	
	if err != nil && err != http.ErrServerClosed {
		log.Printf("ERROR: WebDAV server failed to start: %v", err)
		return err
	}
	return nil
}

// Stop stops the WebDAV server
func (s *Server) Stop() error {
	log.Printf("Stopping WebDAV server at %s", s.httpServer.Addr)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to stop WebDAV server: %v", err)
	}
	return err
}

// DefaultConfig returns the default configuration for the WebDAV server
func DefaultConfig() Config {
	// Use system temp directory as default base path
	defaultBasePath := filepath.Join(os.TempDir(), "herolauncher")
	
	return Config{
		Host:              "0.0.0.0",
		Port:              9999,
		BasePath:          "/",
		FileSystem:        defaultBasePath,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		DebugMode:         false,
		UseAuth:           false,
		Username:          "admin",
		Password:          "1234",
		UseHTTPS:          false,
		CertFile:          "",
		KeyFile:           "",
		AutoGenerateCerts: true,
		CertValidityDays:  365,
		CertOrganization:  "HeroLauncher WebDAV Server",
	}
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && !info.IsDir()
}

// generateCertificate creates a self-signed TLS certificate and key
func generateCertificate(certFile, keyFile, organization string, validityDays int, debugLog func(format string, args ...interface{})) error {
	debugLog("Generating self-signed certificate: certFile=%s, keyFile=%s, organization=%s, validityDays=%d", 
		certFile, keyFile, organization, validityDays)
	
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}
	
	// Prepare certificate template
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(validityDays) * 24 * time.Hour)
	
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}
	
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{organization},
			CommonName:   "localhost",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		DNSNames:              []string{"localhost"},
	}
	
	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}
	
	// Write certificate to file
	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %w", certFile, err)
	}
	defer certOut.Close()
	
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate to file: %w", err)
	}
	
	// Write private key to file
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %w", keyFile, err)
	}
	defer keyOut.Close()
	
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := pem.Encode(keyOut, privateKeyPEM); err != nil {
		return fmt.Errorf("failed to write private key to file: %w", err)
	}
	
	debugLog("Successfully generated self-signed certificate valid for %d days", validityDays)
	return nil
}
