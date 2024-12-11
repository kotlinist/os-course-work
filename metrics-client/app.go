package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"regexp"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx                             context.Context
	connMouseCollector              net.Conn
	connProcessCollector            net.Conn
	connectedToMouseCollector       chan bool
	goCancelMouseCollector          context.CancelFunc
	connectedToPrecessCollector     chan bool
	goCancelProcessCollector        context.CancelFunc
	goCancelMouseCollectorPolling   context.CancelFunc
	goCancelProcessCollectorPolling context.CancelFunc
}

type ConnArgs struct {
	host         string
	port         int
	updateMethod string
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

func NewApp() *App {
	app := &App{connectedToMouseCollector: make(chan bool), connectedToPrecessCollector: make(chan bool)}
	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	//a.connected <- false
	//runtime.EventsOn(a.ctx, "mouseCollectorConnect", func(data ...interface{}) {
	//	runtime.LogPrint(a.ctx, "ConnectEvent")
	//	connArgs := data[0].(ConnArgs)
	//	runtime.LogPrint(a.ctx, connArgs.host)
	//	runtime.LogPrint(a.ctx, strconv.Itoa(connArgs.port))
	//	a.Connect(connArgs.host, connArgs.port, connArgs.updateMethod)
	//})
	//
	//runtime.EventsOn(a.ctx, "disconnect", func(data ...interface{}) {
	//	runtime.LogPrint(a.ctx, "DisconnectEvent")
	//	a.Disconnect()
	//})
}

func (a *App) shutdown(ctx context.Context) {
	a.Disconnect("mouse")
	a.Disconnect("process")
}

func (a *App) Connect(role string, host string, port int, updateMethod string) {
	dest := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", dest)

	if err != nil {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{Type: runtime.ErrorDialog,
			Title:   "Error connecting to server",
			Message: err.Error()})
		//if _, t := err.(*net.OpError); t {
		//	runtime.LogPrint(a.ctx, "Some problem connecting.")
		//} else {
		//	runtime.LogPrint(a.ctx, "Unknown error: "+err.Error())
		//}
		return
		//os.Exit(1)
	}
	if role == "mouse" {
		a.connMouseCollector = conn
	}
	if role == "process" {
		a.connProcessCollector = conn
	}
	runtime.LogPrint(a.ctx, "Connected to "+host)
	runtime.EventsEmit(a.ctx, role+"CollectorConnected", true)
	//a.connected <- true

	request, _ := json.Marshal(Request{Command: "set", UpdateMethod: updateMethod})
	_, err = conn.Write([]byte(fmt.Sprintf("%s\n", request)))
	if err != nil {
		fmt.Println("Error writing to stream.")
	}

	if role == "mouse" {
		_, a.goCancelMouseCollector = context.WithCancel(context.Background())
	}
	if role == "process" {
		_, a.goCancelProcessCollector = context.WithCancel(context.Background())
	}
	go a.readConnection(role, conn)

	if updateMethod == "client_polling" {
		var ctx context.Context
		if role == "mouse" {
			ctx, a.goCancelMouseCollectorPolling = context.WithCancel(context.Background())
			go a.clientPollingMouse(ctx)
		}
		if role == "process" {
			ctx, a.goCancelProcessCollectorPolling = context.WithCancel(context.Background())
			go a.clientPollingProcess(ctx)
		}
	}
}

func (a *App) Disconnect(role string) {
	var err error
	if role == "mouse" {
		if a.goCancelMouseCollectorPolling != nil {
			a.goCancelMouseCollectorPolling()
		}
		a.goCancelMouseCollector()
		err = a.connMouseCollector.Close()
	}
	if role == "process" {
		if a.goCancelProcessCollectorPolling != nil {
			a.goCancelProcessCollectorPolling()
		}
		a.goCancelProcessCollector()
		err = a.connProcessCollector.Close()
	}
	if err != nil {
		return
	}
	runtime.EventsEmit(a.ctx, role+"CollectorConnected", false)
	runtime.LogPrint(a.ctx, "Disconnected")
	//a.connected <- false
}

func (a *App) send(text string, conn net.Conn) {
	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	_, err := conn.Write([]byte(text))
	if err != nil {
		fmt.Println("Error writing to stream.")
	}
}

func (a *App) readConnection(role string, conn net.Conn) {
	fmt.Println("READ_CONN")
	for {
		//if <-a.connected != true {
		//	return
		//}
		scanner := bufio.NewScanner(conn)
		for {
			ok := scanner.Scan()
			text := scanner.Text()

			runtime.LogPrint(a.ctx, "Response: "+text)
			r := &Response{}
			err := json.Unmarshal([]byte(text), r)
			//responseObj := response.(Response)
			if err != nil {
				return
			}
			runtime.EventsEmit(a.ctx, role+"CollectorReceiveData", r)

			//command := a.handleCommands(text)
			//if !command {
			//	runtime.LogPrint(a.ctx, "\b\b** %s\n> "+text)
			//}

			if !ok {
				runtime.LogPrint(a.ctx, "Reached EOF on server connection.")
				return
				//break
			}
		}
	}
}

func (a *App) handleCommands(text string) bool {
	r, err := regexp.Compile("^%.*%$")
	if err == nil {
		if r.MatchString(text) {
			switch {
			case text == "%quit%":
				runtime.LogPrint(a.ctx, "\b\bServer is leaving. Hanging up.")
				os.Exit(0)
			}
			return true
		}
	}
	return false
}

func (a *App) clientPollingMouse(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			a.GetDataManually("mouse")
			time.Sleep(1000 * time.Millisecond)
		}
	}
}

func (a *App) clientPollingProcess(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			a.GetDataManually("process")
			time.Sleep(1000 * time.Millisecond)
		}
	}
}

func (a *App) GetDataManually(role string) {
	request, _ := json.Marshal(Request{Command: "get_" + role})
	var err error
	if role == "mouse" {
		_, err = a.connMouseCollector.Write([]byte(fmt.Sprintf("%s\n", request)))
	} else if role == "process" {
		_, err = a.connProcessCollector.Write([]byte(fmt.Sprintf("%s\n", request)))
	}
	if err != nil {
		fmt.Println("Error writing to stream.")
	}
}
