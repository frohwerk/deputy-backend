package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/lib/pq"
)

func main() {
	listener := database.NewListener()
	defer listener.Close()
	listener.Listen("demo_channel")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill)

	handle := func(n *pq.Notification) {
		fmt.Println("Incoming on channel", n.Channel)
		fmt.Println(">>", n.Extra)
	}

	fmt.Println("Listening on demo_channel")
	for {
		select {
		case n := <-listener.Notify:
			go handle(n)
		case <-sigs:
			return
		case <-time.After(1 * time.Minute):
			go listener.Ping()
		}
	}
}
