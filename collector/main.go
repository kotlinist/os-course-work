package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Client struct {
	conn         net.Conn
	updateMethod string
	lastResponse string
}

type Request struct {
	Command      string
	UpdateMethod string
}

type Response struct {
	Status    string
	Error     string
	Data      interface{}
	Timestamp string
}

func NewResponse(response Response) *Response {
	response.Timestamp = time.Now().Format(time.RFC3339)
	return &response
}

var collectorRole = "mouse"
var ip = "127.0.0.1"
var port = 8123
var serverEventsInterval = 1
var logger *Logger
var inputDevicesFile = "devices"
var rootPipePath = "/dev/shm"

//const PID_PREFIX = "/Users/kotlinist/workspace/projects/golang/collector" // /var/run

func main() {
	flag.StringVar(&collectorRole, "role", collectorRole, "collector role [metrics, mouse, process]")
	flag.StringVar(&ip, "ip", ip, "ip address")
	flag.IntVar(&port, "port", port, "port")
	flag.IntVar(&serverEventsInterval, "se-interval", 1, "serverEvents update interval")
	flag.StringVar(&rootPipePath, "logger-pipe", rootPipePath, "pipe path for logger")
	flag.StringVar(&inputDevicesFile, "input-devices", inputDevicesFile, "input devices file path")
	flag.Parse()

	if !slices.Contains([]string{"metrics", "mouse", "process"}, collectorRole) {
		log.Fatalf("collector role %s invalid\n", collectorRole)
	}

	if pidFileExists() {
		return
	}
	defer deletePidFile()

	logger = &Logger{}
	logger.Init()

	// канал для получения сигналов
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// goroutine для обработки сигналов
	go func() {
		<-sigs
		logger.Logln("Collector shutdown")
		done <- true
	}()

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		logger.Logln(err.Error())
		//deletePidFile()
		return
	}
	logger.Logln("Collector started")

	var connMap = &sync.Map{}

	if collectorRole == "mouse" {
		go serverEventsMouse(connMap)
	}
	if collectorRole == "process" {
		go serverEventsProcess(connMap)
	}

	go acceptConnections(&l, connMap)

	<-done
	logger.Close()
}

func acceptConnections(l *net.Listener, connMap *sync.Map) {
	defer (*l).Close()
	for {
		conn, err := (*l).Accept()
		if err != nil {
			fmt.Println("error accepting connection: ", err.Error())
			return
		}

		id := uuid.New().String()
		client := &Client{conn: conn, updateMethod: "manual"}
		connMap.Store(id, client)
		logger.Logln("Client connected (" + id + ")")

		go handleUserConnection(id, conn, connMap)
	}
}

func serverEventsMouse(connMap *sync.Map) {
	for {
		mice := getMouseInfo()
		connMap.Range(func(id, client interface{}) bool {
			if c, ok := client.(*Client); ok {
				if c.updateMethod != "server_events" {
					return true
				}
				response, _ := json.Marshal(NewResponse(Response{Status: "ok", Data: mice}))
				c.conn.Write([]byte(fmt.Sprintf("%s\n", response)))
			}
			return true
		})
		time.Sleep(time.Duration(serverEventsInterval) * time.Second)
	}
}

func serverEventsProcess(connMap *sync.Map) {
	for {
		processInfo, _ := getProcessInfo()

		connMap.Range(func(id, client interface{}) bool {
			if c, ok := client.(*Client); ok {
				if c.updateMethod != "server_events" {
					return true
				}
				response, _ := json.Marshal(NewResponse(Response{Status: "ok", Data: processInfo}))
				c.conn.Write([]byte(fmt.Sprintf("%s\n", response)))
			}
			return true
		})
		time.Sleep(time.Duration(serverEventsInterval) * time.Second)
	}
}

func handleUserConnection(id string, c net.Conn, connMap *sync.Map) {
	defer func() {
		err := c.Close()
		if err != nil {
			logger.Logln("error reading from client: " + err.Error())
		}
		connMap.Delete(id)
		logger.Logln("Client disconnected (" + id + ")")

	}()

	for {
		userInput, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			logger.Logln("error reading from client: " + err.Error())
			return
		}
		logger.Logln(fmt.Sprintf("[%s]: %s", id, strings.Trim(userInput, "\n")))

		cmd := Request{}
		err = json.Unmarshal([]byte(userInput), &cmd)
		if err != nil {
			logger.Logln("error parsing command: " + err.Error())
		}

		command(id, c, cmd, connMap)
	}
}

