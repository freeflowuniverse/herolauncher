package handlers

import (
	"fmt"
	"log"

	"github.com/freeflowuniverse/herolauncher/pkg/herojobs"
	"github.com/gofiber/fiber/v2"
)

// HeroJobsClientInterface defines the interface for the HeroJobs client
type HeroJobsClientInterface interface {
	Connect() error
	Close() error
	SubmitJob(job *herojobs.Job) (*herojobs.Job, error)
	GetJob(jobID string) (*herojobs.Job, error)
	DeleteJob(jobID string) error
	ListJobs(circleID, topic string) ([]string, error)
	QueueSize(circleID, topic string) (int64, error)
	QueueEmpty(circleID, topic string) error
	QueueGet(circleID, topic string) (*herojobs.Job, error)
	CreateJob(circleID, topic, sessionKey, heroScript, rhaiScript string) (*herojobs.Job, error)
}

// JobHandler handles job-related routes
type JobHandler struct {
	client HeroJobsClientInterface
	logger *log.Logger
}

// NewJobHandler creates a new JobHandler
func NewJobHandler(socketPath string, logger *log.Logger) (*JobHandler, error) {
	client, err := herojobs.NewClient(socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create HeroJobs client: %w", err)
	}

	return &JobHandler{
		client: client,
		logger: logger,
	}, nil
}

// RegisterRoutes registers job API routes
func (h *JobHandler) RegisterRoutes(app *fiber.App) {
	// Register common routes to both API and admin groups
	jobRoutes := func(group fiber.Router) {
		group.Post("/submit", h.submitJob)
		group.Get("/get/:id", h.getJob)
		group.Delete("/delete/:id", h.deleteJob)
		group.Get("/list", h.listJobs)
		group.Get("/queue/size", h.queueSize)
		group.Post("/queue/empty", h.queueEmpty)
		group.Get("/queue/get", h.queueGet)
		group.Post("/create", h.createJob)
	}

	// Apply common routes to API group
	apiJobs := app.Group("/api/jobs")
	jobRoutes(apiJobs)

	// Apply common routes to admin group
	adminJobs := app.Group("/admin/jobs")
	jobRoutes(adminJobs)
}

// @Summary Submit a job
// @Description Submit a new job to the HeroJobs server
// @Tags jobs
// @Accept json
// @Produce json
// @Param job body herojobs.Job true "Job to submit"
// @Success 200 {object} herojobs.Job
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/jobs/submit [post]
// @Router /admin/jobs/submit [post]
func (h *JobHandler) submitJob(c *fiber.Ctx) error {
	// Connect to the HeroJobs server
	if err := h.client.Connect(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to connect to HeroJobs server: %v", err),
		})
	}
	defer h.client.Close()

	// Parse job from request body
	var job herojobs.Job
	if err := c.BodyParser(&job); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to parse job data: %v", err),
		})
	}

	// Submit job
	submittedJob, err := h.client.SubmitJob(&job)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to submit job: %v", err),
		})
	}

	return c.JSON(submittedJob)
}

// @Summary Get a job
// @Description Get a job by ID
// @Tags jobs
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} herojobs.Job
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/jobs/get/{id} [get]
// @Router /admin/jobs/get/{id} [get]
func (h *JobHandler) getJob(c *fiber.Ctx) error {
	// Connect to the HeroJobs server
	if err := h.client.Connect(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to connect to HeroJobs server: %v", err),
		})
	}
	defer h.client.Close()

	// Get job ID from path parameter
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Job ID is required",
		})
	}

	// Get job
	job, err := h.client.GetJob(jobID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get job: %v", err),
		})
	}

	return c.JSON(job)
}

// @Summary Delete a job
// @Description Delete a job by ID
// @Tags jobs
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/jobs/delete/{id} [delete]
// @Router /admin/jobs/delete/{id} [delete]
func (h *JobHandler) deleteJob(c *fiber.Ctx) error {
	// Connect to the HeroJobs server
	if err := h.client.Connect(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to connect to HeroJobs server: %v", err),
		})
	}
	defer h.client.Close()

	// Get job ID from path parameter
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Job ID is required",
		})
	}

	// Delete job
	if err := h.client.DeleteJob(jobID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to delete job: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Job %s deleted successfully", jobID),
	})
}

