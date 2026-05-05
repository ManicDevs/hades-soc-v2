package versioning

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// WriteJSONResponse securely writes JSON response
func WriteJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Handlers for versioning endpoints

// VersionInfoHandler returns comprehensive version information
func (vm *VersionManager) VersionInfoHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"versions":           vm.GetAllVersions(),
		"default_version":    vm.defaultVersion,
		"preferred_version":  vm.preferredVersion,
		"supported_versions": vm.GetSupportedVersions(),
		"versioning_strategies": map[string]string{
			"url_path":     "/api/v{version}/endpoint",
			"header":       "API-Version: v{version}",
			"query_param":  "?version=v{version}",
			"content_type": "application/vnd.hades.v{version}+json",
		},
		"deprecation_policy": vm.config.DeprecationPolicy,
		"migration_paths":    vm.getMigrationPaths(),
		"recommendations": map[string]string{
			"new_users":      vm.preferredVersion,
			"existing_users": "v1 users should migrate to v2",
			"early_adopters": "v3 beta available for testing",
		},
	}

	WriteJSONResponse(w, response)
}

// VersionHandler returns specific version information
func (vm *VersionManager) VersionHandler(w http.ResponseWriter, r *http.Request) {
	version := r.URL.Query().Get("version")
	if version == "" {
		version = vm.preferredVersion
	}

	v, err := vm.GetVersion(version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Add migration information
	migrationInfo := map[string]interface{}{}
	if version == "v1" {
		if path, err := vm.GetMigrationPath("v1", "v2"); err == nil {
			migrationInfo["to_v2"] = path
		}
	}
	if version == "v2" {
		if path, err := vm.GetMigrationPath("v2", "v3"); err == nil {
			migrationInfo["to_v3"] = path
		}
	}

	response := map[string]interface{}{
		"version":         v,
		"migration":       migrationInfo,
		"recommendations": vm.getRecommendations(version),
	}

	WriteJSONResponse(w, response)
}

// MigrationHandler returns migration information
func (vm *VersionManager) MigrationHandler(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" || to == "" {
		http.Error(w, "Both 'from' and 'to' parameters are required", http.StatusBadRequest)
		return
	}

	path, err := vm.GetMigrationPath(from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"migration_path": path,
		"steps":          vm.getMigrationSteps(from, to),
		"compatibility":  vm.getCompatibilityMatrix(from, to),
		"timeline":       vm.getMigrationTimeline(from, to),
	}

	WriteJSONResponse(w, response)
}

// HealthHandler returns version-aware health information
func (vm *VersionManager) HealthHandler(w http.ResponseWriter, r *http.Request) {
	version := r.Context().Value("api_version").(string)

	health := map[string]interface{}{
		"status":         "healthy",
		"timestamp":      time.Now(),
		"version":        version,
		"version_status": vm.GetVersionStatus(version),
		"api_versions": map[string]interface{}{
			"v1": map[string]string{"status": StatusLegacy, "support": "deprecated"},
			"v2": map[string]string{"status": StatusPreferred, "support": "full"},
			"v3": map[string]string{"status": StatusBeta, "support": "preview"},
		},
		"recommendations": map[string]string{
			"upgrade": "v2 is the recommended version",
			"beta":    "v3 is available for testing",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(health); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Helper functions

func (vm *VersionManager) getMigrationPaths() map[string]interface{} {
	paths := make(map[string]interface{})

	// v1 -> v2
	if path, err := vm.GetMigrationPath("v1", "v2"); err == nil {
		paths["v1_to_v2"] = path
	}

	// v2 -> v3
	if path, err := vm.GetMigrationPath("v2", "v3"); err == nil {
		paths["v2_to_v3"] = path
	}

	// v1 -> v3 (not recommended)
	if path, err := vm.GetMigrationPath("v1", "v3"); err == nil {
		path.Complexity = "very_high"
		path.Automated = false
		paths["v1_to_v3"] = path
	}

	return paths
}

func (vm *VersionManager) getRecommendations(version string) []string {
	var recommendations []string

	switch version {
	case "v1":
		recommendations = append(recommendations, "This version is deprecated and will be sunset in 6 months")
		recommendations = append(recommendations, "Migrate to v2 as soon as possible")
		recommendations = append(recommendations, "Use the automated migration guide available")
		recommendations = append(recommendations, "Contact support for migration assistance")

	case "v2":
		recommendations = append(recommendations, "This is the recommended stable version")
		recommendations = append(recommendations, "All features are fully supported")
		recommendations = append(recommendations, "Consider testing v3 for new features")
		recommendations = append(recommendations, "Keep your dependencies up to date")

	case "v3":
		recommendations = append(recommendations, "This is a beta version for testing only")
		recommendations = append(recommendations, "Not recommended for production use")
		recommendations = append(recommendations, "Provide feedback through the feedback channel")
		recommendations = append(recommendations, "Expect breaking changes before stable release")
	}

	return recommendations
}

func (vm *VersionManager) getMigrationSteps(from, to string) []string {
	var steps []string

	if from == "v1" && to == "v2" {
		steps = append(steps, "Review migration guide")
		steps = append(steps, "Update authentication endpoints")
		steps = append(steps, "Migrate data models")
		steps = append(steps, "Update error handling")
		steps = append(steps, "Test new features")
		steps = append(steps, "Deploy to staging")
		steps = append(steps, "Deploy to production")
	}

	if from == "v2" && to == "v3" {
		steps = append(steps, "Review beta documentation")
		steps = append(steps, "Set up beta environment")
		steps = append(steps, "Test ML features")
		steps = append(steps, "Implement automation workflows")
		steps = append(steps, "Test multi-tenant features")
		steps = append(steps, "Provide feedback")
		steps = append(steps, "Monitor for stability")
	}

	return steps
}

func (vm *VersionManager) getCompatibilityMatrix(from, to string) map[string]interface{} {
	compatibility := map[string]interface{}{
		"authentication": "compatible",
		"data_models":    "compatible",
		"endpoints":      "compatible",
		"features":       "enhanced",
	}

	if from == "v1" && to == "v2" {
		compatibility["authentication"] = "enhanced"
		compatibility["data_models"] = "enhanced"
		compatibility["endpoints"] = "enhanced"
	}

	if from == "v2" && to == "v3" {
		compatibility["authentication"] = "compatible"
		compatibility["data_models"] = "enhanced"
		compatibility["endpoints"] = "enhanced"
		compatibility["features"] = "experimental"
	}

	return compatibility
}

func (vm *VersionManager) getMigrationTimeline(from, to string) map[string]interface{} {
	timeline := map[string]interface{}{
		"estimated_duration": "2 hours",
		"downtime_expected":  false,
		"rollback_available": true,
		"testing_required":   true,
		"backup_needed":      true,
	}

	if from == "v1" && to == "v2" {
		timeline["estimated_duration"] = "2 hours"
		timeline["downtime_expected"] = false
	}

	if from == "v2" && to == "v3" {
		timeline["estimated_duration"] = "4 hours"
		timeline["downtime_expected"] = false
		timeline["testing_required"] = true
		timeline["backup_needed"] = true
	}

	return timeline
}
