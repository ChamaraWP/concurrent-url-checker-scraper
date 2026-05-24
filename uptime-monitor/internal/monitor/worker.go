package monitor

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// StartWorker pulls URLs from the jobs channel and executes network checks.
// Note: Function names MUST start with a Capital letter to be visible outside this folder!
func StartWorker(id int, jobs <-chan string, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	client := http.Client{Timeout: 3 * time.Second}

	for url := range jobs {
		resp, err := client.Get(url)
		if err != nil {
			results <- fmt.Sprintf("Worker %d failed to hit %s: %v", id, url, err)
			continue
		}
		resp.Body.Close()
		results <- fmt.Sprintf("Worker %d successfully pinged %s (Status: %d)", id, url, resp.StatusCode)
	}
}
