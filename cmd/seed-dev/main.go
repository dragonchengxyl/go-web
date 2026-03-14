package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/cli/seeder"
)

func main() {
	configFile := flag.String("config", "configs/config.local.yaml", "path to config file")
	flag.Parse()

	cfg, err := configs.Load(*configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := seeder.SeedDemo(context.Background(), cfg, os.Stdout); err != nil {
		log.Fatalf("failed to seed demo data: %v", err)
	}
}
