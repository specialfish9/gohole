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

const defaultConfigPath = "./gohole.conf"
const panicFilePath = "./panic.log"
const dbConnectionAttempts = 10

// logPanic logs the panic message to both slog and a file, then exits the program.
func logPanic(v any) {
	msg := fmt.Sprintf("panic: %v", v)
	slog.Error(msg)

	f, err := os.OpenFile(panicFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("Could not open panic.log:", err)
	} else {
		defer f.Close()
		if _, err := f.Write([]byte(msg + "\n")); err != nil {
			slog.Error("Could not write to panic.log:", err)
		}
	}

	fmt.Println("Bye :O")
	os.Exit(1)
}

func connectToDB(cfg *config.Config) (driver.Conn, error) {
	var dbConn driver.Conn
	var err error

	for i := 0; i < dbConnectionAttempts; i++ {
		dbConn, err = database.Connect(
			cfg.DBAddress,
			cfg.DBName,
			cfg.DBUser,
			cfg.DBPassword,
			false,
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

	domains, err := blocklist.LoadRemote(cfg.BlocklistFile)
	if err != nil {
		logPanic(err.Error())
	}

	if cfg.LocalBlockList != "" {
		localDomains, err := blocklist.LoadLocalFile(cfg.LocalBlockList)
		if err != nil {
			logPanic(err.Error())
		}
		domains = append(domains, localDomains...)
	}

	var allowDomains []string
	if cfg.LocalAllowList != "" {
		allowDomains, err = blocklist.LoadLocalFile(cfg.LocalAllowList)
		if err != nil {
			logPanic(err.Error())
		}
	}

	reg := registry.NewRegistry(domains, allowDomains, dbConn)

	wg := sync.WaitGroup{}

	go dns.Start(&wg, reg, cfg.DNSAddress, cfg.Upstream)
	wg.Add(1)
	go http.Start(&wg, reg, cfg.HTTPAddress, cfg.ServeFrontend)
	wg.Add(1)

	wg.Wait()
}
