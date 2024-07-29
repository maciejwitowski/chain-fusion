package main

import (
	"context"
	"github.com/spf13/cobra"
	"log"
	"os/signal"
	"syscall"
)

func ChainCommand() {
	cmd := cobra.Command{
		Use:   "chain",
		Short: "Application for connecting to various chains",
		Args:  cobra.NoArgs,
	}

	cfg, err := InitConfig()
	if err != nil {
		log.Fatal("Error initialising config")
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return Run(cmd.Context(), cfg)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	err = cmd.ExecuteContext(ctx)
	cancel()
}
