package main

import (
	"github.com/Mrs4s/go-cqhttp/coolq"
	"github.com/tristan-club/kit/log"
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
		log.Info().Msgf("shut down,sign:%v", sig)
		coolq.SentNews.SaveCache()
		log.Info().Msgf("save wall street cache to file")
	}()
}
