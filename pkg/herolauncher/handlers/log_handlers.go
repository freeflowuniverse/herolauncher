package handlers

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

// LogHandler handles log-related routes
type LogHandler struct {
	systemLogger  *logger.Logger
	serviceLogger *logger.Logger
	jobLogger     *logger.Logger
	processLogger *logger.Logger
	logBasePath   string
}

// NewLogHandler creates a new LogHandler
func NewLogHandler(logPath string) (*LogHandler, error) {
	// Create base directories for different log types
	systemLogPath := filepath.Join(logPath, "system")
	serviceLogPath := filepath.Join(logPath, "services")
	jobLogPath := filepath.Join(logPath, "jobs")
	processLogPath := filepath.Join(logPath, "processes")

	// Create logger instances for each type
	systemLogger, err := logger.New(systemLogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create system logger: %w", err)
	}

	serviceLogger, err := logger.New(serviceLogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create service logger: %w", err)
	}

	jobLogger, err := logger.New(jobLogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create job logger: %w", err)
	}

	processLogger, err := logger.New(processLogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create process logger: %w", err)
	}

	fmt.Printf("Log handler created successfully with paths:\n  System: %s\n  Services: %s\n  Jobs: %s\n  Processes: %s\n",
		systemLogPath, serviceLogPath, jobLogPath, processLogPath)

	return &LogHandler{
		systemLogger:  systemLogger,
		serviceLogger: serviceLogger,
		jobLogger:     jobLogger,
		processLogger: processLogger,
		logBasePath:   logPath,
	}, nil
}

// LogType represents the type of logs to retrieve
type LogType string

const (
	LogTypeSystem   LogType = "system"
	LogTypeService  LogType = "service"
	LogTypeJob      LogType = "job"
	LogTypeProcess  LogType = "process"
	LogTypeAll      LogType = "all"     // Special type to retrieve logs from all sources
)

// GetLogs renders the logs page with logs content
func (h *LogHandler) GetLogs(c *fiber.Ctx) error {
	// Check which logger to use based on the log type parameter
	logTypeParam := c.Query("log_type", string(LogTypeSystem))
	
	// Parse query parameters
	category := c.Query("category", "")
	logItemType := parseLogType(c.Query("type", ""))
	maxItems := c.QueryInt("max_items", 100)
	page := c.QueryInt("page", 1)
	itemsPerPage := 20 // Default items per page
	
	// Parse time range
	fromTime := parseTimeParam(c.Query("from", ""))
	toTime := parseTimeParam(c.Query("to", ""))
	
	// Create search arguments
	searchArgs := logger.SearchArgs{
		Category:      category,
		LogType:       logItemType,
		MaxItems:      maxItems,
	}
	
	if !fromTime.IsZero() {
		searchArgs.TimestampFrom = &fromTime
	}
	
	if !toTime.IsZero() {
		searchArgs.TimestampTo = &toTime
	}

	// Variables for logs and error
	var logs []logger.LogItem
	var err error
	var logTypeTitle string
	
	// Check if we want to merge logs from all sources
	if LogType(logTypeParam) == LogTypeAll {
		// Get merged logs from all loggers
		logs, err = h.getMergedLogs(searchArgs)
		logTypeTitle = "All Logs"
	} else {
		// Select the appropriate logger based on the log type
		var selectedLogger *logger.Logger
		
		switch LogType(logTypeParam) {
		case LogTypeService:
			selectedLogger = h.serviceLogger
			logTypeTitle = "Service Logs"
		case LogTypeJob:
			selectedLogger = h.jobLogger
			logTypeTitle = "Job Logs"
		case LogTypeProcess:
			selectedLogger = h.processLogger
			logTypeTitle = "Process Logs"
		default:
			selectedLogger = h.systemLogger
			logTypeTitle = "System Logs"
		}
		
		// Check if the selected logger is properly initialized
		if selectedLogger == nil {
			return c.Render("admin/system/logs", fiber.Map{
				"title": logTypeTitle,
				"error": "Logger not initialized",
				"logTypes": []LogType{LogTypeAll, LogTypeSystem, LogTypeService, LogTypeJob, LogTypeProcess},
				"selectedLogType": logTypeParam,
			})
		}

		// Search for logs using the selected logger
		logs, err = selectedLogger.Search(searchArgs)
	}

	// Handle search error
	if err != nil {
		return c.Render("admin/system/logs", fiber.Map{
			"title": logTypeTitle,
			"error": err.Error(),
			"logTypes": []LogType{LogTypeAll, LogTypeSystem, LogTypeService, LogTypeJob, LogTypeProcess},
			"selectedLogType": logTypeParam,
		})
	}
	
	
	// Calculate total pages
	totalLogs := len(logs)
	totalPages := (totalLogs + itemsPerPage - 1) / itemsPerPage
	
	// Apply pagination
	startIndex := (page - 1) * itemsPerPage
	endIndex := startIndex + itemsPerPage
	if endIndex > totalLogs {
		endIndex = totalLogs
	}
	
	// Slice logs for current page
	pagedLogs := logs
	if startIndex < totalLogs {
		pagedLogs = logs[startIndex:endIndex]
	} else {
		pagedLogs = []logger.LogItem{}
	}
	
	// Convert logs to a format suitable for the UI
	formattedLogs := make([]fiber.Map, 0, len(pagedLogs))
	for _, log := range pagedLogs {
		logTypeStr := "INFO"
		logTypeClass := "log-info"
		if log.LogType == logger.LogTypeError {
			logTypeStr = "ERROR"
			logTypeClass = "log-error"
		}
		
		formattedLogs = append(formattedLogs, fiber.Map{
			"timestamp": log.Timestamp.Format("2006-01-02T15:04:05"),
			"category":  log.Category,
			"message":   log.Message,
			"type":      logTypeStr,
			"typeClass": logTypeClass,
		})
	}
	
	return c.Render("admin/system/logs", fiber.Map{
		"title": logTypeTitle,
		"logTypes": []LogType{LogTypeAll, LogTypeSystem, LogTypeService, LogTypeJob, LogTypeProcess},
		"selectedLogType": logTypeParam,
		"logs": formattedLogs,
		"total": totalLogs,
		"showing": len(formattedLogs),
		"page": page,
		"totalPages": totalPages,
		"categoryParam": category,
		"typeParam": c.Query("type", ""),
		"fromParam": c.Query("from", ""),
		"toParam": c.Query("to", ""),
	})
}

