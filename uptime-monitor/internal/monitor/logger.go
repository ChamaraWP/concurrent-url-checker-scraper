package monitor

import (
	"log"
	"os"
)

// StartFileLogger runs on a single, dedicated background thread.
// It listens to the results channel and commits every incoming string to disk.
func StartFileLogger(results <-chan string, filePath string) {
	for res := range results {
		// 1. Open the file.
		// O_APPEND: Add new text to the bottom.
		// O_CREATE: If the file doesn't exist, build a brand new one.
		// O_WRONLY: Open strictly for writing data.
		// 0644: Standard Linux file permission settings (Owner can read/write, others can read).
		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("[STORAGE ERROR] Failed to open log file: %v", err)
			continue
		}

		// 2. Write the log string followed by a clean newline character.
		_, err = file.WriteString(res + "\n")

		if err != nil {
			log.Printf("[STORAGE ERROR] Failed to write to disk: %v", err)
		}

		// 3. Immediately close the file gate to free the OS resource.
		// Note: We don't use 'defer' here because this is a continuous 'for' loop.
		// Using defer inside a loop would pile up open files until the loop ends!
		file.Close()
	}
}
