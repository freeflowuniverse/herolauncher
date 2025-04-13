package jobsmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient handles Redis operations for jobs
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisClient creates a new Redis client
func NewRedisClient(addr string, isUnixSocket bool) (*RedisClient, error) {
	// Determine network type
	networkType := "tcp"
	if isUnixSocket {
		networkType = "unix"
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Network:      networkType,
		Addr:         addr,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	// Test connection
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		client: client,
		ctx:    ctx,
	}, nil
}

// Close closes the Redis client
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// StoreJob stores a job in Redis
func (r *RedisClient) StoreJob(job *Job) error {
	// Convert job to JSON
	jobJSON, err := job.ToJSON()
	if err != nil {
		return err
	}

	// Store job in Redis
	err = r.client.HSet(r.ctx, job.StorageKey(), "job", jobJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to store job: %w", err)
	}

	return nil
}

// GetJob retrieves a job from Redis by ID
func (r *RedisClient) GetJob(jobID string) (*Job, error) {
	storageKey := fmt.Sprintf("jobsmanager:%s", jobID)

	// Get job from Redis
	jobJSON, err := r.client.HGet(r.ctx, storageKey, "job").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("job not found: %s", jobID)
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Parse job JSON
	job, err := NewJobFromJSON(jobJSON)
	if err != nil {
		return nil, err
	}

	return job, nil
}

// DeleteJob deletes a job from Redis
func (r *RedisClient) DeleteJob(jobID string) error {
	storageKey := fmt.Sprintf("jobsmanager:%s", jobID)

	// Delete job from Redis
	err := r.client.Del(r.ctx, storageKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	return nil
}

// EnqueueJob adds a job to its queue
func (r *RedisClient) EnqueueJob(job *Job) error {
	// Store the job first
	err := r.StoreJob(job)
	if err != nil {
		return err
	}

	// Add job ID to queue
	err = r.client.RPush(r.ctx, job.QueueKey(), job.JobID).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	return nil
}

// QueueSize returns the size of a queue
func (r *RedisClient) QueueSize(circleID, topic string) (int64, error) {
	queueKey := fmt.Sprintf("heroqueue:%s:%s", circleID, topic)

	// Get queue size
	size, err := r.client.LLen(r.ctx, queueKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue size: %w", err)
	}

	return size, nil
}

// QueueEmpty empties a queue and deletes all corresponding jobs
func (r *RedisClient) QueueEmpty(circleID, topic string) error {
	queueKey := fmt.Sprintf("heroqueue:%s:%s", circleID, topic)

	// Get all job IDs from queue
	jobIDs, err := r.client.LRange(r.ctx, queueKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get job IDs from queue: %w", err)
	}

	// Delete all jobs
	for _, jobID := range jobIDs {
		storageKey := fmt.Sprintf("jobsmanager:%s", jobID)
		err := r.client.Del(r.ctx, storageKey).Err()
		if err != nil {
			return fmt.Errorf("failed to delete job %s: %w", jobID, err)
		}
	}

	// Empty queue
	err = r.client.Del(r.ctx, queueKey).Err()
	if err != nil {
		return fmt.Errorf("failed to empty queue: %w", err)
	}

	return nil
}

// QueueGet gets the last job from a queue without removing it
func (r *RedisClient) QueueGet(circleID, topic string) (*Job, error) {
	queueKey := fmt.Sprintf("heroqueue:%s:%s", circleID, topic)

	// Get last job ID from queue
	jobID, err := r.client.LIndex(r.ctx, queueKey, -1).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("queue is empty")
		}
		return nil, fmt.Errorf("failed to get job ID from queue: %w", err)
	}

	// Get job
	return r.GetJob(jobID)
}

// QueueFetch gets and removes the last job from a queue
func (r *RedisClient) QueueFetch(circleID, topic string) (*Job, error) {
	queueKey := fmt.Sprintf("heroqueue:%s:%s", circleID, topic)

	// Get and remove last job ID from queue
	jobID, err := r.client.RPop(r.ctx, queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("queue is empty")
		}
		return nil, fmt.Errorf("failed to fetch job ID from queue: %w", err)
	}

	// Get job
	return r.GetJob(jobID)
}

// ListJobs lists job IDs by circle and topic
func (r *RedisClient) ListJobs(circleID, topic string) ([]string, error) {
	var pattern string
	if circleID != "" && topic != "" {
		pattern = fmt.Sprintf("heroqueue:%s:%s", circleID, topic)
	} else if circleID != "" {
		pattern = fmt.Sprintf("heroqueue:%s:*", circleID)
	} else if topic != "" {
		pattern = fmt.Sprintf("heroqueue:*:%s", topic)
	} else {
		pattern = "heroqueue:*:*"
	}

	// Get all matching queue keys
	queueKeys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list queues: %w", err)
	}

	var jobIDs []string
	for _, queueKey := range queueKeys {
		// Get all job IDs from queue
		ids, err := r.client.LRange(r.ctx, queueKey, 0, -1).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get job IDs from queue %s: %w", queueKey, err)
		}
		jobIDs = append(jobIDs, ids...)
	}

	return jobIDs, nil
}

// UpdateJobStatus updates the status of a job
func (r *RedisClient) UpdateJobStatus(jobID string, status JobStatus) error {
	// Get job
	job, err := r.GetJob(jobID)
	if err != nil {
		return err
	}

	// Update status
	job.Status = status

	// Update timestamps based on status
	now := time.Now().Unix()
	if status == JobStatusActive && job.TimeStart == 0 {
		job.TimeStart = now
	} else if (status == JobStatusDone || status == JobStatusError) && job.TimeEnd == 0 {
		job.TimeEnd = now
	}

	// Store updated job
	return r.StoreJob(job)
}

// UpdateJobResult updates the result of a job
func (r *RedisClient) UpdateJobResult(jobID, result string) error {
	// Get job
	job, err := r.GetJob(jobID)
	if err != nil {
		return err
	}

	// Update result
	job.Result = result
	job.Status = JobStatusDone
	job.TimeEnd = time.Now().Unix()

	// Store updated job
	return r.StoreJob(job)
}

// UpdateJobError updates the error of a job
func (r *RedisClient) UpdateJobError(jobID, errorMsg string) error {
	// Get job
	job, err := r.GetJob(jobID)
	if err != nil {
		return err
	}

	// Update error
	job.Error = errorMsg
	job.Status = JobStatusError
	job.TimeEnd = time.Now().Unix()

	// Store updated job
	return r.StoreJob(job)
}
