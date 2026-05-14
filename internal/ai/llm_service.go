package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type LLMProvider struct {
	Name         string
	Endpoint     string
	Model        string
	EnvKey       string
	Latency      time.Duration
	Working      bool
	RateLimitRPM int
}

type LLMResponse struct {
	Provider string `json:"provider"`
	Latency  int64  `json:"latency_ms"`
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

type LLMService struct {
	mu        sync.RWMutex
	providers []LLMProvider
	current   int
}

var (
	llmService *LLMService
	llmOnce    sync.Once
)

var defaultProviders = []LLMProvider{
	{
		Name:         "Groq",
		Endpoint:     "https://api.groq.com/openai/v1/chat/completions",
		Model:        "llama-3.1-8b-instant",
		EnvKey:       "GROQ_API_KEY",
		RateLimitRPM: 30,
	},
	{
		Name:         "HuggingFace",
		Endpoint:     "https://router.huggingface.co/v1/chat/completions",
		Model:        "meta-llama/Llama-3.1-8B-Instruct",
		EnvKey:       "HUGGINGFACE_API_KEY",
		RateLimitRPM: 100,
	},
	{
		Name:         "Cohere",
		Endpoint:     "https://api.cohere.ai/v1/chat",
		Model:        "command-r7b-12-2024",
		EnvKey:       "COHERE_API_KEY",
		RateLimitRPM: 1000,
	},
	{
		Name:         "Cloudflare",
		Endpoint:     "https://api.cloudflare.com/client/v4/accounts/",
		Model:        "@cf/meta/llama-3.1-8b-instruct",
		EnvKey:       "CLOUDFLARE_API_TOKEN",
		RateLimitRPM: 300,
	},
}

func InitLLMService() *LLMService {
	llmOnce.Do(func() {
		godotenv.Load()
		llmService = &LLMService{
			providers: defaultProviders,
		}
		llmService.checkProviders()
	})
	return llmService
}

func GetLLMService() *LLMService {
	if llmService == nil {
		return InitLLMService()
	}
	return llmService
}

func (s *LLMService) checkProviders() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.providers {
		key := os.Getenv(s.providers[i].EnvKey)
		s.providers[i].Working = key != ""
	}
}

func (s *LLMService) Query(prompt string) (LLMResponse, error) {
	return s.QueryWithProvider("", prompt)
}

func (s *LLMService) QueryWithProvider(providerName, prompt string) (LLMResponse, error) {
	s.mu.RLock()
	if providerName == "" {
		for i := range s.providers {
			if s.providers[i].Working {
				s.mu.RUnlock()
				return s.queryProvider(&s.providers[i], prompt)
			}
		}
		s.mu.RUnlock()
		return LLMResponse{Error: "no working provider"}, fmt.Errorf("no working LLM provider")
	}

	for i := range s.providers {
		if s.providers[i].Name == providerName {
			s.mu.RUnlock()
			return s.queryProvider(&s.providers[i], prompt)
		}
	}
	s.mu.RUnlock()
	return LLMResponse{Error: "provider not found"}, fmt.Errorf("unknown provider: %s", providerName)
}

func (s *LLMService) QueryAll(prompt string) []LLMResponse {
	var responses []LLMResponse
	s.mu.RLock()
	for i := range s.providers {
		if s.providers[i].Working {
			p := s.providers[i]
			go func(p *LLMProvider) {
				resp, _ := s.queryProvider(p, prompt)
				responses = append(responses, resp)
			}(&p)
		}
	}
	s.mu.RUnlock()
	return responses
}

func (s *LLMService) queryProvider(p *LLMProvider, prompt string) (LLMResponse, error) {
	start := time.Now()

	if p.Name == "Cloudflare" {
		return s.queryCloudflare(p, prompt, start)
	}

	key := os.Getenv(p.EnvKey)
	if key == "" {
		return LLMResponse{Provider: p.Name, Error: "no API key"}, fmt.Errorf("no API key for %s", p.Name)
	}

	var body []byte
	var err error

	if p.Name == "Cohere" {
		body, _ = json.Marshal(map[string]interface{}{
			"model":      p.Model,
			"message":    prompt,
			"max_tokens": 1024,
		})
	} else {
		body, _ = json.Marshal(map[string]interface{}{
			"model": p.Model,
			"messages": []map[string]string{
				{"role": "user", "content": prompt},
			},
		})
	}

	req, err := http.NewRequest("POST", p.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return LLMResponse{Provider: p.Name, Error: err.Error()}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	if p.Name == "Cohere" {
		req.Header.Set("Cohere-Version", "2022-12-06")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return LLMResponse{Provider: p.Name, Error: err.Error()}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return LLMResponse{Provider: p.Name, Error: fmt.Sprintf("http %d: %s", resp.StatusCode, string(respBody))}, fmt.Errorf("http %d", resp.StatusCode)
	}

	var result string
	if p.Name == "Cohere" {
		var data map[string]interface{}
		json.Unmarshal(respBody, &data)
		if text, ok := data["text"].(string); ok {
			result = text
		}
	} else {
		var data struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		json.Unmarshal(respBody, &data)
		if len(data.Choices) > 0 {
			result = data.Choices[0].Message.Content
		}
	}

	return LLMResponse{
		Provider: p.Name,
		Latency:  time.Since(start).Milliseconds(),
		Response: result,
	}, nil
}

func (s *LLMService) queryCloudflare(p *LLMProvider, prompt string, start time.Time) (LLMResponse, error) {
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	token := os.Getenv(p.EnvKey)

	if accountID == "" || token == "" {
		return LLMResponse{Provider: p.Name, Error: "missing Cloudflare credentials"}, fmt.Errorf("missing Cloudflare credentials")
	}

	url := p.Endpoint + accountID + "/ai/run/@cf/meta/llama-3.1-8b-instruct"

	body, _ := json.Marshal(map[string]string{"prompt": prompt})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return LLMResponse{Provider: p.Name, Error: err.Error()}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return LLMResponse{Provider: p.Name, Error: err.Error()}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return LLMResponse{Provider: p.Name, Error: fmt.Sprintf("http %d: %s", resp.StatusCode, string(respBody))}, fmt.Errorf("http %d", resp.StatusCode)
	}

	var data map[string]interface{}
	json.Unmarshal(respBody, &data)

	if r, ok := data["result"].(map[string]interface{}); ok {
		if text, ok := r["response"].(string); ok {
			return LLMResponse{
				Provider: p.Name,
				Latency:  time.Since(start).Milliseconds(),
				Response: text,
			}, nil
		}
	}

	return LLMResponse{Provider: p.Name, Error: "empty response"}, fmt.Errorf("empty response")
}

