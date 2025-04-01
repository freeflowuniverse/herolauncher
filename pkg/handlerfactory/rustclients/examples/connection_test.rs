use std::time::Duration;
use std::thread;
use std::io::{Read, Write};
use std::os::unix::net::{UnixListener, UnixStream};
use std::fs;

// Import directly from the lib.rs
use rustclients::FakeHandlerClient;
use rustclients::Result;

// Simple mock server that handles Unix socket connections
fn start_mock_server(socket_path: &str) -> std::thread::JoinHandle<()> {
    let socket_path = socket_path.to_string();
    thread::spawn(move || {
        // Remove the socket file if it exists
        let _ = fs::remove_file(&socket_path);
        
        // Create a Unix socket listener
        let listener = match UnixListener::bind(&socket_path) {
            Ok(listener) => listener,
            Err(e) => {
                println!("Failed to bind to socket {}: {}", socket_path, e);
                return;
            }
        };
        
        println!("Mock server listening on {}", socket_path);
        
        // Accept connections and handle them
        for stream in listener.incoming() {
            match stream {
                Ok(mut stream) => {
                    println!("Mock server: Accepted new connection");
                    
                    // Read from the stream
                    let mut buffer = [0; 1024];
                    match stream.read(&mut buffer) {
                        Ok(n) => {
                            let request = String::from_utf8_lossy(&buffer[0..n]);
                            println!("Mock server received: {}", request);
                            
                            // Send a welcome message first
                            let welcome = "Welcome to the mock server\n";
                            let _ = stream.write_all(welcome.as_bytes());
                            
                            // Send a response
                            let response = "OK: Command processed\n> ";
                            let _ = stream.write_all(response.as_bytes());
                        },
                        Err(e) => println!("Mock server error reading from stream: {}", e),
                    }
                },
                Err(e) => println!("Mock server error accepting connection: {}", e),
            }
        }
    })
}

fn main() -> Result<()> {
    // Define the socket path
    let socket_path = "/tmp/herolauncher/test.sock";
    
    // Start the mock server
    println!("Starting mock server...");
    let server_handle = start_mock_server(socket_path);
    
    // Give the server time to start
    thread::sleep(Duration::from_millis(500));
    
    // Initialize the client
    let client = FakeHandlerClient::new(socket_path)
        .with_timeout(Duration::from_secs(5));
    
    println!("\n--- Test 1: Making first request ---");
    // This should open a new connection
    match client.return_success(Some("Test 1")) {
        Ok(response) => println!("Response: {}", response),
        Err(e) => println!("Error: {}", e),
    }
    
    // Wait a moment
    thread::sleep(Duration::from_millis(500));
    
    println!("\n--- Test 2: Making second request ---");
    // This should open another new connection
    match client.return_success(Some("Test 2")) {
        Ok(response) => println!("Response: {}", response),
        Err(e) => println!("Error: {}", e),
    }
    
    // Wait a moment
    thread::sleep(Duration::from_millis(500));
    
    println!("\n--- Test 3: Making third request ---");
    // This should open yet another new connection
    match client.return_success(Some("Test 3")) {
        Ok(response) => println!("Response: {}", response),
        Err(e) => println!("Error: {}", e),
    }
    
    println!("\nTest completed. Check the debug output to verify that a new connection was opened for each request.");
    
    // Clean up
    let _ = fs::remove_file(socket_path);
    
    // Wait for the server to finish (in a real application, you might want to signal it to stop)
    println!("Waiting for server to finish...");
    // In a real application, we would join the server thread here
    
    Ok(())
}
