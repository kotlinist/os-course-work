package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	pipePath string
	pipe     *os.File
}

func (logger *Logger) Close() {
	logger.pipe.Close()
}

func (logger *Logger) Init() {
	logger.pipePath = rootPipePath + "/collector-log-" + collectorRole + ".pipe"
	//fmt.Println(logger.pipePath)
	//if _, err := os.Stat(logger.pipePath); err == nil {
	//	log.Printf("Pipe %s уже существует. Удаляем его.", logger.pipePath)
	//	if err := os.Remove(logger.pipePath); err != nil {
	//		log.Fatalf("Не удалось удалить существующий pipe: %v", err)
	//	}
	//}
	//
	//err2 := syscall.Mkfifo(logger.pipePath, 0666)
	//if err2 != nil {
	//	log.Fatalf("Ошибка создания именованного канала: %v", err2)
	//}
	//
	//log.Printf("Создан именованный pipe: %s", logger.pipePath)

	var err error
	fmt.Println(logger.pipePath)
	logger.pipe, err = os.OpenFile(logger.pipePath, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		log.Fatalf("Ошибка открытия "+logger.pipePath+" для записи: %v", err)
	}
}

func (logger *Logger) Log(msg string) {
	var err error
	message := time.Now().Format(time.RFC3339) + "> " + msg
	fmt.Print(message)
	_, err = logger.pipe.Write([]byte(message))
	if err != nil {
		log.Fatalf("Ошибка записи в pipe: %v", err)
	}
}
func (logger *Logger) Logln(msg string) {
	var err error
	message := time.Now().Format(time.RFC3339) + "> " + msg + "\n"
	fmt.Print(message)
	_, err = logger.pipe.Write([]byte(message))
	if err != nil {
		log.Fatalf("Ошибка записи в pipe: %v", err)
	}
}
