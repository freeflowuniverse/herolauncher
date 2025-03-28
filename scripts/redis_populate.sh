#!/bin/bash
# Script to populate Redis with test data

# Set Redis port
REDIS_PORT=6378

# Function to check if Redis is running
check_redis() {
  if ! redis-cli -p $REDIS_PORT PING &>/dev/null; then
    echo "Error: Redis server is not running on port $REDIS_PORT"
    echo "Please start the Redis server first"
    exit 1
  fi
  echo "Redis server is running on port $REDIS_PORT"
}

# Function to populate string data
populate_strings() {
  echo "Populating string data..."
  
  # User data
  redis-cli -p $REDIS_PORT SET user:1:name "John Doe"
  redis-cli -p $REDIS_PORT SET user:1:email "john@example.com"
  redis-cli -p $REDIS_PORT SET user:1:age "30"
  
  # Set with expiration
  redis-cli -p $REDIS_PORT SET user:1:session "abc123" EX 60
  
  redis-cli -p $REDIS_PORT SET user:2:name "Jane Smith"
  redis-cli -p $REDIS_PORT SET user:2:email "jane@example.com"
  redis-cli -p $REDIS_PORT SET user:2:age "28"
  
  # Product data
  redis-cli -p $REDIS_PORT SET product:1:name "Laptop"
  redis-cli -p $REDIS_PORT SET product:1:price "999.99"
  redis-cli -p $REDIS_PORT SET product:1:stock "50"
  
  redis-cli -p $REDIS_PORT SET product:2:name "Smartphone"
  redis-cli -p $REDIS_PORT SET product:2:price "499.99"
  redis-cli -p $REDIS_PORT SET product:2:stock "100"
  
  # Configuration data
  redis-cli -p $REDIS_PORT SET config:app:name "HeroLauncher"
  redis-cli -p $REDIS_PORT SET config:app:version "1.0.0"
  redis-cli -p $REDIS_PORT SET config:app:environment "development"
  
  echo "String data population complete"
}

# Function to populate hash data
populate_hashes() {
  echo "Populating hash data..."
  
  # User profiles as hashes
  redis-cli -p $REDIS_PORT HSET user:1 name "John Doe" email "john@example.com" age 30 role "admin" active "true"
  redis-cli -p $REDIS_PORT HSET user:2 name "Jane Smith" email "jane@example.com" age 28 role "developer" active "true"
  redis-cli -p $REDIS_PORT HSET user:3 name "Bob Johnson" email "bob@example.com" age 35 role "manager" active "false"
  
  # Product details as hashes
  redis-cli -p $REDIS_PORT HSET product:1 name "Laptop" price 999.99 stock 50 category "electronics" brand "TechBrand"
  redis-cli -p $REDIS_PORT HSET product:2 name "Smartphone" price 499.99 stock 100 category "electronics" brand "PhoneCo"
  redis-cli -p $REDIS_PORT HSET product:3 name "Headphones" price 149.99 stock 200 category "accessories" brand "AudioTech"
  
  # Configuration settings as hashes
  redis-cli -p $REDIS_PORT HSET config:app name "HeroLauncher" version "1.0.0" environment "development" debug "true"
  redis-cli -p $REDIS_PORT HSET config:server host "localhost" port 9001 maxConnections 100 timeout 30
  
  echo "Hash data population complete"
}

# Function to test Redis commands
test_commands() {
  echo "Testing Redis commands..."
  
  echo -e "\nTesting PING:"
  redis-cli -p $REDIS_PORT PING
  
  echo -e "\nTesting GET:"
  redis-cli -p $REDIS_PORT GET user:1:name
  
  echo -e "\nTesting HGET:"
  redis-cli -p $REDIS_PORT HGET user:1 email
  
  echo -e "\nTesting TYPE command:"
  redis-cli -p $REDIS_PORT TYPE user:1:name
  redis-cli -p $REDIS_PORT TYPE user:1
  
  echo -e "\nTesting SCAN command:"
  redis-cli -p $REDIS_PORT SCAN 0 MATCH user:* COUNT 5
  
  echo -e "\nTesting TTL command:"
  redis-cli -p $REDIS_PORT TTL user:1:name
  redis-cli -p $REDIS_PORT TTL user:1:session
  redis-cli -p $REDIS_PORT TTL nonexistent:key
  
  echo -e "\nTesting EXPIRE command:"
  redis-cli -p $REDIS_PORT EXPIRE user:1:name 120
  redis-cli -p $REDIS_PORT TTL user:1:name
  redis-cli -p $REDIS_PORT EXPIRE nonexistent:key 60
  
  echo -e "\nTesting INFO command:"
  redis-cli -p $REDIS_PORT INFO | head -n 20
  
  echo -e "\nTesting KEYS command:"
  redis-cli -p $REDIS_PORT KEYS user:*
  
  echo "Command testing complete"
}

# Main execution
echo "Redis Data Population Script"
echo "==========================="

check_redis
populate_strings
populate_hashes
test_commands

echo "==========================="
echo "Redis population completed successfully!"
