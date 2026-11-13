package failover

import (
	"testing"
	"time"
)

func TestFailover_CircuitBreakerTrips(t *testing.T) {
	mgr := NewManager(3, 100*time.Millisecond)

	if !mgr.IsProviderHealthy("openai") {
		t.Error("expected initial provider to be healthy")
	}

	// 1st failure
	tripped := mgr.RecordFailure("openai", "HTTP 500")
	if tripped || !mgr.IsProviderHealthy("openai") {
		t.Error("breaker should not trip on 1 failure")
	}

	// 2nd failure
	tripped = mgr.RecordFailure("openai", "HTTP 502")
	if tripped || !mgr.IsProviderHealthy("openai") {
		t.Error("breaker should not trip on 2 failures")
	}

	// 3rd failure - should trip!
	tripped = mgr.RecordFailure("openai", "HTTP 504")
	if !tripped {
		t.Error("breaker should trip on 3rd failure")
	}
	if mgr.IsProviderHealthy("openai") {
		t.Error("provider should be reported unhealthy when breaker is open")
	}
}

func TestFailover_SelectFallback(t *testing.T) {
	mgr := NewManager(2, 500*time.Millisecond)

	// Trip OpenAI
	mgr.RecordFailure("openai", "HTTP 500")
	mgr.RecordFailure("openai", "HTTP 500")

	candidates := []CandidateOption{
		{Provider: "openai", Model: "gpt-4o", CostPerM: 2.50},
		{Provider: "openai", Model: "gpt-4o-mini", CostPerM: 0.15},
		{Provider: "gemini", Model: "gemini-1.5-flash", CostPerM: 0.075},
		{Provider: "anthropic", Model: "claude-3-5-haiku", CostPerM: 0.80},
	}

	selected, err := mgr.SelectFallback("openai", "gpt-4o", candidates)
	if err != nil {
		t.Fatalf("unexpected error selecting fallback: %v", err)
	}

	// Should skip openai because circuit breaker is open, and pick gemini!
	if selected.Provider != "gemini" {
		t.Errorf("expected fallback provider gemini, got %s", selected.Provider)
	}
	if selected.Model != "gemini-1.5-flash" {
		t.Errorf("expected fallback model gemini-1.5-flash, got %s", selected.Model)
	}
}
