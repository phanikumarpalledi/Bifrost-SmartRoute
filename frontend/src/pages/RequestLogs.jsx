import React, { useState } from 'react';
import { 
  ListFilter, 
  Search, 
  Clock, 
  DollarSign, 
  ShieldCheck, 
  AlertTriangle, 
  CheckCircle2, 
  XCircle, 
  X, 
  ArrowRight, 
  FileText 
} from 'lucide-react';

export default function RequestLogs({ logs, onFilterChange }) {
  const [searchTerm, setSearchTerm] = useState('');
  const [providerFilter, setProviderFilter] = useState('all');
  const [statusFilter, setStatusFilter] = useState('all');
  const [selectedLog, setSelectedLog] = useState(null);

  const filteredLogs = logs?.filter((l) => {
    const matchesSearch = 
      l.request_id?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      l.model?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      l.provider?.toLowerCase().includes(searchTerm.toLowerCase());
    
    const matchesProvider = providerFilter === 'all' || l.provider === providerFilter;
    const matchesStatus = statusFilter === 'all' || l.status === statusFilter;

    return matchesSearch && matchesProvider && matchesStatus;
  }) || [];

  const handleProviderSelect = (e) => {
    setProviderFilter(e.target.value);
    if (onFilterChange) onFilterChange(e.target.value, statusFilter);
  };

  const handleStatusSelect = (e) => {
    setStatusFilter(e.target.value);
    if (onFilterChange) onFilterChange(providerFilter, e.target.value);
  };

  const getStatusBadge = (status) => {
    switch (status) {
      case 'success':
        return (
          <span className="flex items-center space-x-1 px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/10 text-emerald-300 border border-emerald-500/20">
            <CheckCircle2 className="w-3 h-3" />
            <span>Success</span>
          </span>
        );
      case 'fallback_recovered':
        return (
          <span className="flex items-center space-x-1 px-2 py-0.5 rounded text-[10px] font-bold bg-amber-500/20 text-amber-300 border border-amber-500/30">
            <AlertTriangle className="w-3 h-3" />
            <span>Fallback Recovered</span>
          </span>
        );
      default:
        return (
          <span className="flex items-center space-x-1 px-2 py-0.5 rounded text-[10px] font-bold bg-rose-500/20 text-rose-300 border border-rose-500/30">
            <XCircle className="w-3 h-3" />
            <span>Failed</span>
          </span>
        );
    }
  };

  return (
    <div className="space-y-6">
      {/* Header and Filter Controls */}
      <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h2 className="text-lg font-bold text-white tracking-tight flex items-center space-x-2">
              <ListFilter className="w-5 h-5 text-purple-400" />
              <span>Gateway Request Logs & Audit Trail</span>
            </h2>
            <p className="text-xs text-slate-400 mt-1">
              Complete historical record of incoming AI requests, latency benchmarks, token usage, and cost savings.
            </p>
          </div>

          <span className="text-xs text-slate-400 font-mono">
            Showing {filteredLogs.length} of {logs?.length || 0} events
          </span>
        </div>

        {/* Filter Bar */}
        <div className="flex flex-col md:flex-row items-center gap-3 pt-2">
          {/* Search Input */}
          <div className="relative flex-1 w-full">
            <Search className="w-4 h-4 text-slate-500 absolute left-3.5 top-3" />
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Search by Request ID, Model, or Provider..."
              className="w-full text-xs pl-10 pr-4 py-2.5 rounded-xl bg-navy-950 border border-navy-700 text-slate-200 placeholder-slate-500 focus:outline-none focus:border-purple-500 font-mono"
            />
          </div>

          {/* Provider Filter */}
          <select
            value={providerFilter}
            onChange={handleProviderSelect}
            className="w-full md:w-48 text-xs px-3 py-2.5 rounded-xl bg-navy-950 border border-navy-700 text-slate-200 focus:outline-none focus:border-purple-500"
          >
            <option value="all">All Providers</option>
            <option value="openai">OpenAI</option>
            <option value="anthropic">Anthropic</option>
            <option value="gemini">Google Gemini</option>
            <option value="mock">Demo Mock</option>
          </select>

          {/* Status Filter */}
          <select
            value={statusFilter}
            onChange={handleStatusSelect}
            className="w-full md:w-48 text-xs px-3 py-2.5 rounded-xl bg-navy-950 border border-navy-700 text-slate-200 focus:outline-none focus:border-purple-500"
          >
            <option value="all">All Statuses</option>
            <option value="success">Success</option>
            <option value="fallback_recovered">Fallback Recovered</option>
            <option value="failed">Failed</option>
          </select>
        </div>
      </div>

      {/* Logs Table */}
      <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="text-[11px] uppercase tracking-wider text-slate-400 border-b border-navy-800">
              <tr>
                <th className="pb-3 font-semibold">Timestamp</th>
                <th className="pb-3 font-semibold">Request ID</th>
                <th className="pb-3 font-semibold">Target Model</th>
                <th className="pb-3 font-semibold">Provider</th>
                <th className="pb-3 font-semibold">Latency</th>
                <th className="pb-3 font-semibold">Tokens (In / Out)</th>
                <th className="pb-3 font-semibold">Actual Cost</th>
                <th className="pb-3 font-semibold">Savings</th>
                <th className="pb-3 font-semibold">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-navy-800/60 font-mono">
              {filteredLogs.map((log) => (
                <tr 
                  key={log.id} 
                  onClick={() => setSelectedLog(log)}
                  className="hover:bg-navy-850/60 cursor-pointer transition-colors group"
                >
                  <td className="py-3 text-slate-400 text-[11px]">
                    {new Date(log.created_at).toLocaleTimeString()}
                  </td>
                  <td className="py-3 font-bold text-purple-300 group-hover:underline">
                    {log.request_id?.slice(0, 10)}...
                  </td>
                  <td className="py-3 text-white font-medium">{log.model}</td>
                  <td className="py-3 uppercase text-slate-400 text-[11px]">{log.provider}</td>
                  <td className="py-3 text-amber-300 font-sans">{log.latency_ms}ms</td>
                  <td className="py-3 text-slate-300">
                    <span className="text-slate-400">{log.prompt_tokens}</span> + <span className="text-slate-200">{log.completion_tokens}</span>
                  </td>
                  <td className="py-3 text-emerald-400">${log.actual_cost.toFixed(5)}</td>
                  <td className="py-3 text-emerald-400 font-bold">
                    {log.savings > 0 ? `+$${log.savings.toFixed(5)}` : '$0.00'}
                  </td>
                  <td className="py-3 font-sans">{getStatusBadge(log.status)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {filteredLogs.length === 0 && (
          <div className="p-8 text-center text-slate-500 text-xs">
            No request logs match the selected filters.
          </div>
        )}
      </div>

      {/* Log Detail Drawer Modal */}
      {selectedLog && (
        <div className="fixed inset-0 bg-navy-950/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-navy-900 border border-purple-600/30 rounded-2xl max-w-lg w-full p-6 shadow-2xl animate-scaleUp">
            <div className="flex items-center justify-between pb-4 border-b border-navy-800">
              <div className="flex items-center space-x-2">
                <FileText className="w-5 h-5 text-purple-400" />
                <h3 className="text-base font-bold text-white">Request Telemetry Audit</h3>
              </div>
              <button 
                onClick={() => setSelectedLog(null)}
                className="p-1.5 rounded-lg bg-navy-800 text-slate-400 hover:text-white transition-colors"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="space-y-3 mt-4 text-xs">
              <div className="flex justify-between py-1.5 border-b border-navy-800/60 font-mono">
                <span className="text-slate-400">Request ID:</span>
                <span className="text-purple-300 font-bold">{selectedLog.request_id}</span>
              </div>

              <div className="flex justify-between py-1.5 border-b border-navy-800/60">
                <span className="text-slate-400">Status:</span>
                <span>{getStatusBadge(selectedLog.status)}</span>
              </div>

              <div className="flex justify-between py-1.5 border-b border-navy-800/60 font-mono">
                <span className="text-slate-400">Model / Provider:</span>
                <span className="text-white font-semibold">{selectedLog.model} ({selectedLog.provider})</span>
              </div>

              <div className="flex justify-between py-1.5 border-b border-navy-800/60 font-mono">
                <span className="text-slate-400">Roundtrip Latency:</span>
                <span className="text-amber-300 font-bold">{selectedLog.latency_ms} ms</span>
              </div>

              <div className="flex justify-between py-1.5 border-b border-navy-800/60 font-mono">
                <span className="text-slate-400">Prompt / Completion Tokens:</span>
                <span className="text-slate-200">{selectedLog.prompt_tokens} in / {selectedLog.completion_tokens} out ({selectedLog.total_tokens} total)</span>
              </div>

              <div className="flex justify-between py-1.5 border-b border-navy-800/60 font-mono">
                <span className="text-slate-400">Actual Cost (SmartRoute):</span>
                <span className="text-emerald-400 font-bold">${selectedLog.actual_cost.toFixed(6)}</span>
              </div>

              <div className="flex justify-between py-1.5 border-b border-navy-800/60 font-mono">
                <span className="text-slate-400">Baseline Model Cost:</span>
                <span className="text-slate-300">${selectedLog.baseline_cost.toFixed(6)}</span>
              </div>

              <div className="flex justify-between py-1.5 border-b border-navy-800/60 font-mono">
                <span className="text-slate-400">Net Dollar Savings:</span>
                <span className="text-emerald-400 font-bold">+${selectedLog.savings.toFixed(6)} ({selectedLog.savings_pct.toFixed(1)}%)</span>
              </div>

              <div className="flex justify-between py-1.5 border-b border-navy-800/60 font-mono">
                <span className="text-slate-400">Data Origin:</span>
                <span className={selectedLog.is_simulated ? 'text-purple-400' : 'text-emerald-400'}>
                  {selectedLog.is_simulated ? 'Demo Simulation' : 'Live Gateway Production'}
                </span>
              </div>

              {selectedLog.error_message && (
                <div className="p-3 rounded-xl bg-rose-950/40 border border-rose-600/30 text-rose-300 font-mono text-[11px]">
                  Error: {selectedLog.error_message}
                </div>
              )}
            </div>

            <button
              onClick={() => setSelectedLog(null)}
              className="mt-6 w-full py-2.5 rounded-xl bg-navy-800 hover:bg-navy-750 text-slate-200 font-semibold text-xs transition-colors"
            >
              Close Inspector
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
