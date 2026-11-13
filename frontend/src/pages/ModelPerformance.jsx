import React from 'react';
import { 
  Zap, 
  Clock, 
  Activity, 
  BarChart3, 
  ShieldCheck, 
  TrendingUp 
} from 'lucide-react';
import { 
  BarChart, 
  Bar, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer, 
  Legend 
} from 'recharts';

export default function ModelPerformance({ providerPerformances, modelUsage }) {
  const latencyData = modelUsage?.map((m) => ({
    name: m.model,
    avgLatency: Math.round(m.avg_latency_ms),
    p95Latency: Math.round(m.avg_latency_ms * 1.35),
    cost: m.total_cost,
  })) || [];

  return (
    <div className="space-y-6">
      {/* Performance Overview Banner */}
      <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl">
        <h2 className="text-lg font-bold text-white tracking-tight flex items-center space-x-2">
          <Zap className="w-5 h-5 text-purple-400" />
          <span>Model Performance & Latency Telemetry</span>
        </h2>
        <p className="text-xs text-slate-400 mt-1">
          Bifrost SmartRoute continually profiles real-world TTFT (time-to-first-token), roundtrip latency, and token throughput to optimize routing decisions.
        </p>
      </div>

      {/* Latency Comparison Chart */}
      <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h3 className="text-base font-bold text-white">Model Response Latency (P50 vs Estimated P95)</h3>
            <p className="text-xs text-slate-400">Milliseconds per full completion roundtrip</p>
          </div>
          <span className="text-xs font-semibold px-2.5 py-1 rounded-full bg-purple-500/10 text-purple-400 border border-purple-500/20">
            Real-Time SLA Tracking
          </span>
        </div>

        <div className="h-72 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={latencyData} margin={{ top: 10, right: 20, left: 0, bottom: 20 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#1E2F56" vertical={false} opacity={0.4} />
              <XAxis dataKey="name" stroke="#64748B" fontSize={11} tickLine={false} angle={-15} textAnchor="end" />
              <YAxis stroke="#64748B" fontSize={11} tickLine={false} tickFormatter={(val) => `${val}ms`} />
              <Tooltip
                contentStyle={{ backgroundColor: '#0B1120', borderColor: '#1E2F56', borderRadius: '12px' }}
                labelStyle={{ color: '#E2E8F0', fontWeight: 'bold' }}
                formatter={(val, name) => [`${val} ms`, name === 'avgLatency' ? 'Avg Latency (P50)' : 'Peak Latency (P95)']}
              />
              <Legend wrapperStyle={{ paddingTop: '15px', fontSize: '12px' }} />
              <Bar dataKey="avgLatency" fill="#8B5CF6" radius={[4, 4, 0, 0]} name="Avg Latency (P50)" />
              <Bar dataKey="p95Latency" fill="#38BDF8" radius={[4, 4, 0, 0]} name="Peak Latency (P95)" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Provider SLA Performance Table */}
      <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl">
        <h3 className="text-sm font-bold text-white mb-4 flex items-center space-x-2">
          <Activity className="w-4 h-4 text-emerald-400" />
          <span>Provider SLA Reliability & Token Volume</span>
        </h3>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="text-[11px] uppercase tracking-wider text-slate-400 border-b border-navy-800">
              <tr>
                <th className="pb-3 font-semibold">Provider</th>
                <th className="pb-3 font-semibold">Total Requests</th>
                <th className="pb-3 font-semibold">Success Rate</th>
                <th className="pb-3 font-semibold">Avg Latency</th>
                <th className="pb-3 font-semibold">P95 Latency</th>
                <th className="pb-3 font-semibold">Total Tokens</th>
                <th className="pb-3 font-semibold">Total Cost</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-navy-800/60 font-mono">
              {providerPerformances?.map((p, idx) => (
                <tr key={idx} className="hover:bg-navy-850/50 transition-colors">
                  <td className="py-3 font-bold text-white uppercase flex items-center space-x-2">
                    <span className="w-2 h-2 rounded-full bg-purple-400"></span>
                    <span>{p.provider}</span>
                  </td>
                  <td className="py-3 text-slate-300">{p.total_calls.toLocaleString()}</td>
                  <td className="py-3">
                    <span className={`px-2 py-0.5 rounded text-[11px] font-bold ${
                      p.success_rate >= 98 
                        ? 'bg-emerald-500/10 text-emerald-300 border border-emerald-500/20' 
                        : 'bg-amber-500/10 text-amber-300 border border-amber-500/20'
                    }`}>
                      {p.success_rate.toFixed(1)}%
                    </span>
                  </td>
                  <td className="py-3 text-cyan-300 font-sans">{Math.round(p.avg_latency_ms)}ms</td>
                  <td className="py-3 text-slate-400 font-sans">{Math.round(p.p95_latency_ms)}ms</td>
                  <td className="py-3 text-slate-300">{p.total_tokens.toLocaleString()}</td>
                  <td className="py-3 text-emerald-400 font-bold">${p.total_cost.toFixed(4)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
