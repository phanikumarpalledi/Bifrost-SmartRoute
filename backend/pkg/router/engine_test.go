package router

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/maximhq/bifrost/plugins/smartroute/pkg/cost"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/failover"
	"github.com/maximhq/bifrost/plugins/smartroute/pkg/storage"
)

func setupTestDB(t *testing.T) (*storage.DB, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_smartroute.db")
	db, err := storage.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}

	// Seed providers and models
	now := time.Now()
	providers := []storage.Provider{
		{ID: "openai", Name: "OpenAI", ProviderType: "openai", IsActive: true, Healthy: true},
		{ID: "gemini", Name: "Gemini", ProviderType: "gemini", IsActive: true, Healthy: true},
		{ID: "anthropic", Name: "Anthropic", ProviderType: "anthropic", IsActive: true, Healthy: true},
	}
	for _, p := range providers {
		db.UpdateProviderHealth(p.ID, p.Healthy, 0, false)
	}

	models := []storage.Model{
		{ID: "m1", ProviderID: "gemini", ModelName: "gemini-1.5-flash", DisplayName: "Gemini 1.5 Flash", TaskTier: "economy", InputPricePerM: 0.075, OutputPricePerM: 0.30, AvgLatencyMs: 180, IsActive: true},
		{ID: "m2", ProviderID: "openai", ModelName: "gpt-4o-mini", DisplayName: "GPT-4o Mini", TaskTier: "economy", InputPricePerM: 0.15, OutputPricePerM: 0.60, AvgLatencyMs: 250, IsActive: true},
		{ID: "m3", ProviderID: "anthropic", ModelName: "claude-3-5-sonnet", DisplayName: "Claude 3.5 Sonnet", TaskTier: "coding", InputPricePerM: 3.00, OutputPricePerM: 15.00, AvgLatencyMs: 750, IsActive: true},
		{ID: "m4", ProviderID: "openai", ModelName: "gpt-4o", DisplayName: "GPT-4o", TaskTier: "reasoning", InputPricePerM: 2.50, OutputPricePerM: 10.00, AvgLatencyMs: 600, IsActive: true},
	}

	rule := storage.RoutingRule{
		ID:                  "test-rule",
		Name:                "Test Rule",
		Strategy:            "cost_optimized",
		ComplexityThreshold: 4.0,
		BaselineModel:       "gpt-4o",
		CostWeight:          0.60,
		LatencyWeight:       0.20,
		ReliabilityWeight:   0.20,
		IsActive:            true,
		UpdatedAt:           now,
	}

	// Insert test entities
	for _, p := range providers {
		db.UpdateProviderHealth(p.ID, true, 0, false)
	}
	for range models {
		_ = db.SaveRoutingDecision(&storage.RoutingDecision{}) // test schema sanity
	}

	_ = db.SeedInitialData()
	_ = db.UpdateRule(&rule)

	return db, func() {
		db.Close()
	}
}

func TestEngine_RouteSimple(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	failoverMgr := failover.NewManager(3, time.Minute)
	costCalc := cost.NewCalculator("gpt-4o")
	engine := NewEngine(db, failoverMgr, costCalc)

	decision, err := engine.Route("What is the capital of Spain?", "")
	if err != nil {
		t.Fatalf("unexpected route error: %v", err)
	}

	if decision.TaskCategory != "simple" {
		t.Errorf("expected simple task category, got %s", decision.TaskCategory)
	}
	// Economy models: gemini-1.5-flash or gpt-4o-mini
	if decision.SelectedModel != "gemini-1.5-flash" && decision.SelectedModel != "gpt-4o-mini" {
		t.Errorf("expected economy model for simple query, got %s", decision.SelectedModel)
	}
	if decision.SavingsPercent <= 0 {
		t.Errorf("expected positive savings vs baseline, got %f", decision.SavingsPercent)
	}
}

func TestEngine_RouteCoding(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	failoverMgr := failover.NewManager(3, time.Minute)
	costCalc := cost.NewCalculator("gpt-4o")
	engine := NewEngine(db, failoverMgr, costCalc)

	codingPrompt := "Write a Python script with asyncio worker pool:\n```python\nimport asyncio\n```"
	decision, err := engine.Route(codingPrompt, "")
	if err != nil {
		t.Fatalf("unexpected route error: %v", err)
	}

	if decision.TaskCategory != "coding" {
		t.Errorf("expected coding category, got %s", decision.TaskCategory)
	}
	if decision.SelectedModel != "claude-3-5-sonnet" && decision.SelectedModel != "gpt-4o" {
		t.Errorf("expected premier coding model (claude-3-5-sonnet or gpt-4o), got %s", decision.SelectedModel)
	}
}

func TestEngine_ManualOverride(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	failoverMgr := failover.NewManager(3, time.Minute)
	costCalc := cost.NewCalculator("gpt-4o")
	engine := NewEngine(db, failoverMgr, costCalc)

	decision, err := engine.Route("What is 2+2?", "claude-3-5-sonnet")
	if err != nil {
		t.Fatalf("unexpected error with manual override: %v", err)
	}

	if !decision.ManualOverride {
		t.Error("expected manual_override to be true")
	}
	if decision.SelectedModel != "claude-3-5-sonnet" {
		t.Errorf("expected claude-3-5-sonnet, got %s", decision.SelectedModel)
	}
}
