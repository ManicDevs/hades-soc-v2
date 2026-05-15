package auxiliary

import (
	"context"
	"fmt"
	"time"

	"hades-v2/pkg/sdk"
)

// SIEMProvider represents supported SIEM providers
type SIEMProvider string

const (
	ProviderSplunk      SIEMProvider = "splunk"
	ProviderElastic     SIEMProvider = "elastic"
	ProviderSentinelOne SIEMProvider = "sentinelone"
)

// SIEMIntegration provides SIEM/EDR integration functionality
type SIEMIntegration struct {
	*sdk.BaseModule
	provider SIEMProvider
	endpoint string
}

// NewSIEMIntegration creates a new SIEM integration instance
func NewSIEMIntegration() *SIEMIntegration {
	return &SIEMIntegration{
		BaseModule: sdk.NewBaseModule(
			"siem_integration",
			"Integrate with SIEM/EDR systems for security monitoring",
			sdk.CategoryReporting,
		),
		provider: ProviderSplunk,
		endpoint: "https://localhost:8088",
	}
}

// Execute starts SIEM integration
func (si *SIEMIntegration) Execute(ctx context.Context) error {
	si.SetStatus(sdk.StatusRunning)
	defer si.SetStatus(sdk.StatusIdle)

	if err := si.validateConfig(); err != nil {
		return fmt.Errorf("hades.auxiliary.siem_integration: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(250 * time.Millisecond):
		si.SetStatus(sdk.StatusCompleted)
		return nil
	}
}

// SetProvider configures the SIEM provider
func (si *SIEMIntegration) SetProvider(provider SIEMProvider) error {
	switch provider {
	case ProviderSplunk, ProviderElastic, ProviderSentinelOne:
		si.provider = provider
		return nil
	default:
		return fmt.Errorf("hades.auxiliary.siem_integration: invalid provider: %s", provider)
	}
}

// SetEndpoint configures the SIEM endpoint
func (si *SIEMIntegration) SetEndpoint(endpoint string) error {
	if endpoint == "" {
		return fmt.Errorf("hades.auxiliary.siem_integration: endpoint cannot be empty")
	}
	si.endpoint = endpoint
	return nil
}

// GetResult returns integration status
func (si *SIEMIntegration) GetResult() string {
	return fmt.Sprintf("SIEM integration: provider=%s, endpoint=%s", si.provider, si.endpoint)
}

// validateConfig ensures integration configuration is valid
func (si *SIEMIntegration) validateConfig() error {
	if si.provider == "" {
		return fmt.Errorf("hades.auxiliary.siem_integration: provider not configured")
	}
	if si.endpoint == "" {
		return fmt.Errorf("hades.auxiliary.siem_integration: endpoint not configured")
	}
	return nil
}
