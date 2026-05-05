package threat

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

// SignatureBasedDetection implements signature-based threat detection
type SignatureBasedDetection struct {
	td *ThreatDetector
}

func (sbd *SignatureBasedDetection) Name() string {
	return "Signature-Based Detection"
}

func (sbd *SignatureBasedDetection) Confidence() float64 {
	return 0.95
}

func (sbd *SignatureBasedDetection) Detect(ctx context.Context, event SecurityEvent) (*ThreatAlert, error) {
	// Check against known signatures
	signature := sbd.generateSignature(event)
	hash := sha256.Sum256([]byte(signature))
	hashStr := hex.EncodeToString(hash[:])

	if sbd.td.knownSignatures[hashStr] {
		return &ThreatAlert{
			ID:          fmt.Sprintf("sig_%d", time.Now().UnixNano()),
			Timestamp:   event.Timestamp,
			ThreatType:  "known_signature",
			Severity:    "high",
			Confidence:  sbd.Confidence(),
			SourceIP:    event.SourceIP,
			DestIP:      event.DestIP,
			Description: "Known malicious signature detected",
			Indicators: []ThreatIndicator{
				{
					Type:        "signature",
					Value:       hashStr,
					Confidence:  sbd.Confidence(),
					Source:      "signature_database",
					Description: "Known threat signature",
				},
			},
			Status: "new",
		}, nil
	}

	return nil, nil
}

func (sbd *SignatureBasedDetection) Train(data []SecurityEvent) error {
	// Add signatures from training data
	for _, event := range data {
		if event.Severity == "high" || event.Severity == "critical" {
			signature := sbd.generateSignature(event)
			hash := sha256.Sum256([]byte(signature))
			hashStr := hex.EncodeToString(hash[:])
			sbd.td.knownSignatures[hashStr] = true
		}
	}
	return nil
}

func (sbd *SignatureBasedDetection) generateSignature(event SecurityEvent) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s",
		event.EventType, event.Protocol, event.Method, event.Path, event.Payload[:min(len(event.Payload), 100)])
}

// AnomalyBasedDetection implements anomaly-based threat detection
type AnomalyBasedDetection struct {
	td       *ThreatDetector
	detector *AnomalyDetector
}

func (abd *AnomalyBasedDetection) Name() string {
	return "Anomaly-Based Detection"
}

func (abd *AnomalyBasedDetection) Confidence() float64 {
	return 0.85
}

func (abd *AnomalyBasedDetection) Detect(ctx context.Context, event SecurityEvent) (*ThreatAlert, error) {
	// Check for anomalies in various metrics
	anomalies := 0
	totalChecks := 0

	// Port anomaly
	if abd.detector.IsAnomaly("port", float64(event.Port)) {
		anomalies++
	}
	totalChecks++

	// Payload length anomaly
	if abd.detector.IsAnomaly("payload_length", float64(len(event.Payload))) {
		anomalies++
	}
	totalChecks++

	// Header count anomaly
	if abd.detector.IsAnomaly("header_count", float64(len(event.Headers))) {
		anomalies++
	}
	totalChecks++

	// If multiple anomalies detected, create alert
	if anomalies >= 2 {
		return &ThreatAlert{
			ID:          fmt.Sprintf("anomaly_%d", time.Now().UnixNano()),
			Timestamp:   event.Timestamp,
			ThreatType:  "anomalous_behavior",
			Severity:    "medium",
			Confidence:  abd.Confidence(),
			SourceIP:    event.SourceIP,
			DestIP:      event.DestIP,
			Description: fmt.Sprintf("Anomalous behavior detected: %d/%d metrics anomalous", anomalies, totalChecks),
			Indicators: []ThreatIndicator{
				{
					Type:        "anomaly",
					Value:       fmt.Sprintf("%d_anomalies", anomalies),
					Confidence:  abd.Confidence(),
					Source:      "anomaly_detector",
					Description: "Multiple anomalous metrics detected",
				},
			},
			Status: "new",
		}, nil
	}

	return nil, nil
}

func (abd *AnomalyBasedDetection) Train(data []SecurityEvent) error {
	// Update baselines with training data
	ports := make([]float64, 0, len(data))
	payloadLengths := make([]float64, 0, len(data))
	headerCounts := make([]float64, 0, len(data))

	for _, event := range data {
		ports = append(ports, float64(event.Port))
		payloadLengths = append(payloadLengths, float64(len(event.Payload)))
		headerCounts = append(headerCounts, float64(len(event.Headers)))
	}

	abd.detector.UpdateBaseline("port", ports)
	abd.detector.UpdateBaseline("payload_length", payloadLengths)
	abd.detector.UpdateBaseline("header_count", headerCounts)

	return nil
}

