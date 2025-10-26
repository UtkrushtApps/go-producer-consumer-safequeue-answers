package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Work represents a work item for consumers to handle.
type Work struct {
	ID   int
	Data string
}

// SafeQueue is a thread-safe FIFO queue for Work items with closure protocol.
type SafeQueue struct {
	mu      sync.Mutex
	cond    *sync.Cond
	queue   []Work
	closed  bool
}

func NewSafeQueue() *SafeQueue {
	q := &SafeQueue{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

// Enqueue adds an item to the queue. Returns error if queue is closed.
func (q *SafeQueue) Enqueue(item Work) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return fmt.Errorf("enqueue on closed queue")
	}
	q.queue = append(q.queue, item)
	q.cond.Signal() // Unblock one waiting consumer.
	return nil
}

// Dequeue retrieves an item from the queue. Blocks if queue empty.
// Returns (item, false) if closed and empty.
func (q *SafeQueue) Dequeue() (Work, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.queue) == 0 && !q.closed {
		q.cond.Wait()
	}
	if len(q.queue) == 0 {
		// Queue empty and closed
		return Work{}, false
	}
	item := q.queue[0]
	q.queue = q.queue[1:]
	return item, true
}

// Close marks the queue as closed and wakes waiting goroutines.
func (q *SafeQueue) Close() {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()
	q.cond.Broadcast() // Wake up all waiting goroutines
}

func main() {
	rand.Seed(time.Now().UnixNano())
	queue := NewSafeQueue()
	var wg sync.WaitGroup
	
	done := make(chan struct{})

	// For unique work IDs
	var workID int64
	var idMu sync.Mutex

	producer := func(pid int) {
		defer wg.Done()
		for {
			select {
			case <-done:
				return // shutdown
			default:
				// continue
			}
			// Simulate producing work
			time.Sleep(time.Duration(rand.Intn(80)+20) * time.Millisecond)
			idMu.Lock()
			workID++
			id := workID
			idMu.Unlock()
			w := Work{ID: int(id), Data: fmt.Sprintf("WorkData-%d-p%d", id, pid)}
			err := queue.Enqueue(w)
			if err != nil {
				// Queue closed
				return
			}
			fmt.Printf("Producer %d: produced %v\n", pid, w)
		}
	}

	consumer := func(cid int) {
		defer wg.Done()
		for {
			item, ok := queue.Dequeue()
			if !ok {
				// Queue closed and drained
				return
			}
			// Simulate processing
			fmt.Printf("  Consumer %d: processing %v\n", cid, item)
			time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)
		}
	}

	// Start 3 producers
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go producer(i)
	}
	// Start 4 consumers
	for i := 1; i <= 4; i++ {
		wg.Add(1)
		go consumer(i)
	}

	// Signal handling for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigs
	fmt.Println("\nShutdown signal received...")

	close(done) // Signal producers
	// Wait a bit for producers to finish enqueueing
	time.Sleep(250 * time.Millisecond)
	queue.Close() // Close the queue, wake consumers

	wg.Wait()
	fmt.Println("All goroutines completed, clean shutdown.")
}