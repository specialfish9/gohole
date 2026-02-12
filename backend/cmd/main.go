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
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

const defaultConfigPath = "./gohole.yaml"
const panicFilePath = "./panic.log"
const dbConnectionAttempts = 10

// logPanic logs the panic message to both slog and a file, then exits the program.
func logPanic(v any) {
	msg := fmt.Sprintf("panic: %v", v)
	slog.Error(msg)

	f, err := os.OpenFile(panicFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("Could not open panic.log:" + err.Error())
	} else {
		defer f.Close()
		if _, err := f.Write([]byte(msg + "\n")); err != nil {
			slog.Error("Could not write to panic.log:" + err.Error())
		}
	}

	fmt.Println("Bye :O")
	os.Exit(1)
}

func connectToDB(cfg *config.Config) (driver.Conn, error) {
	var dbConn driver.Conn
	var err error

	for i := range dbConnectionAttempts {
		dbConn, err = database.Connect(
			cfg.DB.Address,
			cfg.DB.Name,
			cfg.DB.User,
			cfg.DB.Password,
			cfg.App.LogLevel == "debug", // TODO handle log level from config
		)

		if err != nil {
			slog.Error(fmt.Sprintf("DB connection attempt %d failed: %v", i+1, err))
			time.Sleep(2 * time.Second)
		} else {
			return dbConn, nil
		}
	}

	return nil, fmt.Errorf("failed to connect to DB after %d attempts: %w", dbConnectionAttempts, err)
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

	// TODO handle log level from config
	slog.SetLogLoggerLevel(slog.LevelDebug)

	dbConn, err := connectToDB(cfg)
	if err != nil {
		logPanic(err.Error())
	}

	slog.Info("Connected to DB")

	if err := database.Init(context.Background(), dbConn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("created tables")

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

	reg := registry.NewRegistry(domains, allowDomains, cfg.Blocking.FilterStrategy, dbConn)

	wg := sync.WaitGroup{}

	go dns.Start(&wg, reg, cfg.DNS.Address, cfg.DNS.Upstream)
	wg.Add(1)

	shouldServeFrontend := cfg.HTTP.ServeFrontend.Or(false)
	go http.Start(&wg, reg, cfg.HTTP.Address, shouldServeFrontend)
	wg.Add(1)

	wg.Wait()
}
