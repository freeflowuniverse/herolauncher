package telnet

import (
	"bufio"
	"bytes"
	"net"
	"strings"
)

// Client represents a telnet client connection
type Client struct {
	conn          net.Conn
	reader        *bufio.Reader
	interactive   bool
	commandHistory []string
	historyPos    int
	currentLine   bytes.Buffer
	rawMode       bool  // Whether the client is in raw mode (for better keyboard handling)
}

// NewClient creates a new telnet client
func NewClient(conn net.Conn) *Client {
	client := &Client{
		conn:           conn,
		reader:         bufio.NewReader(conn),
		interactive:    false,
		commandHistory: make([]string, 0),
		rawMode:        true,
	}

	// Send telnet negotiation to disable echo on client side and enable character mode
	// IAC WILL ECHO (server will handle echo)
	conn.Write([]byte{255, 251, 1})
	// IAC DO SUPPRESS GO AHEAD (character at a time mode)
	conn.Write([]byte{255, 253, 3})
	// IAC WILL SUPPRESS GO AHEAD (character at a time mode)
	conn.Write([]byte{255, 251, 3})
	// IAC DO LINEMODE (disable line buffering)
	conn.Write([]byte{255, 253, 34})
	// IAC SB LINEMODE MODE 0 IAC SE (set character-at-a-time mode)
	conn.Write([]byte{255, 250, 34, 1, 0, 255, 240})
	// Send a welcome message
	client.Println("Connected to telnet server. Type 'mysecret' to authenticate.")

	return client
}

// Close closes the client connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsInteractive returns whether the client is in interactive mode
func (c *Client) IsInteractive() bool {
	return c.interactive
}

// ToggleInteractive toggles interactive mode
func (c *Client) ToggleInteractive() {
	c.interactive = !c.interactive
	
	// Send notification about the mode change
	if c.interactive {
		c.Write(ColorGreen + "Interactive mode enabled. Using colors for output." + ColorReset + "\n")
		// Send a test color pattern to show available colors
		c.Write("Available colors: " + 
			ColorRed + "Red " + 
			ColorGreen + "Green " + 
			ColorYellow + "Yellow " + 
			ColorBlue + "Blue " + 
			ColorPurple + "Purple " + 
			ColorCyan + "Cyan" + 
			ColorReset + "\n")
	} else {
		c.Write("Interactive mode disabled. Plain text output.\n")
	}
}

// ReadLine reads a line from the client
func (c *Client) ReadLine() (string, error) {
	var line string
	var err error

	// Create a buffer for the current line
	var currentLine bytes.Buffer
	
	for {
		// Read a single byte
		var b [1]byte
		_, err = c.conn.Read(b[:])
		if err != nil {
			return "", err
		}

		// Handle telnet commands (IAC - 255)
		if b[0] == 255 { // IAC
			// Read command byte
			var cmd [1]byte
			_, err = c.conn.Read(cmd[:])
			if err != nil {
				return "", err
			}
			
			// Handle DO/DONT/WILL/WONT
			if cmd[0] >= 251 && cmd[0] <= 254 {
				// Read option byte and ignore
				var opt [1]byte
				c.conn.Read(opt[:])
			}
			continue
		}

		// Check for Ctrl+C (ASCII value 3)
		if b[0] == 3 { // Ctrl+C
			// Echo ^C and return exit command
			c.writeRaw("^C\r\n")
			// Send a proper exit command that the server will recognize
			c.writeRaw("\r\n")
			return "!!exit", nil
		}
		
		// Check for Ctrl+D (ASCII value 4) - EOF
		if b[0] == 4 { // Ctrl+D
			// Echo ^D and return exit command
			c.writeRaw("^D\r\n")
			// Send a proper exit command that the server will recognize
			c.writeRaw("\r\n")
			return "!!exit", nil
		}

		// Handle newlines
		if b[0] == '\r' || b[0] == '\n' {
			// Echo newline
			c.writeRaw("\r\n")
			
			line = currentLine.String()
			break
		}

		// Handle escape sequences
		if b[0] == 27 { // ESC
			// Try to read the next bytes for the escape sequence
			var escapeBuffer [2]byte
			n, _ := c.conn.Read(escapeBuffer[0:1]) // Should be '['
			if n == 0 || escapeBuffer[0] != '[' {
				continue
			}
			
			n, _ = c.conn.Read(escapeBuffer[1:2]) // Should be 'A' for up arrow, 'B' for down arrow
			if n == 0 {
				continue
			}
			
			// Handle up arrow (ESC [ A)
			if escapeBuffer[1] == 'A' {
				// Only process if we have command history
				if len(c.commandHistory) > 0 {
					// Clear current line by sending carriage return and then erase line
					c.writeRaw("\r\033[K") // CR + Erase line
					
					// Navigate command history (backward)
					if c.historyPos > 0 {
						c.historyPos--
					} else if len(c.commandHistory) > 0 {
						c.historyPos = 0
					}
					
					// Display the command from history
					historyCmd := c.commandHistory[c.historyPos]
					c.writeRaw(historyCmd)
					currentLine.Reset()
					currentLine.WriteString(historyCmd)
				}
				continue
			}
			
			// Handle down arrow (ESC [ B)
			if escapeBuffer[1] == 'B' {
				// Clear current line by sending carriage return and then erase line
				c.writeRaw("\r\033[K") // CR + Erase line
				
				// Navigate command history (forward)
				if len(c.commandHistory) > 0 && c.historyPos < len(c.commandHistory)-1 {
					c.historyPos++
					
					// Display the command from history
					historyCmd := c.commandHistory[c.historyPos]
					c.writeRaw(historyCmd)
					currentLine.Reset()
					currentLine.WriteString(historyCmd)
				} else {
					// If at the end of history, show empty line
					c.historyPos = len(c.commandHistory)
					currentLine.Reset()
				}
				continue
			}
			
			// Handle other escape sequences like delete key (ESC [ 3 ~)
			if escapeBuffer[1] == '3' {
				// Read one more byte
				var delByte [1]byte
				n, _ := c.conn.Read(delByte[:])
				if n > 0 && delByte[0] == '~' {
					// Delete key - ignore for now
				}
			}
			
			// Ignore other escape sequences
			continue
		}

		// Handle backspace (ASCII 8 or 127)
		if b[0] == 8 || b[0] == 127 {
			if currentLine.Len() > 0 {
				// Remove last character from buffer
				s := currentLine.String()
				currentLine.Reset()
				currentLine.WriteString(s[:len(s)-1])
				
				// Echo the backspace (move cursor back, print space, move cursor back again)
				c.writeRaw("\b \b")
			} else {
				// Bell sound for backspace at beginning of line
				c.writeRaw("\a")
			}
			continue
		}

		// Echo the character
		c.writeRaw(string(b[0]))
		
		// Add to current line buffer
		currentLine.WriteByte(b[0])
	}

	// Add to command history if not empty and not a duplicate of the last command
	if line != "" && (len(c.commandHistory) == 0 || line != c.commandHistory[len(c.commandHistory)-1]) {
		c.commandHistory = append(c.commandHistory, line)
		c.historyPos = len(c.commandHistory)
	}

	return line, nil
}

