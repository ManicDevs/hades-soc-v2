package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Custom context key types to avoid staticcheck SA1029
type contextKey string

const (
	apiVersionKey         contextKey = "api_version"
	versioningStrategyKey contextKey = "versioning_strategy"
)

// Versioning Strategy Types
type VersioningStrategy string

const (
	URLPath          VersioningStrategy = "url_path"
	Header           VersioningStrategy = "header"
	QueryParam       VersioningStrategy = "query_param"
	ContentType      VersioningStrategy = "content_type"
	Semantic         VersioningStrategy = "semantic"
	DateBased        VersioningStrategy = "date_based"
	FeatureBased     VersioningStrategy = "feature_based"
	EnvironmentBased VersioningStrategy = "environment_based"
)

// Advanced Versioning Implementation
type AdvancedVersioningSystem struct {
	strategies map[VersioningStrategy]VersionStrategy
	config     AdvancedVersioningConfig
	router     *AdvancedVersionRouter
	negotiator *VersionNegotiator
	manager    *DeprecationManager
}

type AdvancedVersioningConfig struct {
	DefaultStrategy   VersioningStrategy `json:"default_strategy"`
	DefaultVersion    string             `json:"default_version"`
	SupportedVersions []string           `json:"supported_versions"`
	VersionHeaders    map[string]string  `json:"version_headers"`
	ContentTypes      map[string]string  `json:"content_types"`
	DeprecationPolicy string             `json:"deprecation_policy"`
	CacheConfig       CacheConfig        `json:"cache_config"`
}

type CacheConfig struct {
	Enabled       bool          `json:"enabled"`
	TTL           time.Duration `json:"ttl"`
	Strategy      string        `json:"strategy"`
	VaryByHeaders []string      `json:"vary_by_headers"`
}

// Enhanced Version Strategy Interface
type VersionStrategy interface {
	ExtractVersion(r *http.Request) string
	ApplyVersion(w http.ResponseWriter, version string)
	IsCacheable() bool
	GetCacheKey(r *http.Request) string
}

// URL Path Versioning Strategy
type URLPathVersioning struct {
	prefix string
}

func (upv *URLPathVersioning) ExtractVersion(r *http.Request) string {
	path := r.URL.Path
	if strings.HasPrefix(path, upv.prefix) {
		parts := strings.Split(path, "/")
		if len(parts) >= 3 && strings.HasPrefix(parts[2], "v") {
			return parts[2]
		}
	}
	return ""
}

func (upv *URLPathVersioning) ApplyVersion(w http.ResponseWriter, version string) {
	w.Header().Set("X-API-Versioning-Strategy", "url_path")
	w.Header().Set("X-API-Version-Source", "path")
}

func (upv *URLPathVersioning) IsCacheable() bool {
	return true
}

func (upv *URLPathVersioning) GetCacheKey(r *http.Request) string {
	return r.URL.Path
}

// Header Versioning Strategy
type HeaderVersioning struct {
	headerName     string
	defaultVersion string
}

func (hv *HeaderVersioning) ExtractVersion(r *http.Request) string {
	version := r.Header.Get(hv.headerName)
	if version != "" {
		return version
	}
	return hv.defaultVersion
}

func (hv *HeaderVersioning) ApplyVersion(w http.ResponseWriter, version string) {
	w.Header().Set(hv.headerName, version)
	w.Header().Set("X-API-Versioning-Strategy", "header")
	w.Header().Set("X-API-Version-Source", "header")
}

func (hv *HeaderVersioning) IsCacheable() bool {
	return false // Header-based versioning affects caching
}

func (hv *HeaderVersioning) GetCacheKey(r *http.Request) string {
	return fmt.Sprintf("%s:%s", r.URL.Path, r.Header.Get(hv.headerName))
}

// Query Parameter Versioning Strategy
type QueryParamVersioning struct {
	paramName string
}

func (qpv *QueryParamVersioning) ExtractVersion(r *http.Request) string {
	return r.URL.Query().Get(qpv.paramName)
}

func (qpv *QueryParamVersioning) ApplyVersion(w http.ResponseWriter, version string) {
	w.Header().Set("X-API-Versioning-Strategy", "query_param")
	w.Header().Set("X-API-Version-Source", "query")
}

