use serde::{Deserialize, Serialize};
use std::time::Duration;

use crate::{Client, Result};

/// Information about a process
#[derive(Debug, Serialize, Deserialize, Default)]
pub struct ProcessInfo {
    #[serde(default)]
    pub name: String,
    #[serde(default)]
    pub command: String,
    #[serde(default)]
    pub status: String,
    #[serde(default)]
    pub pid: i32,
    #[serde(default)]
    pub start_time: String,
    #[serde(default)]
    pub uptime: String,
    #[serde(default)]
    pub cpu: String,
    #[serde(default)]
    pub memory: String,
    #[serde(default)]
    pub cron: Option<String>,
    #[serde(default)]
    pub job_id: Option<String>,
}

/// Client for the process manager
pub struct ProcessManagerClient {
    client: Client,
}

impl ProcessManagerClient {
    /// Create a new process manager client
    pub fn new(socket_path: &str) -> Self {
        Self {
            client: Client::new(socket_path),
        }
    }
    
    /// Set the connection timeout
    pub fn with_timeout(mut self, timeout: Duration) -> Self {
        self.client = self.client.with_timeout(timeout);
        self
    }
    
    /// Set the authentication secret
    pub fn with_secret(mut self, secret: &str) -> Self {
        self.client = self.client.with_secret(secret);
        self
    }
    
    /// Connect to the server
    pub fn connect(&mut self) -> Result<()> {
        self.client.connect()
    }
    
    /// Close the connection
    pub fn close(&mut self) -> Result<()> {
        self.client.close()
    }

    /// Start a new process
    pub fn start(&mut self, name: &str, command: &str, log_enabled: bool, deadline: Option<i32>, cron: Option<&str>, job_id: Option<&str>) -> Result<String> {
        let mut script = format!("!!process.start name:'{}' command:'{}' log:{}", name, command, log_enabled);
        
        if let Some(deadline_val) = deadline {
            script.push_str(&format!(" deadline:{}", deadline_val));
        }
        
        if let Some(cron_val) = cron {
            script.push_str(&format!(" cron:'{}'", cron_val));
        }
        
        if let Some(job_id_val) = job_id {
            script.push_str(&format!(" job_id:'{}'", job_id_val));
        }
        
        self.client.send_command(&script)
    }
    
    /// Stop a running process
    pub fn stop(&mut self, name: &str) -> Result<String> {
        let script = format!("!!process.stop name:'{}'", name);
        self.client.send_command(&script)
    }
    
    /// Restart a process
    pub fn restart(&mut self, name: &str) -> Result<String> {
        let script = format!("!!process.restart name:'{}'", name);
        self.client.send_command(&script)
    }
    
    /// Delete a process
    pub fn delete(&mut self, name: &str) -> Result<String> {
        let script = format!("!!process.delete name:'{}'", name);
        self.client.send_command(&script)
    }
    
    /// List all processes
    pub fn list(&mut self) -> Result<Vec<ProcessInfo>> {
        let script = "!!process.list format:'json'";
        let response = self.client.send_command(&script)?;
        
        // Handle empty responses
        if response.trim().is_empty() {
            return Ok(Vec::new());
        }
        
        // Try to parse the response as JSON
        match serde_json::from_str::<Vec<ProcessInfo>>(&response) {
            Ok(processes) => Ok(processes),
            Err(_) => {
                // If parsing as a list fails, try parsing as a single ProcessInfo
                match serde_json::from_str::<ProcessInfo>(&response) {
                    Ok(process) => Ok(vec![process]),
                    Err(_) => {
                        // If both parsing attempts fail, check if it's a "No processes found" message
                        if response.contains("No processes found") {
                            Ok(Vec::new())
                        } else {
                            // Otherwise, try to send it as JSON
                            self.client.send_command_json(&script)
                        }
                    }
                }
            }
        }
    }
    
    /// Get the status of a specific process
    pub fn status(&mut self, name: &str) -> Result<ProcessInfo> {
        let script = format!("!!process.status name:'{}' format:'json'", name);
        
        // Use the send_command_json method which handles JSON parsing with better error handling
        self.client.send_command_json(&script)
    }
    
    /// Get the logs of a specific process
    pub fn logs(&mut self, name: &str, lines: Option<i32>) -> Result<String> {
        let mut script = format!("!!process.logs name:'{}'", name);
        
        if let Some(lines_val) = lines {
            script.push_str(&format!(" lines:{}", lines_val));
        }
        
        self.client.send_command(&script)
    }
    
    /// Set the logs path for the process manager
    pub fn set_logs_path(&mut self, path: &str) -> Result<String> {
        let script = format!("!!process.set_logs_path path:'{}'", path);
        self.client.send_command(&script)
    }
    
    /// Get help information for the process manager
    pub fn help(&mut self) -> Result<String> {
        let script = "!!process.help";
        self.client.send_command(&script)
    }
}
