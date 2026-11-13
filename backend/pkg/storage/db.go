package storage

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps sql.DB with concurrency safety
type DB struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewDB opens or creates the SQLite database and initializes schema
func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(time.Hour)

	store := &DB{db: conn}
	if err := store.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// Close closes the database connection
func (s *DB) Close() error {
	return s.db.Close()
}

func (s *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS providers (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		provider_type TEXT NOT NULL,
		is_active BOOLEAN NOT NULL DEFAULT 1,
		healthy BOOLEAN NOT NULL DEFAULT 1,
		consecutive_failures INTEGER NOT NULL DEFAULT 0,
		circuit_breaker_open BOOLEAN NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS models (
		id TEXT PRIMARY KEY,
		provider_id TEXT NOT NULL REFERENCES providers(id),
		model_name TEXT NOT NULL,
		display_name TEXT NOT NULL,
		task_tier TEXT NOT NULL,
		input_price_per_m REAL NOT NULL,
		output_price_per_m REAL NOT NULL,
		avg_latency_ms INTEGER NOT NULL DEFAULT 0,
		context_window INTEGER NOT NULL DEFAULT 128000,
		is_active BOOLEAN NOT NULL DEFAULT 1
	);

	CREATE TABLE IF NOT EXISTS routing_decisions (
		id TEXT PRIMARY KEY,
		request_id TEXT NOT NULL,
		prompt_snippet TEXT NOT NULL,
		prompt_tokens_est INTEGER NOT NULL,
		task_category TEXT NOT NULL,
		complexity_score REAL NOT NULL,
		selected_provider TEXT NOT NULL,
		selected_model TEXT NOT NULL,
		fallback_chain TEXT NOT NULL,
		routing_reason TEXT NOT NULL,
		routing_factors TEXT NOT NULL,
		manual_override BOOLEAN NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS request_logs (
		id TEXT PRIMARY KEY,
		request_id TEXT NOT NULL,
		provider TEXT NOT NULL,
		model TEXT NOT NULL,
		status TEXT NOT NULL,
		latency_ms INTEGER NOT NULL,
		prompt_tokens INTEGER NOT NULL,
		completion_tokens INTEGER NOT NULL,
		total_tokens INTEGER NOT NULL,
		actual_cost REAL NOT NULL,
		baseline_cost REAL NOT NULL,
		savings REAL NOT NULL,
		savings_pct REAL NOT NULL,
		is_simulated BOOLEAN NOT NULL DEFAULT 0,
		error_message TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS failover_events (
		id TEXT PRIMARY KEY,
		request_id TEXT NOT NULL,
		failed_provider TEXT NOT NULL,
		failed_model TEXT NOT NULL,
		fallback_provider TEXT NOT NULL,
		fallback_model TEXT NOT NULL,
		error_reason TEXT NOT NULL,
		recovery_latency_ms INTEGER NOT NULL,
		status TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS routing_rules (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		strategy TEXT NOT NULL,
		complexity_threshold REAL NOT NULL,
		baseline_model TEXT NOT NULL DEFAULT 'gpt-4o',
		cost_weight REAL NOT NULL DEFAULT 0.5,
		latency_weight REAL NOT NULL DEFAULT 0.3,
		reliability_weight REAL NOT NULL DEFAULT 0.2,
		is_active BOOLEAN NOT NULL DEFAULT 1,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_logs_created ON request_logs(created_at);
	CREATE INDEX IF NOT EXISTS idx_logs_provider ON request_logs(provider);
	CREATE INDEX IF NOT EXISTS idx_logs_model ON request_logs(model);
	CREATE INDEX IF NOT EXISTS idx_decisions_req ON routing_decisions(request_id);
	`
	_, err := s.db.Exec(schema)
	return err
}

// GetProviders returns all registered AI providers
func (s *DB) GetProviders() ([]Provider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT id, name, provider_type, is_active, healthy, consecutive_failures, circuit_breaker_open, created_at FROM providers ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []Provider
	for rows.Next() {
		var p Provider
		if err := rows.Scan(&p.ID, &p.Name, &p.ProviderType, &p.IsActive, &p.Healthy, &p.ConsecutiveFailures, &p.CircuitBreakerOpen, &p.CreatedAt); err != nil {
			return nil, err
		}
		providers = append(providers, p)
	}
	return providers, nil
}

// UpdateProviderHealth updates health and circuit breaker metrics
func (s *DB) UpdateProviderHealth(providerID string, healthy bool, failures int, breakerOpen bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`UPDATE providers SET healthy = ?, consecutive_failures = ?, circuit_breaker_open = ? WHERE id = ?`,
		healthy, failures, breakerOpen, providerID)
	return err
}

// GetModels returns all models
func (s *DB) GetModels() ([]Model, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT id, provider_id, model_name, display_name, task_tier, input_price_per_m, output_price_per_m, avg_latency_ms, context_window, is_active FROM models ORDER BY display_name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []Model
	for rows.Next() {
		var m Model
		if err := rows.Scan(&m.ID, &m.ProviderID, &m.ModelName, &m.DisplayName, &m.TaskTier, &m.InputPricePerM, &m.OutputPricePerM, &m.AvgLatencyMs, &m.ContextWindow, &m.IsActive); err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, nil
}

// GetActiveModels returns only enabled models
func (s *DB) GetActiveModels() ([]Model, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT m.id, m.provider_id, m.model_name, m.display_name, m.task_tier, m.input_price_per_m, m.output_price_per_m, m.avg_latency_ms, m.context_window, m.is_active FROM models m JOIN providers p ON m.provider_id = p.id WHERE m.is_active = 1 AND p.is_active = 1 AND p.circuit_breaker_open = 0 ORDER BY m.input_price_per_m ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []Model
	for rows.Next() {
		var m Model
		if err := rows.Scan(&m.ID, &m.ProviderID, &m.ModelName, &m.DisplayName, &m.TaskTier, &m.InputPricePerM, &m.OutputPricePerM, &m.AvgLatencyMs, &m.ContextWindow, &m.IsActive); err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, nil
}

// GetModelByName looks up a model by its unique name
func (s *DB) GetModelByName(modelName string) (*Model, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var m Model
	err := s.db.QueryRow(`SELECT id, provider_id, model_name, display_name, task_tier, input_price_per_m, output_price_per_m, avg_latency_ms, context_window, is_active FROM models WHERE model_name = ? LIMIT 1`, modelName).
		Scan(&m.ID, &m.ProviderID, &m.ModelName, &m.DisplayName, &m.TaskTier, &m.InputPricePerM, &m.OutputPricePerM, &m.AvgLatencyMs, &m.ContextWindow, &m.IsActive)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// SaveRoutingDecision writes a routing decision
func (s *DB) SaveRoutingDecision(d *RoutingDecision) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`INSERT INTO routing_decisions (id, request_id, prompt_snippet, prompt_tokens_est, task_category, complexity_score, selected_provider, selected_model, fallback_chain, routing_reason, routing_factors, manual_override, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.RequestID, d.PromptSnippet, d.PromptTokensEst, d.TaskCategory, d.ComplexityScore, d.SelectedProvider, d.SelectedModel, d.FallbackChain, d.RoutingReason, d.RoutingFactors, d.ManualOverride, d.CreatedAt)
	return err
}

// SaveRequestLog writes an execution telemetry log
func (s *DB) SaveRequestLog(l *RequestLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`INSERT INTO request_logs (id, request_id, provider, model, status, latency_ms, prompt_tokens, completion_tokens, total_tokens, actual_cost, baseline_cost, savings, savings_pct, is_simulated, error_message, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		l.ID, l.RequestID, l.Provider, l.Model, l.Status, l.LatencyMs, l.PromptTokens, l.CompletionTokens, l.TotalTokens, l.ActualCost, l.BaselineCost, l.Savings, l.SavingsPct, l.IsSimulated, l.ErrorMessage, l.CreatedAt)
	return err
}

// SaveFailoverEvent logs a failover occurrence
func (s *DB) SaveFailoverEvent(f *FailoverEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`INSERT INTO failover_events (id, request_id, failed_provider, failed_model, fallback_provider, fallback_model, error_reason, recovery_latency_ms, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		f.ID, f.RequestID, f.FailedProvider, f.FailedModel, f.FallbackProvider, f.FallbackModel, f.ErrorReason, f.RecoveryLatencyMs, f.Status, f.CreatedAt)
	return err
}

// GetRecentLogs retrieves request logs with limit and optional filtering
func (s *DB) GetRecentLogs(limit int, providerFilter, statusFilter string) ([]RequestLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, request_id, provider, model, status, latency_ms, prompt_tokens, completion_tokens, total_tokens, actual_cost, baseline_cost, savings, savings_pct, is_simulated, COALESCE(error_message, ''), created_at FROM request_logs WHERE 1=1`
	var args []interface{}

	if providerFilter != "" && providerFilter != "all" {
		query += ` AND provider = ?`
		args = append(args, providerFilter)
	}
	if statusFilter != "" && statusFilter != "all" {
		query += ` AND status = ?`
		args = append(args, statusFilter)
	}

	query += ` ORDER BY created_at DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []RequestLog
	for rows.Next() {
		var l RequestLog
		if err := rows.Scan(&l.ID, &l.RequestID, &l.Provider, &l.Model, &l.Status, &l.LatencyMs, &l.PromptTokens, &l.CompletionTokens, &l.TotalTokens, &l.ActualCost, &l.BaselineCost, &l.Savings, &l.SavingsPct, &l.IsSimulated, &l.ErrorMessage, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// GetRecentFailoverEvents returns recent failover records
func (s *DB) GetRecentFailoverEvents(limit int) ([]FailoverEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT id, request_id, failed_provider, failed_model, fallback_provider, fallback_model, error_reason, recovery_latency_ms, status, created_at FROM failover_events ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []FailoverEvent
	for rows.Next() {
		var f FailoverEvent
		if err := rows.Scan(&f.ID, &f.RequestID, &f.FailedProvider, &f.FailedModel, &f.FallbackProvider, &f.FallbackModel, &f.ErrorReason, &f.RecoveryLatencyMs, &f.Status, &f.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, f)
	}
	return events, nil
}

// GetDashboardMetrics aggregates summary stats
func (s *DB) GetDashboardMetrics() (*DashboardMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := &DashboardMetrics{}

	row := s.db.QueryRow(`
		SELECT 
			COUNT(*),
			COALESCE(SUM(actual_cost), 0.0),
			COALESCE(SUM(baseline_cost), 0.0),
			COALESCE(SUM(savings), 0.0),
			COALESCE(AVG(latency_ms), 0.0),
			COALESCE(SUM(CASE WHEN status != 'failed' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0)
		FROM request_logs
	`)
	if err := row.Scan(&metrics.TotalRequests, &metrics.TotalExpenditure, &metrics.BaselineCost, &metrics.EstimatedSavings, &metrics.AvgResponseTimeMs, &metrics.SuccessfulRequests, &metrics.FailedRequestsCount); err != nil {
		return nil, err
	}

	if metrics.TotalRequests > 0 {
		metrics.SuccessRate = float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests) * 100.0
	} else {
		metrics.SuccessRate = 100.0
	}

	if metrics.BaselineCost > 0 {
		metrics.SavingsPercent = (metrics.EstimatedSavings / metrics.BaselineCost) * 100.0
	}

	s.db.QueryRow(`SELECT COUNT(*) FROM providers WHERE is_active = 1`).Scan(&metrics.ActiveProviders)
	s.db.QueryRow(`SELECT COUNT(*) FROM models WHERE is_active = 1`).Scan(&metrics.ActiveModels)

	return metrics, nil
}

// GetProviderPerformances returns analytics grouped by provider
func (s *DB) GetProviderPerformances() ([]ProviderPerformance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT 
			provider,
			COUNT(*) as total_calls,
			COALESCE(SUM(CASE WHEN status != 'failed' THEN 1 ELSE 0 END), 0) as success_calls,
			COALESCE(AVG(latency_ms), 0.0) as avg_latency,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(actual_cost), 0.0) as total_cost
		FROM request_logs
		GROUP BY provider
		ORDER BY total_calls DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perfs []ProviderPerformance
	for rows.Next() {
		var p ProviderPerformance
		if err := rows.Scan(&p.Provider, &p.TotalCalls, &p.SuccessCalls, &p.AvgLatencyMs, &p.TotalTokens, &p.TotalCost); err != nil {
			return nil, err
		}
		if p.TotalCalls > 0 {
			p.SuccessRate = float64(p.SuccessCalls) / float64(p.TotalCalls) * 100.0
		}
		p.P95LatencyMs = p.AvgLatencyMs * 1.35 // estimated P95
		perfs = append(perfs, p)
	}
	return perfs, nil
}

// GetModelUsageSummaries returns usage volume and cost by model
func (s *DB) GetModelUsageSummaries() ([]ModelUsageSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT 
			rl.model,
			rl.provider,
			COALESCE(m.task_tier, 'balanced') as tier,
			COUNT(*) as request_count,
			COALESCE(SUM(rl.total_tokens), 0) as total_tokens,
			COALESCE(SUM(rl.actual_cost), 0.0) as total_cost,
			COALESCE(AVG(rl.latency_ms), 0.0) as avg_latency
		FROM request_logs rl
		LEFT JOIN models m ON rl.model = m.model_name
		GROUP BY rl.model, rl.provider
		ORDER BY request_count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []ModelUsageSummary
	for rows.Next() {
		var s ModelUsageSummary
		if err := rows.Scan(&s.Model, &s.Provider, &s.Tier, &s.RequestCount, &s.TotalTokens, &s.TotalCost, &s.AvgLatencyMs); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

// GetCostOverTime returns time series spend data
func (s *DB) GetCostOverTime() ([]CostOverTimePoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Group by hour or day depending on data distribution
	rows, err := s.db.Query(`
		SELECT 
			strftime('%Y-%m-%d %H:00', created_at) as bucket,
			COALESCE(SUM(actual_cost), 0.0) as actual_cost,
			COALESCE(SUM(baseline_cost), 0.0) as baseline_cost,
			COALESCE(SUM(savings), 0.0) as savings,
			COUNT(*) as req_count
		FROM request_logs
		GROUP BY bucket
		ORDER BY bucket ASC
		LIMIT 24
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []CostOverTimePoint
	for rows.Next() {
		var p CostOverTimePoint
		if err := rows.Scan(&p.Timestamp, &p.ActualCost, &p.BaselineCost, &p.Savings, &p.RequestCount); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

// GetActiveRule returns current active routing policy rule
func (s *DB) GetActiveRule() (*RoutingRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var r RoutingRule
	err := s.db.QueryRow(`SELECT id, name, strategy, complexity_threshold, baseline_model, cost_weight, latency_weight, reliability_weight, is_active, updated_at FROM routing_rules WHERE is_active = 1 LIMIT 1`).
		Scan(&r.ID, &r.Name, &r.Strategy, &r.ComplexityThreshold, &r.BaselineModel, &r.CostWeight, &r.LatencyWeight, &r.ReliabilityWeight, &r.IsActive, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// GetRules returns all routing rules
func (s *DB) GetRules() ([]RoutingRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT id, name, strategy, complexity_threshold, baseline_model, cost_weight, latency_weight, reliability_weight, is_active, updated_at FROM routing_rules ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []RoutingRule
	for rows.Next() {
		var r RoutingRule
		if err := rows.Scan(&r.ID, &r.Name, &r.Strategy, &r.ComplexityThreshold, &r.BaselineModel, &r.CostWeight, &r.LatencyWeight, &r.ReliabilityWeight, &r.IsActive, &r.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, nil
}

// UpdateRule updates routing policy parameters
func (s *DB) UpdateRule(r *RoutingRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`UPDATE routing_rules SET strategy = ?, complexity_threshold = ?, baseline_model = ?, cost_weight = ?, latency_weight = ?, reliability_weight = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		r.Strategy, r.ComplexityThreshold, r.BaselineModel, r.CostWeight, r.LatencyWeight, r.ReliabilityWeight, r.IsActive, r.ID)
	return err
}