func (qpv *QueryParamVersioning) IsCacheable() bool {
	return true
}

func (qpv *QueryParamVersioning) GetCacheKey(r *http.Request) string {
	return fmt.Sprintf("%s:%s", r.URL.Path, r.URL.Query().Get(qpv.paramName))
}

// Content-Type Versioning Strategy
type ContentTypeVersioning struct {
	vendorPrefix string
}

func (ctv *ContentTypeVersioning) ExtractVersion(r *http.Request) string {
	accept := r.Header.Get("Accept")
	contentType := r.Header.Get("Content-Type")

	// Check Accept header first
	if strings.Contains(accept, ctv.vendorPrefix) {
		parts := strings.Split(accept, ctv.vendorPrefix)
		if len(parts) > 1 {
			version := strings.Split(parts[1], "+")[0]
			return strings.TrimSuffix(version, ";")
		}
	}

	// Check Content-Type header
	if strings.Contains(contentType, ctv.vendorPrefix) {
		parts := strings.Split(contentType, ctv.vendorPrefix)
		if len(parts) > 1 {
			version := strings.Split(parts[1], "+")[0]
			return strings.TrimSuffix(version, ";")
		}
	}

	return ""
}

func (ctv *ContentTypeVersioning) ApplyVersion(w http.ResponseWriter, version string) {
	w.Header().Set("Content-Type", fmt.Sprintf("%s%s+json", ctv.vendorPrefix, version))
	w.Header().Set("X-API-Versioning-Strategy", "content_type")
	w.Header().Set("X-API-Version-Source", "content_negotiation")
}

func (ctv *ContentTypeVersioning) IsCacheable() bool {
	return true
}

func (ctv *ContentTypeVersioning) GetCacheKey(r *http.Request) string {
	return fmt.Sprintf("%s:%s", r.URL.Path, r.Header.Get("Accept"))
}

// Semantic Versioning Strategy
type SemanticVersioning struct {
	semanticVersions map[string]AdvancedSemanticVersion
	defaultVersion   string
}

type AdvancedSemanticVersion struct {
	Major int    `json:"major"`
	Minor int    `json:"minor"`
	Patch int    `json:"patch"`
	Pre   string `json:"pre,omitempty"`
}

func (sv *SemanticVersioning) ExtractVersion(r *http.Request) string {
	// Try multiple sources for semantic version
	version := r.Header.Get("API-Version")
	if version != "" && sv.isValidSemanticVersion(version) {
		return version
	}

	version = r.URL.Query().Get("version")
	if version != "" && sv.isValidSemanticVersion(version) {
		return version
	}

	// Extract from path
	path := r.URL.Path
	if strings.Contains(path, "/api/v") {
		parts := strings.Split(path, "/")
		for _, part := range parts {
			if strings.HasPrefix(part, "v") && sv.isValidSemanticVersion(part) {
				return part
			}
		}
	}

	return sv.defaultVersion
}

func (sv *SemanticVersioning) isValidSemanticVersion(version string) bool {
	if !strings.HasPrefix(version, "v") {
		return false
	}

	version = strings.TrimPrefix(version, "v")
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return false
	}

	for _, part := range parts[:2] {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}

	return true
}

func (sv *SemanticVersioning) ApplyVersion(w http.ResponseWriter, version string) {
	w.Header().Set("API-Version", version)
	w.Header().Set("X-API-Versioning-Strategy", "semantic")
	w.Header().Set("X-API-Version-Format", "semver")
}

func (sv *SemanticVersioning) IsCacheable() bool {
	return true
}

func (sv *SemanticVersioning) GetCacheKey(r *http.Request) string {
	return fmt.Sprintf("%s:%s", r.URL.Path, sv.ExtractVersion(r))
}

// Date-Based Versioning Strategy
type DateBasedVersioning struct {
	format string
}

func (dbv *DateBasedVersioning) ExtractVersion(r *http.Request) string {
	version := r.Header.Get("API-Version")
	if version != "" && dbv.isValidDateVersion(version) {
		return version
	}

	version = r.URL.Query().Get("version")
	if version != "" && dbv.isValidDateVersion(version) {
		return version
	}

	// Default to current date in specified format
	return time.Now().Format(dbv.format)
}

