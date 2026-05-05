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
	orch.Start(ctx)
	time.Sleep(500 * time.Millisecond)

	// Initialize Continuous Deception - Deploy honey files for baseline
	fmt.Println("Initializing Continuous Deception...")
	honeyManager, err := agent.NewHoneyFileManager(eventBus, nil)
	if err != nil {
		log.Printf("Failed to create honey manager: %v", err)
	} else {
		honeyManager.DeployHoneyFiles()
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("Generating Daily Report...")
	report := orch.GenerateDailyReadinessReport()

	os.MkdirAll("reports", 0755)
	content := report.ToMarkdown()
	timestamp := time.Now().Format("20060102")
	os.WriteFile("reports/daily_report_"+timestamp+".md", []byte(content), 0644)
	os.WriteFile("reports/daily_report_latest.md", []byte(content), 0644)

	fmt.Println("\n" + strings.Repeat("=", 79))
	fmt.Println(" DAILY READINESS REPORT - AUDIT")
	fmt.Println(strings.Repeat("=", 79))
	fmt.Println(content)
}
