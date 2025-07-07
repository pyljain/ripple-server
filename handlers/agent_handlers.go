package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"ripple/db"
	"ripple/models"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AgentHandler handles HTTP requests for agent operations
type AgentHandler struct {
	repo *db.AgentRepository
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(repo *db.AgentRepository) *AgentHandler {
	return &AgentHandler{
		repo: repo,
	}
}

// RegisterRoutes registers the agent routes
func (h *AgentHandler) RegisterRoutes(router *mux.Router) {
	// Agent routes
	router.HandleFunc("/api/v1/agents", h.ListAgents).Methods("GET")
	router.HandleFunc("/api/v1/agents/{name}/register", h.RegisterAgent).Methods("POST")

	// Agent version routes
	router.HandleFunc("/api/v1/agents/{agentId}/versions", h.AddAgentVersion).Methods("POST")
	router.HandleFunc("/api/v1/agents/{agentId}/versions", h.GetAgentVersions).Methods("GET")
	router.HandleFunc("/api/v1/agents/{agentId}/versions/{version}", h.GetAgentVersion).Methods("GET")

	// Agent run routes
	router.HandleFunc("/api/v1/agents/{agentId}/versions/{version}/runs", h.AddAgentRun).Methods("POST")
	router.HandleFunc("/api/v1/agents/{agentId}/versions/{version}/runs", h.GetAgentVersionRuns).Methods("GET")
	router.HandleFunc("/api/v1/agents/{agentId}/runs", h.GetAgentRuns).Methods("GET")
}

// ListAgents handles GET /api/v1/agents
func (h *AgentHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := h.repo.ListAgents()
	if err != nil {
		http.Error(w, "Failed to retrieve agents: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, agents)
}

// RegisterAgent handles POST /api/v1/agents/{name}/register
func (h *AgentHandler) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	var req models.RegisterAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Use the name from the URL if not provided in the request
	if req.Name == "" {
		req.Name = name
	} else if req.Name != name {
		http.Error(w, "Agent name in URL and request body must match", http.StatusBadRequest)
		return
	}

	agent := &models.Agent{
		Name:    req.Name,
		Project: req.Project,
	}

	if err := h.repo.CreateAgent(agent); err != nil {
		http.Error(w, "Failed to create agent: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, agent)
}

// AddAgentVersion handles POST /api/v1/agents/{agentId}/versions
func (h *AgentHandler) AddAgentVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentIDStr := vars["agentId"]

	agentID, err := primitive.ObjectIDFromHex(agentIDStr)
	if err != nil {
		http.Error(w, "Invalid agent ID format", http.StatusBadRequest)
		return
	}

	var req models.RegisterAgentVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	version := &models.AgentVersion{
		AgentID:    agentID,
		Version:    req.Version,
		Cluster:    req.Cluster,
		Tools:      req.Tools,
		Models:     req.Models,
		Deployment: req.Deployment,
	}

	if err := h.repo.CreateAgentVersion(version); err != nil {
		http.Error(w, "Failed to create agent version: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, version)
}

// GetAgentVersions handles GET /api/v1/agents/{agentId}/versions
func (h *AgentHandler) GetAgentVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentIDStr := vars["agentId"]

	agentID, err := primitive.ObjectIDFromHex(agentIDStr)
	if err != nil {
		http.Error(w, "Invalid agent ID format", http.StatusBadRequest)
		return
	}

	versions, err := h.repo.GetAgentVersions(agentID)
	if err != nil {
		http.Error(w, "Failed to retrieve agent versions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, versions)
}

// AddAgentRun handles POST /api/v1/agents/{agentId}/versions/{version}/runs
func (h *AgentHandler) AddAgentRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentIDStr := vars["agentId"]
	versionStr := vars["version"]

	agentID, err := primitive.ObjectIDFromHex(agentIDStr)
	if err != nil {
		http.Error(w, "Invalid agent ID format", http.StatusBadRequest)
		return
	}

	// Try to decode as a batch request first
	var batchReq models.RegisterAgentRunBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&batchReq); err != nil || len(batchReq.Runs) == 0 {
		// If it's not a batch request or empty batch, reset the body and try as a single request
		r.Body.Close()
		// Create a new reader from the raw request body
		if r.GetBody != nil {
			r.Body, _ = r.GetBody()
		} else {
			http.Error(w, "Failed to reset request body for single run processing", http.StatusBadRequest)
			return
		}

		// Process as a single run
		var req models.RegisterAgentRunRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Parse the created time
		var createdTime time.Time
		if req.Created != "" {
			createdTime, err = time.Parse(time.RFC3339, req.Created)
			if err != nil {
				createdTime = time.Now()
			}
		} else {
			createdTime = time.Now()
		}

		run := &models.AgentRun{
			AgentID:   agentID,
			Version:   versionStr,
			Created:   createdTime,
			Status:    req.Status,
			TimeTaken: req.TimeTaken,
			Initiator: req.Initiator,
			Tools:     req.Tools,
			Cost:      req.Cost,
			Models:    req.Models,
			RunID:     req.RunID,
			TaskID:    req.TaskID,
		}

		if err := h.repo.CreateAgentRun(run); err != nil {
			http.Error(w, "Failed to create agent run: "+err.Error(), http.StatusInternalServerError)
			return
		}

		respondJSON(w, http.StatusCreated, run)
		return
	}

	// Process as a batch request
	runs := make([]*models.AgentRun, len(batchReq.Runs))
	for i, reqRun := range batchReq.Runs {
		// Parse the created time
		var createdTime time.Time
		if reqRun.Created != "" {
			createdTime, err = time.Parse(time.RFC3339, reqRun.Created)
			if err != nil {
				createdTime = time.Now()
			}
		} else {
			createdTime = time.Now()
		}

		runs[i] = &models.AgentRun{
			AgentID:   agentID,
			Version:   versionStr,
			Created:   createdTime,
			Status:    reqRun.Status,
			TimeTaken: reqRun.TimeTaken,
			Initiator: reqRun.Initiator,
			Tools:     reqRun.Tools,
			Cost:      reqRun.Cost,
			Models:    reqRun.Models,
			RunID:     reqRun.RunID,
			TaskID:    reqRun.TaskID,
		}
	}

	if err := h.repo.CreateAgentRunBatch(runs); err != nil {
		http.Error(w, "Failed to create agent runs batch: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, runs)
}

// GetAgentVersionRuns handles GET /api/v1/agents/{agentId}/versions/{version}/runs
func (h *AgentHandler) GetAgentVersionRuns(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentIDStr := vars["agentId"]
	versionStr := vars["version"]

	agentID, err := primitive.ObjectIDFromHex(agentIDStr)
	if err != nil {
		http.Error(w, "Invalid agent ID format", http.StatusBadRequest)
		return
	}

	runs, err := h.repo.GetAgentVersionRuns(agentID, versionStr)
	if err != nil {
		http.Error(w, "Failed to retrieve agent runs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, runs)
}

// GetAgentRuns handles GET /api/v1/agents/{agentId}/runs
func (h *AgentHandler) GetAgentRuns(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentIDStr := vars["agentId"]

	agentID, err := primitive.ObjectIDFromHex(agentIDStr)
	if err != nil {
		http.Error(w, "Invalid agent ID format", http.StatusBadRequest)
		return
	}

	runs, err := h.repo.GetAgentRuns(agentID)
	if err != nil {
		http.Error(w, "Failed to retrieve agent runs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, runs)
}

// GetAgentVersion handles GET /api/v1/agents/{agentId}/versions/{version}
func (h *AgentHandler) GetAgentVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentIDStr := vars["agentId"]
	versionStr := vars["version"]

	agentID, err := primitive.ObjectIDFromHex(agentIDStr)
	if err != nil {
		http.Error(w, "Invalid agent ID format", http.StatusBadRequest)
		return
	}

	version, err := h.repo.GetAgentVersion(agentID, versionStr)
	if err != nil {
		if err.Error() == "version not found for this agent" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve agent version: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	respondJSON(w, http.StatusOK, version)
}

// Helper function to respond with JSON
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
