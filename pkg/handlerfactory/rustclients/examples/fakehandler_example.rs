use std::time::Duration;

// Import directly from the lib.rs
use rustclients::FakeHandlerClient;
use rustclients::Result;


fn main() -> Result<()> {
    // Create a new fake handler client
    // Replace with the actual socket path used in your environment
    let socket_path = "/tmp/herolauncher/fakehandler.sock";
    
    // Initialize the client with a timeout
    let client = FakeHandlerClient::new(socket_path)
        .with_timeout(Duration::from_secs(5));
    
    println!("Connecting to fake handler at {}", socket_path);
    
    // Connect to the server
    match client.connect() {
        Ok(_) => println!("Successfully connected to fake handler"),
        Err(e) => {
            eprintln!("Failed to connect: {}", e);
            eprintln!("Make sure the fake handler server is running and the socket path is correct");
            return Err(e);
        }
    }
    
    // Test various commands
    
    // 1. Get help information
    println!("\n--- Help Information ---");
    match client.help() {
        Ok(help) => println!("{}", help),
        Err(e) => eprintln!("Error getting help: {}", e),
    }
    
    // 2. Return success message
    println!("\n--- Success Message ---");
    match client.return_success(Some("Custom success message")) {
        Ok(response) => println!("Success response: {}", response),
        Err(e) => eprintln!("Error getting success: {}", e),
    }
    
    // 3. Return JSON response
    println!("\n--- JSON Response ---");
    match client.return_json(Some("JSON message"), Some("success"), Some(200)) {
        Ok(response) => println!("JSON response: {:?}", response),
        Err(e) => eprintln!("Error getting JSON: {}", e),
    }
    
    // 4. Return error message (this will return a ClientError)
    println!("\n--- Error Message ---");
    match client.return_error(Some("Custom error message")) {
        Ok(response) => println!("Error response (unexpected success): {}", response),
        Err(e) => eprintln!("Expected error received: {}", e),
    }
    
    // 5. Return empty response
    println!("\n--- Empty Response ---");
    match client.return_empty() {
        Ok(response) => println!("Empty response (length: {})", response.len()),
        Err(e) => eprintln!("Error getting empty response: {}", e),
    }
    
    // 6. Return large response
    println!("\n--- Large Response ---");
    match client.return_large(Some(10)) {
        Ok(response) => {
            let lines: Vec<&str> = response.lines().collect();
            println!("Large response (first 3 lines of {} total):", lines.len());
            for i in 0..std::cmp::min(3, lines.len()) {
                println!("  {}", lines[i]);
            }
            println!("  ...");
        },
        Err(e) => eprintln!("Error getting large response: {}", e),
    }
    
    // 7. Return invalid JSON (will cause a JSON parsing error)
    println!("\n--- Invalid JSON ---");
    match client.return_invalid_json() {
        Ok(response) => println!("Invalid JSON response (unexpected success): {:?}", response),
        Err(e) => eprintln!("Expected JSON error received: {}", e),
    }
    
    // 8. Return malformed error
    println!("\n--- Malformed Error ---");
    match client.return_malformed_error() {
        Ok(response) => println!("Malformed error response: {}", response),
        Err(e) => eprintln!("Error with malformed error: {}", e),
    }
    
    // Close the connection
    println!("\nClosing connection");
    client.close()?;
    
    println!("Example completed successfully");
    Ok(())
}
