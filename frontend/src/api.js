const API_BASE = '/api/smartroute';

export async function fetchMetrics() {
  const res = await fetch(`${API_BASE}/metrics`);
  if (!res.ok) throw new Error('Failed to fetch dashboard metrics');
  return res.json();
}

export async function fetchCostOverTime() {
  const res = await fetch(`${API_BASE}/analytics/cost`);
  if (!res.ok) throw new Error('Failed to fetch cost trends');
  return res.json();
}

export async function fetchProviderPerformances() {
  const res = await fetch(`${API_BASE}/analytics/providers`);
  if (!res.ok) throw new Error('Failed to fetch provider performance');
  return res.json();
}

export async function fetchModelUsage() {
  const res = await fetch(`${API_BASE}/analytics/models`);
  if (!res.ok) throw new Error('Failed to fetch model usage summaries');
  return res.json();
}

export async function fetchLogs(limit = 50, provider = '', status = '') {
  const params = new URLSearchParams();
  if (limit) params.set('limit', limit);
  if (provider && provider !== 'all') params.set('provider', provider);
  if (status && status !== 'all') params.set('status', status);

  const res = await fetch(`${API_BASE}/logs?${params.toString()}`);
  if (!res.ok) throw new Error('Failed to fetch request logs');
  return res.json();
}

export async function fetchFailovers() {
  const res = await fetch(`${API_BASE}/failovers`);
  if (!res.ok) throw new Error('Failed to fetch failover events');
  return res.json();
}

export async function fetchProviders() {
  const res = await fetch(`${API_BASE}/providers`);
  if (!res.ok) throw new Error('Failed to fetch providers');
  return res.json();
}

export async function simulateFailover(provider, action) {
  const res = await fetch(`${API_BASE}/failover/simulate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ provider, action }),
  });
  if (!res.ok) throw new Error('Failed to simulate failover');
  return res.json();
}

export async function fetchRules() {
  const res = await fetch(`${API_BASE}/rules`);
  if (!res.ok) throw new Error('Failed to fetch routing rules');
  return res.json();
}

export async function updateRule(rule) {
  const res = await fetch(`${API_BASE}/rules`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(rule),
  });
  if (!res.ok) throw new Error('Failed to update routing rule');
  return res.json();
}

export async function fetchSettings() {
  const res = await fetch(`${API_BASE}/settings`);
  if (!res.ok) throw new Error('Failed to fetch settings');
  return res.json();
}

export async function updateSettings(settings) {
  const res = await fetch(`${API_BASE}/settings`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(settings),
  });
  if (!res.ok) throw new Error('Failed to update settings');
  return res.json();
}

export async function routePrompt(prompt, modelOverride = '') {
  const res = await fetch(`${API_BASE}/route`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ prompt, model_override: modelOverride }),
  });
  if (!res.ok) {
    const errText = await res.text();
    throw new Error(errText || 'Failed to route prompt');
  }
  return res.json();
}
