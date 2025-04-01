use std::io::{Read, Write};
use std::os::unix::net::UnixStream;
use std::time::Duration;
use thiserror::Error;
use std::fmt;
use std::error::Error as StdError;

mod processmanager;
mod fakehandler;

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

/// Error type for the client
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
    timeout: Duration,
    secret: Option<String>,
}

impl Client {
    /// Create a new Unix socket client
    pub fn new(socket_path: &str) -> Self {
        Self {
            socket_path: socket_path.to_string(),
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
    
    /// Connect to the Unix socket and return the stream
    fn connect_socket(&self) -> Result<UnixStream> {
        println!("DEBUG: Opening new connection to {}", self.socket_path);
        // Connect to the socket
        let stream = UnixStream::connect(&self.socket_path)
            .map_err(|e| ClientError::ConnectionError(format!("Failed to connect to socket {}: {}", self.socket_path, e)))?;
        
        // Set read timeout
        stream.set_read_timeout(Some(self.timeout))?;
        stream.set_write_timeout(Some(self.timeout))?;
        
        // Read welcome message
        let mut buffer = [0; 4096];
        match stream.try_clone()?.read(&mut buffer) {
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
        if let Some(secret) = &self.secret {
            self.authenticate_stream(&stream, secret)?;
        }
        
        Ok(stream)
    }
    
    /// Authenticate with the server using the provided stream
    fn authenticate_stream(&self, stream: &UnixStream, secret: &str) -> Result<()> {
        let mut stream_clone = stream.try_clone()?;
        let auth_command = format!("auth {}\n\n", secret);
        
        // Send the auth command
        stream_clone.write_all(auth_command.as_bytes())
            .map_err(|e| ClientError::CommandError(format!("Failed to send auth command: {}", e)))?;
        stream_clone.flush()
            .map_err(|e| ClientError::CommandError(format!("Failed to flush auth command: {}", e)))?;
        
        // Add a small delay to ensure the server has time to process the command
        std::thread::sleep(Duration::from_millis(100));
        
        // Read the response
        let mut buffer = [0; 4096];
        let n = stream_clone.read(&mut buffer)
            .map_err(|e| ClientError::CommandError(format!("Failed to read auth response: {}", e)))?;
        
        if n == 0 {
            return Err(ClientError::ConnectionError("Connection closed by server during authentication".to_string()));
        }
        
        let response = String::from_utf8_lossy(&buffer[0..n]).to_string();
        
        // Check for authentication success
        if response.contains("Authentication successful") || response.contains("authenticated") {
            Ok(())
        } else {
            Err(ClientError::ServerError(format!("Authentication failed: {}", response)))
        }
    }
    
    /// Send a command to the server and get the response
    pub fn send_command(&self, command: &str) -> Result<String> {
        // Connect to the socket for this command
        let mut stream = self.connect_socket()?;
        
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
        
        // Close the connection by dropping the stream
        println!("DEBUG: Closing connection to {}", self.socket_path);
        drop(stream);
        
        Ok(response)
    }
    
    /// Send a command and parse the JSON response
    pub fn send_command_json<T: serde::de::DeserializeOwned>(&self, command: &str) -> Result<T> {
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
    
    /// For backward compatibility
    pub fn connect(&self) -> Result<()> {
        // Just verify we can connect
        let stream = self.connect_socket()?;
        drop(stream);
        Ok(())
    }
    
    /// For backward compatibility
    pub fn close(&self) -> Result<()> {
        // No-op since we don't maintain a persistent connection
        Ok(())
    }
    
    /// Authenticate with the server - kept for backward compatibility
    pub fn authenticate(&self, secret: &str) -> Result<()> {
        // Create a temporary connection to authenticate
        let stream = self.connect_socket()?;
        self.authenticate_stream(&stream, secret)
    }
}
