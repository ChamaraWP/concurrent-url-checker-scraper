package main

import (
	"concurrent-url-checker-scraper/uptime-monitor/internal/monitor"
	"concurrent-url-checker-scraper/uptime-monitor/internal/server"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	// 1. Create a root background Context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure that we clean up resources when main exits

	// 2. Create channels for our system pipeline
	jobs := make(chan string, 100)
	results := make(chan string, 100)

	var wg sync.WaitGroup
	numberOfWorkers := 3

	for w := 1; w <= numberOfWorkers; w++ {
		wg.Add(1)
		go monitor.StartWorker(ctx, w, jobs, results, &wg) // Capitalized means Public/Exported
	}

	// 3. WIRING THE NEW STORAGE SUBSYSTEM:
	// Spin up exactly ONE background thread to own the log file.
	// We pass it the 'results' channel out-end, and the target file name.
	go monitor.StartFileLogger(results, "urls.txt")

	// 4. Initialize our HTTP server routes, passing it the 'jobs' channel
	// so the API endpoints can feed the workers dynamically.
	mux := server.RegisterRoutes(jobs)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// 5. Start the HTTP server in a separate Goroutine so it doesn't block.
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("🚀 Uptime Monitor Service running on port :8080...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 6. THE ROADBLOCK: Execution pauses here waiting for a Ctrl+C signal
	<-shutdownChan
	fmt.Println("\n⚠️  Shutdown signal received. Initiating graceful shutdown...")

	// Phase 1: Shut down the HTTP web server so it stops accepting any new requests
	// We give running web requests a 5-second timeout window to finish up
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP server Shutdown failed: %v", err)
	}
	fmt.Println("🛑 Web server stopped. No longer accepting new requests.")

	// Phase 2: Alert all background workers via the context wire to wrap things up
	cancel()  // This sends a shutdown signal to all workers listening to this context
	wg.Wait() // Wait for all workers to finish
	fmt.Println("✅ All background workers have completed their tasks.")

	// Phase 4: Seal the results pipeline so the File Logger drains out completely and exits
	close(results) // This signals the File Logger that no more data is coming, so it can exit cleanly.

	// Small break to let the file logger completely commit the last buffered entries to disk
	time.Sleep(500 * time.Millisecond)
	fmt.Println("📁 File system flushed and ledger file snapped shut. Goodbye!")
}
