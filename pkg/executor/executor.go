package executor

import (
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// Job represents a command execution job
type Job struct {
	ID        string
	Command   string
	Args      []string
	StartTime time.Time
	EndTime   time.Time
	Status    string
	Output    string
	Error     string
}

// Executor handles command execution
type Executor struct {
	jobs     map[string]*Job
	jobsLock sync.RWMutex
}

// NewExecutor creates a new executor instance
func NewExecutor() *Executor {
	return &Executor{
		jobs: make(map[string]*Job),
	}
}

// ExecuteCommand runs a command and returns a job ID
func (e *Executor) ExecuteCommand(command string, args []string) (string, error) {
	job := &Job{
		ID:        fmt.Sprintf("job-%d", time.Now().UnixNano()),
		Command:   command,
		Args:      args,
		StartTime: time.Now(),
		Status:    "running",
	}

	e.jobsLock.Lock()
	e.jobs[job.ID] = job
	e.jobsLock.Unlock()

	go func() {
		cmd := exec.Command(command, args...)
		output, err := cmd.CombinedOutput()
		
		e.jobsLock.Lock()
		defer e.jobsLock.Unlock()
		
		job.EndTime = time.Now()
		job.Output = string(output)
		
		if err != nil {
			job.Status = "failed"
			job.Error = err.Error()
		} else {
			job.Status = "completed"
		}
	}()

	return job.ID, nil
}

// GetJob returns information about a specific job
func (e *Executor) GetJob(jobID string) (*Job, error) {
	e.jobsLock.RLock()
	defer e.jobsLock.RUnlock()
	
	job, exists := e.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}
	
	return job, nil
}

// ListJobs returns all jobs
func (e *Executor) ListJobs() []*Job {
	e.jobsLock.RLock()
	defer e.jobsLock.RUnlock()
	
	jobs := make([]*Job, 0, len(e.jobs))
	for _, job := range e.jobs {
		jobs = append(jobs, job)
	}
	
	return jobs
}
