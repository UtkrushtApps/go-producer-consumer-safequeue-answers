# Solution Steps

1. Define the Work struct to represent a work item for consumers.

2. Implement the thread-safe SafeQueue struct with sync.Mutex, sync.Cond, an internal slice, and a closed flag.

3. Provide NewSafeQueue() constructor that initializes the mutex and condition variable.

4. Implement Enqueue such that: it locks the mutex, checks if closed, appends to the queue, signals waiting consumers, or errors if closed.

5. Implement Dequeue: It blocks (cond.Wait) if the queue is empty and not closed, returns (item, true) if item is available, or (zero, false) if closed and empty.

6. Implement Close: set closed flag under lock and broadcast to wake up all waiting goroutines.

7. In main(), instantiate SafeQueue and a sync.WaitGroup for goroutine tracking.

8. Define a unique work ID system guarded by mutex for safe incrementing by producers.

9. Implement the producer goroutine: loop until done signaled, simulate some work, enqueue items, exit cleanly if queue closed or done.

10. Implement the consumer goroutine: loop dequeuing work, process each work item, exit when queue closed and drained.

11. Start 3 producers and 4 consumers as goroutines, tracking them in WaitGroup.

12. Set up signal handling for SIGINT/SIGTERM, and trigger clean shutdown by closing a done channel.

13. After shutdown signal: close done channel, wait a brief moment, then call queue.Close() to unblock consumers and prevent further enqueues.

14. Wait for all goroutines to finish with WaitGroup.Wait() and print a clean shutdown confirmation message.

