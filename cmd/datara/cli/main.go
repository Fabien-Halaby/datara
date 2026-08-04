package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Fabien-Halaby/datara/internal/audit"
	"github.com/Fabien-Halaby/datara/internal/config"
	"github.com/Fabien-Halaby/datara/internal/core/domain"
	"github.com/Fabien-Halaby/datara/internal/datasource/postgres"
	"github.com/Fabien-Halaby/datara/internal/security/astvalidator"
	"github.com/Fabien-Halaby/datara/internal/transport/mcpstdio"
)

func main() {
	// All diagnostic logging goes to stderr: stdout is reserved for the
	// MCP JSON-RPC stream and must never contain anything else.
	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("datara: config error: %v", err)
	}

	ds, err := postgres.New(ctx, cfg.PostgresDSN, cfg.MaxRows)
	if err != nil {
		log.Fatalf("datara: failed to connect to postgres: %v", err)
	}
	defer ds.Close()

	policy := domain.DefaultReadOnlyPolicy()
	policy.MaxRows = cfg.MaxRows

	validator := astvalidator.NewPostgresValidator(policy)
	auditor := audit.NewStderrLogger()

	server := mcpstdio.New(os.Stdin, os.Stdout, validator, ds, auditor)

	log.Println("datara: ready, listening on stdio")
	if err := server.Run(ctx); err != nil {
		log.Fatalf("datara: server error: %v", err)
	}
}