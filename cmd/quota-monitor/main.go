package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hades-v2/internal/api"
)

func main() {
	var (
		interval = flag.Duration("interval", 5*time.Minute, "Monitoring interval")
		once     = flag.Bool("once", false, "Run once and exit")
		status   = flag.Bool("status", false, "Show current status and exit")
	)
	flag.Parse()

	// Create quota manager
	quotaManager := api.NewQuotaManager()

	// Create quota monitor
	monitor := api.NewQuotaMonitor(quotaManager)

	if *status {
		// Just show current status
		log.Println(monitor.GetStatusString())
		return
	}

	if *once {
		// Update status once and exit
		monitor.UpdateStatus()
		log.Println("Quota status updated")
		return
	}

	// Start continuous monitoring
	monitor.StartMonitoring(*interval)
	defer monitor.StopMonitoring()

	log.Printf("Started quota monitoring with %v interval", *interval)
	log.Println("Press Ctrl+C to stop")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Received interrupt signal, stopping...")
}
