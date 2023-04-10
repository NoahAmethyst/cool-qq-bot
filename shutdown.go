package main

import (
	"github.com/Mrs4s/go-cqhttp/util/top_list"
	"github.com/tristan-club/kit/log"
	"os"
	"os/signal"
	"syscall"
)

// gracefulShutdown waits for termination syscalls and doing clean up operations after received it
func gracefulShutdown() {
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGSTOP, syscall.SIGKILL, syscall.SIGHUP)
	go func() {
		sig := <-signalChannel
		log.Info().Msgf("shut down,sign:%v", sig)
		top_list.SentNews.SaveCache()
	}()
}
