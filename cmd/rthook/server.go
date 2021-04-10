package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
)

type server struct {
	port    int
	handler http.HandlerFunc
}

// Starts a server and waits for SIGINT or SIGTERM
func (s *server) start() {
	server := &http.Server{Addr: fmt.Sprintf(":%v", s.port), Handler: s.handler}
	go func() { server.ListenAndServe() }()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	for {
		select {
		case sig := <-signals:
			switch sig {
			case os.Interrupt:
				fmt.Println("Received SIGINT")
				server.Shutdown(context.Background())
				return
			case os.Kill:
				fmt.Println("Received SIGTERM")
				server.Close()
				return
			default:
				fmt.Println("Received unknown signal:", s)
			}
		}
	}
}
