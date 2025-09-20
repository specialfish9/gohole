package main

import (
	"fmt"
	"gohole/config"
	"gohole/internal/blocklist"
	"gohole/internal/controller/dns"
	"gohole/internal/controller/http"
	"gohole/internal/database"
	"gohole/internal/registry"
	"log/slog"
	"os"
	"sync"
)

const defaultConfigPath = "./gohole.conf"

func logPanic(v any) {
	slog.Error(fmt.Sprintf("%v", v))
	fmt.Println("Bye :O")
	os.Exit(1)
}

func main() {
	fmt.Println("========")
	fmt.Println(" GOHOLE ")
	fmt.Println("========")

	var configPath string
	if len(os.Args) > 1 {
		// The first argument is the config path
		configPath = os.Args[1]
	} else {
		configPath = defaultConfigPath
	}

	cfg, err := config.New(configPath)
	if err != nil {
		logPanic(err.Error())
	}

	slog.Debug("log file is set", "file", cfg.LogFile)
	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logPanic(err.Error())
		}
		defer f.Close()
		slog.SetDefault(slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{AddSource: true})))
	}

	// TODO handle log level from config
	slog.SetLogLoggerLevel(slog.LevelDebug)

	dbConn, err := database.Connect(
		cfg.DBAddress,
		cfg.DBName,
		cfg.DBUser,
		cfg.DBPassword,
		false,
	)
	// if err != nil {
	// 	logPanic(err.Error())
	// }

	slog.Info("Connected to DB")

	// if err := database.Init(context.Background(), dbConn); err != nil {
	// 	logPanic(err.Error())
	// }

	slog.Info("created tables")

	domains, err := blocklist.ReadFromFile(cfg.BlocklistFile)
	if err != nil {
		logPanic(err.Error())
	}

	reg := registry.NewRegistry(domains, dbConn)

	wg := sync.WaitGroup{}

	go dns.Start(&wg, reg, cfg.DNSAddress, cfg.Upstream)
	wg.Add(1)
	go http.Start(&wg, reg, cfg.HTTPAddress, cfg.ServeFrontend)
	wg.Add(1)

	wg.Wait()
}
