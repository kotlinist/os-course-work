package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var rootPipePath = "/dev/shm"
var pipePath string
var collectorRole = "mouse"
var logDir = ""

func main() {
	flag.StringVar(&collectorRole, "role", collectorRole, "collector role [metrics, mouse, process]")
	flag.StringVar(&rootPipePath, "logger-pipe", rootPipePath, "pipe path for logger")
	//flag.StringVar(&logDir, "log-dir", logDir, "log directory path")
	flag.Parse()
	pipePath = rootPipePath + "/collector-log-" + collectorRole + ".pipe"

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println("\nShutdown", sig)
		deleteOldPipe()
		os.Exit(0)
	}()

	create()

	for {
		pipe, err := os.OpenFile(pipePath, os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			log.Fatalf("Error opening pipe %s for reading: %v", pipePath, err)
		}

		log.Printf("The pipe is open for reading: %s", pipePath)

		var logFile *os.File
		logFile, err = os.OpenFile("collector-log-"+collectorRole+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Error opening file: %v", err)
		}

		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			text := scanner.Text() + "\n"
			fmt.Print(text)
			if _, err := logFile.WriteString(text); err != nil {
				log.Fatalf("Error writing to file: %v", err)
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading from pipe: %v", err)
		}

		logFile.Close()
		pipe.Close()

		log.Println("Waiting for data from the collector...")
	}

}

func create() {
	deleteOldPipe()

	err := syscall.Mkfifo(pipePath, 0666)
	if err != nil {
		log.Fatalf("Error creating named pipe: %v", err)
	}

	log.Printf("A named pipe has been created: %s", pipePath)
}

func deleteOldPipe() {
	if _, err := os.Stat(pipePath); err == nil {
		log.Printf("Pipe %s уже существует. Удаляем его.", pipePath)
		if err := os.Remove(pipePath); err != nil {
			log.Fatalf("Failed to delete existing pipe: %v", err)
		}
	}
}
