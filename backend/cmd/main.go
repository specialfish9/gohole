package main

import (
	"context"
	"fmt"
	"gohole/config"
	"gohole/internal/blocklist"
	"gohole/internal/controller/dns"
	"gohole/internal/controller/http"
	"gohole/internal/database"
	"gohole/internal/registry"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const defaultConfigPath = "./gohole.yaml"

func logPanic(v any) {
	msg := fmt.Sprintf("panic: %v", v)
	slog.Error(msg)
	fmt.Println("Bye :O")
	os.Exit(1)
}

func initLogger(cfg *config.Config) {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.NewLeveler(cfg.App.LogLevel),
	})
	slog.SetDefault(slog.New(handler))
}

func main() {
	fmt.Println("=========")
	fmt.Println(" GOHOLE! ")
	fmt.Println("=========")

	var configPath string
	if len(os.Args) > 1 {
		// The first argument is the config path
		configPath = os.Args[1]
	} else {
		configPath = defaultConfigPath
	}

	cfg, err := config.New(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v", err)
		fmt.Fprintf(os.Stderr, "Bye :O")
		os.Exit(1)
	}

	initLogger(cfg)

	dbConn, err := database.Connect(&cfg.DB, 5)
	if err != nil {
		logPanic(err.Error())
	}

	slog.Info("Connected to DB")

	if err := database.Init(context.Background(), dbConn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Created tables")

	domains, err := blocklist.LoadRemote(cfg.Blocking.BlocklistFile)
	if err != nil {
		logPanic(err.Error())
	}

	if cfg.Blocking.LocalBlockList.Ok {
		localDomains, err := blocklist.LoadLocalFile(cfg.Blocking.LocalBlockList.Value)
		if err != nil {
			logPanic(err.Error())
		}
		domains = append(domains, localDomains...)
	}

	var allowDomains []string
	if cfg.Blocking.LocalAllowList.Ok {
		allowDomains, err = blocklist.LoadLocalFile(cfg.Blocking.LocalAllowList.Value)
		if err != nil {
			logPanic(err.Error())
		}
	}

	reg := registry.NewRegistry(domains, allowDomains, cfg.Blocking.FilterStrategy, dbConn, cfg)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM) // SIGINT, SIGTERM

	daemons := []Daemon{
		http.NewServer(&cfg.HTTP, reg.QueryRouter),
		dns.NewServer(&cfg.DNS, reg.DNSHandler),
	}

	for _, d := range daemons {
		go func(d Daemon) {
			if err := d.Start(); err != nil {
				logPanic(fmt.Sprintf("Starting daemon %s: %v", d.ID(), err))
			}
		}(d)
	}

	<-quit

	slog.Info("Shutting down servers…")
	for _, d := range daemons {
		if err := d.Stop(); err != nil {
			slog.Error(fmt.Sprintf("Stopping daemon %s: %v", d.ID(), err))
		}
	}

	slog.Info("Bye :O")
}