func command(id string, conn net.Conn, cmd Request, connMap *sync.Map) {
	switch {
	case cmd.Command == "set":
		if !slices.Contains([]string{"client_polling", "server_events", "manual"}, cmd.UpdateMethod) {
			response, _ := json.Marshal(NewResponse(Response{Status: "error", Error: "Incorrect updateMethod."}))
			conn.Write([]byte(fmt.Sprintf("%s\n", response)))
			return
		}
		oldClient, _ := connMap.Load(id)
		newClient := oldClient.(*Client)
		newClient.updateMethod = cmd.UpdateMethod
		connMap.Store(id, newClient)
		response, _ := json.Marshal(NewResponse(Response{Status: "ok"}))
		conn.Write([]byte(fmt.Sprintf("%s\n", response)))

	case cmd.Command == "get_mouse":
		mice := getMouseInfo()
		data, _ := json.Marshal(mice)
		miceHash := fmt.Sprintf("%x", sha256.Sum256(data))
		client, ok := connMap.Load(id)
		c := client.(*Client)
		if !ok {
			response, _ := json.Marshal(NewResponse(Response{Status: "error", Error: "Client not found"}))
			conn.Write([]byte(fmt.Sprintf("%s\n", response)))
			return
		}
		if c.lastResponse == miceHash {
			response, _ := json.Marshal(NewResponse(Response{Status: "no changes"}))
			c.conn.Write([]byte(fmt.Sprintf("%s\n", response)))
			return
		}
		response, _ := json.Marshal(NewResponse(Response{Status: "ok", Data: mice}))
		c.conn.Write([]byte(fmt.Sprintf("%s\n", response)))
		c.lastResponse = fmt.Sprintf("%x", sha256.Sum256(data))

	case cmd.Command == "get_process":
		process, err := getProcessInfo()
		if err != nil {
			response, _ := json.Marshal(NewResponse(Response{Status: "error", Error: err.Error()}))
			conn.Write([]byte(fmt.Sprintf("%s\n", response)))
			return
		}
		data, _ := json.Marshal(process)
		processHash := fmt.Sprintf("%x", sha256.Sum256(data))
		client, ok := connMap.Load(id)
		c := client.(*Client)
		if !ok {
			response, _ := json.Marshal(NewResponse(Response{Status: "error", Error: "Client not found."}))
			conn.Write([]byte(fmt.Sprintf("%s\n", response)))
			return
		}
		if c.lastResponse == processHash {
			response, _ := json.Marshal(NewResponse(Response{Status: "no changes"}))
			c.conn.Write([]byte(fmt.Sprintf("%s\n", response)))
			return
		}
		response, _ := json.Marshal(NewResponse(Response{Status: "ok", Data: process}))
		c.conn.Write([]byte(fmt.Sprintf("%s\n", response)))
		c.lastResponse = fmt.Sprintf("%x", sha256.Sum256(data))

	default:
		response, _ := json.Marshal(NewResponse(Response{Status: "error", Error: "Incorrect command."}))
		conn.Write([]byte(fmt.Sprintf("%s\n", response)))
		return
	}
}

func pidFileExists() bool {
	//pidFilePath := PID_PREFIX + "/collector_" + collectorRole + ".pid"
	pidFilePath := "collector_" + collectorRole + ".pid"

	if _, err := os.Stat(pidFilePath); err == nil {
		log.Fatalf("Collector with role %s already launched\n", collectorRole)
	} else if !os.IsNotExist(err) {
		log.Fatalf("PID file check error: %v\n", err)
	}

	pid := os.Getpid()

	file, err := os.OpenFile(pidFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Error creating PID file: %v\n", err)
	}
	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(pid))
	if err != nil {
		log.Fatalf("Error writing PID to file: %v\n", err)
	}
	return false
}

func deletePidFile() {
	//pidFilePath := PID_PREFIX + "/collector_" + collectorRole + ".pid"
	pidFilePath := "collector_" + collectorRole + ".pid"
	if err := os.Remove(pidFilePath); err != nil {
		log.Panicf("Error deleting PID file: %s\n", err.Error())
	}
}
