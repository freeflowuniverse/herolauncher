package jobsmanager

import (
	"context"
	"fmt"
	"log"
	"time"
)

// processJob processes a job
func (d *WatchDog) processJob(job *Job) {
	// First, initialize OurDB connection and save the job
	if err := job.initOurDbClient(); err != nil {
		log.Printf("Error initializing OurDB client: %v", err)
	} else {
		// Set job status to active and save it to OurDB
		job.Status = JobStatusActive
		job.TimeStart = time.Now().Unix()

		if err := job.Save(); err != nil {
			log.Printf("Error saving job to OurDB: %v", err)
		}

		// Make sure to close the DB connection when we're done
		defer job.Close()
	}

	// Also update Redis for backward compatibility
	err := d.redisClient.UpdateJobStatus(job.JobID, JobStatusActive)
	if err != nil {
		log.Printf("Error updating job status in Redis: %v", err)
	}

	// Create a context with timeout
	timeout := time.Duration(job.Timeout) * time.Second
	if timeout == 0 {
		timeout = 60 * time.Second // Default timeout: 60 seconds
	}
	ctx, cancel := context.WithTimeout(d.ctx, timeout)
	defer cancel()

	// Process job based on parameters type
	switch job.ParamsType {
	case ParamsTypeHeroScript:
		d.processHeroScript(ctx, job)
	case ParamsTypeRhaiScript:
		d.processRhaiScript(ctx, job)
	case ParamsTypeOpenRPC:
		d.processOpenRPC(ctx, job)
	case ParamsTypeAI:
		d.processAI(ctx, job)
	default:
		// Unknown parameters type
		errMsg := fmt.Sprintf("Unknown parameters type: %s", job.ParamsType)
		log.Printf("Error processing job %s: %s", job.JobID, errMsg)

		// Update job status to error in both OurDB and Redis
		job.Finish(JobStatusError, "", fmt.Errorf(errMsg))

		if err := d.redisClient.UpdateJobError(job.JobID, errMsg); err != nil {
			log.Printf("Error updating job error in Redis: %v", err)
		}
	}
}

// processHeroScript processes a job with HeroScript parameters
func (d *WatchDog) processHeroScript(ctx context.Context, job *Job) {
	// Start a goroutine to process the job
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()

		// TODO: Implement HeroScript processing
		errMsg := "HeroScript processing not implemented yet"
		log.Printf("Error processing job %s: %s", job.JobID, errMsg)

		// Update job status in OurDB
		job.Finish(JobStatusError, "", fmt.Errorf(errMsg))

		// Also update Redis for backward compatibility
		if err := d.redisClient.UpdateJobError(job.JobID, errMsg); err != nil {
			log.Printf("Error updating job error in Redis: %v", err)
		}
	}()
}

// processRhaiScript processes a job with RhaiScript parameters
func (d *WatchDog) processRhaiScript(ctx context.Context, job *Job) {
	// Start a goroutine to process the job
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()

		// TODO: Implement RhaiScript processing
		errMsg := "RhaiScript processing not implemented yet"
		log.Printf("Error processing job %s: %s", job.JobID, errMsg)

		// Update job status in OurDB
		job.Finish(JobStatusError, "", fmt.Errorf(errMsg))

		// Also update Redis for backward compatibility
		if err := d.redisClient.UpdateJobError(job.JobID, errMsg); err != nil {
			log.Printf("Error updating job error in Redis: %v", err)
		}
	}()
}

// processOpenRPC processes a job with OpenRPC parameters
func (d *WatchDog) processOpenRPC(ctx context.Context, job *Job) {
	// Start a goroutine to process the job
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()

		// TODO: Implement OpenRPC processing
		errMsg := "OpenRPC processing not implemented yet"
		log.Printf("Error processing job %s: %s", job.JobID, errMsg)

		// Update job status in OurDB
		job.Finish(JobStatusError, "", fmt.Errorf(errMsg))

		// Also update Redis for backward compatibility
		if err := d.redisClient.UpdateJobError(job.JobID, errMsg); err != nil {
			log.Printf("Error updating job error in Redis: %v", err)
		}
	}()
}

// processAI processes a job with AI parameters
func (d *WatchDog) processAI(ctx context.Context, job *Job) {
	// Start a goroutine to process the job
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()

		// TODO: Implement AI processing
		errMsg := "AI processing not implemented yet"
		log.Printf("Error processing job %s: %s", job.JobID, errMsg)

		// Update job status in OurDB
		job.Finish(JobStatusError, "", fmt.Errorf(errMsg))

		// Also update Redis for backward compatibility
		if err := d.redisClient.UpdateJobError(job.JobID, errMsg); err != nil {
			log.Printf("Error updating job error in Redis: %v", err)
		}
	}()
}
