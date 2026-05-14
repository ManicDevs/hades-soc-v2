package api

import (
	"encoding/json"
	"net/http"

	"hades-v2/internal/hotplug"

	"github.com/gorilla/mux"
)

func RegisterHotplugRoutes(r *mux.Router) {
	api := r.PathPrefix("/api/v2/hotplug").Subrouter()
	api.HandleFunc("/load", handleHotplugLoad).Methods("POST")
	api.HandleFunc("/unload", handleHotplugUnload).Methods("POST")
	api.HandleFunc("/reload", handleHotplugReload).Methods("POST")
	api.HandleFunc("/list", handleHotplugList).Methods("GET")
	api.HandleFunc("/status", handleHotplugStatus).Methods("GET")
	api.HandleFunc("/enable", handleHotplugEnable).Methods("POST")
	api.HandleFunc("/disable", handleHotplugDisable).Methods("POST")
}

func handleHotplugLoad(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	loader := hotplug.GetGlobalLoader()
	module, err := loader.LoadModule(req.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "loaded",
		"name":     module.Name,
		"version":  module.Version,
		"checksum": module.Checksum,
	})
}

func handleHotplugUnload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	loader := hotplug.GetGlobalLoader()
	if err := loader.UnloadModule(req.Path); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "unloaded"})
}

func handleHotplugReload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	loader := hotplug.GetGlobalLoader()
	module, err := loader.ReloadModule(req.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "reloaded",
		"name":    module.Name,
		"version": module.Version,
	})
}

func handleHotplugList(w http.ResponseWriter, r *http.Request) {
	loader := hotplug.GetGlobalLoader()
	modules := loader.ListModules()

	result := make([]map[string]interface{}, 0, len(modules))
	for _, mod := range modules {
		result = append(result, map[string]interface{}{
			"name":      mod.Name,
			"version":   mod.Version,
			"path":      mod.Path,
			"checksum":  mod.Checksum,
			"loaded_at": mod.LoadedAt,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"modules": result,
		"count":   len(result),
	})
}

func handleHotplugStatus(w http.ResponseWriter, r *http.Request) {
	loader := hotplug.GetGlobalLoader()
	status := loader.GetStatus()

	json.NewEncoder(w).Encode(status)
}

func handleHotplugEnable(w http.ResponseWriter, r *http.Request) {
	loader := hotplug.GetGlobalLoader()
	loader.Enable()

	json.NewEncoder(w).Encode(map[string]string{"status": "enabled"})
}

func handleHotplugDisable(w http.ResponseWriter, r *http.Request) {
	loader := hotplug.GetGlobalLoader()
	loader.Disable()

	json.NewEncoder(w).Encode(map[string]string{"status": "disabled"})
}