func (dbv *DateBasedVersioning) isValidDateVersion(version string) bool {
	// Check if version matches date format
	if dbv.format == "2006.01.02" {
		parts := strings.Split(version, ".")
		if len(parts) == 3 {
			_, err1 := strconv.Atoi(parts[0])
			_, err2 := strconv.Atoi(parts[1])
			_, err3 := strconv.Atoi(parts[2])
			return err1 == nil && err2 == nil && err3 == nil
		}
	}

	if dbv.format == "2006-01-02" {
		parts := strings.Split(version, "-")
		if len(parts) == 3 {
			_, err1 := strconv.Atoi(parts[0])
			_, err2 := strconv.Atoi(parts[1])
			_, err3 := strconv.Atoi(parts[2])
			return err1 == nil && err2 == nil && err3 == nil
		}
	}

	return false
}

func (dbv *DateBasedVersioning) ApplyVersion(w http.ResponseWriter, version string) {
	w.Header().Set("API-Version", version)
	w.Header().Set("X-API-Versioning-Strategy", "date_based")
	w.Header().Set("X-API-Version-Format", dbv.format)
}

func (dbv *DateBasedVersioning) IsCacheable() bool {
	return true
}

func (dbv *DateBasedVersioning) GetCacheKey(r *http.Request) string {
	return fmt.Sprintf("%s:%s", r.URL.Path, dbv.ExtractVersion(r))
}

// Feature-Based Versioning Strategy
type FeatureBasedVersioning struct {
	featureVersions map[string]*AdvancedFeatureVersion
	defaultVersion  string
}

type AdvancedFeatureVersion struct {
	Version  string    `json:"version"`
	Features []string  `json:"features"`
	Added    time.Time `json:"added"`
	Stable   bool      `json:"stable"`
}

func (fbv *FeatureBasedVersioning) ExtractVersion(r *http.Request) string {
	// Check for specific feature requirements
	requiredFeatures := r.URL.Query()["feature"]
	if len(requiredFeatures) > 0 {
		for _, version := range fbv.getVersionsWithFeatures(requiredFeatures) {
			return version
		}
	}

	// Fall back to default version detection
	version := r.Header.Get("API-Version")
	if version != "" {
		return version
	}

	return fbv.defaultVersion
}

func (fbv *FeatureBasedVersioning) getVersionsWithFeatures(features []string) []string {
	var versions []string
	for version, fv := range fbv.featureVersions {
		hasAllFeatures := true
		for _, requiredFeature := range features {
			found := false
			for _, availableFeature := range fv.Features {
				if availableFeature == requiredFeature {
					found = true
					break
				}
			}
			if !found {
				hasAllFeatures = false
				break
			}
		}
		if hasAllFeatures {
			versions = append(versions, version)
		}
	}
	return versions
}

func (fbv *FeatureBasedVersioning) ApplyVersion(w http.ResponseWriter, version string) {
	w.Header().Set("API-Version", version)
	w.Header().Set("X-API-Versioning-Strategy", "feature_based")
	if fv, exists := fbv.featureVersions[version]; exists {
		w.Header().Set("X-API-Features", strings.Join(fv.Features, ","))
		w.Header().Set("X-API-Version-Stable", strconv.FormatBool(fv.Stable))
	}
}

func (fbv *FeatureBasedVersioning) IsCacheable() bool {
	return true
}

func (fbv *FeatureBasedVersioning) GetCacheKey(r *http.Request) string {
	features := r.URL.Query()["feature"]
	featureKey := strings.Join(features, ",")
	return fmt.Sprintf("%s:%s:%s", r.URL.Path, fbv.ExtractVersion(r), featureKey)
}

// Environment-Based Versioning Strategy
type EnvironmentBasedVersioning struct {
	environments map[string]*AdvancedEnvironmentVersion
	currentEnv   string
}

type AdvancedEnvironmentVersion struct {
	Environment string   `json:"environment"`
	Version     string   `json:"version"`
	Stable      bool     `json:"stable"`
	Features    []string `json:"features"`
}

