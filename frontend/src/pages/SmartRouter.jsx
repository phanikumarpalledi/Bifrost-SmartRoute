import React, { useState, useEffect } from 'react';
import { 
  Cpu, 
  Send, 
  Sparkles, 
  Clock, 
  DollarSign, 
  ShieldCheck, 
  Copy, 
  Check, 
  Code2, 
  BrainCircuit, 
  MessageSquare, 
  PenTool, 
  SlidersHorizontal 
} from 'lucide-react';
import { routePrompt } from '../api';

export default function SmartRouter({ initialPrompt = '', models = [] }) {
  const [prompt, setPrompt] = useState(initialPrompt);
  const [modelOverride, setModelOverride] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);
  const [error, setError] = useState(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (initialPrompt) {
      setPrompt(initialPrompt);
    }
  }, [initialPrompt]);

  const presetPrompts = [
    {
      category: 'simple',
      label: 'Simple Fact',
      prompt: 'What is the boiling point of water at sea level in Celsius and Fahrenheit?',
      icon: MessageSquare,
    },
    {
      category: 'coding',
      label: 'Go Goroutines',
      prompt: 'Write an idiomatic Go worker pool with channel buffering, error handling, and cancellation using context.Context.',
      icon: Code2,
    },
    {
      category: 'reasoning',
      label: 'Consensus Analysis',
      prompt: 'Analyze the architectural and performance trade-offs between Paxos, Raft, and PBFT consensus protocols in geo-distributed systems.',
      icon: BrainCircuit,
    },
    {
      category: 'creative',
      label: 'Executive Brief',
      prompt: 'Draft an executive briefing summarizing how AI gateway intelligent routing slashes cloud inference costs by 60% without hurting SLAs.',
      icon: PenTool,
    },
  ];

  const handleRoute = async (e) => {
    if (e) e.preventDefault();
    if (!prompt.trim() || loading) return;

    setLoading(true);
    setError(null);
    try {
      const res = await routePrompt(prompt, modelOverride);
      setResult(res);
    } catch (err) {
      setError(err.message || 'Routing failed');
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = () => {
    if (!result?.actual_response) return;
    navigator.clipboard.writeText(result.actual_response);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const getCategoryBadge = (cat) => {
    switch (cat) {
      case 'coding':
        return { label: 'Coding / Technical', bg: 'bg-cyan-500/10 text-cyan-300 border-cyan-500/30' };
      case 'reasoning':
        return { label: 'Complex Reasoning', bg: 'bg-purple-500/20 text-purple-300 border-purple-500/30' };
      case 'simple':
        return { label: 'Simple / Factual', bg: 'bg-emerald-500/10 text-emerald-300 border-emerald-500/30' };
      default:
        return { label: 'Creative / Summary', bg: 'bg-amber-500/10 text-amber-300 border-amber-500/30' };
    }
  };

  return (
    <div className="space-y-6">
      {/* Playground Header Card */}
      <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-5">
          <div>
            <h2 className="text-lg font-bold text-white tracking-tight flex items-center space-x-2">
              <Cpu className="w-5 h-5 text-purple-400" />
              <span>Intelligent Model Routing Playground</span>
            </h2>
            <p className="text-xs text-slate-400 mt-1">
              Submit prompts to test dynamic multi-factor model selection, complexity categorization, and cost projection.
            </p>
          </div>

          {/* Model Override Selector */}
          <div className="flex items-center space-x-2">
            <SlidersHorizontal className="w-4 h-4 text-slate-400" />
            <select
              value={modelOverride}
              onChange={(e) => setModelOverride(e.target.value)}
              className="bg-navy-950 border border-navy-700 text-xs text-purple-200 rounded-xl px-3 py-2 font-medium focus:outline-none focus:border-purple-500"
            >
              <option value="">Auto: Intelligent Routing (Recommended)</option>
              {models?.map((m) => (
                <option key={m.id} value={m.model_name}>
                  Manual Override: {m.display_name} ({m.provider_id})
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Presets */}
        <div className="flex flex-wrap gap-2 mb-4">
          <span className="text-xs font-semibold text-slate-400 self-center mr-1">Sample Prompts:</span>
          {presetPrompts.map((p, idx) => {
            const Icon = p.icon;
            return (
              <button
                key={idx}
                type="button"
                onClick={() => setPrompt(p.prompt)}
                className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-navy-850 hover:bg-navy-800 text-slate-300 hover:text-white border border-navy-700/80 text-xs transition-colors"
              >
                <Icon className="w-3.5 h-3.5 text-purple-400" />
                <span>{p.label}</span>
              </button>
            );
          })}
        </div>

        {/* Prompt Input Form */}
        <form onSubmit={handleRoute} className="space-y-4">
          <textarea
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            placeholder="Type or paste your prompt here to evaluate intelligent routing..."
            rows={5}
            className="w-full p-4 rounded-xl bg-navy-950 border border-navy-700/80 text-slate-100 placeholder-slate-500 focus:outline-none focus:border-purple-500 font-mono text-xs leading-relaxed resize-y"
          />

          <div className="flex justify-between items-center">
            <span className="text-[11px] text-slate-500 font-mono">
              {prompt.length} characters • ~{Math.ceil(prompt.length / 3.8)} estimated tokens
            </span>

            <button
              type="submit"
              disabled={loading || !prompt.trim()}
              className="px-6 py-2.5 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 shadow-md shadow-purple-900/40 transition-all flex items-center space-x-2 disabled:opacity-50"
            >
              {loading ? (
                <>
                  <Sparkles className="w-4 h-4 animate-spin" />
                  <span>Evaluating Bifrost Hooks...</span>
                </>
              ) : (
                <>
                  <Send className="w-4 h-4" />
                  <span>Route & Execute</span>
                </>
              )}
            </button>
          </div>
        </form>

        {error && (
          <div className="mt-4 p-3 rounded-xl bg-rose-950/40 border border-rose-600/30 text-rose-300 text-xs">
            {error}
          </div>
        )}
      </div>

      {/* Decision & Response Results Grid */}
      {result?.decision && (
        <div className="space-y-6 animate-fadeIn">
          {/* Decision Rationale Cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* 1. Category */}
            <div className="p-4 rounded-xl bg-navy-900 border border-navy-800">
              <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block mb-2">Detected Task Category</span>
              <div className="flex items-center space-x-2">
                <span className={`px-2.5 py-1 rounded-lg text-xs font-bold border ${getCategoryBadge(result.decision.task_category).bg}`}>
                  {getCategoryBadge(result.decision.task_category).label}
                </span>
              </div>
            </div>

            {/* 2. Complexity Score */}
            <div className="p-4 rounded-xl bg-navy-900 border border-navy-800">
              <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block mb-2">Complexity Score</span>
              <div className="flex items-baseline space-x-2">
                <span className="text-xl font-extrabold text-white">{result.decision.complexity_score}</span>
                <span className="text-xs text-slate-400">/ 10.0</span>
              </div>
              <div className="w-full bg-navy-950 rounded-full h-1.5 mt-2 overflow-hidden border border-navy-800">
                <div 
                  className={`h-full rounded-full ${
                    result.decision.complexity_score > 6 
                      ? 'bg-purple-500' 
                      : result.decision.complexity_score > 3 
                        ? 'bg-cyan-500' 
                        : 'bg-emerald-400'
                  }`}
                  style={{ width: `${Math.min(100, result.decision.complexity_score * 10)}%` }}
                />
              </div>
            </div>

            {/* 3. Selected Model */}
            <div className="p-4 rounded-xl bg-navy-900 border border-purple-500/30 bg-purple-950/10 shadow-lg shadow-purple-950/20">
              <span className="text-[11px] font-semibold text-purple-300 uppercase tracking-wider block mb-2">Selected Model & Provider</span>
              <div className="font-mono font-bold text-sm text-white truncate">
                {result.decision.selected_model}
              </div>
              <span className="text-xs text-purple-400 uppercase font-semibold">
                Provider: {result.decision.selected_provider}
              </span>
            </div>

            {/* 4. Cost & Savings */}
            <div className="p-4 rounded-xl bg-navy-900 border border-navy-800">
              <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block mb-2">Cost & Projected Savings</span>
              <div className="flex items-baseline space-x-2">
                <span className="text-lg font-extrabold text-emerald-400">${result.decision.estimated_cost?.toFixed(5)}</span>
                <span className="text-xs font-bold text-emerald-400 px-1.5 py-0.2 rounded bg-emerald-500/10 border border-emerald-500/20">
                  +{result.decision.savings_percent?.toFixed(1)}% saved
                </span>
              </div>
              <span className="text-[11px] text-slate-500 mt-1 block">
                vs {result.decision.baseline_model} (${result.decision.baseline_cost?.toFixed(5)})
              </span>
            </div>
          </div>

          {/* Explanation Banner */}
          <div className="p-5 rounded-2xl bg-gradient-to-r from-navy-900 via-purple-950/30 to-navy-900 border border-purple-600/30 shadow-lg">
            <h3 className="text-xs font-bold uppercase tracking-wider text-purple-300 mb-1 flex items-center space-x-1.5">
              <BrainCircuit className="w-4 h-4 text-purple-400" />
              <span>Routing Decision Explanation</span>
            </h3>
            <p className="text-sm text-slate-200 leading-relaxed font-medium">
              {result.decision.explanation}
            </p>

            <div className="mt-3 pt-3 border-t border-purple-500/20 flex flex-wrap items-center justify-between text-xs text-slate-400">
              <div className="flex items-center space-x-4">
                <span>Expected Latency: <strong className="text-white">{result.decision.expected_latency_ms}ms</strong></span>
                <span>Fallback Chain: <strong className="text-slate-300">{result.fallback_status}</strong></span>
              </div>
              <span className="text-[11px] text-purple-300 font-mono">Request ID: {result.decision.request_id}</span>
            </div>
          </div>

          {/* Actual Model Output Response Card */}
          <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center space-x-2">
                <h3 className="text-sm font-bold text-white">Generated Provider Response</h3>
                <span className="text-[11px] px-2 py-0.5 rounded bg-navy-800 text-slate-400 font-mono">
                  {result.execution_ms}ms roundtrip
                </span>
              </div>

              <button
                onClick={handleCopy}
                className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-navy-800 hover:bg-navy-750 text-slate-300 hover:text-white border border-navy-700 text-xs transition-colors"
              >
                {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                <span>{copied ? 'Copied!' : 'Copy Response'}</span>
              </button>
            </div>

            <div className="p-4 rounded-xl bg-navy-950 border border-navy-800 text-slate-200 text-xs font-mono whitespace-pre-wrap leading-relaxed overflow-x-auto">
              {result.actual_response}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
