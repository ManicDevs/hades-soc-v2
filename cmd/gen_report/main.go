package main

import (
	"context"
	"fmt"
	"hades-v2/internal/agent"
	"hades-v2/internal/bus"
	"hades-v2/internal/engine"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	ctx := context.Background()
	eventBus := bus.Default()
	dispatcher := engine.NewDispatcher(nil)
	orch := agent.NewOrchestrator(eventBus, dispatcher, nil)
	if err := orch.Start(ctx); err != nil {
		log.Printf("Failed to start orchestrator: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Initialize Continuous Deception - Deploy honey files for baseline
	fmt.Println("Initializing Continuous Deception...")
	honeyManager, err := agent.NewHoneyFileManager(eventBus, nil)
	if err != nil {
		log.Printf("Failed to create honey manager: %v", err)
	} else {
		if err := honeyManager.DeployHoneyFiles(); err != nil {
			log.Printf("Failed to deploy honey files: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("Generating Daily Report...")
	report := orch.GenerateDailyReadinessReport()

	if err := os.MkdirAll("reports", 0755); err != nil {
		log.Printf("Failed to create reports directory: %v", err)
		return
	}
	content := report.ToMarkdown()
	timestamp := time.Now().Format("20060102")
	if err := os.WriteFile("reports/daily_report_"+timestamp+".md", []byte(content), 0644); err != nil {
		log.Printf("Failed to write report file: %v", err)
	}
	if err := os.WriteFile("reports/daily_report_latest.md", []byte(content), 0644); err != nil {
		log.Printf("Failed to write latest report: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 79))
	fmt.Println(" DAILY READINESS REPORT - AUDIT")
	fmt.Println(strings.Repeat("=", 79))
	fmt.Println(content)
}
