use serde::{Deserialize, Serialize};
use std::io::{BufRead, BufReader, Write};
use std::os::unix::net::UnixStream;
use std::path::Path;
use thiserror::Error;
use uuid::Uuid;

/// Job status enum
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "lowercase")]
pub enum JobStatus {
    New,
    Active,
    Error,
    Done,
}

impl Default for JobStatus {
    fn default() -> Self {
        JobStatus::New
    }
}

/// Job struct representing a job to be processed
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Job {
    pub jobid: String,
    pub sessionkey: String,
    pub circleid: String,
    pub topic: String,
    pub heroscript: String,
    pub rhaiscript: String,
    pub timeout: i64,
    pub status: JobStatus,
    pub time_scheduled: i64,
    pub time_start: i64,
    pub time_end: i64,
    pub error: String,
    pub result: String,
}

impl Default for Job {
    fn default() -> Self {
        let now = chrono::Utc::now().timestamp();
        Self {
            jobid: Uuid::new_v4().to_string(),
            sessionkey: String::new(),
            circleid: String::new(),
            topic: "default".to_string(),
            heroscript: String::new(),
            rhaiscript: String::new(),
            timeout: 0,
            status: JobStatus::default(),
            time_scheduled: now,
            time_start: 0,
            time_end: 0,
            error: String::new(),
            result: String::new(),
        }
    }
}

/// Command types for the HeroJobs server
pub enum Command {
    Put,
    Get,
    Delete,
    List,
    QueueSize,
    QueueEmpty,
    QueueGet,
    QueueFetch,
}

impl Command {
    pub fn as_str(&self) -> &'static str {
        match self {
            Command::Put => "PUT",
            Command::Get => "GET",
            Command::Delete => "DELETE",
            Command::List => "LIST",
            Command::QueueSize => "QUEUESIZE",
            Command::QueueEmpty => "QUEUEEMPTY",
            Command::QueueGet => "QUEUEGET",
            Command::QueueFetch => "QUEUEFETCH",
        }
    }
}

