package api

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"
)

type QuotaOrchestrator struct {
	mu              sync.RWMutex
	providers       map[Provider]*ProviderConfig
	currentProvider Provider
	apiClient       *APIClient
	quotaMonitor    *QuotaMonitor
	strategy        RotationStrategy
	lastSwitchTime  time.Time
	dailyResetTime  time.Time
	rotationCount   int
}

type ProviderConfig struct {
	Name          Provider
	Model         string
	APIKey        string
	DailyLimit    int
	CurrentUsage  int
	Priority      int
	Weight        float64
	LastUsed      time.Time
	ErrorCount    int
	IsHealthy     bool
	CooldownUntil time.Time
}

type RotationStrategy int

const (
	StrategyPriority RotationStrategy = iota
	StrategyWeighted
	StrategyRoundRobin
	StrategyQuotaAware
)

type ProviderScore struct {
	Provider Provider
	Score    float64
	Reason   string
}

func NewQuotaOrchestrator() *QuotaOrchestrator {
	qo := &QuotaOrchestrator{
		providers:      make(map[Provider]*ProviderConfig),
		strategy:       StrategyQuotaAware,
		dailyResetTime: getNextDailyReset(),
	}

	// Initialize free tier providers
	qo.initializeFreeTierProviders()

	// Create API client and monitor (deferred to avoid hanging)
	// apiClient, _ := NewAPIClient()
	// qo.apiClient = apiClient
	// qo.quotaMonitor = NewQuotaMonitor(apiClient.quotaManager)

	// Load API keys for all providers
	qo.loadAPIKeys()

	return qo
}

func (qo *QuotaOrchestrator) loadAPIKeys() {
	for _, config := range qo.providers {
		switch config.Name {
		case ProviderAnthropic:
			config.APIKey = getEnv("ANTHROPIC_API_KEY", "")
		case ProviderGemini:
			config.APIKey = getEnv("GEMINI_API_KEY", "")
		case ProviderOpenAI:
			config.APIKey = getEnv("OPENAI_API_KEY", "")
		case ProviderGroq:
			config.APIKey = getEnv("GROQ_API_KEY", "")
		case ProviderTogether:
			config.APIKey = getEnv("TOGETHER_API_KEY", "")
		case ProviderHuggingFace:
			config.APIKey = getEnv("HUGGINGFACE_API_KEY", "")
		case ProviderReplicate:
			config.APIKey = getEnv("REPLICATE_API_KEY", "")
		case ProviderCohere:
			config.APIKey = getEnv("COHERE_API_KEY", "")
		case ProviderPerplexity:
			config.APIKey = getEnv("PERPLEXITY_API_KEY", "")
		case ProviderMistral:
			config.APIKey = getEnv("MISTRAL_API_KEY", "")
		case ProviderAI21:
			config.APIKey = getEnv("AI21_API_KEY", "")
		}
	}
}

func (qo *QuotaOrchestrator) initializeFreeTierProviders() {
	// Initialize limited free tier providers first
	qo.initializeLimitedProviders()

	// Initialize unlimited providers
	qo.initializeUnlimitedProviders()

	// Set initial provider to highest priority available
	qo.currentProvider = ProviderGroq // Start with unlimited provider
}

func (qo *QuotaOrchestrator) initializeLimitedProviders() {
	// Gemini (20 requests/day free)
	qo.providers[ProviderGemini] = &ProviderConfig{
		Name:       ProviderGemini,
		Model:      "gemini-1.5-flash",
		DailyLimit: 20,
		Priority:   10, // Lower priority due to strict limit
		Weight:     0.1,
		IsHealthy:  true,
	}

	// Anthropic Claude (1000 requests/day free tier)
	qo.providers[ProviderAnthropic] = &ProviderConfig{
		Name:       ProviderAnthropic,
		Model:      "claude-3-5-sonnet-latest",
		DailyLimit: 1000,
		Priority:   5, // Medium priority
		Weight:     0.3,
		IsHealthy:  true,
	}

	// OpenAI GPT (100 requests/day free tier)
	qo.providers[ProviderOpenAI] = &ProviderConfig{
		Name:       ProviderOpenAI,
		Model:      "gpt-4",
		DailyLimit: 100,
		Priority:   8, // Lower priority
		Weight:     0.2,
		IsHealthy:  true,
	}
}

