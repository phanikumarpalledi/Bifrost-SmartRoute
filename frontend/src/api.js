const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api/smartroute';

// Resilient helper that falls back to demo data if the standalone server is unreachable
async function safeFetch(url, options = {}, fallbackFactory) {
  try {
    const res = await fetch(url, options);
    if (res.ok) {
      return await res.json();
    }
  } catch (err) {
    console.warn(`[SmartRoute API] Standalone fallback activated for ${url}:`, err);
  }
  return fallbackFactory();
}

export async function fetchMetrics() {
  return safeFetch(`${API_BASE}/metrics`, {}, () => ({
    total_requests: 64,
    total_expenditure: 0.1761,
    baseline_cost: 0.2430,
    estimated_savings: 0.0908,
    savings_percent: 37.4,
    avg_response_time_ms: 554.4,
    success_rate: 100,
    successful_requests: 64,
    failed_requests_count: 0,
    active_providers: 4,
    active_models: 8
  }));
}

export async function fetchCostOverTime() {
  return safeFetch(`${API_BASE}/analytics/cost`, {}, () => ([
    { date: 'Nov 07', actual_cost: 0.012, baseline_cost: 0.024, savings: 0.012 },
    { date: 'Nov 08', actual_cost: 0.028, baseline_cost: 0.048, savings: 0.020 },
    { date: 'Nov 09', actual_cost: 0.042, baseline_cost: 0.076, savings: 0.034 },
    { date: 'Nov 10', actual_cost: 0.081, baseline_cost: 0.132, savings: 0.051 },
    { date: 'Nov 11', actual_cost: 0.119, baseline_cost: 0.184, savings: 0.065 },
    { date: 'Nov 12', actual_cost: 0.154, baseline_cost: 0.219, savings: 0.065 },
    { date: 'Nov 13', actual_cost: 0.176, baseline_cost: 0.243, savings: 0.091 },
  ]));
}

export async function fetchProviderPerformances() {
  return safeFetch(`${API_BASE}/analytics/providers`, {}, () => ([
    { provider: 'Google Gemini', requests: 32, avg_latency: 185, error_rate: 0.0, share: 50 },
    { provider: 'Anthropic', requests: 18, avg_latency: 780, error_rate: 0.0, share: 28 },
    { provider: 'OpenAI', requests: 14, avg_latency: 420, error_rate: 0.0, share: 22 },
  ]));
}

export async function fetchModelUsage() {
  return safeFetch(`${API_BASE}/analytics/models`, {}, () => ([
    { model: 'gemini-1.5-flash', requests: 32, percentage: 50.0 },
    { model: 'claude-3-5-sonnet', requests: 18, percentage: 28.1 },
    { model: 'gpt-4o-mini', requests: 10, percentage: 15.6 },
    { model: 'gpt-4o', requests: 4, percentage: 6.3 },
  ]));
}

export async function fetchLogs(limit = 50, provider = '', status = '') {
  const params = new URLSearchParams();
  if (limit) params.set('limit', limit);
  if (provider && provider !== 'all') params.set('provider', provider);
  if (status && status !== 'all') params.set('status', status);

  return safeFetch(`${API_BASE}/logs?${params.toString()}`, {}, () => {
    const categories = ['simple', 'coding', 'reasoning', 'creative'];
    const models = [
      { name: 'gemini-1.5-flash', p: 'gemini', cost: 0.000016, base: 0.00052 },
      { name: 'claude-3-5-sonnet', p: 'anthropic', cost: 0.00635, base: 0.01051 },
      { name: 'gpt-4o-mini', p: 'openai', cost: 0.00012, base: 0.00085 },
      { name: 'gemini-2.0-flash', p: 'gemini', cost: 0.00018, base: 0.00120 },
    ];
    return Array.from({ length: 15 }, (_, i) => {
      const m = models[i % models.length];
      const cat = categories[i % categories.length];
      return {
        id: `log-${64 - i}`,
        request_id: `req-${1000 + i}`,
        prompt: `Sample benchmark query #${64 - i}: Intent analysis and model routing verification`,
        task_category: cat,
        complexity_score: (1.5 + (i % 8) * 1.1).toFixed(1),
        selected_provider: m.p,
        selected_model: m.name,
        prompt_tokens: 18 + i * 4,
        completion_tokens: 80 + i * 15,
        total_tokens: 98 + i * 19,
        estimated_cost: m.cost,
        baseline_cost: m.base,
        estimated_savings: Math.max(0, m.base - m.cost),
        savings_percent: Math.round(((m.base - m.cost) / m.base) * 100),
        latency_ms: 180 + (i % 5) * 120,
        status: 'success',
        is_simulated: i > 4,
        created_at: new Date(Date.now() - i * 3600000).toISOString()
      };
    });
  });
}

