package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vaintrub/go-ddd-template/internal/common/db"
)

// HealthHandler provides HTTP handlers for health and readiness checks.
type HealthHandler struct {
	dbPool *pgxpool.Pool
}

// NewHealthHandler creates a new health check handler.
func NewHealthHandler(dbPool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{
		dbPool: dbPool,
	}
}

// HandleHealth handles GET /health requests.
// Returns 200 OK if all systems are healthy, 503 Service Unavailable otherwise.
//
// Response format:
//
//	{
//	  "status": "healthy",
//	  "timestamp": "2025-11-18T12:00:00Z",
//	  "response_time_ms": 15,
//	  "connection_stats": {
//	    "max_connections": 25,
//	    "total_connections": 10,
//	    "idle_connections": 8,
//	    "acquired_connections": 2
//	  }
//	}
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Set a timeout for the health check
	// This ensures the health endpoint responds quickly
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	status := db.CheckHealth(timeoutCtx, h.dbPool)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Set HTTP status code based on health status
	if status.Status == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	// Encode and send response
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Failed to encode health status", http.StatusInternalServerError)
		return
	}
}

// HandleReadiness handles GET /readiness requests.
// This is useful for Kubernetes readiness probes.
// Returns 200 OK if the service is ready to accept traffic, 503 otherwise.
func (h *HealthHandler) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Set a short timeout for readiness check
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Check if database is healthy
	if !db.IsHealthy(timeoutCtx, h.dbPool) {
		http.Error(w, "Service not ready", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
	})
}

// HandleLiveness handles GET /liveness requests.
// This is useful for Kubernetes liveness probes.
// Returns 200 OK if the service is alive (even if database is down).
func (h *HealthHandler) HandleLiveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "alive",
	})
}
