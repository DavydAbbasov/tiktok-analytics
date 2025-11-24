package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"ttanalytic/internal/application"
	_ "ttanalytic/internal/api/docs"
)

func main() {

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	app := application.NewApplication()

	if err := app.Start(ctx); err != nil {
		log.Fatalln("can't start application:", err)
	}
	if err := app.Wait(ctx, cancel); err != nil {
		log.Fatalln("All systems closed with errors. LastError:", err)
	}

	log.Println("All systems closed without errors")
}
