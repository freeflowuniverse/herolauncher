package herohandler

import (
	"fmt"
	"log"

	"github.com/freeflowuniverse/herolauncher/pkg/heroscript/handlerfactory/core"

	// "github.com/freeflowuniverse/herolauncher/pkg/handlerfactory/heroscript/handlerfactory/fakehandler"
	"github.com/freeflowuniverse/herolauncher/pkg/heroscript/handlerfactory/processmanagerhandler"
)

// HeroHandler is the main handler factory that manages all registered handlers
type HeroHandler struct {
	factory      *core.HandlerFactory
	telnetServer *core.TelnetServer
}

var (
	// DefaultInstance is the default HeroHandler instance
	DefaultInstance *HeroHandler
)

// init initializes the default HeroHandler instance
func Init() error {

	factory := core.NewHandlerFactory()
	DefaultInstance = &HeroHandler{
		factory:      factory,
		telnetServer: core.NewTelnetServer(factory),
	}

	log.Println("HeroHandler initialized")

	// Register the process manager handler
	handler := processmanagerhandler.NewProcessManagerHandler()
	if handler == nil {
		log.Fatalf("Failed to create process manager handler")
	}

	if err := DefaultInstance.factory.RegisterHandler(handler); err != nil {
		log.Fatalf("Failed to register process manager handler: %v", err)
	}

	return nil
}

func StartTelnet() error {

	if err := DefaultInstance.StartTelnet("/tmp/hero.sock", "localhost:8023"); err != nil {
		log.Fatalf("Failed to start telnet server: %v", err)
	}
	return nil
}

// StartTelnet starts the telnet server on both Unix socket and TCP port
func (h *HeroHandler) StartTelnet(socketPath string, tcpAddress string, secrets ...string) error {
	// Create a new telnet server with the factory and secrets
	h.telnetServer = core.NewTelnetServer(h.factory, secrets...)

	// Start Unix socket server
	if socketPath != "" {
		if err := h.telnetServer.Start(socketPath); err != nil {
			return fmt.Errorf("failed to start Unix socket telnet server: %v", err)
		}
		log.Printf("Telnet server started on Unix socket: %s", socketPath)
	}

	// Start TCP server
	if tcpAddress != "" {
		if err := h.telnetServer.StartTCP(tcpAddress); err != nil {
			return fmt.Errorf("failed to start TCP telnet server: %v", err)
		}
		log.Printf("Telnet server started on TCP address: %s", tcpAddress)
	}

	return nil
}

// StopTelnet stops the telnet server
func (h *HeroHandler) StopTelnet() error {
	if h.telnetServer == nil {
		return nil
	}
	return h.telnetServer.Stop()
}

// EnableSignalHandling sets up signal handling for graceful shutdown of the telnet server
func (h *HeroHandler) EnableSignalHandling(onShutdown func()) {
	if h.telnetServer == nil {
		return
	}
	h.telnetServer.EnableSignalHandling(onShutdown)
}
