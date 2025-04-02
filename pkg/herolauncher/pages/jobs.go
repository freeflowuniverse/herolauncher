package pages

import (
	"fmt"
	"log"

	"github.com/freeflowuniverse/herolauncher/pkg/herojobs"
	"github.com/gofiber/fiber/v2"
)

// JobDisplayInfo represents information about a job for display purposes
type JobDisplayInfo struct {
	JobID         string `json:"jobid"`
	CircleID      string `json:"circleid"`
	Topic         string `json:"topic"`
	Status        string `json:"status"`
	SessionKey    string `json:"sessionkey"`
	HeroScript    string `json:"heroscript"`
	RhaiScript    string `json:"rhaiscript"`
	Result        string `json:"result"`
	Error         string `json:"error"`
	TimeScheduled int64  `json:"time_scheduled"`
	TimeStart     int64  `json:"time_start"`
	TimeEnd       int64  `json:"time_end"`
	Timeout       int64  `json:"timeout"`
}

// JobHandler handles job-related page routes
type JobHandler struct {
	client *herojobs.Client
	logger *log.Logger
}

// NewJobHandler creates a new job handler with the provided socket path
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

// RegisterRoutes registers job page routes
func (h *JobHandler) RegisterRoutes(app *fiber.App) {
	// Register routes for /jobs
	jobs := app.Group("/jobs")
	jobs.Get("/", h.getJobsPage)
	jobs.Get("/list", h.getJobsList)
	
	// Register the same routes under /admin/jobs for consistency
	adminJobs := app.Group("/admin/jobs")
	adminJobs.Get("/", h.getJobsPage)
	adminJobs.Get("/list", h.getJobsList)
}

// getJobsPage renders the jobs page
func (h *JobHandler) getJobsPage(c *fiber.Ctx) error {
	// Check if we can connect to the HeroJobs server
	var warning string
	if err := h.client.Connect(); err != nil {
		warning = "Could not connect to HeroJobs server: " + err.Error()
		h.logger.Printf("Warning: %s", warning)
	} else {
		h.client.Close()
	}

	return c.Render("admin/jobs", fiber.Map{
		"title":   "Jobs",
		"warning": warning,
		"error":   "",
	})
}

// getJobsList returns the jobs list fragment for AJAX updates
func (h *JobHandler) getJobsList(c *fiber.Ctx) error {
	// Get parameters from query
	circleID := c.Query("circleid", "")
	topic := c.Query("topic", "")

	// Get jobs
	jobs, err := h.getJobsData(circleID, topic)
	if err != nil {
		h.logger.Printf("Error getting jobs: %v", err)
		// Return the error in the template
		return c.Render("admin/jobs_list_fragment", fiber.Map{
			"error": fmt.Sprintf("Failed to get jobs: %v", err),
			"jobs":  []JobDisplayInfo{},
		})
	}

	// Render only the jobs fragment
	return c.Render("admin/jobs_list_fragment", fiber.Map{
		"jobs": jobs,
	})
}

// getJobsData gets job data from the HeroJobs server
func (h *JobHandler) getJobsData(circleID, topic string) ([]JobDisplayInfo, error) {
	// Connect to the HeroJobs server
	if err := h.client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to HeroJobs server: %w", err)
	}
	defer h.client.Close()

	// If circleID and topic are not provided, try to list all jobs
	if circleID == "" && topic == "" {
		// Try to get some default jobs
		defaultCircles := []string{"default", "system"}
		defaultTopics := []string{"default", "system"}
		
		var allJobs []JobDisplayInfo
		
		// Try each combination
		for _, circle := range defaultCircles {
			for _, t := range defaultTopics {
				jobIDs, err := h.client.ListJobs(circle, t)
				if err != nil {
					h.logger.Printf("Could not list jobs for circle=%s, topic=%s: %v", circle, t, err)
					continue
				}
				
				for _, jobID := range jobIDs {
					job, err := h.client.GetJob(jobID)
					if err != nil {
						h.logger.Printf("Error getting job %s: %v", jobID, err)
						continue
					}
					
					allJobs = append(allJobs, JobDisplayInfo{
						JobID:         job.JobID,
						CircleID:      job.CircleID,
						Topic:         job.Topic,
						Status:        string(job.Status),
						SessionKey:    job.SessionKey,
						HeroScript:    job.HeroScript,
						RhaiScript:    job.RhaiScript,
						Result:        job.Result,
						Error:         job.Error,
						TimeScheduled: job.TimeScheduled,
						TimeStart:     job.TimeStart,
						TimeEnd:       job.TimeEnd,
						Timeout:       job.Timeout,
					})
				}
			}
		}
		
		return allJobs, nil
	} else if circleID == "" || topic == "" {
		// If only one of the parameters is provided, we can't list jobs
		return []JobDisplayInfo{}, nil
	}

	// List jobs
	jobIDs, err := h.client.ListJobs(circleID, topic)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	// Get details for each job
	jobsList := make([]JobDisplayInfo, 0, len(jobIDs))
	for _, jobID := range jobIDs {
		job, err := h.client.GetJob(jobID)
		if err != nil {
			h.logger.Printf("Error getting job %s: %v", jobID, err)
			continue
		}

		jobInfo := JobDisplayInfo{
			JobID:         job.JobID,
			CircleID:      job.CircleID,
			Topic:         job.Topic,
			Status:        string(job.Status),
			SessionKey:    job.SessionKey,
			HeroScript:    job.HeroScript,
			RhaiScript:    job.RhaiScript,
			Result:        job.Result,
			Error:         job.Error,
			TimeScheduled: job.TimeScheduled,
			TimeStart:     job.TimeStart,
			TimeEnd:       job.TimeEnd,
			Timeout:       job.Timeout,
		}
		jobsList = append(jobsList, jobInfo)
	}

	return jobsList, nil
}
