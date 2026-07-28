package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/p-mon/agent/internal/collector"
	"github.com/p-mon/agent/internal/config"
	"github.com/p-mon/agent/internal/sender"
)

// Version is injected at build time
var Version = "dev"

func main() {
	configPath := flag.String("config", "", "Path to config file")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("P-mon Agent v%s\n", Version)
		os.Exit(0)
	}

	// 1. Setup minimal logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix("[agent] ")

	fmt.Printf("P-mon Agent v%s — starting\n", Version)

	// 2. Load config
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Fatal config error: %v", err)
	}

	// Setup log file if requested
	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file %q: %v", cfg.LogFile, err)
		}
		defer f.Close()
		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}

	log.Printf("Config loaded successfully! Backend: %s | ServerKey: %s", cfg.BackendURL, cfg.MaskedKey())
	log.Printf("Interval: %ds | Collectors: System:%v Docker:%v PM2:%v Services:%v",
		cfg.IntervalSeconds, cfg.Collect.System, cfg.Collect.Docker, cfg.Collect.PM2, len(cfg.Collect.Services))

	// 3. Graceful shutdown handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 4. Main collection loop
	ticker := time.NewTicker(time.Duration(cfg.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	// Initial forced collection (so we don't wait N seconds for the first payload)
	runCycle(cfg)

	for {
		select {
		case sig := <-sigChan:
			log.Printf("Received signal %v, shutting down gracefully", sig)
			return
		case <-ticker.C:
			runCycle(cfg)
		}
	}
}

func runCycle(cfg *config.Config) {
	start := time.Now()

	// Collect
	payload := collector.Collect(cfg, Version)

	// Send (with 3 retries max)
	err := sender.SendWithRetry(cfg, payload, 3)

	elapsed := time.Since(start).Round(time.Millisecond)

	if err != nil {
		log.Printf("ERROR: Failed to send metrics after retries: %v (took %s)", err, elapsed)
	} else {
		log.Printf("OK: Metrics sent successfully for %q (took %s)", payload.Hostname, elapsed)
	}
}