// BehavioralAnalysis implements behavioral analysis for threat detection
type BehavioralAnalysis struct {
	td *ThreatDetector
	ml *MLModel
}

func (ba *BehavioralAnalysis) Name() string {
	return "Behavioral Analysis"
}

func (ba *BehavioralAnalysis) Confidence() float64 {
	return 0.75
}

func (ba *BehavioralAnalysis) Detect(ctx context.Context, event SecurityEvent) (*ThreatAlert, error) {
	// Use ML model to predict threat probability
	probability := ba.ml.Predict(event)

	if probability > 0.7 {
		severity := "medium"
		if probability > 0.9 {
			severity = "high"
		}

		return &ThreatAlert{
			ID:          fmt.Sprintf("behavior_%d", time.Now().UnixNano()),
			Timestamp:   event.Timestamp,
			ThreatType:  "behavioral_anomaly",
			Severity:    severity,
			Confidence:  probability,
			SourceIP:    event.SourceIP,
			DestIP:      event.DestIP,
			Description: fmt.Sprintf("Behavioral anomaly detected with probability %.2f", probability),
			Indicators: []ThreatIndicator{
				{
					Type:        "behavioral",
					Value:       fmt.Sprintf("%.2f", probability),
					Confidence:  probability,
					Source:      "ml_model",
					Description: "ML model detected behavioral anomaly",
				},
			},
			Status: "new",
		}, nil
	}

	return nil, nil
}

func (ba *BehavioralAnalysis) Train(data []SecurityEvent) error {
	return ba.ml.Train(data)
}

// NetworkTrafficAnalysis implements network traffic analysis
type NetworkTrafficAnalysis struct {
	td *ThreatDetector
}

func (nta *NetworkTrafficAnalysis) Name() string {
	return "Network Traffic Analysis"
}

func (nta *NetworkTrafficAnalysis) Confidence() float64 {
	return 0.80
}

func (nta *NetworkTrafficAnalysis) Detect(ctx context.Context, event SecurityEvent) (*ThreatAlert, error) {
	// Analyze network traffic patterns
	if nta.isSuspiciousTraffic(event) {
		return &ThreatAlert{
			ID:          fmt.Sprintf("network_%d", time.Now().UnixNano()),
			Timestamp:   event.Timestamp,
			ThreatType:  "suspicious_traffic",
			Severity:    "medium",
			Confidence:  nta.Confidence(),
			SourceIP:    event.SourceIP,
			DestIP:      event.DestIP,
			Description: "Suspicious network traffic pattern detected",
			Indicators: []ThreatIndicator{
				{
					Type:        "network",
					Value:       fmt.Sprintf("%s:%d", event.DestIP, event.Port),
					Confidence:  nta.Confidence(),
					Source:      "network_analyzer",
					Description: "Suspicious network traffic",
				},
			},
			Status: "new",
		}, nil
	}

	return nil, nil
}

func (nta *NetworkTrafficAnalysis) isSuspiciousTraffic(event SecurityEvent) bool {
	// Check for suspicious ports
	suspiciousPorts := []int{22, 23, 80, 443, 1433, 3306, 5432, 6379, 27017}
	for _, port := range suspiciousPorts {
		if event.Port == port && event.EventType == "connection_attempt" {
			return true
		}
	}

	// Check for private IP connections to external services
	if nta.isPrivateIP(event.SourceIP) && !nta.isPrivateIP(event.DestIP) {
		return true
	}

	return false
}

func (nta *NetworkTrafficAnalysis) isPrivateIP(ip string) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	for _, rangeStr := range privateRanges {
		_, ipNet, err := net.ParseCIDR(rangeStr)
		if err != nil {
			continue
		}
		if ipNet.Contains(net.ParseIP(ip)) {
			return true
		}
	}

	return false
}

func (nta *NetworkTrafficAnalysis) Train(data []SecurityEvent) error {
	// Network traffic analysis doesn't require training
	return nil
}

// MalwareDetection implements malware detection
type MalwareDetection struct {
	td *ThreatDetector
}

func (md *MalwareDetection) Name() string {
	return "Malware Detection"
}

func (md *MalwareDetection) Confidence() float64 {
	return 0.90
}

