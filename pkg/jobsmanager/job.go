package jobsmanager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/data/ourdb"
	"github.com/freeflowuniverse/herolauncher/pkg/tools"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	// JobStatusNew indicates a newly created job
	JobStatusNew JobStatus = "new"
	// JobStatusActive indicates a job that is currently being processed
	JobStatusActive JobStatus = "active"
	// JobStatusError indicates a job that encountered an error
	JobStatusError JobStatus = "error"
	// JobStatusDone indicates a job that has been completed successfully
	JobStatusDone JobStatus = "done"
)

// ParamsType represents the type of parameters for a job
type ParamsType string

const (
	// ParamsTypeHeroScript indicates parameters in HeroScript format
	ParamsTypeHeroScript ParamsType = "heroscript"
	// ParamsTypeRhaiScript indicates parameters in RhaiScript format
	ParamsTypeRhaiScript ParamsType = "rhaiscript"
	// ParamsTypeOpenRPC indicates parameters in OpenRPC format
	ParamsTypeOpenRPC ParamsType = "openrpc"
	ParamsTypeAI      ParamsType = "ai"
)

// Job represents a job to be processed
type Job struct {
	JobID         uint32     `json:"jobid"`
	SessionKey    string     `json:"sessionkey"`
	CircleID      string     `json:"circleid"`
	Topic         string     `json:"topic"`
	Params        string     `json:"params"`      //can be a script, rpc, etc.
	ParamsType    ParamsType `json:"params_type"` // Type of params: heroscript, rhaiscript, openrpc
	Timeout       int64      `json:"timeout"`
	Status        JobStatus  `json:"status"`
	TimeScheduled int64      `json:"time_scheduled"`
	TimeStart     int64      `json:"time_start"`
	TimeEnd       int64      `json:"time_end"`
	Error         string     `json:"error"`
	Result        string     `json:"result"`
	Log           bool       `json:"log"` // Whether to enable logging for this job

	// OurDB client for job storage
	dbClient *ourdb.Client
}

// NewJob creates a new job with default values
func NewJob() *Job {
	now := time.Now().Unix()
	return &Job{
		JobID:         0, // Default to 0, will be auto-incremented by OurDB
		Topic:         "default",
		Status:        JobStatusNew,
		ParamsType:    ParamsTypeHeroScript, // Default to HeroScript
		TimeScheduled: now,
	}
}

// NewJobFromJSON creates a new job from a JSON string
func NewJobFromJSON(jsonStr string) (*Job, error) {
	job := &Job{}
	err := json.Unmarshal([]byte(jsonStr), job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Set default values if not provided
	if job.JobID == 0 {
		job.JobID = 0 // Default to 0, will be auto-incremented by OurDB
	}
	if job.Topic == "" {
		job.Topic = "default"
	}
	if job.Status == "" {
		job.Status = JobStatusNew
	}
	if job.ParamsType == "" {
		job.ParamsType = ParamsTypeOpenRPC
	}
	if job.TimeScheduled == 0 {
		job.TimeScheduled = time.Now().Unix()
	}

	return job, nil
}

// ToJSON converts the job to a JSON string
func (j *Job) ToJSON() (string, error) {
	bytes, err := json.Marshal(j)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job: %w", err)
	}
	return string(bytes), nil
}

// QueueKey returns the Redis queue key for this job
func (j *Job) QueueKey() string {
	// Apply name fixing to CircleID and Topic
	fixedCircleID := tools.NameFix(j.CircleID)
	fixedTopic := tools.NameFix(j.Topic)
	return fmt.Sprintf("heroqueue:%s:%s", fixedCircleID, fixedTopic)
}

// StorageKey returns the Redis storage key for this job
func (j *Job) StorageKey() string {
	// Apply name fixing to CircleID and Topic
	fixedCircleID := tools.NameFix(j.CircleID)
	fixedTopic := tools.NameFix(j.Topic)
	return fmt.Sprintf("herojobs:%s:%s:%d", fixedCircleID, fixedTopic, j.JobID)
}

// initOurDbClient initializes the OurDB client connection for the job
func (j *Job) initOurDbClient() error {
	if j.dbClient != nil {
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Create path as ~/hero/var/circles/$circleid/$topic/jobsdb
	dbPath := filepath.Join(homeDir, "hero", "var", "circles", j.CircleID, j.Topic, "jobsdb")

	// Ensure directory exists
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Initialize OurDB client
	client, err := ourdb.NewClient(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create OurDB client: %w", err)
	}

	j.dbClient = client
	return nil
}

// Save stores the job in OurDB using auto-incrementing JobID
func (j *Job) Save() error {
	// Ensure OurDB client is initialized
	if err := j.initOurDbClient(); err != nil {
		return err
	}

	// Convert job to JSON
	jobData, err := json.Marshal(j)
	if err != nil {
		return fmt.Errorf("failed to marshal job to JSON: %w", err)
	}

	// If JobID is 0, let OurDB assign an auto-incremented ID
	if j.JobID == 0 {
		// Use OurDB Add method to auto-generate ID
		id, err := j.dbClient.Add(jobData)
		if err != nil {
			return fmt.Errorf("failed to add job to OurDB: %w", err)
		}
		j.JobID = id
	} else {
		// Save to OurDB with specified ID
		if err := j.dbClient.Set(j.JobID, jobData); err != nil {
			return fmt.Errorf("failed to save job to OurDB: %w", err)
		}
	}

	return nil
}

// Load retrieves the job from OurDB using the JobID as the key
func (j *Job) Load() error {
	// Ensure OurDB client is initialized
	if err := j.initOurDbClient(); err != nil {
		return err
	}

	// Validate that JobID is not zero
	if j.JobID == 0 {
		return fmt.Errorf("cannot load job with ID 0: no ID specified")
	}

	// Get from OurDB using the numeric JobID directly
	jobData, err := j.dbClient.Get(j.JobID)
	if err != nil {
		return fmt.Errorf("failed to load job from OurDB: %w", err)
	}

	// Parse job data
	tempJob := &Job{}
	if err := json.Unmarshal(jobData, tempJob); err != nil {
		return fmt.Errorf("failed to unmarshal job data: %w", err)
	}

	// Copy values to current job (except the dbClient)
	dbClient := j.dbClient
	*j = *tempJob
	j.dbClient = dbClient

	return nil
}

// Finish completes the job, updating all properties and saving to OurDB
func (j *Job) Finish(status JobStatus, result string, err error) error {
	// Update job properties
	j.Status = status
	j.TimeEnd = time.Now().Unix()
	j.Result = result

	if err != nil {
		j.Error = err.Error()
	}

	// Save updated job to OurDB
	return j.Save()
}

// Close closes the OurDB client connection
func (j *Job) Close() error {
	if j.dbClient != nil {
		err := j.dbClient.Close()
		j.dbClient = nil
		return err
	}
	return nil
}
