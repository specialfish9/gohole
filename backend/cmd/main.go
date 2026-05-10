package main

import (
	"context"
	"fmt"
	"gohole/config"
	"gohole/internal/blocklist"
	"gohole/internal/database"
	"gohole/internal/database/clickhouse"
	"gohole/internal/database/pg"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const defaultConfigPath = "./gohole.yaml"

func logPanic(v any) {
	msg := fmt.Sprintf("panic: %v", v)
	slog.Error(msg)
	fmt.Println("\nBye :O")
	os.Exit(1)
}

func initLogger(cfg *config.Config) {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.NewLeveler(cfg.App.LogLevel),
	})
	slog.SetDefault(slog.New(handler))
}

func initDatabase(ctx context.Context, cfg *config.Config) (database.Manager, error) {
	var dbManager database.Manager

	switch cfg.DB.Type {
	case database.TypeClickHouse:
		dbManager = clickhouse.NewManager(&cfg.DB)
	case database.TypePostgres:
		dbManager = pg.NewManager(&cfg.DB)
	case database.TypeNone:
		slog.Info("Database storage is disabled (type: none)")
		return database.NewNoOpManager(), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DB.Type)
	}

	if err := database.Connect(ctx, dbManager, &cfg.DB, 5); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Info("Connected to DB")

	if err := dbManager.Init(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	slog.Info("Initialized DB")

	return dbManager, nil
}

func main() {
	fmt.Println("=========")
	fmt.Println(" GOHOLE! ")
	fmt.Println("=========")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM) // SIGINT, SIGTERM

	var configPath string
	if len(os.Args) > 1 {
		// The first argument is the config path
		configPath = os.Args[1]
	} else {
		configPath = defaultConfigPath
	}

	cfg, err := config.New(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		fmt.Fprintf(os.Stderr, "Bye :O\n")
		os.Exit(1)
	}

	initLogger(cfg)

	db, err := initDatabase(context.Background(), cfg)
	if err != nil {
		logPanic(err)
	}

	domains, err := blocklist.LoadRemote(cfg.Blocking.BlocklistFile)
	if err != nil {
		logPanic(err)
	}

	if cfg.Blocking.LocalBlockList.Ok {
		localDomains, err := blocklist.LoadLocalFile(cfg.Blocking.LocalBlockList.Value)
		if err != nil {
			logPanic(err)
		}
		domains = append(domains, localDomains...)
	}

	var allowDomains []string
	if cfg.Blocking.LocalAllowList.Ok {
		allowDomains, err = blocklist.LoadLocalFile(cfg.Blocking.LocalAllowList.Value)
		if err != nil {
			logPanic(err)
		}
	}

	reg, err := NewDaemonRegistry(domains, allowDomains, cfg.Blocking.FilterStrategy, db, cfg)
	if err != nil {
		logPanic(err)
	}

	reg.Start()

	<-quit

	slog.Info("Shutting down servers…")
	reg.Stop()

	slog.Info("Bye :O")
}