func (md *MalwareDetection) Detect(ctx context.Context, event SecurityEvent) (*ThreatAlert, error) {
	// Check for malware signatures in payload
	malwareSignatures := []string{
		"eval(",
		"base64_decode",
		"shell_exec",
		"system(",
		"exec(",
		"passthru",
		"file_get_contents",
		"fopen(",
		"fwrite(",
	}

	payload := strings.ToLower(event.Payload)
	for _, signature := range malwareSignatures {
		if strings.Contains(payload, signature) {
			return &ThreatAlert{
				ID:          fmt.Sprintf("malware_%d", time.Now().UnixNano()),
				Timestamp:   event.Timestamp,
				ThreatType:  "malware",
				Severity:    "critical",
				Confidence:  md.Confidence(),
				SourceIP:    event.SourceIP,
				DestIP:      event.DestIP,
				Description: fmt.Sprintf("Malware signature detected: %s", signature),
				Indicators: []ThreatIndicator{
					{
						Type:        "malware_signature",
						Value:       signature,
						Confidence:  md.Confidence(),
						Source:      "malware_scanner",
						Description: "Malware signature in payload",
					},
				},
				Status: "new",
			}, nil
		}
	}

	return nil, nil
}

func (md *MalwareDetection) Train(data []SecurityEvent) error {
	// Malware detection uses static signatures
	return nil
}

// PhishingDetection implements phishing detection
type PhishingDetection struct {
	td *ThreatDetector
}

func (pd *PhishingDetection) Name() string {
	return "Phishing Detection"
}

func (pd *PhishingDetection) Confidence() float64 {
	return 0.85
}

func (pd *PhishingDetection) Detect(ctx context.Context, event SecurityEvent) (*ThreatAlert, error) {
	// Check for phishing indicators
	phishingKeywords := []string{
		"verify your account",
		"urgent action required",
		"click here immediately",
		"suspended account",
		"security breach",
		"update payment information",
		"confirm your identity",
		"limited time offer",
		"winner",
		"congratulations",
	}

	payload := strings.ToLower(event.Payload)
	userAgent := strings.ToLower(event.UserAgent)
	path := strings.ToLower(event.Path)

	phishingScore := 0
	for _, keyword := range phishingKeywords {
		if strings.Contains(payload, keyword) || strings.Contains(path, keyword) {
			phishingScore++
		}
	}

	// Check for suspicious user agents
	suspiciousUserAgents := []string{
		"bot",
		"crawler",
		"spider",
		"scraper",
	}

	for _, ua := range suspiciousUserAgents {
		if strings.Contains(userAgent, ua) {
			phishingScore++
		}
	}

	if phishingScore >= 2 {
		return &ThreatAlert{
			ID:          fmt.Sprintf("phishing_%d", time.Now().UnixNano()),
			Timestamp:   event.Timestamp,
			ThreatType:  "phishing",
			Severity:    "high",
			Confidence:  pd.Confidence(),
			SourceIP:    event.SourceIP,
			DestIP:      event.DestIP,
			Description: fmt.Sprintf("Phishing attempt detected with %d indicators", phishingScore),
			Indicators: []ThreatIndicator{
				{
					Type:        "phishing_indicators",
					Value:       fmt.Sprintf("%d_keywords", phishingScore),
					Confidence:  pd.Confidence(),
					Source:      "phishing_detector",
					Description: "Multiple phishing keywords detected",
				},
			},
			Status: "new",
		}, nil
	}

	return nil, nil
}

func (pd *PhishingDetection) Train(data []SecurityEvent) error {
	// Phishing detection uses keyword analysis
	return nil
}

// DDoSDetection implements DDoS attack detection
type DDoSDetection struct {
	td *ThreatDetector
}

func (dd *DDoSDetection) Name() string {
	return "DDoS Detection"
}

func (dd *DDoSDetection) Confidence() float64 {
	return 0.95
}

func (dd *DDoSDetection) Detect(ctx context.Context, event SecurityEvent) (*ThreatAlert, error) {
	// Check for DDoS patterns
	// This is a simplified implementation - in production, you'd use rate limiting and traffic analysis

	// Check for high frequency requests from same IP
	if event.EventType == "connection_attempt" {
		// In production, maintain a request counter per IP
		// For now, just check if it's a suspicious pattern
		if dd.isDDoSPattern(event) {
			return &ThreatAlert{
				ID:          fmt.Sprintf("ddos_%d", time.Now().UnixNano()),
				Timestamp:   event.Timestamp,
				ThreatType:  "ddos",
				Severity:    "critical",
				Confidence:  dd.Confidence(),
				SourceIP:    event.SourceIP,
				DestIP:      event.DestIP,
				Description: "Potential DDoS attack detected",
				Indicators: []ThreatIndicator{
					{
						Type:        "ddos_pattern",
						Value:       "high_frequency_requests",
						Confidence:  dd.Confidence(),
						Source:      "ddos_detector",
						Description: "High frequency request pattern",
					},
				},
				Status: "new",
			}, nil
		}
	}

	return nil, nil
}

