package main

import (
	"os"
	"syscall"
)

type Process struct {
	Pid    int
	Uptime int64
}

func getProcessInfo() (Process, error) {
	var usage syscall.Rusage
	err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage)
	userTimeSec := usage.Utime.Sec   // Секунды
	userTimeUsec := usage.Utime.Usec // Микросекунды
	totalMillis := userTimeSec*1000 + int64(userTimeUsec)/1000
	return Process{Pid: os.Getpid(), Uptime: totalMillis}, err
}
