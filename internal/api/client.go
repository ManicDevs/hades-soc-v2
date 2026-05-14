package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Provider string

const (
	ProviderAnthropic Provider = "anthropic"
	ProviderGemini    Provider = "gemini"
	ProviderOpenAI    Provider = "openai"
)

type ModelConfig struct {
	Provider   Provider
	Model      string
	APIKey     string
	MaxRetries int
	BaseDelay  time.Duration
}

type QuotaManager struct {
	mu                sync.RWMutex
	requestCounts     map[string]int
	lastResetTime     time.Time
	dailyLimits       map[string]int
	currentProvider   Provider
	availableModels   []ModelConfig
	currentModelIndex int
}

type APIClient struct {
	quotaManager *QuotaManager
	config       *ModelConfig
	rateLimiter  *RateLimiter
}

type RateLimiter struct {
	tokens     chan struct{}
	refillRate time.Duration
	lastRefill time.Time
	maxTokens  int
	mu         sync.Mutex
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		tokens:     make(chan struct{}, requestsPerMinute),
		refillRate: time.Minute / time.Duration(requestsPerMinute),
		lastRefill: time.Now(),
		maxTokens:  requestsPerMinute,
	}

	// Fill initial tokens
	for i := 0; i < requestsPerMinute; i++ {
		rl.tokens <- struct{}{}
	}

	go rl.refill()
	return rl
}

func (rl *RateLimiter) refill() {
	ticker := time.NewTicker(rl.refillRate)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case rl.tokens <- struct{}{}:
		default:
			// Channel is full, skip
		}
	}
}

func (rl *RateLimiter) Wait() {
	<-rl.tokens
}

func NewQuotaManager() *QuotaManager {
	qm := &QuotaManager{
		requestCounts:   make(map[string]int),
		lastResetTime:   time.Now(),
		dailyLimits:     make(map[string]int),
		currentProvider: ProviderAnthropic,
		availableModels: []ModelConfig{
			{
				Provider:   ProviderGroq,
				Model:      "llama-3.3-70b-versatile",
				MaxRetries: 3,
				BaseDelay:  1 * time.Second,
			},
			{
				Provider:   ProviderTogether,
				Model:      "meta-llama/Llama-3-70b-chat-hf",
				MaxRetries: 3,
				BaseDelay:  1 * time.Second,
			},
			{
				Provider:   ProviderHuggingFace,
				Model:      "meta-llama/Llama-3-70b-chat-hf",
				MaxRetries: 3,
				BaseDelay:  1 * time.Second,
			},
			{
				Provider:   ProviderReplicate,
				Model:      "meta/meta-llama-3-70b-instruct",
				MaxRetries: 3,
				BaseDelay:  1 * time.Second,
			},
			{
				Provider:   ProviderCohere,
				Model:      "command",
				MaxRetries: 3,
				BaseDelay:  1 * time.Second,
			},
			{
				Provider:   ProviderAnthropic,
				Model:      "claude-3-5-sonnet-latest",
				MaxRetries: 3,
				BaseDelay:  1 * time.Second,
			},
			{
				Provider:   ProviderGemini,
				Model:      "gemini-1.5-flash",
				MaxRetries: 2,
				BaseDelay:  2 * time.Second,
			},
			{
				Provider:   ProviderOpenAI,
				Model:      "gpt-4",
				MaxRetries: 3,
				BaseDelay:  1 * time.Second,
			},
		},
	}

	// Set daily limits (free tier limits)
	qm.dailyLimits["llama-3.3-70b-versatile"] = -1        // Unlimited (Groq)
	qm.dailyLimits["meta-llama/Llama-3-70b-chat-hf"] = -1 // Unlimited (Together/HuggingFace)
	qm.dailyLimits["meta-llama/Llama-3-70b-chat-hf"] = -1 // Unlimited (HuggingFace)
	qm.dailyLimits["meta/meta-llama-3-70b-instruct"] = -1 // Unlimited (Replicate)
	qm.dailyLimits["command"] = -1                        // Unlimited (Cohere)
	qm.dailyLimits["gemini-1.5-flash"] = 20
	qm.dailyLimits["claude-3-5-sonnet-latest"] = 1000 // Higher limit for paid tiers
	qm.dailyLimits["gpt-4"] = 100                     // Free tier limit

	return qm
}

func (qm *QuotaManager) CheckQuota(model string) bool {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	// Reset daily counters if needed
	if time.Since(qm.lastResetTime) >= 24*time.Hour {
		qm.requestCounts = make(map[string]int)
		qm.lastResetTime = time.Now()
	}

	currentCount := qm.requestCounts[model]
	limit := qm.dailyLimits[model]

	// Unlimited providers have limit of -1
	if limit == -1 {
		return true
	}

	return currentCount < limit
}

func (qm *QuotaManager) IncrementRequest(model string) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.requestCounts[model]++
	log.Printf("API request count for %s: %d/%d", model, qm.requestCounts[model], qm.dailyLimits[model])
}

func (qm *QuotaManager) GetNextAvailableModel() *ModelConfig {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	// Try current model first
	currentModel := qm.availableModels[qm.currentModelIndex]
	if qm.CheckQuota(currentModel.Model) {
		return &currentModel
	}

	// Try other models
	for i, model := range qm.availableModels {
		if i != qm.currentModelIndex && qm.CheckQuota(model.Model) {
			qm.currentModelIndex = i
			qm.currentProvider = model.Provider
			log.Printf("Switched to backup model: %s (%s)", model.Model, model.Provider)
			return &model
		}
	}

	return nil // All models exhausted
}

