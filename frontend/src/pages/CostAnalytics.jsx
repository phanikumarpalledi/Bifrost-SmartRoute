import React, { useState } from 'react';
import { 
  TrendingDown, 
  DollarSign, 
  PiggyBank, 
  Layers, 
  Sparkles, 
  Sliders 
} from 'lucide-react';
import { 
  BarChart, 
  Bar, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer, 
  PieChart, 
  Pie, 
  Cell, 
  Legend 
} from 'recharts';
import MetricCard from '../components/MetricCard';

export default function CostAnalytics({ 
  metrics, 
  modelUsage, 
  baselineModel = 'gpt-4o', 
  onUpdateBaseline 
}) {
  const [selectedBaseline, setSelectedBaseline] = useState(baselineModel);

  const handleBaselineChange = (e) => {
    const newBase = e.target.value;
    setSelectedBaseline(newBase);
    if (onUpdateBaseline) {
      onUpdateBaseline(newBase);
    }
  };

  const COLORS = ['#8B5CF6', '#10B981', '#38BDF8', '#F59E0B', '#EC4899', '#6366F1'];

  const pieData = modelUsage?.map((m) => ({
    name: m.model,
    value: Number(m.total_cost.toFixed(4)),
  })) || [];

  return (
    <div className="space-y-6">
      {/* Cost Analytics Banner & Baseline Selector */}
      <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h2 className="text-lg font-bold text-white tracking-tight flex items-center space-x-2">
            <TrendingDown className="w-5 h-5 text-emerald-400" />
            <span>Cost Optimization Engine & Expenditure Analytics</span>
          </h2>
          <p className="text-xs text-slate-400 mt-1">
            Track token consumption, compare real provider prices, and quantify exact savings achieved by Bifrost SmartRoute.
          </p>
        </div>

        {/* Baseline Selector */}
        <div className="flex items-center space-x-3 bg-navy-950 p-2 rounded-xl border border-navy-700/80">
          <span className="text-xs text-slate-400 font-medium pl-2">Baseline Model:</span>
          <select
            value={selectedBaseline}
            onChange={handleBaselineChange}
            className="bg-navy-850 border border-navy-700 text-xs text-purple-200 rounded-lg px-3 py-1.5 font-mono font-semibold focus:outline-none focus:border-purple-500"
          >
            <option value="gpt-4o">gpt-4o ($2.50 / $10.00)</option>
            <option value="claude-3-5-sonnet">claude-3-5-sonnet ($3.00 / $15.00)</option>
            <option value="o1-mini">o1-mini ($3.00 / $12.00)</option>
          </select>
        </div>
      </div>

      {/* KPI Highlights */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
        <MetricCard
          title="Total Actual Cost"
          value={`$${metrics?.total_expenditure?.toFixed(4) || '0.0000'}`}
          subtext="Actual provider charges"
          icon={DollarSign}
          color="purple"
        />

        <MetricCard
          title="Baseline Cost"
          value={`$${metrics?.baseline_cost?.toFixed(4) || '0.0000'}`}
          subtext={`If routed 100% to ${selectedBaseline}`}
          icon={Layers}
          color="blue"
        />

        <MetricCard
          title="Total Net Savings"
          value={`$${metrics?.estimated_savings?.toFixed(4) || '0.0000'}`}
          subtext="Retained budget capital"
          icon={PiggyBank}
          color="emerald"
        />

        <MetricCard
          title="Cost Reduction"
          value={`${metrics?.savings_percent?.toFixed(1) || '0.0'}`}
          suffix="%"
          subtext="Overall efficiency gain"
          icon={TrendingDown}
          color="emerald"
        />
      </div>

      {/* Charts Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Cost by Model Bar Chart */}
        <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h3 className="text-sm font-bold text-white">Expenditure by Model</h3>
              <p className="text-xs text-slate-400">Total USD spent per configured model</p>
            </div>
            <span className="text-xs text-slate-400 font-mono">USD ($)</span>
          </div>

          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={modelUsage} margin={{ top: 10, right: 10, left: 0, bottom: 20 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#1E2F56" vertical={false} opacity={0.4} />
                <XAxis dataKey="model" stroke="#64748B" fontSize={10} tickLine={false} angle={-15} textAnchor="end" />
                <YAxis stroke="#64748B" fontSize={11} tickLine={false} tickFormatter={(val) => `$${val}`} />
                <Tooltip
                  contentStyle={{ backgroundColor: '#0B1120', borderColor: '#1E2F56', borderRadius: '12px' }}
                  labelStyle={{ color: '#E2E8F0', fontWeight: 'bold' }}
                  formatter={(val) => [`$${Number(val).toFixed(5)}`, 'Expenditure']}
                />
                <Bar dataKey="total_cost" fill="#10B981" radius={[4, 4, 0, 0]} name="Expenditure" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Cost Distribution Donut Chart */}
        <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h3 className="text-sm font-bold text-white">Spend Allocation Share</h3>
              <p className="text-xs text-slate-400">Proportional budget breakdown</p>
            </div>
            <span className="text-xs font-semibold px-2 py-0.5 rounded bg-purple-500/10 text-purple-300 border border-purple-500/20">
              Fleet Cost Share
            </span>
          </div>

          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={85}
                  paddingAngle={4}
                  dataKey="value"
                >
                  {pieData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{ backgroundColor: '#0B1120', borderColor: '#1E2F56', borderRadius: '12px' }}
                  formatter={(val) => [`$${Number(val).toFixed(5)}`, 'Share']}
                />
                <Legend wrapperStyle={{ fontSize: '11px', paddingTop: '10px' }} />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* Model Unit Cost Table */}
      <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl">
        <h3 className="text-sm font-bold text-white mb-4">Unit Pricing & Token Volume Breakdown</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="text-[11px] uppercase tracking-wider text-slate-400 border-b border-navy-800">
              <tr>
                <th className="pb-3 font-semibold">Model</th>
                <th className="pb-3 font-semibold">Tier</th>
                <th className="pb-3 font-semibold">Total Requests</th>
                <th className="pb-3 font-semibold">Tokens Processed</th>
                <th className="pb-3 font-semibold">Actual Spend</th>
                <th className="pb-3 font-semibold">Cost per 1k Requests</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-navy-800/60 font-mono">
              {modelUsage?.map((m, idx) => {
                const costPer1k = m.request_count > 0 ? (m.total_cost / m.request_count) * 1000 : 0;
                return (
                  <tr key={idx} className="hover:bg-navy-850/50 transition-colors">
                    <td className="py-3 font-bold text-white">{m.model}</td>
                    <td className="py-3">
                      <span className="px-2 py-0.5 rounded text-[10px] uppercase font-sans font-bold bg-navy-800 text-slate-300 border border-navy-700">
                        {m.tier}
                      </span>
                    </td>
                    <td className="py-3 text-slate-300">{m.request_count.toLocaleString()}</td>
                    <td className="py-3 text-slate-300">{m.total_tokens.toLocaleString()}</td>
                    <td className="py-3 text-emerald-400 font-bold">${m.total_cost.toFixed(5)}</td>
                    <td className="py-3 text-purple-300">${costPer1k.toFixed(3)}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
