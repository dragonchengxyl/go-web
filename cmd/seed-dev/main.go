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
	mode := flag.String("mode", "demo", "seed mode: demo or bulk")
	profile := flag.String("profile", "medium", "bulk seed profile: small, medium, large")
	namespace := flag.String("namespace", "bulk", "namespace prefix used for bulk seed determinism")
	seed := flag.Int64("seed", 0, "bulk seed random seed override; 0 derives from namespace")
	flag.Parse()

	cfg, err := configs.Load(*configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	switch *mode {
	case "demo":
		if err := seeder.SeedDemo(context.Background(), cfg, os.Stdout); err != nil {
			log.Fatalf("failed to seed demo data: %v", err)
		}
	case "bulk":
		if err := seeder.SeedBulk(context.Background(), cfg, seeder.BulkSeedOptions{
			Profile:   *profile,
			Namespace: *namespace,
			Seed:      *seed,
		}, os.Stdout); err != nil {
			log.Fatalf("failed to seed bulk data: %v", err)
		}
	default:
		log.Fatalf("unsupported seed mode: %s", *mode)
	}
}
