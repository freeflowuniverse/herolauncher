package jobsmanager

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// JobProcessor is a function type that processes a job
// WatchDog handles the processing of jobs from queues
type WatchDog struct {
	redisClient    *RedisClient
	processorMutex sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// NewWatchDog creates a new watchdog
func NewWatchDog(redisClient *RedisClient) *WatchDog {
	ctx, cancel := context.WithCancel(context.Background())
	return &WatchDog{
		redisClient: redisClient,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start starts the watchdog
func (d *WatchDog) Start() {
	d.wg.Add(1)
	go d.processQueues()
}

// Stop stops the watchdog
func (d *WatchDog) Stop() {
	d.cancel()
	d.wg.Wait()
}

// processQueues processes all queues
func (d *WatchDog) processQueues() {
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
func (d *WatchDog) getQueues() ([]queueInfo, error) {
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
