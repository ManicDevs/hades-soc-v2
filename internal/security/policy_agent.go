package security

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"hades-v2/internal/quantum"
)

type PolicyAgent struct {
	failedLogins    map[string]*LoginAttemptTracker
	quantumEngine   *quantum.CryptographyEngine
	logPath         string
	mu              sync.RWMutex
	stopChan        chan struct{}
	wg              sync.WaitGroup
	maxAttempts     int
	windowDuration  time.Duration
	rotationEnabled bool
}

type LoginAttemptTracker struct {
	Username     string
	Attempts     []time.Time
	LastRotation time.Time
	SessionKeyID string
}

type PolicyAction struct {
	Type       string
	TargetUser string
	Timestamp  time.Time
	Details    string
	Success    bool
}

var (
	defaultPolicyAgent *PolicyAgent
	once               sync.Once
)

func NewPolicyAgent(logPath string, quantumEngine *quantum.CryptographyEngine) *PolicyAgent {
	return &PolicyAgent{
		failedLogins:    make(map[string]*LoginAttemptTracker),
		quantumEngine:   quantumEngine,
		logPath:         logPath,
		stopChan:        make(chan struct{}),
		maxAttempts:     5,
		windowDuration:  10 * time.Minute,
		rotationEnabled: true,
	}
}

func DefaultPolicyAgent() *PolicyAgent {
	once.Do(func() {
		defaultPolicyAgent = NewPolicyAgent("hades.log", nil)
	})
	return defaultPolicyAgent
}

func (pa *PolicyAgent) Start(ctx context.Context) error {
	if pa.logPath == "" {
		pa.logPath = "hades.log"
	}

	log.Printf("PolicyAgent: Starting log monitoring for %s", pa.logPath)

	pa.wg.Add(1)
	go pa.monitorLog(ctx)

	pa.wg.Add(1)
	go pa.cleanupOldAttempts(ctx)

	log.Printf("PolicyAgent: Started with max attempts=%d, window=%v",
		pa.maxAttempts, pa.windowDuration)

	return nil
}

func (pa *PolicyAgent) Stop() {
	close(pa.stopChan)
	pa.wg.Wait()
	log.Println("PolicyAgent: Stopped")
}

func (pa *PolicyAgent) monitorLog(ctx context.Context) {
	defer pa.wg.Done()

	file, err := os.Open(pa.logPath)
	if err != nil {
		log.Printf("PolicyAgent: Failed to open log file: %v", err)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: failed to close log file: %v", err)
		}
	}()

	stat, err := file.Stat()
	if err != nil {
		log.Printf("PolicyAgent: Failed to stat log file: %v", err)
		return
	}

	initialSize := stat.Size()

	reader := bufio.NewReader(file)

	for {
		select {
		case <-pa.stopChan:
			return
		case <-ctx.Done():
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			if int64(len(line)) > initialSize {
				initialSize += int64(len(line))
			}

			pa.processLogLine(line)
		}
	}
}

var (
	failedLoginRegex = regexp.MustCompile(`(?i)(failed|invalid|incorrect).*(login|authentication|password).*user[=:\s]+(\w+)`)
	loginFailRegex   = regexp.MustCompile(`(?i)login\s+failed.*user[=:\s]+(\w+)`)
	authFailRegex    = regexp.MustCompile(`(?i)authentication\s+failed.*user[=:\s]+(\w+)`)
)

func (pa *PolicyAgent) processLogLine(line string) {
	var username string

	if match := loginFailRegex.FindStringSubmatch(line); len(match) > 1 {
		username = match[1]
	} else if match := authFailRegex.FindStringSubmatch(line); len(match) > 1 {
		username = match[1]
	} else if match := failedLoginRegex.FindStringSubmatch(line); len(match) > 2 {
		username = match[2]
	}

	if username != "" {
		pa.recordFailedAttempt(username)
	}
}

