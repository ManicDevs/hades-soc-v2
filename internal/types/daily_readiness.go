// Package types provides the Daily Readiness Report structure
// for autonomous 24-hour security operations summaries.
package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DailyReadinessReport contains all metrics for the 24h autonomous readiness report
type DailyReadinessReport struct {
	// Report Metadata
	ReportDate      time.Time `json:"report_date"`
	SystemUptime    string    `json:"system_uptime"`
	GlobalRiskLevel float64   `json:"global_risk_level"`

	// Agentic Performance (Section 1)
	TotalCascadesTriggered      int           `json:"total_cascades_triggered"`
	SuccessfulRemediations      int           `json:"successful_remediations"`
	SafetyGovernorInterventions int           `json:"safety_governor_interventions"`
	MeanTimeToRespond           time.Duration `json:"mean_time_to_respond"`

	// Discovery & Recon (Section 2)
	NewAssetsIdentified       int      `json:"new_assets_identified"`
	HighPriorityTargetsMapped int      `json:"high_priority_targets_mapped"`
	OSINTFindings             []string `json:"osint_findings"`

	// Defensive Actions (Section 3)
	QuantumShieldActivations int    `json:"quantum_shield_activations"`
	TopBruteForceSource      string `json:"top_brute_force_source"`
	ZeroTrustIsolations      int    `json:"zero_trust_isolations"`
	BruteForceMitigations    int    `json:"brute_force_mitigations"`

	// Self-Healing & Integrity (Section 4)
	AutonomousPatchesApplied int      `json:"autonomous_patches_applied"`
	PatchExamples            []string `json:"patch_examples"`
	AllPatchesVerified       bool     `json:"all_patches_verified"`
	EntropyPercentage        float64  `json:"entropy_percentage"`

	// Identity Deception - Honey Token Tripwires (Section 6)
	HoneyTokenTripwires int      `json:"honey_token_tripwires"`
	HoneyTokenDetails   []string `json:"honey_token_details"`

	// Critical Alerts (Section 5)
	CriticalAlerts []CriticalAlert `json:"critical_alerts"`
}

// CriticalAlert represents an alert requiring human review
type CriticalAlert struct {
	EventID     string `json:"event_id"`
	Description string `json:"description"`
	Prevented   bool   `json:"prevented"`
	Reason      string `json:"reason"`
}

// NewDailyReadinessReport creates a new empty report with current timestamp
func NewDailyReadinessReport() *DailyReadinessReport {
	return &DailyReadinessReport{
		ReportDate:         time.Now(),
		OSINTFindings:      make([]string, 0),
		PatchExamples:      make([]string, 0),
		CriticalAlerts:     make([]CriticalAlert, 0),
		AllPatchesVerified: true,
		EntropyPercentage:  100.0,
	}
}

// WithUptime sets the system uptime
func (r *DailyReadinessReport) WithUptime(uptime string) *DailyReadinessReport {
	r.SystemUptime = uptime
	return r
}

// WithRiskLevel sets the global risk level
func (r *DailyReadinessReport) WithRiskLevel(score float64) *DailyReadinessReport {
	r.GlobalRiskLevel = score
	return r
}

// AddOSINTFinding adds an OSINT finding
func (r *DailyReadinessReport) AddOSINTFinding(finding string) *DailyReadinessReport {
	r.OSINTFindings = append(r.OSINTFindings, finding)
	return r
}

// AddPatchExample adds a patch example
func (r *DailyReadinessReport) AddPatchExample(example string) *DailyReadinessReport {
	r.PatchExamples = append(r.PatchExamples, example)
	return r
}

// AddCriticalAlert adds a critical alert
func (r *DailyReadinessReport) AddCriticalAlert(eventID, description string, prevented bool, reason string) *DailyReadinessReport {
	r.CriticalAlerts = append(r.CriticalAlerts, CriticalAlert{
		EventID:     eventID,
		Description: description,
		Prevented:   prevented,
		Reason:      reason,
	})
	return r
}