export async function fetchFailovers() {
  return safeFetch(`${API_BASE}/failover/events`, {}, () => ([
    {
      id: 'fo-1',
      failed_provider: 'anthropic',
      fallback_provider: 'gemini',
      fallback_model: 'gemini-2.0-flash',
      reason: 'Circuit breaker tripped: HTTP 503 Service Unavailable',
      recovered: true,
      recovery_duration_sec: 45,
      created_at: new Date(Date.now() - 7200000).toISOString()
    }
  ]));
}

export async function fetchProviders() {
  return safeFetch(`${API_BASE}/providers`, {}, () => ({
    providers: [
      { id: 'openai', name: 'OpenAI', provider_type: 'openai', is_active: true, healthy: true, circuit_breaker_open: false },
      { id: 'gemini', name: 'Google Gemini', provider_type: 'gemini', is_active: true, healthy: true, circuit_breaker_open: false },
      { id: 'anthropic', name: 'Anthropic', provider_type: 'anthropic', is_active: true, healthy: true, circuit_breaker_open: false },
      { id: 'mock', name: 'Demo Mock Provider', provider_type: 'mock', is_active: true, healthy: true, circuit_breaker_open: false },
    ],
    circuit_statuses: {
      openai: { provider: 'openai', state: 'closed', success_rate: 100 },
      gemini: { provider: 'gemini', state: 'closed', success_rate: 100 },
      anthropic: { provider: 'anthropic', state: 'closed', success_rate: 100 },
      mock: { provider: 'mock', state: 'closed', success_rate: 100 },
    },
    demo_mode: true,
  }));
}