func (pa *PolicyAgent) recordFailedAttempt(username string) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	now := time.Now()

	tracker, exists := pa.failedLogins[username]
	if !exists {
		tracker = &LoginAttemptTracker{
			Username: username,
		}
		pa.failedLogins[username] = tracker
	}

	tracker.Attempts = append(tracker.Attempts, now)

	pa.pruneOldAttempts(tracker)

	if len(tracker.Attempts) >= pa.maxAttempts {
		go pa.triggerKeyRotation(username)
	}
}

func (pa *PolicyAgent) pruneOldAttempts(tracker *LoginAttemptTracker) {
	cutoff := time.Now().Add(-pa.windowDuration)

	validAttempts := make([]time.Time, 0)
	for _, attempt := range tracker.Attempts {
		if attempt.After(cutoff) {
			validAttempts = append(validAttempts, attempt)
		}
	}
	tracker.Attempts = validAttempts
}

func (pa *PolicyAgent) triggerKeyRotation(username string) {
	if !pa.rotationEnabled {
		log.Printf("PolicyAgent: Key rotation disabled, skipping for user %s", username)
		return
	}

	if pa.quantumEngine == nil {
		log.Printf("PolicyAgent: No quantum engine configured, skipping key rotation for user %s", username)
		return
	}

	log.Printf("PolicyAgent: Triggering quantum-resistant key rotation for user %s due to %d failed login attempts",
		username, pa.maxAttempts)

	key, err := pa.quantumEngine.GenerateKey("Kyber768", "session")
	if err != nil {
		log.Printf("PolicyAgent: Failed to generate new quantum key for user %s: %v", username, err)
		return
	}

	pa.mu.Lock()
	if tracker, exists := pa.failedLogins[username]; exists {
		tracker.LastRotation = time.Now()
		tracker.SessionKeyID = key.ID
		tracker.Attempts = nil
	}
	pa.mu.Unlock()

	log.Printf("PolicyAgent: Successfully rotated session key for user %s, new key ID: %s", username, key.ID)

	pa.logPolicyAction(PolicyAction{
		Type:       "key_rotation",
		TargetUser: username,
		Timestamp:  time.Now(),
		Details:    fmt.Sprintf("Rotated session key %s after %d failed login attempts", key.ID, pa.maxAttempts),
		Success:    true,
	})
}

func (pa *PolicyAgent) cleanupOldAttempts(ctx context.Context) {
	defer pa.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-pa.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			pa.mu.Lock()
			for username, tracker := range pa.failedLogins {
				pa.pruneOldAttempts(tracker)
				if len(tracker.Attempts) == 0 && time.Since(tracker.LastRotation) > 1*time.Hour {
					delete(pa.failedLogins, username)
				}
			}
			pa.mu.Unlock()
		}
	}
}

func (pa *PolicyAgent) logPolicyAction(action PolicyAction) {
	log.Printf("PolicyAgent: ACTION %s - User: %s, Time: %s, Details: %s, Success: %v",
		action.Type, action.TargetUser, action.Timestamp.Format(time.RFC3339), action.Details, action.Success)
}

func (pa *PolicyAgent) GetFailedAttempts(username string) int {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	if tracker, exists := pa.failedLogins[username]; exists {
		return len(tracker.Attempts)
	}
	return 0
}

func (pa *PolicyAgent) GetAllTrackedUsers() []string {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	users := make([]string, 0, len(pa.failedLogins))
	for username := range pa.failedLogins {
		users = append(users, username)
	}
	return users
}

func (pa *PolicyAgent) SetMaxAttempts(max int) {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.maxAttempts = max
}

func (pa *PolicyAgent) SetWindowDuration(duration time.Duration) {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.windowDuration = duration
}

func (pa *PolicyAgent) SetRotationEnabled(enabled bool) {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.rotationEnabled = enabled
}

func (pa *PolicyAgent) GetStats() map[string]interface{} {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	totalAttempts := 0
	for _, tracker := range pa.failedLogins {
		totalAttempts += len(tracker.Attempts)
	}

	return map[string]interface{}{
		"tracked_users":    len(pa.failedLogins),
		"total_attempts":   totalAttempts,
		"max_attempts":     pa.maxAttempts,
		"window_duration":  pa.windowDuration.String(),
		"rotation_enabled": pa.rotationEnabled,
	}
}
