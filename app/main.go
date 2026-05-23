package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type results struct {
	url      string
	status   int
	duration time.Duration
	err      error
}

func worker(id int, jobs <-chan string, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	client := http.Client{Timeout: 3 * time.Second}

	// This loop blocks and waits until a job enters the channel.
	// It automatically terminates when the jobs channel is closed.
	for url := range jobs {
		fmt.Printf("[Worker %d] Started checking: %s\n", id, url)
		resp, err := client.Get(url)
		if err != nil {
			results <- fmt.Sprintf("[Worker %d] DOWN: %s (Error: %v)", id, url, err)
			continue
		}
		resp.Body.Close()
		results <- fmt.Sprintf("[Worker %d] UP  : %s (Status: %d)", id, url, resp.StatusCode)
	}

}

func main() {
	urls := []string{
		"https://www.google.com",
		"https://www.facebook.com",
		"https://www.twitter.com",
		"https://www.linkedin.com",
		"https://www.github.com",
	}

	// Create a WaitGroup to track our background workers.
	var wg sync.WaitGroup
	jobs := make(chan string, len(urls))    // Buffered channel to hold results from workers.
	results := make(chan string, len(urls)) // Buffered channel to hold results from workers.
	numberOfWorkers := 3

	// Start a fixed number of worker Goroutines.

	fmt.Printf("Starting concurrent URL checks...")
	totalStart := time.Now()

	for w := 1; w <= numberOfWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	for _, url := range urls {
		jobs <- url
	}

	close(jobs)

	go func() {
		wg.Wait()      // 4. Block execution here until the WaitGroup counter drops back down to 0.
		close(results) // Close the tube so the receiver knows no more data is coming.
	}()

	for res := range results {
		fmt.Println(res)
	}

	fmt.Printf("\nAll checks completed in %v!\n", time.Since(totalStart))
}
