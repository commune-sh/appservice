package main

import (
	"commune/app"
	"flag"
	"log"
	"os"
	"os/signal"
)

func main() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)

	go func() {
		<-sc

		log.Println("Shutting down server")
		os.Exit(1)
	}()

	config := flag.String("config", "config.toml", "Commune configuration file")

	flag.Parse()

	req := &app.StartRequest{
		Config: *config,
	}

	if len(os.Args) > 1 {
		command := os.Args[1]

		switch command {
		case "join":
			req.JoinPublicRooms = true
		}
	}

	app.Start(req)
}