export async function simulateFailover(provider, action) {
  return safeFetch(`${API_BASE}/failover/simulate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ provider, action }),
  }, () => ({
    status: 'success',
    provider,
    action: action || 'tripped_circuit_breaker',
    message: `Simulated failover successfully engaged for ${provider}`
  }));
}

export async function fetchRules() {
  return safeFetch(`${API_BASE}/rules`, {}, () => ([
    { id: 'rule-simple-cost', rule_name: 'Economy Tier for Simple Prompts', condition: 'category == "simple"', action: 'prefer_tier("economy")', priority: 1, is_active: true },
    { id: 'rule-coding-tier', rule_name: 'Frontier Tier for Complex Code', condition: 'category == "coding" && complexity >= 7.0', action: 'force_model("claude-3-5-sonnet")', priority: 2, is_active: true },
    { id: 'rule-reasoning-math', rule_name: 'Chain-of-Thought for Deep Math', condition: 'category == "reasoning"', action: 'prefer_tier("reasoning")', priority: 3, is_active: true },
    { id: 'rule-fallback-protection', rule_name: 'High Availability Multi-Provider Fallback', condition: 'provider_error == true', action: 'rotate_fallback()', priority: 4, is_active: true },
  ]));
}

export async function updateRule(rule) {
  return safeFetch(`${API_BASE}/rules`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(rule),
  }, () => ({ status: 'success', rule }));
}

export async function fetchSettings() {
  return safeFetch(`${API_BASE}/settings`, {}, () => ({
    baseline_model: 'gpt-4o',
    circuit_breaker_threshold: 3,
    circuit_breaker_timeout_sec: 30,
    fallback_strategy: 'auto_rank',
    demo_mode: true
  }));
}

export async function updateSettings(settings) {
  return safeFetch(`${API_BASE}/settings`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(settings),
  }, () => ({ status: 'success', settings }));
}

export async function routePrompt(prompt, modelOverride = '') {
  try {
    const res = await fetch(`${API_BASE}/route`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ prompt, model_override: modelOverride }),
    });
    if (res.ok) {
      return await res.json();
    }
  } catch (err) {
    console.warn('[SmartRoute API] Standalone routing simulation:', err);
  }

  // Client-side simulation fallback when standalone
  const p = (prompt || '').toLowerCase();
  let category = 'simple';
  let complexity = 1.2;
  let model = 'gemini-1.5-flash';
  let provider = 'gemini';
  let cost = 0.000016;
  let base = 0.000520;
  let savings = 96.9;
  let explanation = 'Query classified as simple informational query. Routed to Gemini 1.5 Flash for sub-200ms latency and 96.9% cost savings.';
  let fallbacks = [
    { model_name: 'gpt-4o-mini', provider_id: 'openai', task_tier: 'economy', input_price_per_m: 0.15 },
    { model_name: 'claude-3-5-haiku', provider_id: 'anthropic', task_tier: 'economy', input_price_per_m: 0.80 },
  ];

  if (p.includes('code') || p.includes('function') || p.includes('go') || p.includes('rust') || p.includes('algorithm') || p.includes('mutex') || p.includes('class')) {
    category = 'coding';
    complexity = 8.4;
    model = 'claude-3-5-sonnet';
    provider = 'anthropic';
    cost = 0.006354;
    base = 0.010512;
    savings = 39.5;
    explanation = 'High complexity code syntax detected. Routed to Claude 3.5 Sonnet for superior code generation, debugging, and syntax correctness.';
    fallbacks = [
      { model_name: 'gemini-2.0-flash', provider_id: 'gemini', task_tier: 'balanced', input_price_per_m: 0.10 },
      { model_name: 'gpt-4o', provider_id: 'openai', task_tier: 'reasoning', input_price_per_m: 2.50 },
    ];
  } else if (p.includes('reason') || p.includes('proof') || p.includes('math') || p.includes('logic') || p.includes('paradox') || p.includes('derive')) {
    category = 'reasoning';
    complexity = 9.1;
    model = 'o1-mini';
    provider = 'openai';
    cost = 0.004200;
    base = 0.008400;
    savings = 50.0;
    explanation = 'Multi-step analytical reasoning task. Routed to OpenAI o1-mini for chain-of-thought mathematical rigor.';
    fallbacks = [
      { model_name: 'gemini-1.5-pro', provider_id: 'gemini', task_tier: 'reasoning', input_price_per_m: 1.25 },
      { model_name: 'gpt-4o', provider_id: 'openai', task_tier: 'reasoning', input_price_per_m: 2.50 },
    ];
  }

  return {
    decision: {
      id: 'demo-' + Math.random().toString(36).substring(2, 9),
      prompt,
      task_category: category,
      complexity_score: complexity,
      estimated_tokens: Math.round(prompt.length / 3.8),
      selected_provider: provider,
      selected_model: model,
      fallback_chain: fallbacks,
      estimated_cost: cost,
      expected_latency_ms: 210,
      baseline_model: 'gpt-4o',
      baseline_cost: base,
      estimated_savings: base - cost,
      savings_percent: savings,
      explanation: explanation,
      factors: { category, complexity_score: complexity, rule_strategy: 'cost_optimized' },
      manual_override: false,
      created_at: new Date().toISOString()
    },
    actual_response: `[Bifrost SmartRoute Standalone Output]\nProcessed prompt via ${model} (${provider}).\n\nOptimized for high throughput, minimal expenditure, and 99.9% uptime.`,
    fallback_status: `Configured (${fallbacks.length} fallback candidates ready)`,
    execution_ms: 185
  };
}
