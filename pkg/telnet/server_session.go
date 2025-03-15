package telnet

import (
	"bufio"
	"net"
	"strings"
)

// Session represents a telnet session with a connected client
type Session struct {
	conn           net.Conn
	reader         *bufio.Reader
	interactive    bool
	commandHistory []string
	historyPos     int
	rawMode        bool // Whether the session is in raw mode (for better keyboard handling)
}

// NewSession creates a new telnet session
func NewSession(conn net.Conn) *Session {
	session := &Session{
		conn:           conn,
		reader:         bufio.NewReader(conn),
		interactive:    false,
		commandHistory: make([]string, 0),
		rawMode:        true,
	}

	// Send telnet negotiation to enable local echo on client side
	// IAC WONT ECHO (client should handle echo)
	conn.Write([]byte{255, 252, 1})
	// IAC DO ECHO (client should echo)
	conn.Write([]byte{255, 253, 1})
	// IAC DO SUPPRESS GO AHEAD (character at a time mode)
	conn.Write([]byte{255, 253, 3})
	// IAC WILL SUPPRESS GO AHEAD (character at a time mode)
	conn.Write([]byte{255, 251, 3})
	// Send a welcome message
	session.Println("Connected to telnet server.")

	return session
}

// Close closes the session connection
func (s *Session) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

// IsInteractive returns whether the client is in interactive mode
func (s *Session) IsInteractive() bool {
	return s.interactive
}

// ToggleInteractive toggles interactive mode
func (s *Session) ToggleInteractive() {
	s.interactive = !s.interactive

	// Send notification about the mode change
	if s.interactive {
		s.Write(ColorGreen + "Interactive mode enabled. Using colors for output." + ColorReset + "\n")
		// Send a test color pattern to show available colors
		s.Write("Available colors: " +
			ColorRed + "Red " +
			ColorGreen + "Green " +
			ColorYellow + "Yellow " +
			ColorBlue + "Blue " +
			ColorPurple + "Purple " +
			ColorCyan + "Cyan" +
			ColorReset + "\n")
	} else {
		s.Write("Interactive mode disabled. Plain text output.\n")
	}
}

// ReadLine reads a line from the client
// If suppressEcho is true, the input should not be displayed (for passwords)
// However, we're currently ignoring this parameter for better usability
func (s *Session) ReadLine(suppressEcho bool) (string, error) {
	// Write a prompt to make it clear we're waiting for input
	s.writeRaw("\r\n> ")

	// Read a line directly from the reader using the built-in bufio.ReadString method
	line, err := s.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Trim CR and LF
	line = strings.TrimRight(line, "\r\n")

	// Add to command history if not empty and not a duplicate of the last command
	if line != "" && (len(s.commandHistory) == 0 || line != s.commandHistory[len(s.commandHistory)-1]) {
		s.commandHistory = append(s.commandHistory, line)
		s.historyPos = len(s.commandHistory)
	}

	return line, nil
}

// writeRaw writes raw data to the connection
func (s *Session) writeRaw(data string) {
	s.conn.Write([]byte(data))
}

// Write writes data to the client
func (s *Session) Write(data string) {
	// In non-interactive mode, strip all ANSI color codes
	if !s.interactive {
		// Strip ANSI color codes
		data = stripAnsiCodes(data)
	} else {
		// In interactive mode, ensure color codes are properly formatted
		// Replace common color placeholders with actual ANSI codes if needed
		data = strings.ReplaceAll(data, "{{RED}}", ColorRed)
		data = strings.ReplaceAll(data, "{{GREEN}}", ColorGreen)
		data = strings.ReplaceAll(data, "{{YELLOW}}", ColorYellow)
		data = strings.ReplaceAll(data, "{{BLUE}}", ColorBlue)
		data = strings.ReplaceAll(data, "{{PURPLE}}", ColorPurple)
		data = strings.ReplaceAll(data, "{{CYAN}}", ColorCyan)
		data = strings.ReplaceAll(data, "{{RESET}}", ColorReset)
		data = strings.ReplaceAll(data, "{{BOLD}}", Bold)
	}

	// Convert all \n to \r\n for proper telnet line endings
	data = strings.ReplaceAll(data, "\n", "\r\n")

	s.writeRaw(data)
}

// Println writes a line to the client
func (s *Session) Println(text string) {
	s.Write(text + "\n")
}