func (dd *DDoSDetection) isDDoSPattern(event SecurityEvent) bool {
	// Simplified DDoS detection
	// In production, implement proper rate limiting and traffic analysis
	return event.EventType == "connection_attempt" &&
		(event.Port == 80 || event.Port == 443) &&
		len(event.Payload) < 100 // Small requests typical of DDoS
}

func (dd *DDoSDetection) Train(data []SecurityEvent) error {
	// DDoS detection uses traffic analysis
	return nil
}

// SQLInjectionDetection implements SQL injection detection
type SQLInjectionDetection struct {
	td *ThreatDetector
}

func (sqli *SQLInjectionDetection) Name() string {
	return "SQL Injection Detection"
}

func (sqli *SQLInjectionDetection) Confidence() float64 {
	return 0.90
}

func (sqli *SQLInjectionDetection) Detect(ctx context.Context, event SecurityEvent) (*ThreatAlert, error) {
	// Check for SQL injection patterns
	sqlInjectionPatterns := []string{
		`(?i)union\s+select`,
		`(?i)select\s+.*\s+from`,
		`(?i)insert\s+into`,
		`(?i)update\s+.*\s+set`,
		`(?i)delete\s+from`,
		`(?i)drop\s+table`,
		`(?i)create\s+table`,
		`(?i)alter\s+table`,
		`(?i)exec\s*\(`,
		`(?i)script\s+`,
		`(?i)or\s+1\s*=\s*1`,
		`(?i)and\s+1\s*=\s*1`,
		`(?i)'or'1'='1`,
		`(?i)"or"1"="1`,
		`(?i)--`,
		`(?i)#`,
		`(?i)\/\*`,
		`(?i)\*\/`,
	}

	payload := event.Query + event.Payload
	for _, pattern := range sqlInjectionPatterns {
		if matched, err := regexp.MatchString(pattern, payload); err == nil && matched {
			return &ThreatAlert{
				ID:          fmt.Sprintf("sqli_%d", time.Now().UnixNano()),
				Timestamp:   event.Timestamp,
				ThreatType:  "sql_injection",
				Severity:    "critical",
				Confidence:  sqli.Confidence(),
				SourceIP:    event.SourceIP,
				DestIP:      event.DestIP,
				Description: "SQL injection attempt detected",
				Indicators: []ThreatIndicator{
					{
						Type:        "sql_injection_pattern",
						Value:       pattern,
						Confidence:  sqli.Confidence(),
						Source:      "sqli_detector",
						Description: "SQL injection pattern detected",
					},
				},
				Status: "new",
			}, nil
		}
	}

	return nil, nil
}

func (sqli *SQLInjectionDetection) Train(data []SecurityEvent) error {
	// SQL injection detection uses pattern matching
	return nil
}

// XSSDetection implements XSS attack detection
type XSSDetection struct {
	td *ThreatDetector
}

func (xss *XSSDetection) Name() string {
	return "XSS Detection"
}

func (xss *XSSDetection) Confidence() float64 {
	return 0.85
}

func (xss *XSSDetection) Detect(ctx context.Context, event SecurityEvent) (*ThreatAlert, error) {
	// Check for XSS patterns
	xssPatterns := []string{
		`(?i)<script`,
		`(?i)javascript:`,
		`(?i)onload=`,
		`(?i)onerror=`,
		`(?i)onclick=`,
		`(?i)onmouseover=`,
		`(?i)alert\s*\(`,
		`(?i)document\.`,
		`(?i)window\.`,
		`(?i)eval\s*\(`,
		`(?i)expression\s*\(`,
		`(?i)vbscript:`,
		`(?i)data:text/html`,
		`(?i)<iframe`,
		`(?i)<object`,
		`(?i)<embed`,
	}

	payload := event.Query + event.Payload
	for _, pattern := range xssPatterns {
		if matched, err := regexp.MatchString(pattern, payload); err == nil && matched {
			return &ThreatAlert{
				ID:          fmt.Sprintf("xss_%d", time.Now().UnixNano()),
				Timestamp:   event.Timestamp,
				ThreatType:  "xss",
				Severity:    "high",
				Confidence:  xss.Confidence(),
				SourceIP:    event.SourceIP,
				DestIP:      event.DestIP,
				Description: "XSS attack attempt detected",
				Indicators: []ThreatIndicator{
					{
						Type:        "xss_pattern",
						Value:       pattern,
						Confidence:  xss.Confidence(),
						Source:      "xss_detector",
						Description: "XSS pattern detected",
					},
				},
				Status: "new",
			}, nil
		}
	}

	return nil, nil
}

func (xss *XSSDetection) Train(data []SecurityEvent) error {
	// XSS detection uses pattern matching
	return nil
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
