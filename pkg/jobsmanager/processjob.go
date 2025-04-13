package jobsmanager

import (
	"context"
	"log"
	"time"
)

// processJob processes a job
func (d *WatchDog) processJob(job *Job) {
	// Update job status to active
	err := d.redisClient.UpdateJobStatus(job.JobID, JobStatusActive)
	if err != nil {
		log.Printf("Error updating job status: %v", err)
		return
	}

	// If job has a HeroScript, process it
	if job.HeroScript != "" {
		// Create a context with timeout
		timeout := time.Duration(job.Timeout) * time.Second
		if timeout == 0 {
			timeout = 60 * time.Second // Default timeout: 60 seconds
		}
		_, cancel := context.WithTimeout(d.ctx, timeout)

		if err := d.redisClient.UpdateJobError(job.JobID, "Heroscript processing not implemented"); err != nil {
			log.Printf("Error updating job error: %v", err)
		}

		// Start a goroutine to process the job
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			defer cancel()

			// // Register the job processor
			// d.registerJobProcessor(job.JobID, job.CircleID, job.Topic, func(j *Job) (string, error) {
			// 	// Initialize hero handler
			// 	if err := herohandler.Init(); err != nil {
			// 		return "", fmt.Errorf("failed to initialize hero handler: %w", err)
			// 	}

			// 	// Process the HeroScript
			// 	result, err := herohandler.DefaultInstance.GetFactory().ProcessHeroscript(job.HeroScript)
			// 	if err != nil {
			// 		return "", err
			// 	}
			// 	return result, nil
			// })

			// // Process the job
			// result, err := d.executeJobWithTimeout(ctx, job)

			// // Update job status based on result
			// if err != nil {
			// 	log.Printf("Error processing job %s: %v", job.JobID, err)
			// 	if err := d.redisClient.UpdateJobError(job.JobID, err.Error()); err != nil {
			// 		log.Printf("Error updating job error: %v", err)
			// 	}
			// } else {
			// 	if err := d.redisClient.UpdateJobResult(job.JobID, result); err != nil {
			// 		log.Printf("Error updating job result: %v", err)
			// 	}
			// }
			if err := d.redisClient.UpdateJobError(job.JobID, "RhaiScript processing not implemented"); err != nil {
				log.Printf("Error updating job error: %v", err)
			}			

			
	} else if job.RhaiScript != "" {
		// Process RhaiScript (not implemented in this version)
		log.Printf("RhaiScript processing not implemented yet")
		if err := d.redisClient.UpdateJobError(job.JobID, "RhaiScript processing not implemented"); err != nil {
			log.Printf("Error updating job error: %v", err)
		}
	} else {
		// No script to process
		if err := d.redisClient.UpdateJobError(job.JobID, "No script provided"); err != nil {
			log.Printf("Error updating job error: %v", err)
		}
	}
}
