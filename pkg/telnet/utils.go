package telnet

import (
	"fmt"
	"strings"
)

// FormatResult formats a result with RESULT/ENDRESULT markers
func FormatResult(content string, jobID string, interactive bool) string {
	var result strings.Builder
	
	if interactive {
		result.WriteString(fmt.Sprintf("%s%s**RESULT** %s%s\n", ColorCyan, Bold, jobID, ColorReset))
	} else {
		result.WriteString(fmt.Sprintf("**RESULT** %s\n", jobID))
	}
	
	result.WriteString(content)
	
	if interactive {
		result.WriteString(ColorCyan + Bold + "**ENDRESULT**" + ColorReset + "\n")
	} else {
		result.WriteString("**ENDRESULT**\n")
	}
	
	return result.String()
}

// FormatError formats an error message
func FormatError(err error, interactive bool) string {
	if interactive {
		return ColorRed + "Error: " + err.Error() + ColorReset + "\n"
	}
	return "Error: " + err.Error() + "\n"
}

// FormatSuccess formats a success message
func FormatSuccess(message string, interactive bool) string {
	if interactive {
		return ColorGreen + message + ColorReset + "\n"
	}
	return message + "\n"
}

// FormatTable creates a formatted table from data
func FormatTable(headers []string, rows [][]string, interactive bool) string {
	var result strings.Builder
	
	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}
	
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}
	
	// Format header
	if interactive {
		result.WriteString(Bold)
	}
	
	for i, header := range headers {
		format := fmt.Sprintf("%%-%ds", colWidths[i]+2)
		result.WriteString(fmt.Sprintf(format, header))
	}
	
	if interactive {
		result.WriteString(ColorReset)
	}
	result.WriteString("\n")
	
	// Add separator line
	for _, width := range colWidths {
		result.WriteString(strings.Repeat("-", width+2))
	}
	result.WriteString("\n")
	
	// Format rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				format := fmt.Sprintf("%%-%ds", colWidths[i]+2)
				result.WriteString(fmt.Sprintf(format, cell))
			}
		}
		result.WriteString("\n")
	}
	
	return result.String()
}
