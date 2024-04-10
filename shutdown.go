package main

import (
	"github.com/NoahAmethyst/go-cqhttp/coolq"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

// gracefulShutdown waits for termination syscalls and doing clean up operations after received it
func gracefulShutdown() {
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGSTOP, syscall.SIGKILL, syscall.SIGHUP)
	go func() {
		sig := <-signalChannel
		coolq.ShutdownNotify <- struct{}{}
		log.Infof("shut down,sign:%+v", sig)

	}()
}