// ToMarkdown generates the report in the exact markdown format specified
func (r *DailyReadinessReport) ToMarkdown() string {
	var sb strings.Builder

	// Header
	sb.WriteString("# ⚡ Hades SOC: 24h Autonomous Readiness Report\n")
	fmt.Fprintf(&sb, "**Report Date:** %s | **System Uptime:** %s | **Global Risk Level:** %.1f%%\n\n",
		r.ReportDate.Format("2006-01-02 15:04 MST"),
		r.SystemUptime,
		r.GlobalRiskLevel)

	// Section 1: Agentic Performance
	sb.WriteString("## 🤖 1. Agentic Performance Summary\n")
	fmt.Fprintf(&sb, "*   **Total Cascades Triggered:** %d (e.g., Recon -> Scan)\n", r.TotalCascadesTriggered)
	fmt.Fprintf(&sb, "*   **Successful Remediations:** %d\n", r.SuccessfulRemediations)
	fmt.Fprintf(&sb, "*   **Safety Governor Interventions:** %d (High-risk events paused)\n", r.SafetyGovernorInterventions)
	fmt.Fprintf(&sb, "*   **Mean Time to Respond (MTTR):** %d ms\n\n", r.MeanTimeToRespond.Milliseconds())

	// Section 2: Discovery & Recon
	sb.WriteString("## 🔍 2. Discovery & Recon (Phase 1)\n")
	fmt.Fprintf(&sb, "*   **New Assets Identified:** %d\n", r.NewAssetsIdentified)
	fmt.Fprintf(&sb, "*   **High-Priority Targets Mapped:** %d\n", r.HighPriorityTargetsMapped)
	sb.WriteString("*   **Autonomous OSINT Findings:** ")
	if len(r.OSINTFindings) > 0 {
		sb.WriteString("[")
		for i, finding := range r.OSINTFindings {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(finding)
		}
		sb.WriteString("]\n")
	} else {
		sb.WriteString("[None]\n")
	}
	sb.WriteString("\n")

	// Section 3: Defensive Actions
	sb.WriteString("## 🛡️ 3. Defensive Actions (Phase 2 & 3)\n")
	fmt.Fprintf(&sb, "*   **Quantum Shield Activations:** %d\n", r.QuantumShieldActivations)
	if r.TopBruteForceSource != "" {
		fmt.Fprintf(&sb, "    - *Reasoning:* %s\n", r.TopBruteForceSource)
	}
	fmt.Fprintf(&sb, "*   **Zero-Trust Isolations:** %d\n", r.ZeroTrustIsolations)
	fmt.Fprintf(&sb, "*   **Brute-Force Mitigations:** %d\n\n", r.BruteForceMitigations)

	// Section 4: Self-Healing & Integrity
	sb.WriteString("## 🧬 4. Self-Healing & Integrity (Phase 5)\n")
	fmt.Fprintf(&sb, "*   **Autonomous Patches Applied:** %d\n", r.AutonomousPatchesApplied)
	if len(r.PatchExamples) > 0 {
		fmt.Fprintf(&sb, "    - *Example:* %s\n", r.PatchExamples[0])
	}
	if r.AllPatchesVerified {
		sb.WriteString("*   **Verification Status:** ✅ ALL PATCHES VERIFIED\n")
	} else {
		sb.WriteString("*   **Verification Status:** ⚠️ PENDING VERIFICATION\n")
	}
	fmt.Fprintf(&sb, "*   **Entropy Check:** Quantum entropy source health at %.1f%%\n\n", r.EntropyPercentage)

	// Section 6: Honey Token Tripwires (Identity Deception)
	sb.WriteString("## 🍯 6. Identity Deception - Honey Token Tripwires\n")
	fmt.Fprintf(&sb, "*   **Honey-Token Traps Triggered:** %d\n", r.HoneyTokenTripwires)
	if r.HoneyTokenTripwires > 0 {
		sb.WriteString("*   **🚨 ACTIVE THREATS DETECTED:**\n")
		for _, detail := range r.HoneyTokenDetails {
			fmt.Fprintf(&sb, "    - %s\n", detail)
		}
		sb.WriteString("*   **Immediate Actions Taken:** Source IPs isolated, sessions revoked, Sentinel team alerted\n")
	} else {
		sb.WriteString("*   ✅ No honey-token traps triggered - no malicious lateral movement detected\n")
	}
	sb.WriteString("\n")

	// Section 5: Critical Alerts
	sb.WriteString("## ⚠️ 7. Critical Alerts for Human Review\n")
	if len(r.CriticalAlerts) == 0 {
		sb.WriteString("*   ✅ No critical alerts requiring human review\n")
	} else {
		for _, alert := range r.CriticalAlerts {
			preventedText := ""
			if alert.Prevented {
				preventedText = " (Prevented by Safety Governor)"
			}
			fmt.Fprintf(&sb, "*   [ ] Event ID #%s: %s%s\n", alert.EventID, alert.Description, preventedText)
			if alert.Reason != "" {
				fmt.Fprintf(&sb, "    - *Reason:* %s\n", alert.Reason)
			}
		}
	}
	sb.WriteString("\n")

	// Footer
	sb.WriteString("---\n")
	sb.WriteString("*Generated autonomously by Hades Orchestrator v2.0*\n")

	return sb.String()
}

// ToJSON returns the report as a JSON string
func (r *DailyReadinessReport) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
