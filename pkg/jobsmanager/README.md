# HeroJobs Package

The HeroJobs package provides a robust job queue and processing system for the HeroLauncher project. It allows for asynchronous execution of scripts and API calls through a persistence and queue system.

## Overview

HeroJobs uses a dual storage system combining Redis and OurDB. Redis handles live queue management and provides fast access to running jobs, while OurDB offers persistent storage with auto-incrementing IDs. Jobs can be submitted to topic-specific queues within a circle, processed by a watchdog, and their results can be retrieved asynchronously.

## Components

### Job

A `Job` represents a unit of work to be processed. Each job has:

- **JobID**: Unique identifier for the job (auto-incremented by OurDB)
- **SessionKey**: Session identifier for authentication
- **CircleID**: Circle identifier for organization
- **Topic**: Topic for categorization
- **Params**: The parameters to be executed (script or API call)
- **ParamsType**: Type of parameters (HeroScript, RhaiScript, OpenRPC, AI)
- **Status**: Current status (new, active, error, done)
- **Timeout**: Maximum execution time in seconds
- **TimeScheduled/TimeStart/TimeEnd**: Timestamps for job lifecycle
- **Error/Result**: Error message or result data
- **Log**: Whether to enable logging for this job

### Storage

#### Dual Storage System

Jobs are stored in both Redis and OurDB:

#### Redis
- Handles all queue operations (adding/removing jobs)
- Stores all running jobs for fast access
- Used for real-time operations and status updates
- **Job Storage**: `herojobs:<circleID>:<topic>:<jobID>` or legacy `jobsmanager:<jobID>`
- **Queue**: `heroqueue:<circleID>:<topic>`

#### OurDB
- Provides persistent storage with auto-incrementing IDs
- Stores job details in a file-based database
- Path: `~/hero/var/circles/<circleid>/<topic>/jobsdb`
- Used for long-term storage and job history

### WatchDog

The `WatchDog` processes jobs from queues in the background:

- Runs as a goroutine that continuously checks all Redis queues
- When a job is found, it's processed in a separate goroutine
- Monitors execution time and enforces timeouts
- Updates job status and results in both OurDB and Redis
- Routes jobs to appropriate processors based on ParamsType (HeroScript, RhaiScript, OpenRPC, AI)

## Job Processing Flow

1. **Job Creation**: A job is created with parameters and saved to OurDB to get a unique ID
2. **Enqueuing**: The job is stored in Redis and its ID is added to its Redis queue
3. **Processing**: The watchdog fetches the job ID from the Redis queue and loads the job
4. **Execution**: The job is processed based on its ParamsType
5. **Concurrent Updates**: The job status and result are updated in both OurDB and Redis during processing
6. **Completion**: Final status and results are stored in both systems

## Timeout Handling

The watchdog implements timeout handling to prevent jobs from running indefinitely:

- Each job has a configurable timeout (default: 60 seconds)
- If a job exceeds its timeout, it's terminated and marked as error
- Timeouts are enforced using Go's context package

## Concurrency Management

The watchdog uses Go's concurrency primitives to safely manage multiple jobs:

- Each job is processed in a separate goroutine
- A wait group tracks all running goroutines
- A mutex protects access to shared data structures
- Context cancellation provides clean shutdown

## Usage Examples

### Starting the WatchDog

```go
// Initialize Redis client
redisClient, err := jobsmanager.NewRedisClient("localhost:6379", false)
if err != nil {
    log.Fatalf("Failed to connect to Redis: %v", err)
}
defer redisClient.Close()

// Create and start watchdog
watchdog := jobsmanager.NewWatchDog(redisClient)
watchdog.Start()

// Handle shutdown
defer watchdog.Stop()
```

### Submitting a Job

```go
// Create a new job
job := jobsmanager.NewJob()
job.CircleID = "myCircle"
job.Topic = "myTopic"
job.Params = `
!!fake.return_success
    message: "This is a test job"
`
job.ParamsType = jobsmanager.ParamsTypeHeroScript
job.Timeout = 30 // 30 seconds timeout

// Save the job to OurDB to get an ID
if err := job.Save(); err != nil {
    log.Printf("Failed to save job: %v", err)
    return
}

// Store and enqueue the job in Redis
if err := redisClient.StoreJob(job); err != nil {
    log.Printf("Failed to store job in Redis: %v", err)
    return
}
if err := redisClient.EnqueueJob(job); err != nil {
    log.Printf("Failed to enqueue job: %v", err)
    return
}

// Remember the job ID for later retrieval
jobID := job.JobID
```

### Checking Job Status

```go
// For currently running jobs, use Redis for fastest access
job, err := redisClient.GetJob(jobID)
if err != nil {
    // If not found in Redis, try OurDB for historical jobs
    job = &jobsmanager.Job{JobID: jobID}
    if err := job.Load(); err != nil {
        log.Printf("Failed to load job: %v", err)
        return
    }
}

// Check job status
switch job.Status {
case jobsmanager.JobStatusNew:
    fmt.Println("Job is waiting to be processed")
case jobsmanager.JobStatusActive:
    fmt.Println("Job is currently being processed")
case jobsmanager.JobStatusDone:
    fmt.Printf("Job completed successfully: %s\n", job.Result)
case jobsmanager.JobStatusError:
    fmt.Printf("Job failed: %s\n", job.Error)
}
```

## Processing Implementation

The watchdog supports different types of job parameters processing:

### HeroScript Processing
Jobs with ParamsType = ParamsTypeHeroScript are processed by passing the script to a HeroScript handler.

### RhaiScript Processing
Jobs with ParamsType = ParamsTypeRhaiScript are processed by the Rhai scripting engine.

### OpenRPC Processing
Jobs with ParamsType = ParamsTypeOpenRPC are processed by passing the parameters to an OpenRPC handler.

### AI Processing
Jobs with ParamsType = ParamsTypeAI are processed by an AI system (experimental).

This modular design allows for extensible script and API call processing through appropriate handlers.