// writeRaw writes raw data to the connection
func (c *Client) writeRaw(data string) {
	c.conn.Write([]byte(data))
}

// Write writes data to the client
func (c *Client) Write(data string) {
	// In non-interactive mode, strip all ANSI color codes
	if !c.interactive {
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
	
	c.writeRaw(data)
}

// Println writes a line to the client
func (c *Client) Println(text string) {
	c.Write(text + "\n")
}

// PrintlnRed writes a red line to the client
func (c *Client) PrintlnRed(text string) {
	c.Write(ColorRed + text + ColorReset + "\n")
}

// PrintlnGreen writes a green line to the client
func (c *Client) PrintlnGreen(text string) {
	c.Write(ColorGreen + text + ColorReset + "\n")
}

// PrintlnYellow writes a yellow line to the client
func (c *Client) PrintlnYellow(text string) {
	c.Write(ColorYellow + text + ColorReset + "\n")
}

// PrintlnBlue writes a blue line to the client
func (c *Client) PrintlnBlue(text string) {
	c.Write(ColorBlue + text + ColorReset + "\n")
}

// PrintlnCyan writes a cyan line to the client
func (c *Client) PrintlnCyan(text string) {
	c.Write(ColorCyan + text + ColorReset + "\n")
}

// PrintlnPurple writes a purple line to the client
func (c *Client) PrintlnPurple(text string) {
	c.Write(ColorPurple + text + ColorReset + "\n")
}

// PrintlnBold writes a bold line to the client
func (c *Client) PrintlnBold(text string) {
	c.Write(Bold + text + ColorReset + "\n")
}

// PrintHelp prints the help text
func (c *Client) PrintHelp() {
	// Save current interactive state to restore it later
	origInteractive := c.interactive
	
	// Force interactive mode for help display
	c.interactive = true
	
	c.Println("\n" + ColorCyan + Bold + "===== Telnet Client Help =====" + ColorReset + "\n")
	
	// Special commands section
	c.Println(ColorYellow + Bold + "Special Commands:" + ColorReset)
	c.Println(ColorGreen + "  !!help, ? or h" + ColorReset + " - Show this help text")
	c.Println(ColorGreen + "  !!interactive or !!i" + ColorReset + " - Toggle interactive mode with colors")
	c.Println(ColorGreen + "  !!exit or !!quit or q" + ColorReset + " - Close the connection")
	
	// Keyboard shortcuts section
	c.Println("\n" + ColorYellow + Bold + "Keyboard Shortcuts:" + ColorReset)
	c.Println(ColorGreen + "  Ctrl+C" + ColorReset + " - Interrupt current operation/exit")
	c.Println(ColorGreen + "  Ctrl+D" + ColorReset + " - Send EOF/exit")
	c.Println(ColorGreen + "  Up/Down arrows" + ColorReset + " - Navigate command history")
	c.Println(ColorGreen + "  Backspace/Delete" + ColorReset + " - Delete characters")
	
	// Color codes section (only if in interactive mode)
	if c.interactive {
		c.Println("\n" + ColorYellow + Bold + "Available Colors:" + ColorReset)
		c.Println(ColorRed + "  Red" + ColorReset + " - Error messages")
		c.Println(ColorGreen + "  Green" + ColorReset + " - Success messages")
		c.Println(ColorYellow + "  Yellow" + ColorReset + " - Warnings and important information")
		c.Println(ColorBlue + "  Blue" + ColorReset + " - Informational messages")
		c.Println(ColorPurple + "  Purple" + ColorReset + " - System messages")
		c.Println(ColorCyan + "  Cyan" + ColorReset + " - Prompts and highlights")
	}
	
	c.Println("\n" + ColorCyan + Bold + "============================" + ColorReset + "\n")
	
	// Restore original interactive state
	c.interactive = origInteractive
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
