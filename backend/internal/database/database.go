package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Manager interface {
	Connect(ctx context.Context) error
	Init(ctx context.Context) error
	Repository() Repository
}

func Connect(ctx context.Context, manager Manager, config *Config, attempts int) error {
	for i := range attempts {
		err := manager.Connect(ctx)
		if err != nil {
			slog.Error("DB connection attempt failed", "attempt", i+1, "error", err)
			time.Sleep(2 * time.Second)
		} else {
			return nil
		}
	}

	return fmt.Errorf("db: unable to connect to database after %d attempts", attempts)
}
