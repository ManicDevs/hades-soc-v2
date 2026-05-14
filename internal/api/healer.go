package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

type AutoHealer struct {
	mu          sync.RWMutex
	Enabled     bool
	Interval    time.Duration
	AutoRestart bool
	LLMProvider string
	HealthLog   []HealthRecord
	FixedLog    []IssueFix
	stopChan    chan struct{}
}

type HealthRecord struct {
	Component string    `json:"component"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type IssueFix struct {
	Issue     string    `json:"issue"`
	Fix       string    `json:"fix"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
}

type HealerStatus struct {
	Enabled         bool          `json:"enabled"`
	Interval        time.Duration `json:"interval"`
	RecentHealthy   int           `json:"recent_healthy"`
	RecentUnhealthy int           `json:"recent_unhealthy"`
	IssuesFixed     int           `json:"issues_fixed"`
}

var autoHealer *AutoHealer

func initAutoHealer(intervalSec int) *AutoHealer {
	autoHealer = &AutoHealer{
		Enabled:     true,
		Interval:    time.Duration(intervalSec) * time.Second,
		AutoRestart: true,
		LLMProvider: "Groq",
		stopChan:    make(chan struct{}),
	}
	return autoHealer
}

func (a *AutoHealer) Start() {
	go a.runLoop()
	fmt.Println("[auto-heal] Background service started")
}

func (a *AutoHealer) Stop() {
	close(a.stopChan)
}

func (a *AutoHealer) runLoop() {
	ticker := time.NewTicker(a.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopChan:
			return
		case <-ticker.C:
			a.check()
		}
	}
}

func (a *AutoHealer) check() {
	a.mu.Lock()
	defer a.mu.Unlock()

	checks := []struct {
		name string
		fn   func() (bool, string)
	}{
		{"disk", checkDisk},
		{"memory", checkMemory},
		{"cpu", checkCPU},
		{"hades", checkHadesRunning},
	}

	for _, c := range checks {
		ok, msg := c.fn()
		status := "healthy"
		if !ok {
			status = "unhealthy"
			fmt.Printf("[auto-heal] [!] %s: %s\n", c.name, msg)
			if a.AutoRestart {
				a.heal(c.name, msg)
			}
		} else {
			fmt.Printf("[auto-heal] [+] %s: OK\n", c.name)
		}
		a.HealthLog = append(a.HealthLog, HealthRecord{c.name, status, msg, time.Now()})
	}

	if len(a.HealthLog) > 100 {
		a.HealthLog = a.HealthLog[len(a.HealthLog)-100:]
	}
}

func checkDisk() (bool, string) {
	out, err := exec.Command("df", "-h", "/").Output()
	if err != nil {
		return false, "disk check failed"
	}
	m := regexp.MustCompile(`(\d+)%`).FindSubmatch(out)
	if len(m) > 1 {
		var u int
		fmt.Sscanf(string(m[1]), "%d", &u)
		if u > 90 {
			return false, fmt.Sprintf("disk at %d%%", u)
		}
	}
	return true, "disk OK"
}

func checkMemory() (bool, string) {
	out, err := exec.Command("free", "-m").Output()
	if err != nil {
		return false, "memory check failed"
	}
	var tot, avail int
	fmt.Sscanf(string(out), "Mem: %d %d", &tot, &avail)
	if avail < 100 {
		return false, fmt.Sprintf("only %dMB free", avail)
	}
	return true, fmt.Sprintf("%dMB available", avail)
}

func checkCPU() (bool, string) {
	out, err := exec.Command("uptime").Output()
	if err != nil {
		return false, "cpu check failed"
	}
	return true, strings.TrimSpace(string(out))
}

func checkHadesRunning() (bool, string) {
	out, err := exec.Command("pgrep", "-f", "hades").Output()
	if err != nil || len(out) == 0 {
		return false, "hades not running"
	}
	return true, "hades running"
}

func (a *AutoHealer) heal(component, issue string) {
	prompt := fmt.Sprintf(`Component: %s
Issue: %s
Fix system issue. Return ONLY JSON:
{"fix": "command", "verify": "verify command"}`, component, issue)

	resp, err := a.callLLM(prompt)
	if err != nil {
		fmt.Printf("[auto-heal] LLM error: %v\n", err)
		return
	}

	var fix struct {
		Fix    string `json:"fix"`
		Verify string `json:"verify"`
	}
	if err := json.Unmarshal([]byte(resp), &fix); err != nil {
		return
	}

	if fix.Fix != "" {
		fmt.Printf("[auto-heal] [*] Running: %s\n", fix.Fix)
		out, err := exec.Command("bash", "-c", fix.Fix).CombinedOutput()
		if err != nil {
			fmt.Printf("[auto-heal] [!] Failed: %s\n", string(out))
			a.FixedLog = append(a.FixedLog, IssueFix{fmt.Sprintf("%s: %s", component, issue), fix.Fix, false, time.Now()})
		} else {
			fmt.Printf("[auto-heal] [+] Applied\n")
			a.FixedLog = append(a.FixedLog, IssueFix{fmt.Sprintf("%s: %s", component, issue), fix.Fix, true, time.Now()})
		}
	}

	if len(a.FixedLog) > 50 {
		a.FixedLog = a.FixedLog[len(a.FixedLog)-50:]
	}
}

func (a *AutoHealer) callLLM(prompt string) (string, error) {
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
			result, err := a.queryProvider(p.name, key, p.url, p.model, prompt)
			if err == nil {
				return result, nil
			}
		}
	}
	return "", fmt.Errorf("no provider")
}

func (a *AutoHealer) queryProvider(name, key, url, model, prompt string) (string, error) {
	var body []byte
	if name == "Cohere" {
		body, _ = json.Marshal(map[string]interface{}{"model": model, "message": prompt, "max_tokens": 512})
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

	out, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http %d", resp.StatusCode)
	}

	if name == "Cohere" {
		var data map[string]interface{}
		json.Unmarshal(out, &data)
		if text, ok := data["text"].(string); ok {
			return text, nil
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
	return "", fmt.Errorf("empty")
}

func (a *AutoHealer) GetStatus() HealerStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	healthy := 0
	unhealthy := 0
	for _, h := range a.HealthLog {
		if h.Status == "healthy" {
			healthy++
		} else {
			unhealthy++
		}
	}

	fixed := 0
	for _, f := range a.FixedLog {
		if f.Success {
			fixed++
		}
	}

	return HealerStatus{
		Enabled:         a.Enabled,
		Interval:        a.Interval,
		RecentHealthy:   healthy,
		RecentUnhealthy: unhealthy,
		IssuesFixed:     fixed,
	}
}

func GetAutoHealer() *AutoHealer {
	return autoHealer
}
