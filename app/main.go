package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// checkURL fetches a single URL and prints whether it is up or down.
func checkURL(url string, wg *sync.WaitGroup) {
	// 3. Decrement the counter by 1 when this function finishes completely.
	defer wg.Done()

	// Configure a quick 5-second timeout so we don't hang forever.
	client := http.Client { Timeout: 5 * time.Second }

	start := time.Now()
	resp, err := client.Get(url)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("[DOWN] %s (Error: %v)\n", url,err)
		return
	}

	// Clean up the network connection resource.
	defer resp.Body.Close()

	fmt.Printf("[UP]   %s (Status: %d) took %v\n", url, resp.StatusCode, duration)
}

func main() {
	urls  := []string {
		"https://www.google.com",
		"https://www.facebook.com",
		"https://www.twitter.com",
		"https://www.linkedin.com",
		"https://www.github.com",
	}

	// Create a WaitGroup to track our background workers.
	var wg sync.WaitGroup

	fmt.Printf("Starting concurrent URL checks...")
	totalStart := time.Now()

	for _,url := range urls {
		// 1. Increment the WaitGroup counter for each worker we start.
		wg.Add(1)
		// 2. The 'go' keyword spins up a concurrent Goroutine instantly.
		go checkURL(url,&wg)
	}
	// 4. Block execution here until the WaitGroup counter drops back down to 0.
	wg.Wait()

	fmt.Printf("\nAll checks completed in %v!\n", time.Since(totalStart))
}