func (s *LLMService) GetProviders() []LLMProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.providers
}

func (s *LLMService) GetWorkingProviders() []LLMProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var working []LLMProvider
	for _, p := range s.providers {
		if p.Working {
			working = append(working, p)
		}
	}
	return working
}

func (s *LLMService) Diagnose(issue string) (string, error) {
	prompt := fmt.Sprintf(`You are a security expert and system administrator. Diagnose this issue and provide a fix:

Issue: %s

Provide a JSON response:
{
  "diagnosis": "brief diagnosis",
  "cause": "root cause",
  "fix": "command to fix (or empty if no simple fix)",
  "verify": "command to verify"
}`, issue)

	resp, err := s.Query(prompt)
	if err != nil {
		return "", err
	}
	return resp.Response, nil
}

func (s *LLMService) AnalyzeSecurityLog(logEntry string) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Analyze this security log entry and provide structured JSON:

Log: %s

Provide JSON:
{
  "threat_level": "low/medium/high/critical",
  "category": "brute_force/malware/suspicious/normal",
  "action": "block/alert/investigate/ignore",
  "summary": "brief summary"
}`, logEntry)

	resp, err := s.Query(prompt)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Response), &result); err != nil {
		result = map[string]interface{}{
			"threat_level": "unknown",
			"category":     "unknown",
			"action":       "investigate",
			"summary":      resp.Response,
		}
	}
	return result, nil
}

func (s *LLMService) AnalyzeNetworkTraffic(trafficData string) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Analyze this network traffic data and identify threats:

Data: %s

Provide JSON:
{
  "threats": ["list of detected threats"],
  "risk_score": 0-100,
  "recommendations": ["list of actions"]
}`, trafficData)

	resp, err := s.Query(prompt)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Response), &result); err != nil {
		result = map[string]interface{}{
			"threats":         []string{},
			"risk_score":      0,
			"recommendations": []string{resp.Response},
		}
	}
	return result, nil
}

func (s *LLMService) GenerateReport(data map[string]interface{}) (string, error) {
	prompt := fmt.Sprintf(`Generate a security report from this data:

Data: %v

Format as a professional security report.`, data)

	resp, err := s.Query(prompt)
	if err != nil {
		return "", err
	}
	return resp.Response, nil
}

func (s *LLMService) ExplainVulnerability(cve string) (string, error) {
	prompt := fmt.Sprintf(`Explain this vulnerability and provide mitigation steps:

%s

Be concise and actionable.`, cve)

	resp, err := s.Query(prompt)
	if err != nil {
		return "", err
	}
	return resp.Response, nil
}

func (s *LLMService) SuggestRemediation(issue string) (string, error) {
	prompt := fmt.Sprintf(`Suggest remediation steps for this security issue:

%s

Provide numbered steps.`, issue)

	resp, err := s.Query(prompt)
	if err != nil {
		return "", err
	}
	return resp.Response, nil
}

func (s *LLMService) CompareThreatIntel(intel1, intel2 string) (string, error) {
	prompt := fmt.Sprintf(`Compare these two threat intelligence sources:

Source 1: %s

Source 2: %s

Which is more credible? Why?`, intel1, intel2)

	resp, err := s.Query(prompt)
	if err != nil {
		return "", err
	}
	return resp.Response, nil
}

func (s *LLMService) PredictAttack(attackHistory string) (string, error) {
	prompt := fmt.Sprintf(`Based on this attack history, predict the likely next attack:

%s

Be specific about timing and method.`, attackHistory)

	resp, err := s.Query(prompt)
	if err != nil {
		return "", err
	}
	return resp.Response, nil
}
