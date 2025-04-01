use serde::{Deserialize, Serialize};
use std::time::Duration;

use crate::{Client, Result, ClientError};

/// Response from the fake handler
#[derive(Debug, Serialize, Deserialize, Default)]
pub struct FakeResponse {
    #[serde(default)]
    pub message: String,
    #[serde(default)]
    pub status: String,
    #[serde(default)]
    pub code: i32,
}

/// Client for the fake handler
pub struct FakeHandlerClient {
    client: Client,
}

impl FakeHandlerClient {
    /// Create a new fake handler client
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
    
    /// Connect to the server
    pub fn connect(&mut self) -> Result<()> {
        self.client.connect()
    }
    
    /// Close the connection
    pub fn close(&mut self) -> Result<()> {
        self.client.close()
    }
    
    /// Return a success message
    pub fn return_success(&mut self, message: Option<&str>) -> Result<String> {
        let mut script = "!!fake.return_success".to_string();
        
        if let Some(msg) = message {
            script.push_str(&format!(" message:'{}'", msg));
        }
        
        self.client.send_command(&script)
    }
    
    /// Return an error message
    pub fn return_error(&mut self, message: Option<&str>) -> Result<String> {
        let mut script = "!!fake.return_error".to_string();
        
        if let Some(msg) = message {
            script.push_str(&format!(" message:'{}'", msg));
        }
        
        // This will return a ClientError::ServerError with the error message
        self.client.send_command(&script)
    }
    
    /// Return a JSON response
    pub fn return_json(&mut self, message: Option<&str>, status: Option<&str>, code: Option<i32>) -> Result<FakeResponse> {
        let mut script = "!!fake.return_json".to_string();
        
        if let Some(msg) = message {
            script.push_str(&format!(" message:'{}'", msg));
        }
        
        if let Some(status_val) = status {
            script.push_str(&format!(" status:'{}'", status_val));
        }
        
        if let Some(code_val) = code {
            script.push_str(&format!(" code:{}", code_val));
        }
        
        let response = self.client.send_command(&script)?;
        
        // Parse the JSON response
        match serde_json::from_str::<FakeResponse>(&response) {
            Ok(result) => Ok(result),
            Err(e) => Err(ClientError::JsonError(e)),
        }
    }
    
    /// Return an invalid JSON response
    pub fn return_invalid_json(&mut self) -> Result<FakeResponse> {
        let script = "!!fake.return_invalid_json";
        let response = self.client.send_command(&script)?;
        
        // This should fail with a JSON parsing error
        match serde_json::from_str::<FakeResponse>(&response) {
            Ok(result) => Ok(result),
            Err(e) => Err(ClientError::JsonError(e)),
        }
    }
    
    /// Return an empty response
    pub fn return_empty(&mut self) -> Result<String> {
        let script = "!!fake.return_empty";
        self.client.send_command(&script)
    }
    
    /// Return a large response
    pub fn return_large(&mut self, size: Option<i32>) -> Result<String> {
        let mut script = "!!fake.return_large".to_string();
        
        if let Some(size_val) = size {
            script.push_str(&format!(" size:{}", size_val));
        }
        
        self.client.send_command(&script)
    }
    
    /// Return a malformed error message
    pub fn return_malformed_error(&mut self) -> Result<String> {
        let script = "!!fake.return_malformed_error";
        self.client.send_command(&script)
    }
    
    /// Get help information
    pub fn help(&mut self) -> Result<String> {
        let script = "!!fake.help";
        self.client.send_command(&script)
    }
}
