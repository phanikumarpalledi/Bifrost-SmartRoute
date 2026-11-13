package cost

import (
	"math"
	"testing"
)

func TestCalculator_CalculateCost(t *testing.T) {
	calc := NewCalculator("gpt-4o")

	// 1,000,000 prompt tokens + 1,000,000 completion tokens on GPT-4o
	// 2.50 + 10.00 = $12.50
	cost := calc.CalculateCost("gpt-4o", 1_000_000, 1_000_000)
	if math.Abs(cost-12.50) > 0.001 {
		t.Errorf("expected $12.50, got %f", cost)
	}

	// 10,000 prompt tokens + 20,000 completion tokens on GPT-4o-mini
	// (10,000 * 0.15 + 20,000 * 0.60) / 1M = (1.50 + 12.00) / 1000 = 0.0135
	costMini := calc.CalculateCost("gpt-4o-mini", 10_000, 20_000)
	if math.Abs(costMini-0.0135) > 0.0001 {
		t.Errorf("expected $0.0135, got %f", costMini)
	}
}

func TestCalculator_CostComparison(t *testing.T) {
	calc := NewCalculator("gpt-4o")

	// Routing to Gemini 1.5 Flash vs baseline GPT-4o
	cmp := calc.CompareCost("gemini-1.5-flash", "gpt-4o", 1000, 2000, false)

	if cmp.Savings <= 0 {
		t.Errorf("expected positive savings for gemini-1.5-flash vs gpt-4o, got %f", cmp.Savings)
	}
	if cmp.SavingsPercent < 90.0 {
		t.Errorf("expected >90%% savings for gemini-1.5-flash vs gpt-4o, got %f%%", cmp.SavingsPercent)
	}
	if cmp.PricingLabel != "Real Provider Pricing" {
		t.Errorf("expected Real Provider Pricing label, got %s", cmp.PricingLabel)
	}
}