func (ebv *EnvironmentBasedVersioning) ExtractVersion(r *http.Request) string {
	// Check environment header
	env := r.Header.Get("X-Environment")
	if env == "" {
		env = ebv.currentEnv
	}

	if envVersion, exists := ebv.environments[env]; exists {
		return envVersion.Version
	}

	return ebv.environments["production"].Version
}

func (ebv *EnvironmentBasedVersioning) ApplyVersion(w http.ResponseWriter, version string) {
	w.Header().Set("API-Version", version)
	w.Header().Set("X-API-Versioning-Strategy", "environment_based")

	// Find which environment this version belongs to
	for env, envVersion := range ebv.environments {
		if envVersion.Version == version {
			w.Header().Set("X-API-Environment", env)
			w.Header().Set("X-API-Version-Stable", strconv.FormatBool(envVersion.Stable))
			if len(envVersion.Features) > 0 {
				w.Header().Set("X-API-Features", strings.Join(envVersion.Features, ","))
			}
			break
		}
	}
}

func (ebv *EnvironmentBasedVersioning) IsCacheable() bool {
	return true
}

func (ebv *EnvironmentBasedVersioning) GetCacheKey(r *http.Request) string {
	env := r.Header.Get("X-Environment")
	if env == "" {
		env = ebv.currentEnv
	}
	return fmt.Sprintf("%s:%s:%s", r.URL.Path, env, ebv.ExtractVersion(r))
}

// Advanced Version Router
type AdvancedVersionRouter struct {
	strategies      map[VersioningStrategy]VersionStrategy
	defaultStrategy VersioningStrategy
	cache           *VersionCache
}

type VersionCache struct {
	enabled bool
	ttl     time.Duration
	store   map[string]*CacheEntry
}

type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
	Headers   http.Header
}

func NewAdvancedVersioningSystem(config AdvancedVersioningConfig) *AdvancedVersioningSystem {
	avs := &AdvancedVersioningSystem{
		strategies: make(map[VersioningStrategy]VersionStrategy),
		config:     config,
		router:     NewAdvancedVersionRouter(config),
		negotiator: NewVersionNegotiator(config.SupportedVersions, config.DefaultVersion),
		manager:    NewDeprecationManager(),
	}

	// Initialize all strategies
	avs.initializeStrategies()

	return avs
}

func (avs *AdvancedVersioningSystem) initializeStrategies() {
	// URL Path Versioning
	avs.strategies[URLPath] = &URLPathVersioning{
		prefix: "/api",
	}

	// Header Versioning
	avs.strategies[Header] = &HeaderVersioning{
		headerName:     "API-Version",
		defaultVersion: avs.config.DefaultVersion,
	}

	// Query Parameter Versioning
	avs.strategies[QueryParam] = &QueryParamVersioning{
		paramName: "version",
	}

	// Content-Type Versioning
	avs.strategies[ContentType] = &ContentTypeVersioning{
		vendorPrefix: "application/vnd.hades.",
	}

	// Semantic Versioning
	avs.strategies[Semantic] = &SemanticVersioning{
		semanticVersions: map[string]AdvancedSemanticVersion{
			"v2.1.3": {Major: 2, Minor: 1, Patch: 3},
			"v2.1.0": {Major: 2, Minor: 1, Patch: 0},
			"v2.0.0": {Major: 2, Minor: 0, Patch: 0},
		},
		defaultVersion: "v2.0.0",
	}

	// Date-Based Versioning
	avs.strategies[DateBased] = &DateBasedVersioning{
		format: "2006.01.02",
	}

	// Feature-Based Versioning
	avs.strategies[FeatureBased] = &FeatureBasedVersioning{
		featureVersions: map[string]*AdvancedFeatureVersion{
			"analytics-enabled": {
				Version:  "analytics-enabled",
				Features: []string{"real-time-analytics", "custom-dashboards"},
				Added:    time.Now().Add(-30 * 24 * time.Hour),
				Stable:   true,
			},
			"ml-integration": {
				Version:  "ml-integration",
				Features: []string{"threat-prediction", "anomaly-detection"},
				Added:    time.Now().Add(-7 * 24 * time.Hour),
				Stable:   false,
			},
		},
		defaultVersion: "analytics-enabled",
	}

	// Environment-Based Versioning
	avs.strategies[EnvironmentBased] = &EnvironmentBasedVersioning{
		environments: map[string]*AdvancedEnvironmentVersion{
			"development": {
				Environment: "development",
				Version:     "dev-latest",
				Stable:      false,
				Features:    []string{"debug-mode", "verbose-logging"},
			},
			"staging": {
				Environment: "staging",
				Version:     "staging-v2.1",
				Stable:      true,
				Features:    []string{"pre-release-features"},
			},
			"production": {
				Environment: "production",
				Version:     "prod-v2.0",
				Stable:      true,
				Features:    []string{"optimized-performance"},
			},
		},
		currentEnv: "production",
	}
}