func (qo *QuotaOrchestrator) MakeRequest(ctx context.Context, prompt string) (string, error) {
	qo.mu.Lock()
	defer qo.mu.Unlock()

	// Initialize API client on first use
	if qo.apiClient == nil {
		apiClient, err := NewAPIClient()
		if err != nil {
			return "", fmt.Errorf("failed to create API client: %w", err)
		}
		qo.apiClient = apiClient
		qo.quotaMonitor = NewQuotaMonitor(apiClient.quotaManager)
	}

	// Check for daily reset
	if time.Now().After(qo.dailyResetTime) {
		qo.performDailyReset()
	}

	// Get best provider for this request
	provider, err := qo.selectBestProvider()
	if err != nil {
		return "", fmt.Errorf("no available providers: %w", err)
	}

	// Make the request
	response, err := qo.makeRequestWithProvider(ctx, prompt, provider)
	if err != nil {
		qo.handleProviderError(provider, err)

		// Try next best provider
		if nextProvider, nextErr := qo.selectBestProvider(); nextErr == nil {
			return qo.makeRequestWithProvider(ctx, prompt, nextProvider)
		}

		return "", err
	}

	// Update provider stats
	qo.updateProviderStats(provider, true)

	return response, nil
}

func (qo *QuotaOrchestrator) selectBestProvider() (Provider, error) {
	var availableProviders []Provider

	// Filter available providers
	for provider, config := range qo.providers {
		if qo.isProviderAvailable(provider, config) {
			availableProviders = append(availableProviders, provider)
		}
	}

	if len(availableProviders) == 0 {
		return "", fmt.Errorf("all providers exhausted or in cooldown")
	}

	// Score providers based on strategy
	scores := qo.scoreProviders(availableProviders)

	// Sort by score (descending)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	bestProvider := scores[0].Provider

	// Switch providers if needed
	if bestProvider != qo.currentProvider {
		qo.switchProvider(bestProvider, scores[0].Reason)
	}

	return bestProvider, nil
}

func (qo *QuotaOrchestrator) isProviderAvailable(provider Provider, config *ProviderConfig) bool {
	// Check cooldown
	if time.Now().Before(config.CooldownUntil) {
		return false
	}

	// Check health
	if !config.IsHealthy {
		return false
	}

	// Check API key
	if config.APIKey == "" {
		return false
	}

	// Unlimited providers are always available (if healthy, not in cooldown, has key)
	if config.DailyLimit == -1 {
		return true
	}

	// Check quota for limited providers
	remaining := config.DailyLimit - config.CurrentUsage
	if remaining <= 0 {
		return false
	}

	return true
}

func (qo *QuotaOrchestrator) scoreProviders(providers []Provider) []ProviderScore {
	var scores []ProviderScore

	for _, provider := range providers {
		config := qo.providers[provider]
		score := qo.calculateProviderScore(config)
		scores = append(scores, score)
	}

	return scores
}

func (qo *QuotaOrchestrator) calculateProviderScore(config *ProviderConfig) ProviderScore {
	score := 0.0
	reason := ""

	// Base score from remaining quota percentage
	remainingPercent := float64(config.DailyLimit-config.CurrentUsage) / float64(config.DailyLimit)
	score += remainingPercent * 100
	reason += fmt.Sprintf("Quota: %.1f%%, ", remainingPercent*100)

	// Weight factor
	score += config.Weight * 50
	reason += fmt.Sprintf("Weight: %.1f, ", config.Weight)

	// Error penalty
	errorPenalty := float64(config.ErrorCount) * 10
	score -= errorPenalty
	if config.ErrorCount > 0 {
		reason += fmt.Sprintf("Errors: %d, ", config.ErrorCount)
	}

	// Time since last used (prefer less recently used)
	timeSinceLastUse := time.Since(config.LastUsed).Minutes()
	if timeSinceLastUse > 0 {
		timeBonus := math.Min(timeSinceLastUse/60.0, 10.0) // Max 10 points
		score += timeBonus
		reason += fmt.Sprintf("Time bonus: %.1f, ", timeBonus)
	}

	// Priority bonus
	priorityBonus := float64(4-config.Priority) * 5
	score += priorityBonus
	reason += fmt.Sprintf("Priority: %d", config.Priority)

	return ProviderScore{
		Provider: config.Name,
		Score:    score,
		Reason:   reason,
	}
}

func (qo *QuotaOrchestrator) switchProvider(newProvider Provider, reason string) {
	oldProvider := qo.currentProvider
	qo.currentProvider = newProvider
	qo.lastSwitchTime = time.Now()
	qo.rotationCount++

	log.Printf("🔄 Switched providers: %s -> %s (Reason: %s)", oldProvider, newProvider, reason)
	log.Printf("📊 Total rotations today: %d", qo.rotationCount)
}

