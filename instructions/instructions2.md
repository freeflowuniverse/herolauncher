
create a module jobmanager

## which has a key value store in 

## which has a queue

based on following implementation 


```go
package main

import "fmt"

type Queue struct {
	items []string
}

func (q *Queue) Enqueue(item string) {
	q.items = append(q.items, item)
}

func (q *Queue) Dequeue() (string, bool) {
	if len(q.items) == 0 {
		return "", false
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, true
}

func (q *Queue) IsEmpty() bool {
	return len(q.items) == 0
}

func main() {
	q := Queue{}
	q.Enqueue("apple")
	q.Enqueue("banana")

	item, _ := q.Dequeue()
	fmt.Println(item) // "apple"
}
```