/// Error type for HeroJobs client
#[derive(Error, Debug)]
pub enum HeroJobsError {
    #[error("IO error: {0}")]
    IoError(#[from] std::io::Error),

    #[error("JSON error: {0}")]
    JsonError(#[from] serde_json::Error),

    #[error("Server error: {0}")]
    ServerError(String),

    #[error("Not connected to server")]
    NotConnected,

    #[error("Invalid response: {0}")]
    InvalidResponse(String),
}

/// Result type for HeroJobs client
pub type Result<T> = std::result::Result<T, HeroJobsError>;

/// HeroJobs client for interacting with the HeroJobs server
pub struct HeroJobsClient {
    socket_path: String,
    stream: Option<UnixStream>,
}

impl HeroJobsClient {
    /// Create a new HeroJobs client
    pub fn new<P: AsRef<Path>>(socket_path: P) -> Self {
        Self {
            socket_path: socket_path.as_ref().to_string_lossy().to_string(),
            stream: None,
        }
    }

    /// Connect to the HeroJobs server
    pub fn connect(&mut self) -> Result<()> {
        let stream = UnixStream::connect(&self.socket_path)?;
        self.stream = Some(stream);
        Ok(())
    }

    /// Check if the client is connected
    pub fn is_connected(&self) -> bool {
        self.stream.is_some()
    }

    /// Send a command to the server
    fn send_command<T: Serialize>(&mut self, command: Command, data: Option<&T>) -> Result<String> {
        let stream = match &mut self.stream {
            Some(stream) => stream,
            None => return Err(HeroJobsError::NotConnected),
        };

        // Send command
        writeln!(stream, "{}", command.as_str())?;

        // Send data as JSON if provided
        if let Some(data) = data {
            let json_data = serde_json::to_string(data)?;
            writeln!(stream, "{}", json_data)?;
        }

        // Send empty line to mark end of data
        writeln!(stream, "")?;

        // Read response
        let mut reader = BufReader::new(stream);
        let mut response = String::new();
        reader.read_line(&mut response)?;

        Ok(response.trim().to_string())
    }

    /// Create and submit a job
    pub fn create_job(
        &mut self,
        circle_id: &str,
        topic: &str,
        session_key: &str,
        hero_script: &str,
        rhai_script: &str,
    ) -> Result<Job> {
        let mut job = Job::default();
        job.circleid = circle_id.to_string();
        job.topic = topic.to_string();
        job.sessionkey = session_key.to_string();
        job.heroscript = hero_script.to_string();
        job.rhaiscript = rhai_script.to_string();

        self.submit_job(&job)
    }

    /// Submit a job to the server
    pub fn submit_job(&mut self, job: &Job) -> Result<Job> {
        let response = self.send_command(Command::Put, Some(job))?;
        let job: Job = serde_json::from_str(&response).map_err(|e| {
            HeroJobsError::InvalidResponse(format!("Failed to parse job response: {}", e))
        })?;
        Ok(job)
    }

    /// Get a job by ID
    pub fn get_job(&mut self, job_id: &str) -> Result<Job> {
        let data = serde_json::json!({ "jobid": job_id });
        let response = self.send_command(Command::Get, Some(&data))?;
        let job: Job = serde_json::from_str(&response).map_err(|e| {
            HeroJobsError::InvalidResponse(format!("Failed to parse job response: {}", e))
        })?;
        Ok(job)
    }

    /// Delete a job by ID
    pub fn delete_job(&mut self, job_id: &str) -> Result<()> {
        let data = serde_json::json!({ "jobid": job_id });
        let response = self.send_command(Command::Delete, Some(&data))?;
        
        let response: serde_json::Value = serde_json::from_str(&response).map_err(|e| {
            HeroJobsError::InvalidResponse(format!("Failed to parse delete response: {}", e))
        })?;
        
        if let Some(status) = response.get("status").and_then(|s| s.as_str()) {
            if status == "success" {
                return Ok(());
            }
        }
        
        if let Some(error) = response.get("error").and_then(|e| e.as_str()) {
            return Err(HeroJobsError::ServerError(error.to_string()));
        }
        
        Err(HeroJobsError::InvalidResponse(format!(
            "Unexpected response: {}",
            response
        )))
    }

    /// List jobs by circle and topic
    pub fn list_jobs(&mut self, circle_id: Option<&str>, topic: Option<&str>) -> Result<Vec<String>> {
        let data = serde_json::json!({
            "circleid": circle_id.unwrap_or(""),
            "topic": topic.unwrap_or("")
        });
        
        let response = self.send_command(Command::List, Some(&data))?;
        let response: serde_json::Value = serde_json::from_str(&response).map_err(|e| {
            HeroJobsError::InvalidResponse(format!("Failed to parse list response: {}", e))
        })?;
        
        if let Some(jobs) = response.get("jobs").and_then(|j| j.as_array()) {
            let job_ids = jobs
                .iter()
                .filter_map(|j| j.as_str().map(|s| s.to_string()))
                .collect();
            return Ok(job_ids);
        }
        
        if let Some(error) = response.get("error").and_then(|e| e.as_str()) {
            return Err(HeroJobsError::ServerError(error.to_string()));
        }
        
        Ok(vec![])
    }

    /// Get queue size
    pub fn queue_size(&mut self, circle_id: &str, topic: &str) -> Result<i64> {
        let data = serde_json::json!({
            "circleid": circle_id,
            "topic": topic
        });
        
        let response = self.send_command(Command::QueueSize, Some(&data))?;
        let response: serde_json::Value = serde_json::from_str(&response).map_err(|e| {
            HeroJobsError::InvalidResponse(format!("Failed to parse queue size response: {}", e))
        })?;
        
        if let Some(size) = response.get("size").and_then(|s| s.as_i64()) {
            return Ok(size);
        }
        
        if let Some(error) = response.get("error").and_then(|e| e.as_str()) {
            return Err(HeroJobsError::ServerError(error.to_string()));
        }
        
        Err(HeroJobsError::InvalidResponse(format!(
            "Unexpected response: {}",
            response
        )))
    }

    /// Empty a queue
    pub fn queue_empty(&mut self, circle_id: &str, topic: &str) -> Result<()> {
        let data = serde_json::json!({
            "circleid": circle_id,
            "topic": topic
        });
        
        let response = self.send_command(Command::QueueEmpty, Some(&data))?;
        let response: serde_json::Value = serde_json::from_str(&response).map_err(|e| {
            HeroJobsError::InvalidResponse(format!("Failed to parse queue empty response: {}", e))
        })?;
        
        if let Some(status) = response.get("status").and_then(|s| s.as_str()) {
            if status == "success" {
                return Ok(());
            }
        }
        
        if let Some(error) = response.get("error").and_then(|e| e.as_str()) {
            return Err(HeroJobsError::ServerError(error.to_string()));
        }
        
        Err(HeroJobsError::InvalidResponse(format!(
            "Unexpected response: {}",
            response
        )))
    }

    /// Get a job from a queue without removing it
    pub fn queue_get(&mut self, circle_id: &str, topic: &str) -> Result<Job> {
        let data = serde_json::json!({
            "circleid": circle_id,
            "topic": topic
        });
        
        let response = self.send_command(Command::QueueGet, Some(&data))?;
        let job: Job = serde_json::from_str(&response).map_err(|e| {
            HeroJobsError::InvalidResponse(format!("Failed to parse job response: {}", e))
        })?;
        
        Ok(job)
    }

    /// Get and remove a job from a queue
    pub fn queue_fetch(&mut self, circle_id: &str, topic: &str) -> Result<Job> {
        let data = serde_json::json!({
            "circleid": circle_id,
            "topic": topic
        });
        
        let response = self.send_command(Command::QueueFetch, Some(&data))?;
        let job: Job = serde_json::from_str(&response).map_err(|e| {
            HeroJobsError::InvalidResponse(format!("Failed to parse job response: {}", e))
        })?;
        
        Ok(job)
    }
}

impl Drop for HeroJobsClient {
    fn drop(&mut self) {
        if let Some(stream) = &self.stream {
            let _ = stream.shutdown(std::net::Shutdown::Both);
        }
    }
}