func (qo *QuotaOrchestrator) makeRequestWithProvider(ctx context.Context, prompt string, provider Provider) (string, error) {
	config := qo.providers[provider]

	// Update last used time
	config.LastUsed = time.Now()

	// Increment usage (will be decremented if request fails)
	config.CurrentUsage++

	// Build model config for this provider
	modelConfig := &ModelConfig{
		Provider:   provider,
		Model:      config.Model,
		APIKey:     config.APIKey,
		MaxRetries: 2,
		BaseDelay:  1 * time.Second,
	}

	// Make the actual API call using the provider-specific endpoint
	response, err := qo.apiClient.callAPI(ctx, prompt, modelConfig)

	if err != nil {
		// Decrement usage on failure
		config.CurrentUsage--
		return "", fmt.Errorf("provider %s failed: %w", provider, err)
	}

	log.Printf("✅ Request successful via %s (model: %s) (Usage: %d/%d)", provider, config.Model, config.CurrentUsage, config.DailyLimit)

	return response, nil
}

func (qo *QuotaOrchestrator) handleProviderError(provider Provider, err error) {
	config := qo.providers[provider]
	config.ErrorCount++

	// Check if this is a quota error
	if isQuotaError(err) {
		log.Printf("⚠️  Provider %s quota exhausted", provider)
		// Set long cooldown until daily reset
		config.CooldownUntil = qo.dailyResetTime
	} else {
		log.Printf("❌ Provider %s error: %v", provider, err)
		// Set short cooldown for other errors
		config.CooldownUntil = time.Now().Add(5 * time.Minute)
	}

	// Mark as unhealthy if too many errors
	if config.ErrorCount >= 3 {
		config.IsHealthy = false
		log.Printf("🚫 Provider %s marked as unhealthy", provider)
	}
}

func (qo *QuotaOrchestrator) updateProviderStats(provider Provider, success bool) {
	config := qo.providers[provider]

	if success {
		// Reset error count on success
		config.ErrorCount = 0
		config.IsHealthy = true
	}
}

func (qo *QuotaOrchestrator) performDailyReset() {
	log.Printf("🌅 Performing daily quota reset")

	for _, config := range qo.providers {
		config.CurrentUsage = 0
		config.ErrorCount = 0
		config.IsHealthy = true
		config.CooldownUntil = time.Time{}
	}

	qo.dailyResetTime = getNextDailyReset()
	qo.rotationCount = 0

	log.Printf("✅ Daily reset complete. Next reset: %s", qo.dailyResetTime.Format("2006-01-02 15:04:05"))
}

func (qo *QuotaOrchestrator) StartContinuousMonitoring(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			qo.performHealthCheck()
		}
	}()
}

func (qo *QuotaOrchestrator) performHealthCheck() {
	qo.mu.RLock()
	defer qo.mu.RUnlock()

	log.Printf("🏥 Performing provider health check")

	for provider, config := range qo.providers {
		remaining := config.DailyLimit - config.CurrentUsage
		health := "✅ Healthy"

		if !config.IsHealthy {
			health = "🚫 Unhealthy"
		} else if remaining <= 0 {
			health = "❌ Exhausted"
		} else if remaining < 5 {
			health = "⚠️  Low"
		}

		log.Printf("  %s: %s (%d/%d used, %d errors)",
			provider, health, config.CurrentUsage, config.DailyLimit, config.ErrorCount)
	}
}

func (qo *QuotaOrchestrator) GetStatus() map[string]interface{} {
	qo.mu.RLock()
	defer qo.mu.RUnlock()

	status := make(map[string]interface{})
	status["current_provider"] = qo.currentProvider
	status["strategy"] = qo.strategy
	status["rotation_count"] = qo.rotationCount
	status["last_switch_time"] = qo.lastSwitchTime
	status["next_daily_reset"] = qo.dailyResetTime

	providers := make(map[string]interface{})
	for provider, config := range qo.providers {
		providers[string(provider)] = map[string]interface{}{
			"model":          config.Model,
			"daily_limit":    config.DailyLimit,
			"current_usage":  config.CurrentUsage,
			"remaining":      config.DailyLimit - config.CurrentUsage,
			"priority":       config.Priority,
			"weight":         config.Weight,
			"is_healthy":     config.IsHealthy,
			"error_count":    config.ErrorCount,
			"last_used":      config.LastUsed,
			"cooldown_until": config.CooldownUntil,
		}
	}
	status["providers"] = providers

	return status
}

func getNextDailyReset() time.Time {
	now := time.Now()
	// Reset at midnight UTC
	nextReset := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	return nextReset
}
