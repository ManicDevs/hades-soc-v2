package recon

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"hades-v2/internal/bus"
	"hades-v2/internal/types"
	"hades-v2/pkg/sdk"
)

// TargetType represents OSINT target types
type TargetType string

const (
	TargetEmail    TargetType = "email"
	TargetDomain   TargetType = "domain"
	TargetUsername TargetType = "username"
	TargetIP       TargetType = "ip"
)

// OSINTFinding represents an OSINT discovery
type OSINTFinding struct {
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Value       string                 `json:"value"`
	Description string                 `json:"description"`
	Confidence  int                    `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// OSINTScanner performs Open Source Intelligence gathering
type OSINTScanner struct {
	*sdk.BaseModule
	targetType TargetType
	target     string
	findings   []OSINTFinding
	httpClient *http.Client
}

// NewOSINTScanner creates a new OSINT scanner instance
func NewOSINTScanner() *OSINTScanner {
	return &OSINTScanner{
		BaseModule: sdk.NewBaseModule(
			"osint_scanner",
			"Open Source Intelligence gathering from public sources",
			sdk.CategoryReconnaissance,
		),
		findings: make([]OSINTFinding, 0),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Execute runs the OSINT scanner
func (os *OSINTScanner) Execute(ctx context.Context) error {
	os.SetStatus(sdk.StatusRunning)
	defer os.SetStatus(sdk.StatusIdle)

	if err := os.validateConfig(); err != nil {
		return fmt.Errorf("hades.recon.osint_scanner: %w", err)
	}

	os.findings = make([]OSINTFinding, 0)

	var err error
	switch os.targetType {
	case TargetEmail:
		err = os.scanEmail(ctx)
	case TargetDomain:
		err = os.scanDomain(ctx)
	case TargetUsername:
		err = os.scanUsername(ctx)
	case TargetIP:
		err = os.scanIP(ctx)
	default:
		return fmt.Errorf("hades.recon.osint_scanner: unsupported target type: %s", os.targetType)
	}

	if err == nil {
		bus.Default().Publish(bus.Event{
			Type:   bus.EventTypeReconComplete,
			Source: "osint_scanner",
			Target: fmt.Sprintf("%s:%s", os.targetType, os.target),
			Payload: map[string]interface{}{
				"target_type": os.targetType,
				"findings":    os.findings,
				"total_found": len(os.findings),
				"scanned_at":  time.Now().Unix(),
			},
		})

		os.publishIPDiscoveryEvents()
	}

	return err
}

func (os *OSINTScanner) publishIPDiscoveryEvents() {
	discoveredIPs := make(map[string]bool)

	for _, finding := range os.findings {
		if finding.Type == "a_record" || finding.Type == "geolocation" || finding.Type == "isp_info" {
			ip := finding.Value
			if _, ok := discoveredIPs[ip]; !ok && os.isValidIP(ip) {
				discoveredIPs[ip] = true

				// Create and publish NewAssetEvent using standardized types
				assetEvent := types.NewNewAssetEvent("osint_scanner", ip).
					WithDomain(os.target).
					WithMetadata("discovered_from", os.target).
					WithMetadata("source_type", string(os.targetType)).
					WithMetadata("finding_type", finding.Type).
					WithMetadata("confidence", finding.Confidence)

				envelope, _ := types.WrapEvent(types.EventTypeNewAsset, assetEvent)
				bus.Default().Publish(bus.Event{
					Type:    bus.EventType(envelope.Type),
					Source:  assetEvent.SourceModule,
					Target:  ip,
					Payload: map[string]interface{}{"data": envelope.Payload},
				})

				// Publish LogEvent with reasoning
				reasoning := fmt.Sprintf("Discovered new asset IP %s from %s OSINT scan. This IP will trigger recon-to-scan cascade.", ip, os.target)
				logEvent := types.NewLogEvent("osint_scanner", fmt.Sprintf("New asset discovered: %s", ip), reasoning)
				logEnvelope, _ := types.WrapEvent(types.EventTypeLog, logEvent)
				bus.Default().Publish(bus.Event{
					Type:    bus.EventType(logEnvelope.Type),
					Source:  logEvent.SourceModule,
					Target:  ip,
					Payload: map[string]interface{}{"data": logEnvelope.Payload},
				})

				log.Printf("OSINTScanner: Published NewAssetEvent for %s (from %s) with reasoning: %s", ip, os.target, reasoning)
			}
		}
	}
}

// SetTargetType configures the target type
func (os *OSINTScanner) SetTargetType(targetType TargetType) error {
	switch targetType {
	case TargetEmail, TargetDomain, TargetUsername, TargetIP:
		os.targetType = targetType
		return nil
	default:
		return fmt.Errorf("hades.recon.osint_scanner: invalid target type: %s", targetType)
	}
}

// SetTarget configures the scan target
func (os *OSINTScanner) SetTarget(target string) error {
	if target == "" {
		return fmt.Errorf("hades.recon.osint_scanner: target cannot be empty")
	}
	os.target = target
	return nil
}

// GetFindings returns discovered OSINT findings
func (os *OSINTScanner) GetFindings() []OSINTFinding {
	result := make([]OSINTFinding, len(os.findings))
	copy(result, os.findings)
	return result
}

// GetResult returns scan results as formatted string
func (os *OSINTScanner) GetResult() string {
	if len(os.findings) == 0 {
		return fmt.Sprintf("No OSINT findings for %s %s", os.targetType, os.target)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "OSINT Findings for %s %s:\n", os.targetType, os.target)

	for i, finding := range os.findings {
		fmt.Fprintf(&sb, "%d. [%s] %s - %s (Confidence: %d%%)\n",
			i+1, finding.Source, finding.Type, finding.Description, finding.Confidence)
		if finding.Metadata != nil {
			for k, v := range finding.Metadata {
				fmt.Fprintf(&sb, "   %s: %v\n", k, v)
			}
		}
	}

	return sb.String()
}

// validateConfig ensures scanner configuration is valid
func (os *OSINTScanner) validateConfig() error {
	if os.targetType == "" {
		return fmt.Errorf("hades.recon.osint_scanner: target type not configured")
	}
	if os.target == "" {
		return fmt.Errorf("hades.recon.osint_scanner: target not configured")
	}
	return nil
}

// scanEmail performs email-based OSINT
func (os *OSINTScanner) scanEmail(ctx context.Context) error {
	email := os.target

	// Basic email validation
	if !os.isValidEmail(email) {
		return fmt.Errorf("hades.recon.osint_scanner: invalid email format")
	}

	// Extract domain for additional analysis
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		domain := parts[1]
		os.addFinding(OSINTFinding{
			Type:        "domain",
			Source:      "email_analysis",
			Value:       domain,
			Description: "Domain extracted from email address",
			Confidence:  100,
			Metadata: map[string]interface{}{
				"email": email,
			},
		})

		// Check domain reputation (simulated)
		os.checkDomainReputation(ctx, domain)
	}

	// Check common breach patterns (simulated)
	os.checkEmailBreaches(ctx, email)

	os.SetStatus(sdk.StatusCompleted)
	return nil
}

// scanDomain performs domain-based OSINT
func (os *OSINTScanner) scanDomain(ctx context.Context) error {
	domain := os.target

	// DNS resolution check
	if err := os.checkDNSResolution(ctx, domain); err != nil {
		return err
	}

	// Subdomain enumeration (simulated)
	os.enumerateSubdomains(ctx, domain)

	// Technology stack detection (simulated)
	os.detectTechnologies(ctx, domain)

	// Social media presence check
	os.checkSocialMedia(ctx, domain)

	os.SetStatus(sdk.StatusCompleted)
	return nil
}

// scanUsername performs username-based OSINT
func (os *OSINTScanner) scanUsername(ctx context.Context) error {
	username := os.target

	// Check common social platforms
	platforms := []string{
		"github.com", "twitter.com", "linkedin.com", "instagram.com",
		"reddit.com", "youtube.com", "tiktok.com", "facebook.com",
	}

	for _, platform := range platforms {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			os.checkUsernameOnPlatform(ctx, username, platform)
		}
	}

	os.SetStatus(sdk.StatusCompleted)
	return nil
}

// scanIP performs IP-based OSINT
func (os *OSINTScanner) scanIP(ctx context.Context) error {
	ip := os.target

	// Basic IP validation
	if !os.isValidIP(ip) {
		return fmt.Errorf("hades.recon.osint_scanner: invalid IP address")
	}

	// Geolocation lookup (simulated)
	os.lookupGeolocation(ctx, ip)

	// ISP and hosting information
	os.lookupISP(ctx, ip)

	// Check if IP is in known malicious ranges
	os.checkMaliciousIP(ctx, ip)

	os.SetStatus(sdk.StatusCompleted)
	return nil
}

// Helper methods
func (os *OSINTScanner) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func (os *OSINTScanner) isValidIP(ip string) bool {
	ipRegex := regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	return ipRegex.MatchString(ip)
}

func (os *OSINTScanner) addFinding(finding OSINTFinding) {
	os.findings = append(os.findings, finding)
}

func (os *OSINTScanner) checkDomainReputation(ctx context.Context, domain string) {
	// Simulated domain reputation check
	os.addFinding(OSINTFinding{
		Type:        "domain_reputation",
		Source:      "reputation_analysis",
		Value:       domain,
		Description: "Domain reputation analysis completed",
		Confidence:  75,
		Metadata: map[string]interface{}{
			"reputation_score": 85,
			"category":         "low_risk",
		},
	})
}

func (os *OSINTScanner) checkEmailBreaches(ctx context.Context, email string) {
	// Simulated breach check
	os.addFinding(OSINTFinding{
		Type:        "breach_check",
		Source:      "breach_database",
		Value:       email,
		Description: "Email found in public breach data",
		Confidence:  90,
		Metadata: map[string]interface{}{
			"breach_count": 2,
			"last_breach":  "2023-01-15",
		},
	})
}

func (os *OSINTScanner) checkDNSResolution(ctx context.Context, domain string) error {
	// Simulated DNS resolution
	os.addFinding(OSINTFinding{
		Type:        "dns_record",
		Source:      "dns_query",
		Value:       domain,
		Description: "DNS resolution successful",
		Confidence:  100,
		Metadata: map[string]interface{}{
			"a_record":  "192.168.1.1",
			"mx_record": "mail." + domain,
		},
	})
	return nil
}

func (os *OSINTScanner) enumerateSubdomains(ctx context.Context, domain string) {
	// Simulated subdomain enumeration
	subdomains := []string{"www", "mail", "api", "blog", "dev", "test"}

	for _, subdomain := range subdomains {
		fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)
		os.addFinding(OSINTFinding{
			Type:        "subdomain",
			Source:      "subdomain_enumeration",
			Value:       fullDomain,
			Description: "Discovered subdomain",
			Confidence:  80,
			Metadata: map[string]interface{}{
				"parent_domain": domain,
				"method":        "dictionary_attack",
			},
		})
	}
}

func (os *OSINTScanner) detectTechnologies(ctx context.Context, domain string) {
	// Simulated technology detection
	technologies := []string{"nginx", "wordpress", "cloudflare", "google_analytics"}

	for _, tech := range technologies {
		os.addFinding(OSINTFinding{
			Type:        "technology",
			Source:      "technology_detection",
			Value:       tech,
			Description: fmt.Sprintf("Detected technology: %s", tech),
			Confidence:  70,
			Metadata: map[string]interface{}{
				"domain":  domain,
				"version": "latest",
			},
		})
	}
}

func (os *OSINTScanner) checkSocialMedia(ctx context.Context, domain string) {
	// Simulated social media presence check
	platforms := []string{"twitter", "linkedin", "facebook"}

	for _, platform := range platforms {
		os.addFinding(OSINTFinding{
			Type:        "social_media",
			Source:      "social_search",
			Value:       fmt.Sprintf("%s.com/%s", platform, domain),
			Description: fmt.Sprintf("Potential %s presence", platform),
			Confidence:  60,
			Metadata: map[string]interface{}{
				"platform": platform,
				"domain":   domain,
			},
		})
	}
}

func (os *OSINTScanner) checkUsernameOnPlatform(ctx context.Context, username, platform string) {
	// Simulated username check
	url := fmt.Sprintf("https://%s/%s", platform, username)

	os.addFinding(OSINTFinding{
		Type:        "social_profile",
		Source:      "platform_search",
		Value:       url,
		Description: fmt.Sprintf("Potential profile on %s", platform),
		Confidence:  65,
		Metadata: map[string]interface{}{
			"platform": platform,
			"username": username,
			"url":      url,
		},
	})
}

func (os *OSINTScanner) lookupGeolocation(ctx context.Context, ip string) {
	// Simulated geolocation lookup
	os.addFinding(OSINTFinding{
		Type:        "geolocation",
		Source:      "geo_database",
		Value:       ip,
		Description: "IP geolocation information",
		Confidence:  85,
		Metadata: map[string]interface{}{
			"country":   "United States",
			"city":      "San Francisco",
			"latitude":  37.7749,
			"longitude": -122.4194,
			"isp":       "Example ISP",
		},
	})
}

func (os *OSINTScanner) lookupISP(ctx context.Context, ip string) {
	// Simulated ISP lookup
	os.addFinding(OSINTFinding{
		Type:        "isp_info",
		Source:      "whois_database",
		Value:       ip,
		Description: "ISP and hosting information",
		Confidence:  90,
		Metadata: map[string]interface{}{
			"isp":          "Example ISP Inc.",
			"organization": "Example Organization",
			"asn":          "AS12345",
			"hosting":      true,
		},
	})
}

func (os *OSINTScanner) checkMaliciousIP(ctx context.Context, ip string) {
	// Simulated malicious IP check
	os.addFinding(OSINTFinding{
		Type:        "reputation",
		Source:      "threat_intelligence",
		Value:       ip,
		Description: "IP reputation check",
		Confidence:  95,
		Metadata: map[string]interface{}{
			"malicious":        false,
			"reputation_score": 92,
			"categories":       []string{"benign"},
		},
	})
}