// @Summary List jobs
// @Description List jobs by circle ID and topic
// @Tags jobs
// @Produce json
// @Param circleid query string true "Circle ID"
// @Param topic query string true "Topic"
// @Success 200 {object} map[string][]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/jobs/list [get]
// @Router /admin/jobs/list [get]
func (h *JobHandler) listJobs(c *fiber.Ctx) error {
	// Connect to the HeroJobs server
	if err := h.client.Connect(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to connect to HeroJobs server: %v", err),
		})
	}
	defer h.client.Close()

	// Get parameters from query
	circleID := c.Query("circleid")
	if circleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Circle ID is required",
		})
	}

	topic := c.Query("topic")
	if topic == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Topic is required",
		})
	}

	// List jobs
	jobs, err := h.client.ListJobs(circleID, topic)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to list jobs: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"jobs":   jobs,
	})
}

// @Summary Get queue size
// @Description Get the size of a job queue by circle ID and topic
// @Tags jobs
// @Produce json
// @Param circleid query string true "Circle ID"
// @Param topic query string true "Topic"
// @Success 200 {object} map[string]int64
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/jobs/queue/size [get]
// @Router /admin/jobs/queue/size [get]
func (h *JobHandler) queueSize(c *fiber.Ctx) error {
	// Connect to the HeroJobs server
	if err := h.client.Connect(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to connect to HeroJobs server: %v", err),
		})
	}
	defer h.client.Close()

	// Get parameters from query
	circleID := c.Query("circleid")
	if circleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Circle ID is required",
		})
	}

	topic := c.Query("topic")
	if topic == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Topic is required",
		})
	}

	// Get queue size
	size, err := h.client.QueueSize(circleID, topic)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get queue size: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"size":   size,
	})
}

// @Summary Empty queue
// @Description Empty a job queue by circle ID and topic
// @Tags jobs
// @Accept json
// @Produce json
// @Param body body object true "Queue parameters"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/jobs/queue/empty [post]
// @Router /admin/jobs/queue/empty [post]
func (h *JobHandler) queueEmpty(c *fiber.Ctx) error {
	// Connect to the HeroJobs server
	if err := h.client.Connect(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to connect to HeroJobs server: %v", err),
		})
	}
	defer h.client.Close()

	// Parse parameters from request body
	var params struct {
		CircleID string `json:"circleid"`
		Topic    string `json:"topic"`
	}
	if err := c.BodyParser(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to parse parameters: %v", err),
		})
	}

	if params.CircleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Circle ID is required",
		})
	}

	if params.Topic == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Topic is required",
		})
	}

	// Empty queue
	if err := h.client.QueueEmpty(params.CircleID, params.Topic); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to empty queue: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Queue for circle %s and topic %s emptied successfully", params.CircleID, params.Topic),
	})
}

// @Summary Get job from queue
// @Description Get a job from a queue without removing it
// @Tags jobs
// @Produce json
// @Param circleid query string true "Circle ID"
// @Param topic query string true "Topic"
// @Success 200 {object} herojobs.Job
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/jobs/queue/get [get]
// @Router /admin/jobs/queue/get [get]
func (h *JobHandler) queueGet(c *fiber.Ctx) error {
	// Connect to the HeroJobs server
	if err := h.client.Connect(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to connect to HeroJobs server: %v", err),
		})
	}
	defer h.client.Close()

	// Get parameters from query
	circleID := c.Query("circleid")
	if circleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Circle ID is required",
		})
	}

	topic := c.Query("topic")
	if topic == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Topic is required",
		})
	}

	// Get job from queue
	job, err := h.client.QueueGet(circleID, topic)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get job from queue: %v", err),
		})
	}

	return c.JSON(job)
}

// @Summary Create job
// @Description Create a new job with the given parameters
// @Tags jobs
// @Accept json
// @Produce json
// @Param body body object true "Job parameters"
// @Success 200 {object} herojobs.Job
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/jobs/create [post]
// @Router /admin/jobs/create [post]
func (h *JobHandler) createJob(c *fiber.Ctx) error {
	// Connect to the HeroJobs server
	if err := h.client.Connect(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to connect to HeroJobs server: %v", err),
		})
	}
	defer h.client.Close()

	// Parse parameters from request body
	var params struct {
		CircleID   string `json:"circleid"`
		Topic      string `json:"topic"`
		SessionKey string `json:"sessionkey"`
		HeroScript string `json:"heroscript"`
		RhaiScript string `json:"rhaiscript"`
	}
	if err := c.BodyParser(&params); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to parse parameters: %v", err),
		})
	}

	if params.CircleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Circle ID is required",
		})
	}

	if params.Topic == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Topic is required",
		})
	}

	// Create job
	job, err := h.client.CreateJob(
		params.CircleID,
		params.Topic,
		params.SessionKey,
		params.HeroScript,
		params.RhaiScript,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create job: %v", err),
		})
	}

	return c.JSON(job)
}
