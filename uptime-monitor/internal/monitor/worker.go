package monitor

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// StartWorker pulls URLs from the jobs channel and executes network checks.
// Note: Function names MUST start with a Capital letter to be visible outside this folder!
func StartWorker(ctx context.Context, id int, jobs <-chan string, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	client := http.Client{Timeout: 3 * time.Second}

	for {
		select {
		// WIRING POINT A: Check if the main loop has triggered a shutdown signal
		case <-ctx.Done():
			fmt.Printf("[Worker %d] Shutdown signal received.Exiting loops cleanly.\n", id)
			return // Drops out of the function safely, triggering the deferred wg.Done()

		// WIRING POINT B: If no shutdown signal, try to read a job from the channel
		case url, ok := <-jobs:
			if !ok {
				// The channel was closed, no more work is left
				return
			}

			resp, err := client.Get(url)
			if err != nil {
				results <- fmt.Sprintf("[Worker %d] DOWN: %s (Error: %v)", id, url, err)
				continue
			}
			resp.Body.Close()
			results <- fmt.Sprintf("[Worker %d] UP  : %s (Status: %d)", id, url, resp.StatusCode)
		}
	}
}
