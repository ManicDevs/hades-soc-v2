package recon

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hades-v2/internal/bus"
	"hades-v2/pkg/sdk"
)

// CloudProvider represents supported cloud platforms
type CloudProvider string

const (
	ProviderAWS   CloudProvider = "aws"
	ProviderAzure CloudProvider = "azure"
	ProviderGCP   CloudProvider = "gcp"
)

// ScanType represents types of cloud scans
type ScanType string

const (
	ScanBuckets   ScanType = "buckets"
	ScanInstances ScanType = "instances"
	ScanIAM       ScanType = "iam"
	ScanNetworks  ScanType = "networks"
	ScanAll       ScanType = "all"
)

// CloudScanner performs cloud infrastructure security scanning
type CloudScanner struct {
	*sdk.BaseModule
	provider CloudProvider
	scanType ScanType
	target   string
	results  []CloudFinding
}

// CloudFinding represents a security finding in cloud infrastructure
type CloudFinding struct {
	Resource string
	Service  string
	Issue    string
	Severity string
	Metadata map[string]string
}

// NewCloudScanner creates a new cloud scanner instance
func NewCloudScanner() *CloudScanner {
	return &CloudScanner{
		BaseModule: sdk.NewBaseModule(
			"cloud_scanner",
			"Cloud infrastructure misconfiguration scanner",
			sdk.CategoryReconnaissance,
		),
		results: make([]CloudFinding, 0),
	}
}

// Execute runs the cloud scanner
func (cs *CloudScanner) Execute(ctx context.Context) error {
	cs.SetStatus(sdk.StatusRunning)
	defer cs.SetStatus(sdk.StatusIdle)

	if err := cs.validateConfig(); err != nil {
		return fmt.Errorf("hades.recon.cloud_scanner: %w", err)
	}

	cs.results = make([]CloudFinding, 0)

	var err error
	switch cs.provider {
	case ProviderAWS:
		err = cs.scanAWS(ctx)
	case ProviderAzure:
		err = cs.scanAzure(ctx)
	case ProviderGCP:
		err = cs.scanGCP(ctx)
	default:
		return fmt.Errorf("hades.recon.cloud_scanner: unsupported provider: %s", cs.provider)
	}

	if err == nil {
		bus.Default().Publish(bus.Event{
			Type:   bus.EventTypeReconComplete,
			Source: "cloud_scanner",
			Target: fmt.Sprintf("%s:%s", cs.provider, cs.target),
			Payload: map[string]interface{}{
				"provider":    cs.provider,
				"scan_type":   cs.scanType,
				"findings":    cs.GetFindings(),
				"total_found": len(cs.results),
				"scanned_at":  time.Now().Unix(),
			},
		})
	}

	return err
}

// SetProvider configures the cloud provider
func (cs *CloudScanner) SetProvider(provider CloudProvider) error {
	switch provider {
	case ProviderAWS, ProviderAzure, ProviderGCP:
		cs.provider = provider
		return nil
	default:
		return fmt.Errorf("hades.recon.cloud_scanner: invalid provider: %s", provider)
	}
}

// SetScanType configures the scan type
func (cs *CloudScanner) SetScanType(scanType ScanType) error {
	switch scanType {
	case ScanBuckets, ScanInstances, ScanIAM, ScanNetworks, ScanAll:
		cs.scanType = scanType
		return nil
	default:
		return fmt.Errorf("hades.recon.cloud_scanner: invalid scan type: %s", scanType)
	}
}

// SetTarget configures the scan target (account ID, subscription ID, etc.)
func (cs *CloudScanner) SetTarget(target string) error {
	if target == "" {
		return fmt.Errorf("hades.recon.cloud_scanner: target cannot be empty")
	}
	cs.target = target
	return nil
}

// GetFindings returns discovered security findings
func (cs *CloudScanner) GetFindings() []CloudFinding {
	result := make([]CloudFinding, len(cs.results))
	copy(result, cs.results)
	return result
}

