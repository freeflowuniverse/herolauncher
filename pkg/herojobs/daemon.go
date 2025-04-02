package herojobs

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory/herohandler"
)

// JobProcessor is a function type that processes a job
type JobProcessor func(job *Job) (string, error)

// Daemon handles the processing of jobs from queues
type Daemon struct {
	redisClient    *RedisClient
	jobProcessors  map[string]map[string]map[string]JobProcessor // map[circleID]map[topic]map[jobID]processor
	processorMutex sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// NewDaemon creates a new daemon
func NewDaemon(redisClient *RedisClient) *Daemon {
	ctx, cancel := context.WithCancel(context.Background())
	return &Daemon{
		redisClient:   redisClient,
		jobProcessors: make(map[string]map[string]map[string]JobProcessor),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start starts the daemon
func (d *Daemon) Start() {
	d.wg.Add(1)
	go d.processQueues()
}

// Stop stops the daemon
func (d *Daemon) Stop() {
	d.cancel()
	d.wg.Wait()
}

// processQueues processes all queues
func (d *Daemon) processQueues() {
	defer d.wg.Done()

	for {
		select {
		case <-d.ctx.Done():
			return
		default:
			// Get all queues
			queues, err := d.getQueues()
			if err != nil {
				log.Printf("Error getting queues: %v", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Process each queue
			processed := false
			for _, queue := range queues {
				circleID, topic := queue.circleID, queue.topic
				job, err := d.redisClient.QueueFetch(circleID, topic)
				if err != nil {
					// Queue is empty, continue to next queue
					continue
				}

				// Process the job
				processed = true
				d.processJob(job)
			}

			// If no jobs were processed, wait before checking again
			if !processed {
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// queueInfo represents a queue
type queueInfo struct {
	circleID string
	topic    string
}

// getQueues returns all queues
func (d *Daemon) getQueues() ([]queueInfo, error) {
	// Get all queue keys from Redis
	queueKeys, err := d.redisClient.client.Keys(d.redisClient.ctx, "heroqueue:*:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list queues: %w", err)
	}

	var queues []queueInfo
	for _, queueKey := range queueKeys {
		// Parse queue key (format: heroqueue:<circleID>:<topic>)
		// Split the key by ":"
		parts := strings.Split(queueKey, ":")
		if len(parts) != 3 {
			log.Printf("Invalid queue key format: %s", queueKey)
			continue
		}

		queues = append(queues, queueInfo{
			circleID: parts[1],
			topic:    parts[2],
		})
	}

	return queues, nil
}

// processJob processes a job
func (d *Daemon) processJob(job *Job) {
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
		ctx, cancel := context.WithTimeout(d.ctx, timeout)

		// Start a goroutine to process the job
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			defer cancel()

			// Register the job processor
			d.registerJobProcessor(job.JobID, job.CircleID, job.Topic, func(j *Job) (string, error) {
				// Initialize hero handler
				if err := herohandler.Init(); err != nil {
					return "", fmt.Errorf("failed to initialize hero handler: %w", err)
				}

				// Process the HeroScript
				result, err := herohandler.DefaultInstance.GetFactory().ProcessHeroscript(job.HeroScript)
				if err != nil {
					return "", err
				}
				return result, nil
			})

			// Process the job
			result, err := d.executeJobWithTimeout(ctx, job)

			// Update job status based on result
			if err != nil {
				log.Printf("Error processing job %s: %v", job.JobID, err)
				if err := d.redisClient.UpdateJobError(job.JobID, err.Error()); err != nil {
					log.Printf("Error updating job error: %v", err)
				}
			} else {
				if err := d.redisClient.UpdateJobResult(job.JobID, result); err != nil {
					log.Printf("Error updating job result: %v", err)
				}
			}

			// Unregister the job processor
			d.unregisterJobProcessor(job.JobID, job.CircleID, job.Topic)
		}()
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

// executeJobWithTimeout executes a job with a timeout
func (d *Daemon) executeJobWithTimeout(ctx context.Context, job *Job) (string, error) {
	resultCh := make(chan struct {
		result string
		err    error
	})

	// Get the job processor
	processor, err := d.getJobProcessor(job.JobID, job.CircleID, job.Topic)
	if err != nil {
		return "", err
	}

	// Execute the job in a goroutine
	go func() {
		result, err := processor(job)
		resultCh <- struct {
			result string
			err    error
		}{result, err}
	}()

	// Wait for the job to complete or timeout
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("job execution timed out after %d seconds", job.Timeout)
	case res := <-resultCh:
		return res.result, res.err
	}
}

// registerJobProcessor registers a job processor
func (d *Daemon) registerJobProcessor(jobID, circleID, topic string, processor JobProcessor) {
	d.processorMutex.Lock()
	defer d.processorMutex.Unlock()

	// Initialize maps if they don't exist
	if _, ok := d.jobProcessors[circleID]; !ok {
		d.jobProcessors[circleID] = make(map[string]map[string]JobProcessor)
	}
	if _, ok := d.jobProcessors[circleID][topic]; !ok {
		d.jobProcessors[circleID][topic] = make(map[string]JobProcessor)
	}

	// Register the processor
	d.jobProcessors[circleID][topic][jobID] = processor
}

// unregisterJobProcessor unregisters a job processor
func (d *Daemon) unregisterJobProcessor(jobID, circleID, topic string) {
	d.processorMutex.Lock()
	defer d.processorMutex.Unlock()

	// Check if maps exist
	if _, ok := d.jobProcessors[circleID]; !ok {
		return
	}
	if _, ok := d.jobProcessors[circleID][topic]; !ok {
		return
	}

	// Unregister the processor
	delete(d.jobProcessors[circleID][topic], jobID)

	// Clean up empty maps
	if len(d.jobProcessors[circleID][topic]) == 0 {
		delete(d.jobProcessors[circleID], topic)
	}
	if len(d.jobProcessors[circleID]) == 0 {
		delete(d.jobProcessors, circleID)
	}
}

// getJobProcessor gets a job processor
func (d *Daemon) getJobProcessor(jobID, circleID, topic string) (JobProcessor, error) {
	d.processorMutex.RLock()
	defer d.processorMutex.RUnlock()

	// Check if maps exist
	if _, ok := d.jobProcessors[circleID]; !ok {
		return nil, fmt.Errorf("no processors for circle %s", circleID)
	}
	if _, ok := d.jobProcessors[circleID][topic]; !ok {
		return nil, fmt.Errorf("no processors for topic %s in circle %s", topic, circleID)
	}

	// Get the processor
	processor, ok := d.jobProcessors[circleID][topic][jobID]
	if !ok {
		return nil, fmt.Errorf("no processor for job %s in topic %s, circle %s", jobID, topic, circleID)
	}

	return processor, nil
}

// TestWithFakeHandler tests the daemon with a fake handler
func TestWithFakeHandler(redisClient *RedisClient) error {
	// Initialize hero handler
	if err := herohandler.Init(); err != nil {
		return fmt.Errorf("failed to initialize hero handler: %w", err)
	}

	// Create a test job
	job := NewJob()
	job.CircleID = "test"
	job.Topic = "test"
	job.HeroScript = `
!!fake.return_success
	message: "Test job processed successfully"
`

	// Enqueue the job
	if err := redisClient.EnqueueJob(job); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	// Create and start the daemon
	daemon := NewDaemon(redisClient)
	daemon.Start()

	// Wait for the job to be processed
	time.Sleep(1 * time.Second)

	// Stop the daemon
	daemon.Stop()

	// Check the job status
	processedJob, err := redisClient.GetJob(job.JobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if processedJob.Status != JobStatusDone {
		return fmt.Errorf("job status is %s, expected %s", processedJob.Status, JobStatusDone)
	}

	log.Printf("Test job processed successfully: %s", processedJob.Result)
	return nil
}
