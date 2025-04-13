package jobsmanager

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
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
	JobID         string     `json:"jobid"`
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
}

// NewJob creates a new job with default values
func NewJob() *Job {
	now := time.Now().Unix()
	return &Job{
		JobID:         uuid.New().String(),
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
	if job.JobID == "" {
		job.JobID = uuid.New().String()
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
	return fmt.Sprintf("heroqueue:%s:%s", j.CircleID, j.Topic)
}

// StorageKey returns the Redis storage key for this job
func (j *Job) StorageKey() string {
	return fmt.Sprintf("herojobs:%s", j.JobID)
}