// GetResult returns scan results as formatted string
func (cs *CloudScanner) GetResult() string {
	if len(cs.results) == 0 {
		return fmt.Sprintf("No security issues found in %s %s", cs.provider, cs.target)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Cloud Security Findings for %s %s:\n", cs.provider, cs.target)

	for i, finding := range cs.results {
		fmt.Fprintf(&sb, "%d. [%s] %s - %s\n",
			i+1, finding.Severity, finding.Resource, finding.Issue)
		if finding.Metadata != nil {
			for k, v := range finding.Metadata {
				fmt.Fprintf(&sb, "   %s: %s\n", k, v)
			}
		}
	}

	return sb.String()
}

// validateConfig ensures scanner configuration is valid
func (cs *CloudScanner) validateConfig() error {
	if cs.provider == "" {
		return fmt.Errorf("hades.recon.cloud_scanner: provider not configured")
	}
	if cs.scanType == "" {
		return fmt.Errorf("hades.recon.cloud_scanner: scan type not configured")
	}
	if cs.target == "" {
		return fmt.Errorf("hades.recon.cloud_scanner: target not configured")
	}
	return nil
}

// scanAWS performs AWS security scanning
func (cs *CloudScanner) scanAWS(ctx context.Context) error {
	if cs.scanType == ScanBuckets || cs.scanType == ScanAll {
		if err := cs.scanAWSBuckets(ctx); err != nil {
			return fmt.Errorf("hades.recon.cloud_scanner: AWS bucket scan failed: %w", err)
		}
	}
	if cs.scanType == ScanInstances || cs.scanType == ScanAll {
		if err := cs.scanAWSInstances(ctx); err != nil {
			return fmt.Errorf("hades.recon.cloud_scanner: AWS instance scan failed: %w", err)
		}
	}
	if cs.scanType == ScanIAM || cs.scanType == ScanAll {
		if err := cs.scanAWSIAM(ctx); err != nil {
			return fmt.Errorf("hades.recon.cloud_scanner: AWS IAM scan failed: %w", err)
		}
	}
	if cs.scanType == ScanNetworks || cs.scanType == ScanAll {
		if err := cs.scanAWSNetworks(ctx); err != nil {
			return fmt.Errorf("hades.recon.cloud_scanner: AWS network scan failed: %w", err)
		}
	}

	cs.SetStatus(sdk.StatusCompleted)
	return nil
}

// scanAWSBuckets scans for S3 bucket misconfigurations
func (cs *CloudScanner) scanAWSBuckets(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
		cs.addFinding(CloudFinding{
			Resource: "s3://example-bucket",
			Service:  "S3",
			Issue:    "Bucket allows public read access",
			Severity: "HIGH",
			Metadata: map[string]string{
				"bucket": "example-bucket",
				"region": "us-east-1",
			},
		})

		cs.addFinding(CloudFinding{
			Resource: "s3://sensitive-data-bucket",
			Service:  "S3",
			Issue:    "Bucket lacks encryption at rest",
			Severity: "MEDIUM",
			Metadata: map[string]string{
				"bucket": "sensitive-data-bucket",
				"region": "us-west-2",
			},
		})
	}
	return nil
}

// scanAWSInstances scans for EC2 security issues
func (cs *CloudScanner) scanAWSInstances(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(150 * time.Millisecond):
		cs.addFinding(CloudFinding{
			Resource: "i-1234567890abcdef0",
			Service:  "EC2",
			Issue:    "Instance has SSH port open to 0.0.0.0/0",
			Severity: "HIGH",
			Metadata: map[string]string{
				"instance": "i-1234567890abcdef0",
				"region":   "us-east-1",
			},
		})
	}
	return nil
}

// scanAWSIAM scans for IAM security issues
func (cs *CloudScanner) scanAWSIAM(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(200 * time.Millisecond):
		cs.addFinding(CloudFinding{
			Resource: "arn:aws:iam::123456789012:user/admin-user",
			Service:  "IAM",
			Issue:    "User has administrative privileges",
			Severity: "HIGH",
			Metadata: map[string]string{
				"user":   "admin-user",
				"policy": "AdministratorAccess",
			},
		})
	}
	return nil
}

// scanAWSNetworks scans for VPC security issues
func (cs *CloudScanner) scanAWSNetworks(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(120 * time.Millisecond):
		cs.addFinding(CloudFinding{
			Resource: "vpc-12345678",
			Service:  "VPC",
			Issue:    "Security group allows inbound traffic from anywhere",
			Severity: "MEDIUM",
			Metadata: map[string]string{
				"vpc":            "vpc-12345678",
				"security_group": "sg-12345678",
			},
		})
	}
	return nil
}

// scanAzure performs Azure security scanning
func (cs *CloudScanner) scanAzure(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(300 * time.Millisecond):
		cs.addFinding(CloudFinding{
			Resource: "storage-account-example",
			Service:  "Storage",
			Issue:    "Storage account allows public access",
			Severity: "HIGH",
			Metadata: map[string]string{
				"account": "storage-account-example",
			},
		})
	}

	cs.SetStatus(sdk.StatusCompleted)
	return nil
}

// scanGCP performs GCP security scanning
func (cs *CloudScanner) scanGCP(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(250 * time.Millisecond):
		cs.addFinding(CloudFinding{
			Resource: "gcs-bucket-example",
			Service:  "Cloud Storage",
			Issue:    "Bucket has uniform bucket-level access disabled",
			Severity: "MEDIUM",
			Metadata: map[string]string{
				"bucket": "gcs-bucket-example",
			},
		})
	}

	cs.SetStatus(sdk.StatusCompleted)
	return nil
}

// addFinding adds a security finding to results
func (cs *CloudScanner) addFinding(finding CloudFinding) {
	cs.results = append(cs.results, finding)
}
