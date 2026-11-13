package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/maximhq/bifrost/plugins/smartroute/pkg/bifrost_plugin"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/cost"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/failover"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/router"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/storage"
)

func setupTestServer(t *testing.T) (http.Handler, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "api_test.db")
	db, err := storage.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	_ = db.SeedInitialData()

	failoverMgr := failover.NewManager(3, time.Minute)
	costCalc := cost.NewCalculator("gpt-4o")
	engine := router.NewEngine(db, failoverMgr, costCalc)
	plugin := bifrost_plugin.NewSmartRoutePlugin(db, engine, costCalc, failoverMgr, true)

	handler := NewAPIHandler(db, engine, costCalc, failoverMgr, plugin)
	routes := SetupRouter(handler)

	return routes, func() {
		db.Close()
	}
}

func TestAPI_SmartRoute(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	body := map[string]string{
		"prompt": "What is the distance to Mars?",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/smartroute/route", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp RouteResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Decision == nil {
		t.Fatal("expected decision to be non-nil")
	}
	if resp.Decision.TaskCategory != "simple" {
		t.Errorf("expected simple task category, got %s", resp.Decision.TaskCategory)
	}
	if resp.ActualResponse == "" {
		t.Error("expected non-empty actual response")
	}
}

func TestAPI_ChatCompletions(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	body := map[string]interface{}{
		"model": "auto",
		"messages": []map[string]string{
			{"role": "user", "content": "Write a python quicksort algorithm"},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var openAIResp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&openAIResp); err != nil {
		t.Fatalf("failed to decode OpenAI completion response: %v", err)
	}

	if openAIResp["model"] == nil || openAIResp["model"] == "" {
		t.Error("expected model to be populated")
	}
	if rec.Header().Get("X-SmartRoute-Category") != "coding" {
		t.Errorf("expected X-SmartRoute-Category 'coding', got '%s'", rec.Header().Get("X-SmartRoute-Category"))
	}
}

func TestAPI_DashboardMetrics(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/smartroute/metrics", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	var m storage.DashboardMetrics
	if err := json.NewDecoder(rec.Body).Decode(&m); err != nil {
		t.Fatalf("failed to decode metrics: %v", err)
	}

	if m.TotalRequests < 50 {
		t.Errorf("expected >=50 seeded requests, got %d", m.TotalRequests)
	}
	if m.EstimatedSavings <= 0 {
		t.Errorf("expected positive estimated savings, got %f", m.EstimatedSavings)
	}
}

func TestAPI_SimulateFailover(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	body := map[string]string{
		"provider": "anthropic",
		"action":   "fail",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/smartroute/failover/simulate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	// Verify failovers endpoint has the new failover event
	reqGet := httptest.NewRequest(http.MethodGet, "/api/smartroute/failovers", nil)
	recGet := httptest.NewRecorder()
	server.ServeHTTP(recGet, reqGet)

	var events []storage.FailoverEvent
	json.NewDecoder(recGet.Body).Decode(&events)
	if len(events) == 0 {
		t.Error("expected at least 1 failover event recorded")
	}
}
