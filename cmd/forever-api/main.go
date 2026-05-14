package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"hades-v2/internal/api"
)

func main() {
	var (
		requestInterval = flag.Duration("interval", 30*time.Second, "Interval between requests")
		maxRequests     = flag.Int("max", 0, "Maximum requests to make (0 = infinite)")
		monitorInterval = flag.Duration("monitor", 2*time.Minute, "Health monitoring interval")
		promptFile      = flag.String("prompts", "", "File with prompts (one per line)")
		verbose         = flag.Bool("verbose", false, "Verbose logging")
	)
	flag.Parse()

	log.Printf("🚀 Starting Forever API - Free Tier Maximizer")
	log.Printf("📊 Request interval: %v", *requestInterval)
	log.Printf("🔍 Health monitoring: %v", *monitorInterval)
	if *maxRequests > 0 {
		log.Printf("🎯 Max requests: %d", *maxRequests)
	} else {
		log.Printf("♾️  Infinite mode: ON")
	}

	// Create orchestrator
	orchestrator := api.NewQuotaOrchestrator()

	// Start health monitoring
	orchestrator.StartContinuousMonitoring(*monitorInterval)

	// Get prompts
	prompts := getPrompts(*promptFile)
	if len(prompts) == 0 {
		prompts = getDefaultPrompts()
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start request loop
	go func() {
		requestCount := 0
		promptIndex := 0

		ticker := time.NewTicker(*requestInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if *maxRequests > 0 && requestCount >= *maxRequests {
					log.Printf("🎯 Reached max requests limit (%d)", *maxRequests)
					cancel()
					return
				}

				prompt := prompts[promptIndex%len(prompts)]
				promptIndex++

				err := makeRequest(ctx, orchestrator, prompt, requestCount, *verbose)
				if err != nil {
					log.Printf("❌ Request %d failed: %v", requestCount, err)
				} else {
					requestCount++
					if *verbose || requestCount%10 == 0 {
						log.Printf("✅ Completed request #%d", requestCount)
					}
				}
			}
		}
	}()

	// Status reporter
	go func() {
		statusTicker := time.NewTicker(5 * time.Minute)
		defer statusTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-statusTicker.C:
				printStatus(orchestrator)
			}
		}
	}()

	// Wait for shutdown
	<-sigChan
	log.Printf("🛑 Received shutdown signal")
	cancel()

	// Final status report
	printStatus(orchestrator)
	log.Printf("👋 Forever API stopped")
}

func makeRequest(ctx context.Context, orchestrator *api.QuotaOrchestrator, prompt string, requestNum int, verbose bool) error {
	if verbose {
		log.Printf("🔄 Making request #%d: %s", requestNum, prompt[:min(50, len(prompt))])
	}

	response, err := orchestrator.MakeRequest(ctx, prompt)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if verbose {
		log.Printf("📝 Response: %s", response[:min(100, len(response))])
	}

	return nil
}

func printStatus(orchestrator *api.QuotaOrchestrator) {
	status := orchestrator.GetStatus()

	log.Printf("📊 === STATUS REPORT ===")
	log.Printf("Current Provider: %v", status["current_provider"])
	log.Printf("Total Rotations: %v", status["rotation_count"])
	log.Printf("Next Reset: %v", status["next_daily_reset"])

	if providers, ok := status["providers"].(map[string]interface{}); ok {
		for provider, config := range providers {
			if configMap, ok := config.(map[string]interface{}); ok {
				remaining := configMap["remaining"].(int)
				usage := configMap["current_usage"].(int)
				limit := configMap["daily_limit"].(int)
				healthy := configMap["is_healthy"].(bool)

				status := "✅"
				if !healthy {
					status = "🚫"
				} else if remaining <= 0 {
					status = "❌"
				} else if remaining < 5 {
					status = "⚠️"
				}

				log.Printf("  %s %s: %d/%d (%d remaining) %s",
					status, provider, usage, limit, remaining,
					map[bool]string{true: "Healthy", false: "Unhealthy"}[healthy])
			}
		}
	}
	log.Printf("📊 ===================")
}

func getPrompts(filename string) []string {
	if filename == "" {
		return nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("⚠️  Could not read prompt file %s: %v", filename, err)
		return nil
	}

	prompts := splitLines(string(data))
	log.Printf("📚 Loaded %d prompts from %s", len(prompts), filename)
	return prompts
}

func getDefaultPrompts() []string {
	return []string{
		"What is the current time?",
		"Explain quantum computing in one sentence",
		"Write a haiku about programming",
		"What are the benefits of Go programming language?",
		"Explain machine learning simply",
		"Write a joke about computers",
		"What is the meaning of life?",
		"Explain blockchain technology",
		"Write a short story about AI",
		"What makes good software architecture?",
		"Explain the concept of recursion",
		"Write a poem about technology",
		"What is cloud computing?",
		"Explain API design principles",
		"Write about the future of AI",
		"What is cybersecurity?",
		"Explain database normalization",
		"Write about software testing",
		"What is DevOps?",
		"Explain microservices architecture",
	}
}

func splitLines(text string) []string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