// GetLogsAPI returns logs in JSON format for API consumption
func (h *LogHandler) GetLogsAPI(c *fiber.Ctx) error {
	// Check which logger to use based on the log type parameter
	logTypeParam := c.Query("log_type", string(LogTypeSystem))
	
	// Parse query parameters
	category := c.Query("category", "")
	logItemType := parseLogType(c.Query("type", ""))
	maxItems := c.QueryInt("max_items", 100)
	
	// Parse time range
	fromTime := parseTimeParam(c.Query("from", ""))
	toTime := parseTimeParam(c.Query("to", ""))
	
	// Create search arguments
	searchArgs := logger.SearchArgs{
		Category:      category,
		LogType:       logItemType,
		MaxItems:      maxItems,
	}
	
	if !fromTime.IsZero() {
		searchArgs.TimestampFrom = &fromTime
	}
	
	if !toTime.IsZero() {
		searchArgs.TimestampTo = &toTime
	}

	// Variables for logs and error
	var logs []logger.LogItem
	var err error
	
	// Check if we want to merge logs from all sources
	if LogType(logTypeParam) == LogTypeAll {
		// Get merged logs from all loggers
		logs, err = h.getMergedLogs(searchArgs)
	} else {
		// Select the appropriate logger based on the log type
		var selectedLogger *logger.Logger
		
		switch LogType(logTypeParam) {
		case LogTypeService:
			selectedLogger = h.serviceLogger
		case LogTypeJob:
			selectedLogger = h.jobLogger
		case LogTypeProcess:
			selectedLogger = h.processLogger
		default:
			selectedLogger = h.systemLogger
		}
		
		// Check if the selected logger is properly initialized
		if selectedLogger == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Logger not initialized",
			})
		}

		// Search for logs using the selected logger
		logs, err = selectedLogger.Search(searchArgs)
	}

	// Handle search error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	// Convert logs to a format suitable for the UI
	response := make([]fiber.Map, 0, len(logs))
	for _, log := range logs {
		logTypeStr := "INFO"
		if log.LogType == logger.LogTypeError {
			logTypeStr = "ERROR"
		}
		
		response = append(response, fiber.Map{
			"timestamp": log.Timestamp.Format(time.RFC3339),
			"category":  log.Category,
			"message":   log.Message,
			"type":      logTypeStr,
		})
	}
	
	return c.JSON(fiber.Map{
		"logs": response,
		"total": len(logs),
	})
}

