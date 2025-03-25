package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Search finds log entries matching the given criteria
func (l *Logger) Search(args SearchArgs) ([]LogItem, error) {
	// Set default max items if not specified
	if args.MaxItems <= 0 {
		args.MaxItems = 10000
	}

	// Protect concurrent use
	l.mu.Lock()
	defer l.mu.Unlock()

	// Format category (max 10 chars, ASCII only)
	category := formatName(args.Category)
	if len(category) > 10 {
		return nil, fmt.Errorf("category cannot be longer than 10 chars")
	}

	// Set default time range if not specified
	fromTime := time.Time{}
	if args.TimestampFrom != nil {
		fromTime = *args.TimestampFrom
	}

	toTime := time.Now()
	if args.TimestampTo != nil {
		toTime = *args.TimestampTo
	}

	// Get time range as Unix timestamps
	fromUnix := fromTime.Unix()
	toUnix := toTime.Unix()
	if fromUnix > toUnix {
		return nil, fmt.Errorf("from_time cannot be after to_time: %d < %d", fromUnix, toUnix)
	}

	var result []LogItem

	// Find log files in time range
	files, err := os.ReadDir(l.Path)
	if err != nil {
		return nil, err
	}

	// Sort files by name (which is by date)
	fileNames := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".log") {
			fileNames = append(fileNames, file.Name())
		}
	}
	// Sort fileNames in chronological order
	sort.Strings(fileNames)

	for _, fileName := range fileNames {
		// Parse date-hour from filename
		dayHour := strings.TrimSuffix(fileName, ".log")
		fileTime, err := time.ParseInLocation("2006-01-02-15", dayHour, time.Local)
		if err != nil {
			continue // Skip files with invalid names
		}

		var currentItem LogItem
		var currentTime time.Time
		collecting := false

		// Read and parse log file
		content, err := os.ReadFile(filepath.Join(l.Path, fileName))
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if len(result) >= args.MaxItems {
				return result, nil
			}

			lineTrim := strings.TrimSpace(line)
			if lineTrim == "" {
				continue
			}

			// Check if this is a timestamp line
			if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "E ") {
				// Parse timestamp line
				t, err := time.Parse("15:04:05", lineTrim)
				if err != nil {
					continue
				}

				// Create a full timestamp by combining the file date with the line time
				currentTime = time.Date(
					fileTime.Year(), fileTime.Month(), fileTime.Day(),
					t.Hour(), t.Minute(), t.Second(), 0,
					fileTime.Location(),
				)

				if collecting {
					processLogItem(&result, currentItem, args, fromUnix, toUnix)
				}
				collecting = false
				continue
			}

			if collecting && len(line) > 14 && line[13] == '-' {
				processLogItem(&result, currentItem, args, fromUnix, toUnix)
				collecting = false
			}

			// Handle error log continuations
			if collecting && strings.HasPrefix(line, "E ") && len(line) > 14 && line[13] != '-' {
				// Continuation line for error log
				if len(line) > 15 {
					currentItem.Message += "\n" + strings.TrimSpace(line[15:])
				}
				continue
			}

			// Parse log line
			isError := strings.HasPrefix(line, "E ")
			if !collecting {
				// Start new item
				logType := LogTypeStdout
				if isError {
					logType = LogTypeError
				}

				// Extract category and message
				var cat, msg string

				startPos := 1 // Default for normal logs (" category - message")
				if isError {
					startPos = 2 // For error logs ("E category - message")
				}

				if len(line) > startPos+10 {
					cat = strings.TrimSpace(line[startPos : startPos+10])
				}

				if len(line) > startPos+13 && line[startPos+10:startPos+13] == " - " {
					msg = strings.TrimSpace(line[startPos+13:])
				}

				currentItem = LogItem{
					Timestamp: currentTime,
					Category:  cat,
					Message:   msg,
					LogType:   logType,
				}
				collecting = true
			} else {
				// Continuation line
				if len(lineTrim) < 16 {
					currentItem.Message += "\n"
				} else {
					if len(line) > 14 {
						currentItem.Message += "\n" + strings.TrimSpace(line[14:])
					}
				}
			}
		}

		// Add last item if collecting
		if collecting {
			processLogItem(&result, currentItem, args, fromUnix, toUnix)
		}
	}

	return result, nil
}

func processLogItem(result *[]LogItem, item LogItem, args SearchArgs, fromTime, toTime int64) {
	// Add item if it matches filters
	logEpoch := item.Timestamp.Unix()
	if logEpoch < fromTime || logEpoch > toTime {
		return
	}

	// Trim spaces from category for comparison
	itemCategory := strings.TrimSpace(item.Category)
	argsCategory := strings.TrimSpace(args.Category)

	categoryMatches := argsCategory == "" || itemCategory == argsCategory
	messageMatches := args.Message == "" || strings.Contains(item.Message, args.Message)
	typeMatches := (args.LogType == LogTypeError && item.LogType == LogTypeError) ||
		(args.LogType == LogTypeStdout && item.LogType == LogTypeStdout) ||
		(args.LogType == 0) // Assuming 0 as the default or "any" type

	if categoryMatches && messageMatches && typeMatches {
		*result = append(*result, item)
	}
}
