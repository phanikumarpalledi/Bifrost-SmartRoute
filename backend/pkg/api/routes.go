package api

import (
	"net/http"
)

// SetupRouter initializes the HTTP mux with CORS and routing
func SetupRouter(h *APIHandler) http.Handler {
	mux := http.NewServeMux()

	// OpenAI-compatible Chat Completions
	mux.HandleFunc("/v1/chat/completions", h.HandleChatCompletions)

	// SmartRoute Core APIs
	mux.HandleFunc("/api/smartroute/route", h.HandleSmartRoute)
	mux.HandleFunc("/api/smartroute/metrics", h.HandleGetMetrics)
	mux.HandleFunc("/api/smartroute/analytics/cost", h.HandleGetCostOverTime)
	mux.HandleFunc("/api/smartroute/analytics/providers", h.HandleGetProviderPerformances)
	mux.HandleFunc("/api/smartroute/analytics/models", h.HandleGetModelUsage)
	mux.HandleFunc("/api/smartroute/logs", h.HandleGetLogs)
	mux.HandleFunc("/api/smartroute/failovers", h.HandleGetFailovers)
	mux.HandleFunc("/api/smartroute/providers", h.HandleGetProviders)
	mux.HandleFunc("/api/smartroute/failover/simulate", h.HandleSimulateFailover)
	mux.HandleFunc("/api/smartroute/rules", h.HandleRules)
	mux.HandleFunc("/api/smartroute/settings", h.HandleSettings)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"bifrost-smartroute","version":"1.0.0"}`))
	})

	// Wrap with permissive CORS for React frontend development
	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-SmartRoute-Override")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