func NewAPIClient() (*APIClient, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	qm := NewQuotaManager()

	// Try to find an available model, but don't fail if none available
	// The orchestrator handles provider selection at request time
	model := qm.GetNextAvailableModel()

	var apiKey string
	if model != nil {
		switch model.Provider {
		case ProviderAnthropic:
			apiKey = getEnv("ANTHROPIC_API_KEY", "")
		case ProviderGemini:
			apiKey = getEnv("GEMINI_API_KEY", "")
		case ProviderOpenAI:
			apiKey = getEnv("OPENAI_API_KEY", "")
		case ProviderGroq:
			apiKey = getEnv("GROQ_API_KEY", "")
		case ProviderTogether:
			apiKey = getEnv("TOGETHER_API_KEY", "")
		case ProviderHuggingFace:
			apiKey = getEnv("HUGGINGFACE_API_KEY", "")
		case ProviderReplicate:
			apiKey = getEnv("REPLICATE_API_KEY", "")
		case ProviderCohere:
			apiKey = getEnv("COHERE_API_KEY", "")
		case ProviderPerplexity:
			apiKey = getEnv("PERPLEXITY_API_KEY", "")
		case ProviderMistral:
			apiKey = getEnv("MISTRAL_API_KEY", "")
		case ProviderAI21:
			apiKey = getEnv("AI21_API_KEY", "")
		}
		model.APIKey = apiKey
	}

	return &APIClient{
		quotaManager: qm,
		config:       model,
		rateLimiter:  NewRateLimiter(60),
	}, nil
}

func (client *APIClient) MakeRequest(ctx context.Context, prompt string) (string, error) {
	// Simple direct API call using Groq as primary provider
	apiKey := getEnv("GROQ_API_KEY", "")
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY not set")
	}

	return client.callOpenAICompatibleDirect(ctx, prompt, "llama-3.3-70b-versatile", apiKey, "https://api.groq.com/openai/v1/chat/completions")
}

func (client *APIClient) callOpenAICompatibleDirect(ctx context.Context, prompt, model, apiKey, endpoint string) (string, error) {
	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	type ChatRequest struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
	}

	type ChatResponse struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}

	reqBody := ChatRequest{
		Model: model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("http error %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) > 0 {
		return chatResp.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("empty choices returned")
}

func (client *APIClient) makeRequestWithRetry(ctx context.Context, prompt string, model *ModelConfig) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= model.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter
			delay := time.Duration(float64(model.BaseDelay) * math.Pow(2, float64(attempt-1)))
			jitter := time.Duration(float64(delay) * 0.1 * (2.0*float64(time.Now().UnixNano()%1000)/1000.0 - 1.0))
			time.Sleep(delay + jitter)

			log.Printf("Retry attempt %d for model %s", attempt, model.Model)
		}

		response, err := client.callAPI(ctx, prompt, model)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Check if this is a quota error
		if isQuotaError(err) {
			log.Printf("Quota error detected for %s, trying next model", model.Model)
			if nextModel := client.quotaManager.GetNextAvailableModel(); nextModel != nil {
				model = nextModel
				continue
			}
			break
		}

		// For other errors, continue retrying
		log.Printf("API error (attempt %d): %v", attempt+1, err)
	}

	return "", fmt.Errorf("all retries exhausted. Last error: %w", lastErr)
}

func (client *APIClient) callAPI(ctx context.Context, prompt string, model *ModelConfig) (string, error) {
	// This is a placeholder for the actual API call
	// In a real implementation, this would make HTTP requests to the respective APIs

	switch model.Provider {
	case ProviderAnthropic:
		return client.callAnthropic(ctx, prompt, model)
	case ProviderGemini:
		return client.callGemini(ctx, prompt, model)
	case ProviderOpenAI:
		return client.callOpenAI(ctx, prompt, model)
	case ProviderGroq:
		return client.callGroqExtended(ctx, prompt, model)
	case ProviderTogether:
		return client.callTogetherExtended(ctx, prompt, model)
	case ProviderHuggingFace:
		return client.callHuggingFaceExtended(ctx, prompt, model)
	case ProviderReplicate:
		return client.callReplicateExtended(ctx, prompt, model)
	case ProviderCohere:
		return client.callCohereExtended(ctx, prompt, model)
	case ProviderPerplexity:
		return client.callPerplexityExtended(ctx, prompt, model)
	case ProviderMistral:
		return client.callMistralExtended(ctx, prompt, model)
	case ProviderAI21:
		return client.callAI21Extended(ctx, prompt, model)
	default:
		return "", fmt.Errorf("unsupported provider: %s", model.Provider)
	}
}

func (client *APIClient) callAnthropic(ctx context.Context, prompt string, model *ModelConfig) (string, error) {
	return client.callAnthropicHelper(ctx, prompt, model, "https://api.anthropic.com/v1/messages")
}

func (client *APIClient) callGemini(ctx context.Context, prompt string, model *ModelConfig) (string, error) {
	return client.callOpenAICompatible(ctx, prompt, model, "https://generativelanguage.googleapis.com/v1beta/models/"+model.Model+":generateContent?key="+model.APIKey)
}

func (client *APIClient) callOpenAI(ctx context.Context, prompt string, model *ModelConfig) (string, error) {
	return client.callOpenAICompatible(ctx, prompt, model, "https://api.openai.com/v1/chat/completions")
}

func isQuotaError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "quota exceeded") ||
		contains(errStr, "rate limit") ||
		contains(errStr, "RESOURCE_EXHAUSTED") ||
		contains(errStr, "429")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
