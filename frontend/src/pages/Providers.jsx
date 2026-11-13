import React, { useState } from 'react';
import { 
  Server, 
  ShieldCheck, 
  AlertTriangle, 
  RefreshCw, 
  Zap, 
  Activity, 
  CheckCircle2, 
  XCircle, 
  ExternalLink 
} from 'lucide-react';
import { simulateFailover } from '../api';

export default function Providers({ providersData, failovers, onRefresh }) {
  const [simulating, setSimulating] = useState(false);
  const [actionMessage, setActionMessage] = useState(null);

  const handleSimulate = async (providerId, action) => {
    setSimulating(true);
    setActionMessage(null);
    try {
      await simulateFailover(providerId, action);
      setActionMessage({
        type: 'success',
        text: `Successfully ${action === 'fail' ? 'simulated failure on' : 'restored health for'} provider ${providerId.toUpperCase()}.`,
      });
      if (onRefresh) onRefresh();
    } catch (err) {
      setActionMessage({
        type: 'error',
        text: err.message || 'Simulation action failed',
      });
    } finally {
      setSimulating(false);
    }
  };

  const providers = providersData?.providers || [];
  const models = providersData?.models || [];
  const circuitStatuses = providersData?.circuit_statuses || {};

  return (
    <div className="space-y-6">
      {/* Overview Banner */}
      <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h2 className="text-lg font-bold text-white tracking-tight flex items-center space-x-2">
            <Server className="w-5 h-5 text-purple-400" />
            <span>AI Provider Fleet & Automatic Failover</span>
          </h2>
          <p className="text-xs text-slate-400 mt-1">
            Bifrost monitors provider SLAs and latency thresholds. If a provider fails or trips its circuit breaker, traffic is rerouted automatically.
          </p>
        </div>

        <div className="flex items-center space-x-3 text-xs">
          <span className="flex items-center space-x-1.5 text-emerald-400 bg-emerald-500/10 px-3 py-1.5 rounded-lg border border-emerald-500/20 font-semibold">
            <ShieldCheck className="w-4 h-4" />
            <span>Circuit Breaker Active</span>
          </span>
        </div>
      </div>

      {actionMessage && (
        <div className={`p-4 rounded-xl text-xs flex items-center justify-between border ${
          actionMessage.type === 'success' 
            ? 'bg-emerald-950/40 text-emerald-300 border-emerald-600/30' 
            : 'bg-rose-950/40 text-rose-300 border-rose-600/30'
        }`}>
          <span>{actionMessage.text}</span>
          <button onClick={() => setActionMessage(null)} className="text-slate-400 hover:text-white">✕</button>
        </div>
      )}

      {/* Provider Cards Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
        {providers.map((p) => {
          const isBreakerOpen = p.circuit_breaker_open;
          const status = circuitStatuses[p.id];
          const providerModels = models.filter((m) => m.provider_id === p.id);

          return (
            <div 
              key={p.id}
              className={`p-5 rounded-2xl border transition-all duration-200 flex flex-col justify-between ${
                isBreakerOpen
                  ? 'bg-rose-950/20 border-rose-600/40 shadow-lg shadow-rose-950/30'
                  : 'bg-navy-900 border-navy-800 hover:border-purple-600/30'
              }`}
            >
              <div>
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center space-x-2">
                    <span className="font-extrabold text-white text-base">{p.name}</span>
                  </div>
                  <span className={`w-3 h-3 rounded-full ${isBreakerOpen ? 'bg-rose-500 animate-pulse' : 'bg-emerald-400'}`} />
                </div>

                <div className="mb-4">
                  <span className={`text-[11px] font-bold px-2 py-0.5 rounded border uppercase tracking-wider ${
                    isBreakerOpen
                      ? 'bg-rose-500/20 text-rose-300 border-rose-500/40'
                      : 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30'
                  }`}>
                    {isBreakerOpen ? 'Circuit Breaker Open' : 'Healthy / Operating'}
                  </span>
                </div>

                <div className="space-y-2 text-xs text-slate-400 mb-4">
                  <div className="flex justify-between">
                    <span>Configured Models:</span>
                    <span className="text-white font-medium">{providerModels.length}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Consecutive Failures:</span>
                    <span className={p.consecutive_failures > 0 ? 'text-amber-400 font-bold' : 'text-slate-300'}>
                      {p.consecutive_failures}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span>Failover SLA:</span>
                    <span className="text-slate-300">3 retries • 3000ms</span>
                  </div>
                </div>
              </div>

              {/* Simulation Controls */}
              <div className="pt-3 border-t border-navy-800/80 flex items-center space-x-2">
                {isBreakerOpen ? (
                  <button
                    onClick={() => handleSimulate(p.id, 'recover')}
                    disabled={simulating}
                    className="w-full py-2 px-3 rounded-xl bg-emerald-600/20 hover:bg-emerald-600/30 text-emerald-300 border border-emerald-500/30 text-xs font-semibold transition-all disabled:opacity-50"
                  >
                    Restore Health
                  </button>
                ) : (
                  <button
                    onClick={() => handleSimulate(p.id, 'fail')}
                    disabled={simulating}
                    className="w-full py-2 px-3 rounded-xl bg-rose-600/20 hover:bg-rose-600/30 text-rose-300 border border-rose-500/30 text-xs font-semibold transition-all disabled:opacity-50"
                  >
                    Simulate Outage
                  </button>
                )}
              </div>
            </div>
          );
        })}
      </div>

      {/* Models Configuration Table */}
      <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl">
        <h3 className="text-sm font-bold text-white mb-4">Fleet Model Catalog & Pricing Rates</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="text-[11px] uppercase tracking-wider text-slate-400 border-b border-navy-800">
              <tr>
                <th className="pb-3 font-semibold">Model Name</th>
                <th className="pb-3 font-semibold">Provider</th>
                <th className="pb-3 font-semibold">Tier</th>
                <th className="pb-3 font-semibold">Input / 1M Tokens</th>
                <th className="pb-3 font-semibold">Output / 1M Tokens</th>
                <th className="pb-3 font-semibold">Avg Latency</th>
                <th className="pb-3 font-semibold">Context Window</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-navy-800/60 font-mono">
              {models.map((m) => (
                <tr key={m.id} className="hover:bg-navy-850/50 transition-colors">
                  <td className="py-3 font-bold text-white">{m.display_name}</td>
                  <td className="py-3 text-purple-300 uppercase">{m.provider_id}</td>
                  <td className="py-3">
                    <span className="px-2 py-0.5 rounded text-[10px] uppercase font-sans font-bold bg-navy-800 text-slate-300 border border-navy-700">
                      {m.task_tier}
                    </span>
                  </td>
                  <td className="py-3 text-emerald-400">${m.input_price_per_m.toFixed(3)}</td>
                  <td className="py-3 text-emerald-400">${m.output_price_per_m.toFixed(2)}</td>
                  <td className="py-3 text-amber-300 font-sans">{m.avg_latency_ms}ms</td>
                  <td className="py-3 text-slate-400">{(m.context_window / 1000).toLocaleString()}k</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Recent Failover Audit Trail */}
      <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-sm font-bold text-white flex items-center space-x-2">
            <AlertTriangle className="w-4 h-4 text-amber-400" />
            <span>Automatic Failover Audit Events</span>
          </h3>
          <span className="text-xs text-slate-400 font-mono">{failovers?.length || 0} events recorded</span>
        </div>

        {failovers && failovers.length > 0 ? (
          <div className="space-y-3">
            {failovers.slice(0, 5).map((f) => (
              <div 
                key={f.id} 
                className="p-3.5 rounded-xl bg-navy-950/70 border border-navy-800 flex flex-col md:flex-row md:items-center justify-between text-xs gap-3"
              >
                <div>
                  <div className="flex items-center space-x-2 mb-1">
                    <span className="px-2 py-0.5 rounded bg-rose-500/20 text-rose-300 font-bold border border-rose-500/30 uppercase text-[10px]">
                      Failed: {f.failed_provider}
                    </span>
                    <span className="text-slate-400">➔</span>
                    <span className="px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-300 font-bold border border-emerald-500/30 uppercase text-[10px]">
                      Recovered: {f.fallback_provider} ({f.fallback_model})
                    </span>
                  </div>
                  <p className="text-[11px] text-slate-400 font-mono">{f.error_reason}</p>
                </div>

                <div className="text-right text-[11px] text-slate-500">
                  <div className="text-slate-300 font-mono">Recovery overhead: {f.recovery_latency_ms}ms</div>
                  <span>{new Date(f.created_at).toLocaleTimeString()}</span>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="p-8 text-center text-slate-500 text-xs">
            No failover events recorded yet. Click "Simulate Outage" above to test the automated failover engine.
          </div>
        )}
      </div>
    </div>
  );
}
