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

// checkURL fetches a single URL and prints whether it is up or down.
func checkURL(url string, ch chan<- results, wg *sync.WaitGroup) {
	// 3. Decrement the counter by 1 when this function finishes completely.
	defer wg.Done()

	// Configure a quick 5-second timeout so we don't hang forever.
	client := http.Client{Timeout: 5 * time.Second}

	start := time.Now()
	resp, err := client.Get(url)
	duration := time.Since(start)

	if err != nil {
		ch <- results{url: url, err: err}
		return
	}

	// Clean up the network connection resource.
	defer resp.Body.Close()
	ch <- results{url: url, status: resp.StatusCode, duration: duration}
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
	ch := make(chan results)

	fmt.Printf("Starting concurrent URL checks...")
	totalStart := time.Now()

	for _, url := range urls {
		// 1. Increment the WaitGroup counter for each worker we start.
		wg.Add(1)
		// 2. The 'go' keyword spins up a concurrent Goroutine instantly.
		go checkURL(url, ch, &wg)
	}

	go func() {
		wg.Wait() // 4. Block execution here until the WaitGroup counter drops back down to 0.
		close(ch) // Close the tube so the receiver knows no more data is coming.
	}()

	for res := range ch {
		if res.err != nil {
			fmt.Printf("[DOWN] %s (Error: %v)\n", res.url, res.err)
		} else {
			fmt.Printf("[UP]   %s (Status: %d, Time: %v)\n", res.url, res.status, res.duration)
		}
	}

	fmt.Printf("\nAll checks completed in %v!\n", time.Since(totalStart))
}
