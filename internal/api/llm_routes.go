package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"hades-v2/internal/ai"
)

func RegisterLLMRoutes(router *mux.Router) {
	router.HandleFunc("/api/v2/llm/providers", handleLLMProviders).Methods("GET")
	router.HandleFunc("/api/v2/llm/query", handleLLMQuery).Methods("POST")
	router.HandleFunc("/api/v2/llm/compare", handleLLMCompare).Methods("POST")
	router.HandleFunc("/api/v2/llm/benchmark", handleLLMBenchmark).Methods("GET")
	router.HandleFunc("/api/v2/llm/threat/analyze", handleLLMThreatAnalyze).Methods("POST")
	router.HandleFunc("/api/v2/llm/threat/predict", handleLLMThreatPredict).Methods("POST")
	router.HandleFunc("/api/v2/llm/threat/trend", handleLLMThreatTrend).Methods("POST")
	router.HandleFunc("/api/v2/llm/log/enrich", handleLLMLogEnrich).Methods("POST")
}

func handleLLMProviders(w http.ResponseWriter, r *http.Request) {
	providers := []map[string]interface{}{
		{"name": "Groq", "model": "llama-3.1-8b-instant", "rpm": 30, "key": "GROQ_API_KEY"},
		{"name": "HuggingFace", "model": "meta-llama/Llama-3.1-8B-Instruct", "rpm": 100, "key": "HUGGINGFACE_API_KEY"},
		{"name": "Cohere", "model": "command-r7b-12-2024", "rpm": 1000, "key": "COHERE_API_KEY"},
		{"name": "Cloudflare", "model": "@cf/meta/llama-3.1-8b-instruct", "rpm": 300, "key": "CLOUDFLARE_API_TOKEN"},
	}

	for i := range providers {
		providers[i]["configured"] = os.Getenv(providers[i]["key"].(string)) != ""
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    providers,
	})
}

func handleLLMQuery(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prompt   string `json:"prompt"`
		Provider string `json:"provider"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Prompt == "" {
		http.Error(w, "prompt required", http.StatusBadRequest)
		return
	}

	resp, err := queryLLMAPI(req.Provider, req.Prompt)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    resp,
	})
}

func handleLLMCompare(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prompt string `json:"prompt"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Prompt == "" {
		http.Error(w, "prompt required", http.StatusBadRequest)
		return
	}

	results := []map[string]interface{}{}
	for _, p := range []string{"Groq", "Cohere", "HuggingFace"} {
		resp, _ := queryLLMAPI(p, req.Prompt)
		results = append(results, resp)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    results,
	})
}

func handleLLMBenchmark(w http.ResponseWriter, r *http.Request) {
	results := []map[string]interface{}{}
	for _, p := range []struct {
		name  string
		url   string
		model string
		key   string
	}{
		{"Groq", "https://api.groq.com/openai/v1/chat/completions", "llama-3.1-8b-instant", "GROQ_API_KEY"},
		{"Cohere", "https://api.cohere.ai/v1/chat", "command-r7b-12-2024", "COHERE_API_KEY"},
		{"HuggingFace", "https://router.huggingface.co/v1/chat/completions", "meta-llama/Llama-3.1-8B-Instruct", "HUGGINGFACE_API_KEY"},
	} {
		if key := os.Getenv(p.key); key != "" {
			start := time.Now()
			_, err := callProviderAPI(p.name, key, p.url, p.model, "OK")
			results = append(results, map[string]interface{}{
				"provider": p.name,
				"working":  err == nil,
				"latency":  time.Since(start).Milliseconds(),
			})
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    results,
	})
}

func queryLLMAPI(provider, prompt string) (map[string]interface{}, error) {
	if provider == "" || provider == "auto" {
		provider = "Groq"
	}

	providers := map[string]struct {
		URL   string
		Model string
		Key   string
	}{
		"Groq":        {"https://api.groq.com/openai/v1/chat/completions", "llama-3.1-8b-instant", "GROQ_API_KEY"},
		"Cohere":      {"https://api.cohere.ai/v1/chat", "command-r7b-12-2024", "COHERE_API_KEY"},
		"HuggingFace": {"https://router.huggingface.co/v1/chat/completions", "meta-llama/Llama-3.1-8B-Instruct", "HUGGINGFACE_API_KEY"},
		"Cloudflare":  {"https://api.cloudflare.com/client/v4/accounts/", "@cf/meta/llama-3.1-8b-instruct", "CLOUDFLARE_API_TOKEN"},
	}

	p, ok := providers[provider]
	if !ok {
		return nil, &LLMError{"unknown provider: " + provider}
	}

	key := os.Getenv(p.Key)
	if key == "" {
		return nil, &LLMError{"no API key for " + provider}
	}

	if provider == "Cloudflare" {
		accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
		if accountID == "" {
			return nil, &LLMError{"CLOUDFLARE_ACCOUNT_ID not set"}
		}
		p.URL = p.URL + accountID + "/ai/run/@cf/meta/llama-3.1-8b-instruct"
	}

	start := time.Now()
	result, err := callProviderAPI(provider, key, p.URL, p.Model, prompt)

	return map[string]interface{}{
		"provider": provider,
		"response": result,
		"latency":  time.Since(start).Milliseconds(),
	}, err
}

