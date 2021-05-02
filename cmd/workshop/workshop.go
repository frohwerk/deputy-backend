package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/frohwerk/deputy-backend/cmd/workshop/task"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	tasks := []task.Task{task.CreateTask(1, work), task.CreateTask(2, work), task.CreateTask(3, work)}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill)

	for _, task := range tasks {
		go task.Start()
	}

	for {
		switch s := <-sigs; s {
		case os.Interrupt:
			log.Println("SIGINT")
			for _, t := range tasks {
				t.Stop()
			}
			for _, t := range tasks {
				t.Wait()
			}
			os.Exit(0)
		case os.Kill:
			log.Println("SIGTERM")
			os.Exit(0)
		}
	}
}

func work(cancel <-chan interface{}) error {
	if rand.Intn(1) > 0 {
		return fmt.Errorf("crash on startup")
	}
	ticker := time.NewTicker(time.Duration(1) * time.Second)
	defer ticker.Stop()
	for i := 0; i < rand.Intn(15)+1; i++ {
		select {
		case <-cancel:
			return fmt.Errorf("got canceled")
		case <-ticker.C:
		}
		log.Printf("working hard: %d", i)
		if rand.Intn(15) > 13 {
			return fmt.Errorf("crash during work")
		}
	}
	return nil
}
