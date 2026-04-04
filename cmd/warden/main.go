package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"server-monitor/internal/app"
	"server-monitor/internal/config"
)

func main() {
	cfg := config.Parse()
	if err := cfg.Validate(); err != nil {
		fmt.Println(err.Error())
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app.Run(ctx, cfg)
}
