use std::io::{Read, Write};
use std::os::unix::net::UnixStream;
use std::time::Duration;
use thiserror::Error;
use std::fmt;
use std::error::Error as StdError;

mod client;
mod processmanager;
mod fakehandler;

pub use client::{Client, ClientError, Result};
pub use processmanager::ProcessManagerClient;
pub use fakehandler::FakeHandlerClient;

/// Standard error response from the telnet server
#[derive(Debug, Clone)]
pub struct ServerError {
    pub message: String,
    pub raw_response: String,
}

impl fmt::Display for ServerError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.message)
    }
}

impl StdError for ServerError {}

/// Error type for the improved client
#[derive(Error, Debug)]
pub enum ClientError {
    #[error("IO error: {0}")]
    IoError(#[from] std::io::Error),
    
    #[error("JSON parsing error: {0}")]
    JsonError(#[from] serde_json::Error),
    
    #[error("Connection error: {0}")]
    ConnectionError(String),
    
    #[error("Command error: {0}")]
    CommandError(String),
    
    #[error("Server error: {0}")]
    ServerError(String),
}

pub type Result<T> = std::result::Result<T, ClientError>;

/// A client for connecting to a Unix socket server with improved error handling
pub struct Client {
    socket_path: String,
    stream: Option<UnixStream>,
    timeout: Duration,
    secret: Option<String>,
}

impl Client {
    /// Create a new Unix socket client
    pub fn new(socket_path: &str) -> Self {
        Self {
            socket_path: socket_path.to_string(),
            stream: None,
            timeout: Duration::from_secs(10),
            secret: None,
        }
    }
    
    /// Set the connection timeout
    pub fn with_timeout(mut self, timeout: Duration) -> Self {
        self.timeout = timeout;
        self
    }
    
    /// Set the authentication secret
    pub fn with_secret(mut self, secret: &str) -> Self {
        self.secret = Some(secret.to_string());
        self
    }
    
    /// Connect to the Unix socket
    pub fn connect(&mut self) -> Result<()> {
        // Close existing connection if any
        self.close()?;
        
        // Connect to the socket
        let stream = UnixStream::connect(&self.socket_path)
            .map_err(|e| ClientError::ConnectionError(format!("Failed to connect to socket {}: {}", self.socket_path, e)))?;
        
        // Set read timeout
        stream.set_read_timeout(Some(self.timeout))?;
        stream.set_write_timeout(Some(self.timeout))?;
        
        self.stream = Some(stream);
        
        // Read welcome message
        let mut buffer = [0; 4096];
        let stream = self.stream.as_mut().unwrap();
        match stream.read(&mut buffer) {
            Ok(n) => {
                let welcome = String::from_utf8_lossy(&buffer[0..n]);
                if !welcome.contains("Welcome") {
                    return Err(ClientError::ConnectionError("Invalid welcome message".to_string()));
                }
            },
            Err(e) => {
                return Err(ClientError::IoError(e));
            }
        }
        
        // Authenticate if a secret is provided
        let secret_clone = self.secret.clone();
        if let Some(secret) = secret_clone {
            self.authenticate(&secret)?;
        }
        
        Ok(())
    }
    
    /// Close the connection
    pub fn close(&mut self) -> Result<()> {
        if let Some(stream) = self.stream.take() {
            drop(stream);
        }
        Ok(())
    }
    
    /// Send a command to the server and get the response
    pub fn send_command(&mut self, command: &str) -> Result<String> {
        let stream = self.stream.as_mut()
            .ok_or_else(|| ClientError::ConnectionError("Not connected".to_string()))?;
        
        // Ensure command ends with double newlines to execute it
        let command = if command.ends_with("\n\n") {
            command.to_string()
        } else if command.ends_with('\n') {
            format!("{}\n", command)
        } else {
            format!("{}\n\n", command)
        };
        
        // Send the command
        stream.write_all(command.as_bytes())
            .map_err(|e| ClientError::CommandError(format!("Failed to send command: {}", e)))?;
        stream.flush()
            .map_err(|e| ClientError::CommandError(format!("Failed to flush command: {}", e)))?;
        
        // Add a small delay to ensure the server has time to process the command
        std::thread::sleep(Duration::from_millis(100));
        
        // Read the response
        let mut buffer = [0; 8192]; // Use a larger buffer for large responses
        let n = stream.read(&mut buffer)
            .map_err(|e| ClientError::CommandError(format!("Failed to read response: {}", e)))?;
        
        if n == 0 {
            return Err(ClientError::ConnectionError("Connection closed by server".to_string()));
        }
        
        let response = String::from_utf8_lossy(&buffer[0..n]).to_string();
        
        // Remove the prompt if present
        let response = response.trim_end_matches("> ").trim().to_string();
        
        // Check for standard error format
        if response.starts_with("Error:") {
            return Err(ClientError::ServerError(response));
        }
        
        Ok(response)
    }
    
    /// Send a command and parse the JSON response
    pub fn send_command_json<T: serde::de::DeserializeOwned>(&mut self, command: &str) -> Result<T> {
        let response = self.send_command(command)?;
        
        // If the response is empty, return an error
        if response.trim().is_empty() {
            return Err(ClientError::CommandError("Empty response".to_string()));
        }
        
        // Handle "action not supported" errors specially
        if response.contains("action not supported") {
            return Err(ClientError::ServerError(response));
        }
        
        // Try to parse the JSON response
        match serde_json::from_str::<T>(&response) {
            Ok(result) => Ok(result),
            Err(e) => {
                // If parsing fails, check if it's an error message
                if response.starts_with("Error:") || response.contains("error") || response.contains("failed") {
                    Err(ClientError::ServerError(response))
                } else {
                    Err(ClientError::JsonError(e))
                }
            },
        }
    }
    
    /// Authenticate with the server
    fn authenticate(&mut self, secret: &str) -> Result<()> {
        let stream = self.stream.as_mut()
            .ok_or_else(|| ClientError::ConnectionError("Not connected".to_string()))?;
        
        // Send the secret
        stream.write_all(format!("{secret}\n").as_bytes())
            .map_err(|e| ClientError::CommandError(format!("Failed to send secret: {e}")))?;
        
        // Read the authentication response
        let mut buffer = [0; 4096];
        match stream.read(&mut buffer) {
            Ok(n) => {
                let response = String::from_utf8_lossy(&buffer[0..n]);
                if response.contains("Authentication successful") {
                    Ok(())
                } else {
                    Err(ClientError::CommandError(format!("Authentication failed: {response}")))
                }
            },
            Err(e) => Err(ClientError::IoError(e)),
        }
    }
}