// PrintlnRed writes a red line to the client
func (s *Session) PrintlnRed(text string) {
	s.Write(ColorRed + text + ColorReset + "\n")
}

// PrintlnGreen writes a green line to the client
func (s *Session) PrintlnGreen(text string) {
	s.Write(ColorGreen + text + ColorReset + "\n")
}

// PrintlnYellow writes a yellow line to the client
func (s *Session) PrintlnYellow(text string) {
	s.Write(ColorYellow + text + ColorReset + "\n")
}

// PrintlnBlue writes a blue line to the client
func (s *Session) PrintlnBlue(text string) {
	s.Write(ColorBlue + text + ColorReset + "\n")
}

// PrintlnCyan writes a cyan line to the client
func (s *Session) PrintlnCyan(text string) {
	s.Write(ColorCyan + text + ColorReset + "\n")
}

// PrintlnPurple writes a purple line to the client
func (s *Session) PrintlnPurple(text string) {
	s.Write(ColorPurple + text + ColorReset + "\n")
}

// PrintlnBold writes a bold line to the client
func (s *Session) PrintlnBold(text string) {
	s.Write(Bold + text + ColorReset + "\n")
}

// PrintHelp prints the help text
func (s *Session) PrintHelp() {
	// Save current interactive state to restore it later
	origInteractive := s.interactive

	// Force interactive mode for help display
	s.interactive = true

	s.Println("\n" + ColorCyan + Bold + "===== Telnet Client Help =====" + ColorReset + "\n")

	// Special commands section
	s.Println(ColorYellow + Bold + "Special Commands:" + ColorReset)
	s.Println(ColorGreen + "  !!help, ? or h" + ColorReset + " - Show this help text")
	s.Println(ColorGreen + "  !!interactive or !!i" + ColorReset + " - Toggle interactive mode with colors")
	s.Println(ColorGreen + "  !!exit or !!quit or q" + ColorReset + " - Close the connection")

	// Custom commands section
	s.Println("\n" + ColorYellow + Bold + "Available Commands:" + ColorReset)
	s.Println(ColorGreen + "  hello" + ColorReset + " - Simple greeting")
	s.Println(ColorGreen + "  status" + ColorReset + " - Show system status")
	s.Println(ColorGreen + "  error" + ColorReset + " - Example error formatting")
	s.Println(ColorGreen + "  success" + ColorReset + " - Example success formatting")
	s.Println(ColorGreen + "  result" + ColorReset + " - Example result formatting")
	s.Println(ColorGreen + "  !!echo" + ColorReset + " - Enter echo mode")

	// Keyboard shortcuts section
	s.Println("\n" + ColorYellow + Bold + "Keyboard Shortcuts:" + ColorReset)
	s.Println(ColorGreen + "  Ctrl+C" + ColorReset + " - Interrupt current operation/exit")
	s.Println(ColorGreen + "  Ctrl+D" + ColorReset + " - Send EOF/exit")
	s.Println(ColorGreen + "  Up/Down arrows" + ColorReset + " - Navigate command history")
	s.Println(ColorGreen + "  Backspace/Delete" + ColorReset + " - Delete characters")

	// Color codes section (only if in interactive mode)
	if s.interactive {
		s.Println("\n" + ColorYellow + Bold + "Available Colors:" + ColorReset)
		s.Println(ColorRed + "  Red" + ColorReset + " - Error messages")
		s.Println(ColorGreen + "  Green" + ColorReset + " - Success messages")
		s.Println(ColorYellow + "  Yellow" + ColorReset + " - Warnings and important information")
		s.Println(ColorBlue + "  Blue" + ColorReset + " - Informational messages")
		s.Println(ColorPurple + "  Purple" + ColorReset + " - System messages")
		s.Println(ColorCyan + "  Cyan" + ColorReset + " - Prompts and highlights")
	}

	s.Println("\n" + ColorCyan + Bold + "============================" + ColorReset + "\n")

	// Restore original interactive state
	s.interactive = origInteractive
}

// stripAnsiCodes removes ANSI color codes from a string
func stripAnsiCodes(s string) string {
	r := strings.NewReplacer(
		ColorReset, "",
		ColorRed, "",
		ColorGreen, "",
		ColorYellow, "",
		ColorBlue, "",
		ColorPurple, "",
		ColorCyan, "",
		ColorWhite, "",
		Bold, "",
	)
	return r.Replace(s)
}