// GetLogsFragment returns logs in HTML format for Unpoly partial updates
func (h *LogHandler) GetLogsFragment(c *fiber.Ctx) error {
	// This is a fragment template for Unpoly updates

	// Check which logger to use based on the log type parameter
	logTypeParam := c.Query("log_type", string(LogTypeSystem))
	
	// Parse query parameters
	category := c.Query("category", "")
	logItemType := parseLogType(c.Query("type", ""))
	maxItems := c.QueryInt("max_items", 100)
	page := c.QueryInt("page", 1)
	itemsPerPage := 20 // Default items per page
	
	// Parse time range
	fromTime := parseTimeParam(c.Query("from", ""))
	toTime := parseTimeParam(c.Query("to", ""))
	
	// Create search arguments
	searchArgs := logger.SearchArgs{
		Category:      category,
		LogType:       logItemType,
		MaxItems:      maxItems,
	}
	
	if !fromTime.IsZero() {
		searchArgs.TimestampFrom = &fromTime
	}
	
	if !toTime.IsZero() {
		searchArgs.TimestampTo = &toTime
	}

	// Variables for logs and error
	var logs []logger.LogItem
	var err error
	var logTypeTitle string
	
	// Check if we want to merge logs from all sources
	if LogType(logTypeParam) == LogTypeAll {
		// Get merged logs from all loggers
		logs, err = h.getMergedLogs(searchArgs)
		logTypeTitle = "All Logs"
	} else {
		// Select the appropriate logger based on the log type
		var selectedLogger *logger.Logger
		
		switch LogType(logTypeParam) {
		case LogTypeService:
			selectedLogger = h.serviceLogger
			logTypeTitle = "Service Logs"
		case LogTypeJob:
			selectedLogger = h.jobLogger
			logTypeTitle = "Job Logs"
		case LogTypeProcess:
			selectedLogger = h.processLogger
			logTypeTitle = "Process Logs"
		default:
			selectedLogger = h.systemLogger
			logTypeTitle = "System Logs"
		}
		
		// Check if the selected logger is properly initialized
		if selectedLogger == nil {
			return c.Render("admin/system/logs_fragment", fiber.Map{
				"title": logTypeTitle,
				"error": "Logger not initialized",
				"logTypes": []LogType{LogTypeAll, LogTypeSystem, LogTypeService, LogTypeJob, LogTypeProcess},
				"selectedLogType": logTypeParam,
			})
		}

		// Search for logs using the selected logger
		logs, err = selectedLogger.Search(searchArgs)
	}

	// Handle search error
	if err != nil {
		return c.Render("admin/system/logs_fragment", fiber.Map{
			"title": logTypeTitle,
			"error": err.Error(),
			"logTypes": []LogType{LogTypeAll, LogTypeSystem, LogTypeService, LogTypeJob, LogTypeProcess},
			"selectedLogType": logTypeParam,
		})
	}
	
	// Calculate total pages
	totalLogs := len(logs)
	totalPages := (totalLogs + itemsPerPage - 1) / itemsPerPage
	
	// Apply pagination
	startIndex := (page - 1) * itemsPerPage
	endIndex := startIndex + itemsPerPage
	if endIndex > totalLogs {
		endIndex = totalLogs
	}
	
	// Slice logs for current page
	pagedLogs := logs
	if startIndex < totalLogs {
		pagedLogs = logs[startIndex:endIndex]
	} else {
		pagedLogs = []logger.LogItem{}
	}
	
	// Convert logs to a format suitable for the UI
	formattedLogs := make([]fiber.Map, 0, len(pagedLogs))
	for _, log := range pagedLogs {
		logTypeStr := "INFO"
		logTypeClass := "log-info"
		if log.LogType == logger.LogTypeError {
			logTypeStr = "ERROR"
			logTypeClass = "log-error"
		}
		
		formattedLogs = append(formattedLogs, fiber.Map{
			"timestamp": log.Timestamp.Format("2006-01-02T15:04:05"),
			"category":  log.Category,
			"message":   log.Message,
			"type":      logTypeStr,
			"typeClass": logTypeClass,
		})
	}
	
	// Set layout to empty to disable the layout for fragment responses
	return c.Render("admin/system/logs_fragment", fiber.Map{
		"title": logTypeTitle,
		"logTypes": []LogType{LogTypeAll, LogTypeSystem, LogTypeService, LogTypeJob, LogTypeProcess},
		"selectedLogType": logTypeParam,
		"logs": formattedLogs,
		"total": totalLogs,
		"showing": len(formattedLogs),
		"page": page,
		"totalPages": totalPages,
		"layout": "", // Disable layout for partial template
	})
}

