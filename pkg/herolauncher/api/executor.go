package api

import (
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/executor"

	"github.com/gofiber/fiber/v2"
)

// ExecutorHandler handles executor-related API endpoints
type ExecutorHandler struct {
	executor *executor.Executor
}

// NewExecutorHandler creates a new executor handler
func NewExecutorHandler(exec *executor.Executor) *ExecutorHandler {
	return &ExecutorHandler{
		executor: exec,
	}
}

// RegisterRoutes registers executor routes to the fiber app
func (h *ExecutorHandler) RegisterRoutes(app *fiber.App) {
	group := app.Group("/api/executor")

	// @Summary Execute a command
	// @Description Execute a command and return a job ID
	// @Tags executor
	// @Accept json
	// @Produce json
	// @Param command body ExecuteCommandRequest true "Command to execute"
	// @Success 200 {object} ExecuteCommandResponse
	// @Failure 400 {object} ErrorResponse
	// @Router /api/executor/execute [post]
	group.Post("/execute", h.executeCommand)

	// @Summary List all jobs
	// @Description Get a list of all command execution jobs
	// @Tags executor
	// @Produce json
	// @Success 200 {array} JobResponse
	// @Router /api/executor/jobs [get]
	group.Get("/jobs", h.listJobs)

	// @Summary Get job details
	// @Description Get details of a specific job by ID
	// @Tags executor
	// @Produce json
	// @Param id path string true "Job ID"
	// @Success 200 {object} JobResponse
	// @Failure 404 {object} ErrorResponse
	// @Router /api/executor/jobs/{id} [get]
	group.Get("/jobs/:id", h.getJob)
}

// @Summary Execute a command
// @Description Execute a command and return a job ID
// @Tags executor
// @Accept json
// @Produce json
// @Param command body ExecuteCommandRequest true "Command to execute"
// @Success 200 {object} ExecuteCommandResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/executor/execute [post]
func (h *ExecutorHandler) executeCommand(c *fiber.Ctx) error {
	var req ExecuteCommandRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid request: " + err.Error(),
		})
	}

	jobID, err := h.executor.ExecuteCommand(req.Command, req.Args)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to execute command: " + err.Error(),
		})
	}

	return c.JSON(ExecuteCommandResponse{
		JobID: jobID,
	})
}

// @Summary List all jobs
// @Description Get a list of all command execution jobs
// @Tags executor
// @Produce json
// @Success 200 {array} JobResponse
// @Router /api/executor/jobs [get]
func (h *ExecutorHandler) listJobs(c *fiber.Ctx) error {
	jobs := h.executor.ListJobs()

	response := make([]JobResponse, 0, len(jobs))
	for _, job := range jobs {
		var endTime time.Time
		if job.Status == "completed" || job.Status == "failed" {
			endTime = job.EndTime
		}
		response = append(response, JobResponse{
			ID:        job.ID,
			Command:   job.Command,
			Args:      job.Args,
			StartTime: job.StartTime,
			EndTime:   endTime,
			Status:    job.Status,
			Output:    job.Output,
			Error:     job.Error,
		})
	}

	return c.JSON(response)
}

// @Summary Get job details
// @Description Get details of a specific job by ID
// @Tags executor
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} JobResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/executor/jobs/{id} [get]
func (h *ExecutorHandler) getJob(c *fiber.Ctx) error {
	jobID := c.Params("id")

	job, err := h.executor.GetJob(jobID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	var endTime time.Time
	if job.Status == "completed" || job.Status == "failed" {
		endTime = job.EndTime
	}

	return c.JSON(JobResponse{
		ID:        job.ID,
		Command:   job.Command,
		Args:      job.Args,
		StartTime: job.StartTime,
		EndTime:   endTime,
		Status:    job.Status,
		Output:    job.Output,
		Error:     job.Error,
	})
}


