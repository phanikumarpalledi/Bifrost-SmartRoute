import React, { useState } from 'react';
import { 
  DollarSign, 
  TrendingDown, 
  Clock, 
  Activity, 
  Zap, 
  ShieldCheck, 
  ArrowRight, 
  Cpu, 
  CheckCircle2, 
  AlertCircle 
} from 'lucide-react';
import { 
  AreaChart, 
  Area, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer, 
  BarChart, 
  Bar, 
  Legend 
} from 'recharts';
import MetricCard from '../components/MetricCard';

export default function Overview({ 
  metrics, 
  costOverTime, 
  providers, 
  modelUsage, 
  onNavigate 
}) {
  const [quickPrompt, setQuickPrompt] = useState('');

  const handleQuickTest = (e) => {
    e.preventDefault();
    if (!quickPrompt.trim()) return;
    onNavigate('smart-router', { prompt: quickPrompt });
  };

  return (
    <div className="space-y-6">
      {/* KPI Cards Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
        <MetricCard
          title="Total API Requests"
          value={metrics?.total_requests?.toLocaleString() || '0'}
          subtext={`${metrics?.successful_requests || 0} successful`}
          trend={`${metrics?.success_rate?.toFixed(1) || 100}% reliability`}
          trendPositive={true}
          icon={Activity}
          color="purple"
        />

        <MetricCard
          title="Estimated Cost Savings"
          value={`$${metrics?.estimated_savings?.toFixed(4) || '0.0000'}`}
          subtext={`vs Baseline ($${metrics?.baseline_cost?.toFixed(4) || '0.00'})`}
          trend={`+${metrics?.savings_percent?.toFixed(1) || 0}% saved`}
          trendPositive={true}
          icon={TrendingDown}
          color="emerald"
        />

        <MetricCard
          title="Total API Expenditure"
          value={`$${metrics?.total_expenditure?.toFixed(4) || '0.0000'}`}
          subtext="Actual provider billing"
          trend="Optimized"
          trendPositive={true}
          icon={DollarSign}
          color="blue"
        />

        <MetricCard
          title="Avg Gateway Latency"
          value={`${metrics?.avg_response_time_ms ? Math.round(metrics.avg_response_time_ms) : 0}`}
          suffix="ms"
          subtext="Under 300ms SLA target"
          trend="High Throughput"
          trendPositive={true}
          icon={Clock}
          color="amber"
        />
      </div>

      {/* Main Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left 2 Cols: Cost Over Time Dual-Area Chart */}
        <div className="lg:col-span-2 p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-base font-bold text-white tracking-tight">Cumulative Cost & Savings Trend</h2>
              <p className="text-xs text-slate-400 mt-0.5">Real-time expenditure comparison against baseline model</p>
            </div>
            <span className="text-xs font-semibold px-2.5 py-1 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              Live Optimization
            </span>
          </div>

          <div className="h-72 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={costOverTime} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                <defs>
                  <linearGradient id="baselineColor" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#7C3AED" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="#7C3AED" stopOpacity={0}/>
                  </linearGradient>
                  <linearGradient id="actualColor" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#10B981" stopOpacity={0.4}/>
                    <stop offset="95%" stopColor="#10B981" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#1E2F56" vertical={false} opacity={0.4} />
                <XAxis 
                  dataKey="timestamp" 
                  stroke="#64748B" 
                  fontSize={11} 
                  tickLine={false} 
                  tickFormatter={(val) => val.split(' ')[1] || val}
                />
                <YAxis stroke="#64748B" fontSize={11} tickLine={false} tickFormatter={(val) => `$${val}`} />
                <Tooltip
                  contentStyle={{ backgroundColor: '#0B1120', borderColor: '#1E2F56', borderRadius: '12px' }}
                  labelStyle={{ color: '#E2E8F0', fontWeight: 'bold' }}
                  formatter={(val, name) => [`$${Number(val).toFixed(5)}`, name === 'baseline_cost' ? 'Baseline Cost (GPT-4o)' : 'Actual Cost (SmartRoute)']}
                />
                <Area 
                  type="monotone" 
                  dataKey="baseline_cost" 
                  stroke="#8B5CF6" 
                  strokeWidth={2} 
                  strokeDasharray="4 4"
                  fillOpacity={1} 
                  fill="url(#baselineColor)" 
                  name="baseline_cost"
                />
                <Area 
                  type="monotone" 
                  dataKey="actual_cost" 
                  stroke="#10B981" 
                  strokeWidth={2.5} 
                  fillOpacity={1} 
                  fill="url(#actualColor)" 
                  name="actual_cost"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
          <div className="flex items-center justify-center space-x-6 mt-3 text-xs text-slate-400">
            <div className="flex items-center space-x-2">
              <span className="w-3 h-0.5 bg-purple-400 border border-purple-400 border-dashed"></span>
              <span>Baseline Model Expenditure</span>
            </div>
            <div className="flex items-center space-x-2">
              <span className="w-3 h-2 rounded bg-emerald-400"></span>
              <span className="text-white font-medium">SmartRoute Actual Cost</span>
            </div>
          </div>
        </div>

        {/* Right 1 Col: Quick Test Console */}
        <div className="p-6 rounded-2xl bg-gradient-to-b from-navy-900 to-navy-950 border border-purple-600/20 shadow-xl flex flex-col justify-between">
          <div>
            <div className="flex items-center space-x-2 mb-2">
              <div className="p-1.5 rounded-lg bg-purple-600/20 text-purple-400">
                <Cpu className="w-4 h-4" />
              </div>
              <h2 className="text-base font-bold text-white tracking-tight">Test Smart Router</h2>
            </div>
            <p className="text-xs text-slate-400 leading-relaxed mb-4">
              Enter any prompt to see real-time complexity analysis, cost projection, and intelligent model routing.
            </p>

            <form onSubmit={handleQuickTest} className="space-y-3">
              <textarea
                value={quickPrompt}
                onChange={(e) => setQuickPrompt(e.target.value)}
                placeholder="e.g. Write a Go goroutine worker pool with context timeout..."
                rows={4}
                className="w-full text-xs p-3 rounded-xl bg-navy-950 border border-navy-700/80 text-slate-100 placeholder-slate-500 focus:outline-none focus:border-purple-500 transition-colors resize-none font-mono"
              />

              <div className="flex flex-wrap gap-1.5">
                {[
                  { label: 'Simple Q&A', text: 'What is the freezing point of water in Fahrenheit?' },
                  { label: 'Coding Query', text: 'Write a Go worker pool with select and cancellation context.' },
                  { label: 'Deep Reasoning', text: 'Analyze the architectural trade-offs between Paxos and Raft.' },
                ].map((sample, idx) => (
                  <button
                    key={idx}
                    type="button"
                    onClick={() => setQuickPrompt(sample.text)}
                    className="text-[11px] px-2 py-1 rounded bg-navy-800 hover:bg-navy-700/80 text-purple-300 border border-navy-700 transition-colors"
                  >
                    {sample.label}
                  </button>
                ))}
              </div>

              <button
                type="submit"
                disabled={!quickPrompt.trim()}
                className="w-full mt-3 flex items-center justify-center space-x-2 py-2.5 px-4 rounded-xl font-semibold text-xs text-white bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 shadow-md shadow-purple-900/40 transition-all disabled:opacity-50"
              >
                <span>Run Intelligent Route</span>
                <ArrowRight className="w-3.5 h-3.5" />
              </button>
            </form>
          </div>

          <div className="mt-4 pt-4 border-t border-navy-800/80 flex items-center justify-between text-xs text-slate-400">
            <span>Dynamic Failover SLA:</span>
            <span className="font-semibold text-emerald-400">Enabled (3s timeout)</span>
          </div>
        </div>
      </div>

      {/* Model Usage Breakdown & Provider Status Table */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Model Distribution */}
        <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-base font-bold text-white tracking-tight">Model Request Volume</h2>
            <span className="text-xs text-slate-400 font-medium">By Request Count</span>
          </div>

          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={modelUsage?.slice(0, 6)} layout="vertical" margin={{ top: 5, right: 30, left: 40, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#1E2F56" horizontal={false} opacity={0.4} />
                <XAxis type="number" stroke="#64748B" fontSize={11} tickLine={false} />
                <YAxis dataKey="model" type="category" stroke="#94A3B8" fontSize={11} tickLine={false} />
                <Tooltip
                  contentStyle={{ backgroundColor: '#0B1120', borderColor: '#1E2F56', borderRadius: '12px' }}
                  labelStyle={{ color: '#E2E8F0', fontWeight: 'bold' }}
                />
                <Bar dataKey="request_count" fill="#8B5CF6" radius={[0, 6, 6, 0]} name="Requests" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Live Provider Health Table */}
        <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-base font-bold text-white tracking-tight">Active AI Provider Matrix</h2>
              <button 
                onClick={() => onNavigate('providers')}
                className="text-xs text-purple-400 hover:text-purple-300 font-medium flex items-center space-x-1"
              >
                <span>Manage Failover</span>
                <ArrowRight className="w-3 h-3" />
              </button>
            </div>

            <div className="space-y-3">
              {providers?.providers?.map((p) => {
                const isTripped = p.circuit_breaker_open;
                return (
                  <div 
                    key={p.id}
                    className="p-3.5 rounded-xl bg-navy-850/70 border border-navy-800 flex items-center justify-between"
                  >
                    <div className="flex items-center space-x-3">
                      <div className={`w-3 h-3 rounded-full ${isTripped ? 'bg-rose-500 animate-pulse' : 'bg-emerald-400'}`} />
                      <div>
                        <h4 className="text-xs font-bold text-white uppercase tracking-wider">{p.name}</h4>
                        <span className="text-[11px] text-slate-400">Type: {p.provider_type}</span>
                      </div>
                    </div>

                    <div className="flex items-center space-x-4 text-xs">
                      <span className={`px-2 py-0.5 rounded text-[11px] font-semibold border ${
                        isTripped 
                          ? 'bg-rose-950/60 text-rose-300 border-rose-600/40' 
                          : 'bg-emerald-950/50 text-emerald-300 border-emerald-600/30'
                      }`}>
                        {isTripped ? 'Circuit Breaker Open' : 'Healthy (Operational)'}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          <div className="mt-4 pt-3 border-t border-navy-800 text-[11px] text-slate-500 flex justify-between">
            <span>Bifrost PreLLMHook Health Observation</span>
            <span className="text-slate-400">Auto-recovery: 30s timeout</span>
          </div>
        </div>
      </div>
    </div>
  );
}
