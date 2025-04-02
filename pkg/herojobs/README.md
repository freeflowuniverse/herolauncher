# HeroJobs Package

The HeroJobs package provides a robust job queue and processing system for the HeroLauncher project. It allows for asynchronous execution of HeroScripts and RhaiScripts through a Redis-backed queue system.

## Overview

HeroJobs uses Redis as a backend for storing jobs and managing queues. Jobs can be submitted to topic-specific queues within a circle, processed by a daemon, and their results can be retrieved asynchronously.

## Components

### Job

A `Job` represents a unit of work to be processed. Each job has:

- **JobID**: Unique identifier for the job
- **SessionKey**: Session identifier for authentication
- **CircleID**: Circle identifier for organization
- **Topic**: Topic for categorization
- **HeroScript/RhaiScript**: The script to be executed
- **Status**: Current status (new, active, error, done)
- **Timeout**: Maximum execution time in seconds
- **TimeScheduled/TimeStart/TimeEnd**: Timestamps for job lifecycle
- **Error/Result**: Error message or result data

### RedisClient

The `RedisClient` handles all Redis operations for job storage and queue management:

- **StoreJob**: Stores a job in Redis
- **GetJob**: Retrieves a job by ID
- **DeleteJob**: Deletes a job
- **EnqueueJob**: Adds a job to its queue
- **QueueFetch**: Gets and removes the first job from a queue
- **UpdateJobStatus/Result/Error**: Updates job properties

### Daemon

The `Daemon` processes jobs from queues in the background:

- Runs as a goroutine that continuously checks all queues
- When a job is found, it's processed in a separate goroutine
- Monitors execution time and enforces timeouts
- Updates job status and results
- Handles HeroScripts by passing them to the HeroHandler

## Queue Structure

Jobs are organized in Redis using the following key patterns:

- **Job Storage**: `herojobs:<jobID>`
- **Queue**: `heroqueue:<circleID>:<topic>`

This allows for efficient retrieval of jobs by ID and processing of jobs by circle and topic.

## Job Processing Flow

1. **Job Creation**: A job is created with a unique ID and script
2. **Enqueuing**: The job is stored in Redis and added to its queue
3. **Processing**: The daemon fetches the job from the queue and processes it
4. **Execution**: The script is executed by the appropriate handler
5. **Completion**: The job status and result are updated in Redis

## Timeout Handling

The daemon implements timeout handling to prevent jobs from running indefinitely:

- Each job has a configurable timeout (default: 60 seconds)
- If a job exceeds its timeout, it's terminated and marked as error
- Timeouts are enforced using Go's context package

## Concurrency Management

The daemon uses Go's concurrency primitives to safely manage multiple jobs:

- Each job is processed in a separate goroutine
- A wait group tracks all running goroutines
- A mutex protects access to shared data structures
- Context cancellation provides clean shutdown

## Testing

The package includes a test utility that demonstrates the functionality:

- **TestWithFakeHandler**: Tests the daemon with a fake handler
- **cmd/daemontest**: Command-line tool for testing the daemon

## Usage Examples

### Starting the Daemon

```go
// Initialize Redis client
redisClient, err := herojobs.NewRedisClient("localhost:6379", false)
if err != nil {
    log.Fatalf("Failed to connect to Redis: %v", err)
}
defer redisClient.Close()

// Create and start daemon
daemon := herojobs.NewDaemon(redisClient)
daemon.Start()

// Handle shutdown
defer daemon.Stop()
```

### Submitting a Job

```go
// Create a new job
job := herojobs.NewJob()
job.CircleID = "myCircle"
job.Topic = "myTopic"
job.HeroScript = `
!!fake.return_success
    message: "This is a test job"
`
job.Timeout = 30 // 30 seconds timeout

// Enqueue the job
if err := redisClient.EnqueueJob(job); err != nil {
    log.Printf("Failed to enqueue job: %v", err)
    return
}

// Remember the job ID for later retrieval
jobID := job.JobID
```

### Checking Job Status

```go
// Get the job by ID
job, err := redisClient.GetJob(jobID)
if err != nil {
    log.Printf("Failed to get job: %v", err)
    return
}

// Check job status
switch job.Status {
case herojobs.JobStatusNew:
    fmt.Println("Job is waiting to be processed")
case herojobs.JobStatusActive:
    fmt.Println("Job is currently being processed")
case herojobs.JobStatusDone:
    fmt.Printf("Job completed successfully: %s\n", job.Result)
case herojobs.JobStatusError:
    fmt.Printf("Job failed: %s\n", job.Error)
}
```

## Integration with HeroHandler

The daemon integrates with the HeroHandler to process HeroScripts:

1. The HeroHandler is initialized
2. The script is passed to the handler factory
3. The factory routes the script to the appropriate actor handler
4. The result is captured and stored in the job

This allows for extensible script processing through the handler system.