// Helper functions

// parseLogType converts a string log type to the appropriate LogType enum
func parseLogType(logTypeStr string) logger.LogType {
	switch logTypeStr {
	case "error":
		return logger.LogTypeError
	default:
		return logger.LogTypeStdout
	}
}

// parseTimeParam parses a time string in ISO format
func parseTimeParam(timeStr string) time.Time {
	if timeStr == "" {
		return time.Time{}
	}
	
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}
	}
	
	return t
}

// getMergedLogs retrieves and merges logs from all available loggers
func (h *LogHandler) getMergedLogs(args logger.SearchArgs) ([]logger.LogItem, error) {
	// Create a slice to hold all logs
	allLogs := make([]logger.LogItem, 0)
	
	// Create a map to track errors
	errors := make(map[string]error)
	
	// Get logs from system logger if available
	if h.systemLogger != nil {
		systemLogs, err := h.systemLogger.Search(args)
		if err != nil {
			errors["system"] = err
		} else {
			// Add source information to each log item
			for i := range systemLogs {
				systemLogs[i].Category = fmt.Sprintf("system:%s", systemLogs[i].Category)
			}
			allLogs = append(allLogs, systemLogs...)
		}
	}
	
	// Get logs from service logger if available
	if h.serviceLogger != nil {
		serviceLogs, err := h.serviceLogger.Search(args)
		if err != nil {
			errors["service"] = err
		} else {
			// Add source information to each log item
			for i := range serviceLogs {
				serviceLogs[i].Category = fmt.Sprintf("service:%s", serviceLogs[i].Category)
			}
			allLogs = append(allLogs, serviceLogs...)
		}
	}
	
	// Get logs from job logger if available
	if h.jobLogger != nil {
		jobLogs, err := h.jobLogger.Search(args)
		if err != nil {
			errors["job"] = err
		} else {
			// Add source information to each log item
			for i := range jobLogs {
				jobLogs[i].Category = fmt.Sprintf("job:%s", jobLogs[i].Category)
			}
			allLogs = append(allLogs, jobLogs...)
		}
	}
	
	// Get logs from process logger if available
	if h.processLogger != nil {
		processLogs, err := h.processLogger.Search(args)
		if err != nil {
			errors["process"] = err
		} else {
			// Add source information to each log item
			for i := range processLogs {
				processLogs[i].Category = fmt.Sprintf("process:%s", processLogs[i].Category)
			}
			allLogs = append(allLogs, processLogs...)
		}
	}
	
	// Check if we have any logs
	if len(allLogs) == 0 && len(errors) > 0 {
		// Combine error messages
		errorMsgs := make([]string, 0, len(errors))
		for source, err := range errors {
			errorMsgs = append(errorMsgs, fmt.Sprintf("%s: %s", source, err.Error()))
		}
		return nil, fmt.Errorf("failed to retrieve logs: %s", strings.Join(errorMsgs, "; "))
	}
	
	// Sort logs by timestamp (newest first)
	sort.Slice(allLogs, func(i, j int) bool {
		return allLogs[i].Timestamp.After(allLogs[j].Timestamp)
	})
	
	// Apply max items limit if specified
	if args.MaxItems > 0 && len(allLogs) > args.MaxItems {
		allLogs = allLogs[:args.MaxItems]
	}
	
	return allLogs, nil
}
