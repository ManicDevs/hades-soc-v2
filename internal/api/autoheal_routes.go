package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterAutoHealRoutes(router *mux.Router) {
	router.HandleFunc("/api/v2/auto-heal/status", handleAutoHealStatus).Methods("GET")
	router.HandleFunc("/api/v2/auto-heal/health", handleAutoHealHealth).Methods("GET")
	router.HandleFunc("/api/v2/auto-heal/diagnose", handleAutoHealDiagnose).Methods("POST")
	router.HandleFunc("/api/v2/auto-heal/fix", handleAutoHealFix).Methods("POST")
}

func handleAutoHealStatus(w http.ResponseWriter, r *http.Request) {
	h := GetAutoHealer()
	if h == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "auto-healer not initialized",
		})
		return
	}

	status := h.GetStatus()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"enabled":          status.Enabled,
			"interval":         status.Interval.String(),
			"recent_healthy":   status.RecentHealthy,
			"recent_unhealthy": status.RecentUnhealthy,
			"issues_fixed":     status.IssuesFixed,
		},
	})
}

func handleAutoHealHealth(w http.ResponseWriter, r *http.Request) {
	h := GetAutoHealer()
	if h == nil {
		http.Error(w, "not initialized", http.StatusServiceUnavailable)
		return
	}

	status := h.GetStatus()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"enabled":   status.Enabled,
			"interval":  status.Interval,
			"healthy":   status.RecentHealthy,
			"unhealthy": status.RecentUnhealthy,
			"fixed":     status.IssuesFixed,
		},
	})
}

func handleAutoHealDiagnose(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Issue string `json:"issue"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Issue == "" {
		http.Error(w, "issue required", http.StatusBadRequest)
		return
	}

	resp, err := queryLLMAPI("auto", "Diagnose this system issue and provide a fix:\n"+req.Issue)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"diagnosis": resp["response"],
		"provider":  resp["provider"],
	})
}

func handleAutoHealFix(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Issue string `json:"issue"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Issue == "" {
		http.Error(w, "issue required", http.StatusBadRequest)
		return
	}

	resp, err := queryLLMAPI("auto", `Fix this issue automatically. Return ONLY JSON:
{"fix": "command to run", "verify": "verify command"}`+"\n\nIssue: "+req.Issue)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var fix struct {
		Fix    string `json:"fix"`
		Verify string `json:"verify"`
	}

	if err := json.Unmarshal([]byte(resp["response"].(string)), &fix); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"raw":     resp["response"],
		})
		return
	}

	if fix.Fix != "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"fix":     fix.Fix,
			"verify":  fix.Verify,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": "No fix available",
	})
}
