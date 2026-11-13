import React from 'react';
import { RefreshCw, Sparkles, CheckCircle2, AlertTriangle } from 'lucide-react';

export default function Header({ 
  title, 
  subtitle, 
  demoMode, 
  onToggleDemo, 
  baselineModel = 'gpt-4o', 
  onRefresh, 
  refreshing 
}) {
  return (
    <header className="px-8 py-5 border-b border-navy-800/80 bg-navy-950/60 backdrop-blur-md sticky top-0 z-10 flex items-center justify-between">
      <div>
        <h1 className="text-xl font-extrabold text-white tracking-tight flex items-center space-x-2.5">
          <span>{title}</span>
        </h1>
        {subtitle && (
          <p className="text-xs text-slate-400 mt-0.5 font-medium">{subtitle}</p>
        )}
      </div>

      <div className="flex items-center space-x-4">
        {/* Baseline Model Badge */}
        <div className="hidden md:flex items-center space-x-2 px-3 py-1.5 rounded-lg bg-navy-850 border border-navy-700/60 text-xs">
          <span className="text-slate-400">Baseline Model:</span>
          <span className="font-mono font-semibold text-purple-300">{baselineModel}</span>
        </div>

        {/* Demo Mode Toggle Badge */}
        <button
          onClick={onToggleDemo}
          className={`flex items-center space-x-2 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all border ${
            demoMode
              ? 'bg-purple-950/50 text-purple-300 border-purple-600/40 hover:bg-purple-900/40 shadow-sm shadow-purple-950'
              : 'bg-emerald-950/40 text-emerald-300 border-emerald-600/40 hover:bg-emerald-900/30'
          }`}
          title="Toggle between Turnkey Demo Mode (simulated responses) and Live Provider Mode"
        >
          {demoMode ? (
            <>
              <Sparkles className="w-3.5 h-3.5 text-purple-400" />
              <span>Demo Mode: Simulated Telemetry</span>
            </>
          ) : (
            <>
              <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
              <span>Live Provider Gateway Mode</span>
            </>
          )}
        </button>

        {/* Refresh Button */}
        <button
          onClick={onRefresh}
          disabled={refreshing}
          className="p-2 rounded-lg bg-navy-850 hover:bg-navy-800 border border-navy-700/60 text-slate-300 hover:text-white transition-all disabled:opacity-50"
          title="Refresh telemetry and charts"
        >
          <RefreshCw className={`w-4 h-4 ${refreshing ? 'animate-spin text-purple-400' : ''}`} />
        </button>
      </div>
    </header>
  );
}