// Advanced Middleware
func (avs *AdvancedVersioningSystem) AdvancedVersioningMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Determine version using multiple strategies
		version := avs.determineVersion(r)

		// Apply version headers
		avs.applyVersionHeaders(w, r, version)

		// Check deprecation
		avs.checkDeprecation(w, version)

		// Add version to context
		ctx := context.WithValue(r.Context(), apiVersionKey, version)
		ctx = context.WithValue(ctx, versioningStrategyKey, avs.getActiveStrategy(r))

		// Add performance metrics
		startTime := time.Now()

		// Call next handler
		next.ServeHTTP(w, r.WithContext(ctx))

		// Record metrics
		duration := time.Since(startTime)
		avs.recordMetrics(version, duration, r)
	})
}

func (avs *AdvancedVersioningSystem) determineVersion(r *http.Request) string {
	// Try strategies in order of preference
	strategies := []VersioningStrategy{URLPath, Header, ContentType, QueryParam, Semantic, DateBased, FeatureBased, EnvironmentBased}

	for _, strategy := range strategies {
		if strat, exists := avs.strategies[strategy]; exists {
			if version := strat.ExtractVersion(r); version != "" {
				return version
			}
		}
	}

	return avs.config.DefaultVersion
}

func (avs *AdvancedVersioningSystem) applyVersionHeaders(w http.ResponseWriter, r *http.Request, version string) {
	// Add standard headers
	w.Header().Set("API-Version", version)
	w.Header().Set("X-API-Supported-Versions", strings.Join(avs.config.SupportedVersions, ", "))
	w.Header().Set("X-API-Versioning-Timestamp", time.Now().Format(time.RFC3339))

	// Add strategy-specific headers
	strategy := avs.getActiveStrategy(r)
	if strategy != nil {
		strategy.ApplyVersion(w, version)
	}

	// Add caching headers if applicable
	if avs.config.CacheConfig.Enabled && strategy.IsCacheable() {
		cacheKey := strategy.GetCacheKey(r)
		w.Header().Set("X-API-Cache-Key", cacheKey)
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", int(avs.config.CacheConfig.TTL.Seconds())))
	}
}

func (avs *AdvancedVersioningSystem) getActiveStrategy(r *http.Request) VersionStrategy {
	for _, strategy := range []VersioningStrategy{URLPath, Header, ContentType, QueryParam} {
		if strat, exists := avs.strategies[strategy]; exists {
			if strat.ExtractVersion(r) != "" {
				return strat
			}
		}
	}
	return avs.strategies[URLPath] // Default fallback
}

func (avs *AdvancedVersioningSystem) checkDeprecation(w http.ResponseWriter, version string) {
	// Implementation would check version against deprecation policies
	// For now, assume no versions are deprecated
}

func (avs *AdvancedVersioningSystem) recordMetrics(version string, duration time.Duration, r *http.Request) {
	// Implementation would record metrics for monitoring
	// This would integrate with your analytics system
}

// Helper functions
func NewAdvancedVersionRouter(config AdvancedVersioningConfig) *AdvancedVersionRouter {
	return &AdvancedVersionRouter{
		strategies:      make(map[VersioningStrategy]VersionStrategy),
		defaultStrategy: config.DefaultStrategy,
		cache: &VersionCache{
			enabled: config.CacheConfig.Enabled,
			ttl:     config.CacheConfig.TTL,
			store:   make(map[string]*CacheEntry),
		},
	}
}

func NewVersionNegotiator(supportedVersions []string, defaultVersion string) *VersionNegotiator {
	return &VersionNegotiator{
		supportedVersions: supportedVersions,
		defaultVersion:    defaultVersion,
	}
}