func callProviderAPI(name, key, url, model, prompt string) (string, error) {
	var body []byte
	if name == "Cohere" {
		body, _ = json.Marshal(map[string]interface{}{"model": model, "message": prompt, "max_tokens": 1024})
	} else if name == "Cloudflare" {
		body, _ = json.Marshal(map[string]string{"prompt": prompt})
	} else {
		body, _ = json.Marshal(map[string]interface{}{
			"model":    model,
			"messages": []map[string]string{{"role": "user", "content": prompt}},
		})
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	if name == "Cohere" {
		req.Header.Set("Cohere-Version", "2022-12-06")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	out, _ := readAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", &LLMError{fmt.Sprintf("http %d: %s", resp.StatusCode, string(out))}
	}

	if name == "Cohere" {
		var data map[string]interface{}
		json.Unmarshal(out, &data)
		if text, ok := data["text"].(string); ok {
			return text, nil
		}
	} else if name == "Cloudflare" {
		var data map[string]interface{}
		json.Unmarshal(out, &data)
		if r, ok := data["result"].(map[string]interface{}); ok {
			if text, ok := r["response"].(string); ok {
				return text, nil
			}
		}
	} else {
		var data struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		json.Unmarshal(out, &data)
		if len(data.Choices) > 0 {
			return data.Choices[0].Message.Content, nil
		}
	}
	return "", &LLMError{"empty response"}
}

type LLMError struct {
	msg string
}

func (e *LLMError) Error() string {
	return e.msg
}

func readAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

type ThreatAnalysisRequest struct {
	EventID   string                 `json:"event_id"`
	Type      string                 `json:"type"`
	Severity  string                 `json:"severity"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target"`
	Signature string                 `json:"signature"`
	Features  map[string]interface{} `json:"features"`
	Metadata  map[string]interface{} `json:"metadata"`
	UseLLM    bool                   `json:"use_llm"`
}

func handleLLMThreatAnalyze(w http.ResponseWriter, r *http.Request) {
	var req ThreatAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	llm := ai.GetLLMService()
	if llm == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "LLM service not initialized",
		})
		return
	}

	event := ai.SecurityEvent{
		ID:        req.EventID,
		Type:      req.Type,
		Severity:  req.Severity,
		Source:    req.Source,
		Target:    req.Target,
		Signature: req.Signature,
		Features:  req.Features,
		Metadata:  req.Metadata,
		Timestamp: time.Now(),
	}

	threatEngine, err := ai.NewAIThreatEngine()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	threatEngine.IntegrateWithLLM(llm, req.UseLLM)

	ctx := context.Background()
	var assessment *ai.ThreatAssessment
	var err2 error

	if req.UseLLM {
		assessment, err2 = threatEngine.AnalyzeThreatWithLLM(ctx, event)
	} else {
		assessment, err2 = threatEngine.AnalyzeThreat(ctx, event)
	}

	if err2 != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err2.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"data":     assessment,
		"llm_used": req.UseLLM,
	})
}

type ThreatPredictRequest struct {
	EventHistory string `json:"event_history"`
}

func handleLLMThreatPredict(w http.ResponseWriter, r *http.Request) {
	var req ThreatPredictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	llm := ai.GetLLMService()
	if llm == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "LLM service not initialized",
		})
		return
	}

	threatEngine, err := ai.NewAIThreatEngine()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	threatEngine.IntegrateWithLLM(llm, true)

	predictions, err := threatEngine.GenerateLLMPredictions(req.EventHistory)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    predictions,
	})
}

type ThreatTrendRequest struct {
	Trends []ai.SecurityEvent `json:"trends"`
}

func handleLLMThreatTrend(w http.ResponseWriter, r *http.Request) {
	var req ThreatTrendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	llm := ai.GetLLMService()
	if llm == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "LLM service not initialized",
		})
		return
	}

	threatEngine, err := ai.NewAIThreatEngine()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	threatEngine.IntegrateWithLLM(llm, true)

	analysis, err := threatEngine.AnalyzeThreatTrend(req.Trends)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"trend_analysis": analysis,
		},
	})
}

type LogEnrichRequest struct {
	LogEntry string `json:"log_entry"`
}

func handleLLMLogEnrich(w http.ResponseWriter, r *http.Request) {
	var req LogEnrichRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.LogEntry == "" {
		http.Error(w, "log_entry required", http.StatusBadRequest)
		return
	}

	llm := ai.GetLLMService()
	if llm == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "LLM service not initialized",
		})
		return
	}

	threatEngine, err := ai.NewAIThreatEngine()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	threatEngine.IntegrateWithLLM(llm, true)

	result, err := threatEngine.EnrichLogWithLLM(req.LogEntry)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}
