package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
)

var fifoInitialized bool        /* If file watching is initialized */
var fifoWmClass string          /* Current wm class */
var fifoReady = make(chan bool) /* If the new wmclass is already read */

const fifoPath = "/tmp/bamboo_introspector.fifo"

func fifoGetLatestFocusWindowClass() string {
	<-fifoReady // Wait for new WM CLASS
	return fifoWmClass
}

func fifoCheckHasFile() bool {
	var _, err = os.Stat(fifoPath)
	if err == nil {
		return true
	}
	return false
}

// Initialize FIFO watching
func fifoWatchInitialize() {

	// If the pipe exists, just remove it
	if fifoCheckHasFile() {
		os.Remove(fifoPath)
	}

	// Create a new FIFO
	var err = syscall.Mkfifo(fifoPath, 0600)
	if err != nil {
		log.Fatalf("Failed to create FIFO: %v", err)
	} else {
		fmt.Printf("FIFO created successfully at: %s\n", fifoPath)
	}

	// Try reading the fifo file
	var file *os.File
	file, err = os.OpenFile(fifoPath, os.O_RDWR, 0600)
	if err != nil {
		fmt.Printf("Failed to open FIFO: %v\n", err)
	}

	var scanner = bufio.NewScanner(file)
	for {
		if scanner.Scan() {
			// Process each incoming line
			var line = strings.Trim(scanner.Text(), "\r\n ")
			fifoWmClass = line
			fifoReady <- true
		} else {
			// Handle scanning errors or unexpected closures
			if err := scanner.Err(); err != nil {
				log.Printf("Scanner error: %v", err)
			}
			// If it hits an absolute EOF despite O_RDWR, break out
			file.Close()
			break
		}
	}
}
