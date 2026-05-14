package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Unlimited free tier providers
const (
	ProviderGroq        Provider = "groq"
	ProviderTogether    Provider = "together"
	ProviderHuggingFace Provider = "huggingface"
	ProviderReplicate   Provider = "replicate"
	ProviderCohere      Provider = "cohere"
	ProviderPerplexity  Provider = "perplexity"
	ProviderMistral     Provider = "mistral"
	ProviderAI21        Provider = "ai21"
)

// New unlimited provider configurations
func (qm *QuotaOrchestrator) initializeUnlimitedProviders() {
	// Groq - Unlimited requests, 30/minute
	qm.providers[ProviderGroq] = &ProviderConfig{
		Name:       ProviderGroq,
		Model:      "llama-3.3-70b-versatile",
		DailyLimit: -1, // Unlimited
		Priority:   0,  // Highest priority
		Weight:     1.0,
		IsHealthy:  true,
	}

	// Together AI - Unlimited requests, 60/minute
	qm.providers[ProviderTogether] = &ProviderConfig{
		Name:       ProviderTogether,
		Model:      "meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo",
		DailyLimit: -1, // Unlimited
		Priority:   1,
		Weight:     0.9,
		IsHealthy:  true,
	}

	// Hugging Face - Unlimited requests, 300/hour
	qm.providers[ProviderHuggingFace] = &ProviderConfig{
		Name:       ProviderHuggingFace,
		Model:      "meta-llama/Meta-Llama-3.1-8B-Instruct",
		DailyLimit: -1, // Unlimited
		Priority:   2,
		Weight:     0.8,
		IsHealthy:  true,
	}

	// Replicate - Unlimited requests, 100/minute
	qm.providers[ProviderReplicate] = &ProviderConfig{
		Name:       ProviderReplicate,
		Model:      "meta/meta-llama-3-70b-instruct",
		DailyLimit: -1, // Unlimited
		Priority:   3,
		Weight:     0.7,
		IsHealthy:  true,
	}

	// Cohere - Unlimited requests, 100/minute
	qm.providers[ProviderCohere] = &ProviderConfig{
		Name:       ProviderCohere,
		Model:      "command",
		DailyLimit: -1, // Unlimited
		Priority:   4,
		Weight:     0.6,
		IsHealthy:  true,
	}

	// Perplexity - 5,000 requests/month
	qm.providers[ProviderPerplexity] = &ProviderConfig{
		Name:       ProviderPerplexity,
		Model:      "llama-3-sonar-small-32k-online",
		DailyLimit: 167, // ~5,000/month
		Priority:   5,
		Weight:     0.5,
		IsHealthy:  true,
	}

	// Mistral AI - 1,000 requests/month
	qm.providers[ProviderMistral] = &ProviderConfig{
		Name:       ProviderMistral,
		Model:      "mistral-large-latest",
		DailyLimit: 33, // ~1,000/month
		Priority:   6,
		Weight:     0.4,
		IsHealthy:  true,
	}

	// AI21 - 1,000 requests/month
	qm.providers[ProviderAI21] = &ProviderConfig{
		Name:       ProviderAI21,
		Model:      "j2-ultra",
		DailyLimit: 33, // ~1,000/month
		Priority:   7,
		Weight:     0.3,
		IsHealthy:  true,
	}
}

// Unified structs for standard OpenAI-compatible endpoints
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

// Anthropic specific structures
type AnthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

type AnthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

// Replicate specific structures
type ReplicateRequest struct {
	Input struct {
		Prompt string `json:"prompt"`
	} `json:"input"`
}

type ReplicateResponse struct {
	Output []string `json:"output"`
}

// Helper function for OpenAI-compatible providers
func (client *APIClient) callOpenAICompatible(ctx context.Context, prompt string, model *ModelConfig, endpoint string) (string, error) {
	log.Printf("🚀 Making OpenAI-compatible API call to %s with model: %s", endpoint, model.Model)

	reqBody := ChatRequest{
		Model: model.Model,
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
	req.Header.Set("Authorization", "Bearer "+model.APIKey)

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

// Helper function for Anthropic
func (client *APIClient) callAnthropicHelper(ctx context.Context, prompt string, model *ModelConfig, endpoint string) (string, error) {
	log.Printf("🚀 Making Anthropic API call with model: %s", model.Model)

	reqBody := AnthropicRequest{
		Model:     model.Model,
		MaxTokens: 1024,
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
	req.Header.Set("x-api-key", model.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

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

	var anthropicResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(anthropicResp.Content) > 0 {
		return anthropicResp.Content[0].Text, nil
	}
	return "", fmt.Errorf("empty text block returned")
}

// Helper function for Replicate
func (client *APIClient) callReplicateHelper(ctx context.Context, prompt string, model *ModelConfig, endpoint string) (string, error) {
	log.Printf("🚀 Making Replicate API call with model: %s", model.Model)

	var reqBody ReplicateRequest
	reqBody.Input.Prompt = prompt

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+model.APIKey)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("http error %d: %s", resp.StatusCode, string(body))
	}

	var replicateResp ReplicateResponse
	if err := json.NewDecoder(resp.Body).Decode(&replicateResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	var output string
	for _, token := range replicateResp.Output {
		output += token
	}

	if output != "" {
		return output, nil
	}
	return "", fmt.Errorf("empty prediction output")
}

// Extended API client methods for unlimited providers
func (client *APIClient) callGroqExtended(ctx context.Context, prompt string, model *ModelConfig) (string, error) {
	return client.callOpenAICompatible(ctx, prompt, model, "https://api.groq.com/openai/v1/chat/completions")
}

func (client *APIClient) callTogetherExtended(ctx context.Context, prompt string, model *ModelConfig) (string, error) {
	return client.callOpenAICompatible(ctx, prompt, model, "https://api.together.xyz/v1/chat/completions")
}

func (client *APIClient) callHuggingFaceExtended(ctx context.Context, prompt string, model *ModelConfig) (string, error) {
	return client.callOpenAICompatible(ctx, prompt, model, "https://api-inference.huggingface.co/models/"+model.Model)
}

func (client *APIClient) callReplicateExtended(ctx context.Context, prompt string, model *ModelConfig) (string, error) {
	return client.callReplicateHelper(ctx, prompt, model, "https://api.replicate.com/v1/predictions")
}

func (client *APIClient) callCohereExtended(ctx context.Context, prompt string, model *ModelConfig) (string, error) {
	return client.callOpenAICompatible(ctx, prompt, model, "https://api.cohere.com/v1/chat")
}

func (client *APIClient) callPerplexityExtended(ctx context.Context, prompt string, model *ModelConfig) (string, error) {
	return client.callOpenAICompatible(ctx, prompt, model, "https://api.perplexity.ai/chat/completions")
}

func (client *APIClient) callMistralExtended(ctx context.Context, prompt string, model *ModelConfig) (string, error) {
	return client.callOpenAICompatible(ctx, prompt, model, "https://api.mistral.ai/v1/chat/completions")
}

func (client *APIClient) callAI21Extended(ctx context.Context, prompt string, model *ModelConfig) (string, error) {
	return client.callOpenAICompatible(ctx, prompt, model, "https://api.ai21.com/studio/v1/chat/completions")
}
