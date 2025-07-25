package handlers

import (
	"net/http"

	"ripple/db"

	"github.com/gorilla/mux"
)

// UIHandler handles HTTP requests for UI-related operations
type UIHandler struct {
	repo *db.UIRepository
}

// NewUIHandler creates a new UI handler
func NewUIHandler(repo *db.UIRepository) *UIHandler {
	return &UIHandler{
		repo: repo,
	}
}

// RegisterRoutes registers the UI routes
func (h *UIHandler) RegisterRoutes(router *mux.Router) {
	// Create a subrouter for UI endpoints
	uiRouter := router.PathPrefix("/api/v1/ui").Subrouter()

	// Register UI routes
	uiRouter.HandleFunc("/stats", h.GetDashboardStats).Methods("GET")
	uiRouter.HandleFunc("/recent_activity", h.GetRecentActivity).Methods("GET")
	uiRouter.HandleFunc("/agent_versions", h.GetAgentVersions).Methods("GET")

}

// GetDashboardStats handles GET /api/v1/ui/stats
func (h *UIHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.GetDashboardStats()
	if err != nil {
		http.Error(w, "Failed to retrieve dashboard stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// GetRecentActivity handles GET /api/v1/ui/recent_activity
func (h *UIHandler) GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	activities, err := h.repo.GetRecentActivity()
	if err != nil {
		http.Error(w, "Failed to retrieve recent activity: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, activities)
}

func (h *UIHandler) GetAgentVersions(w http.ResponseWriter, r *http.Request) {
	agents, err := h.repo.GetAgentVersions(r.Context())
	if err != nil {
		http.Error(w, "Failed to get agents: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, agents)
}
