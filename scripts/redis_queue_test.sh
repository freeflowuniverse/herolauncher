#!/bin/bash

# Redis Queue Test Script
# This script tests the Redis queue commands (LPUSH, RPUSH, LPOP, RPOP, LLEN, LRANGE)

echo "Redis Queue Test Script"
echo "======================="

# Check if Redis server is running
redis-cli -p 6378 PING > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "Error: Redis server is not running on port 6378"
    echo "Please start the server before running this script"
    exit 1
fi

echo "Redis server is running on port 6378"

# Clear any existing test keys
redis-cli -p 6378 DEL test:queue > /dev/null
redis-cli -p 6378 DEL test:queue2 > /dev/null

# Test LPUSH and LLEN
echo -e "\nTesting LPUSH and LLEN:"
echo "Adding items to the queue from the left side..."
redis-cli -p 6378 LPUSH test:queue "item1"
redis-cli -p 6378 LPUSH test:queue "item2"
redis-cli -p 6378 LPUSH test:queue "item3"

echo "Queue length:"
redis-cli -p 6378 LLEN test:queue

# Test LRANGE
echo -e "\nTesting LRANGE:"
echo "Queue contents (all items):"
redis-cli -p 6378 LRANGE test:queue 0 -1

# Test RPUSH
echo -e "\nTesting RPUSH:"
echo "Adding items to the queue from the right side..."
redis-cli -p 6378 RPUSH test:queue "item4"
redis-cli -p 6378 RPUSH test:queue "item5"

echo "Queue length after RPUSH:"
redis-cli -p 6378 LLEN test:queue

echo "Queue contents after RPUSH:"
redis-cli -p 6378 LRANGE test:queue 0 -1

# Test LPOP
echo -e "\nTesting LPOP:"
echo "Removing item from the left side of the queue..."
echo "Popped item: $(redis-cli -p 6378 LPOP test:queue)"

echo "Queue length after LPOP:"
redis-cli -p 6378 LLEN test:queue

echo "Queue contents after LPOP:"
redis-cli -p 6378 LRANGE test:queue 0 -1

# Test RPOP
echo -e "\nTesting RPOP:"
echo "Removing item from the right side of the queue..."
echo "Popped item: $(redis-cli -p 6378 RPOP test:queue)"

echo "Queue length after RPOP:"
redis-cli -p 6378 LLEN test:queue

echo "Queue contents after RPOP:"
redis-cli -p 6378 LRANGE test:queue 0 -1

# Test queue as a FIFO (First In, First Out)
echo -e "\nTesting queue as FIFO (using RPUSH and LPOP):"
echo "Creating a new queue..."
redis-cli -p 6378 RPUSH test:queue2 "first"
redis-cli -p 6378 RPUSH test:queue2 "second"
redis-cli -p 6378 RPUSH test:queue2 "third"

echo "Queue contents:"
redis-cli -p 6378 LRANGE test:queue2 0 -1

echo "Dequeuing items in FIFO order:"
echo "First out: $(redis-cli -p 6378 LPOP test:queue2)"
echo "Second out: $(redis-cli -p 6378 LPOP test:queue2)"
echo "Third out: $(redis-cli -p 6378 LPOP test:queue2)"

# Test queue as a LIFO (Last In, First Out) / Stack
echo -e "\nTesting queue as LIFO/Stack (using LPUSH and LPOP):"
echo "Creating a new stack..."
redis-cli -p 6378 LPUSH test:queue2 "bottom"
redis-cli -p 6378 LPUSH test:queue2 "middle"
redis-cli -p 6378 LPUSH test:queue2 "top"

echo "Stack contents:"
redis-cli -p 6378 LRANGE test:queue2 0 -1

echo "Popping items from stack:"
echo "First out (top): $(redis-cli -p 6378 LPOP test:queue2)"
echo "Second out (middle): $(redis-cli -p 6378 LPOP test:queue2)"
echo "Third out (bottom): $(redis-cli -p 6378 LPOP test:queue2)"

echo -e "\nQueue tests completed successfully!"
echo "======================="
