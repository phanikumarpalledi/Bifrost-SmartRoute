import React, { useState, useEffect } from 'react';
import { 
  Sliders, 
  Save, 
  CheckCircle2, 
  Sparkles, 
  Cpu, 
  Clock, 
  DollarSign, 
  ShieldCheck 
} from 'lucide-react';
import { updateRule } from '../api';

export default function RoutingRules({ rules, onRefresh }) {
  const [activeRule, setActiveRule] = useState(null);
  const [costWeight, setCostWeight] = useState(0.55);
  const [latencyWeight, setLatencyWeight] = useState(0.25);
  const [reliabilityWeight, setReliabilityWeight] = useState(0.20);
  const [complexityThreshold, setComplexityThreshold] = useState(4.0);
  const [baselineModel, setBaselineModel] = useState('gpt-4o');
  const [strategy, setStrategy] = useState('cost_optimized');
  const [saving, setSaving] = useState(false);
  const [savedSuccess, setSavedSuccess] = useState(false);

  useEffect(() => {
    if (rules && rules.length > 0) {
      const active = rules.find((r) => r.is_active) || rules[0];
      setActiveRule(active);
      setCostWeight(active.cost_weight);
      setLatencyWeight(active.latency_weight);
      setReliabilityWeight(active.reliability_weight);
      setComplexityThreshold(active.complexity_threshold);
      setBaselineModel(active.baseline_model || 'gpt-4o');
      setStrategy(active.strategy);
    }
  }, [rules]);

  const handleStrategyPreset = (preset) => {
    setStrategy(preset);
    switch (preset) {
      case 'cost_optimized':
        setCostWeight(0.60);
        setLatencyWeight(0.20);
        setReliabilityWeight(0.20);
        setComplexityThreshold(4.0);
        break;
      case 'balanced':
        setCostWeight(0.40);
        setLatencyWeight(0.35);
        setReliabilityWeight(0.25);
        setComplexityThreshold(5.0);
        break;
      case 'performance_first':
        setCostWeight(0.10);
        setLatencyWeight(0.70);
        setReliabilityWeight(0.20);
        setComplexityThreshold(6.0);
        break;
      default:
        break;
    }
  };

  const handleSave = async (e) => {
    e.preventDefault();
    if (!activeRule) return;

    setSaving(true);
    setSavedSuccess(false);

    try {
      const payload = {
        id: activeRule.id,
        name: activeRule.name,
        strategy: strategy,
        complexity_threshold: parseFloat(complexityThreshold),
        baseline_model: baselineModel,
        cost_weight: parseFloat(costWeight),
        latency_weight: parseFloat(latencyWeight),
        reliability_weight: parseFloat(reliabilityWeight),
        is_active: true,
      };

      await updateRule(payload);
      setSavedSuccess(true);
      if (onRefresh) onRefresh();
      setTimeout(() => setSavedSuccess(false), 3000);
    } catch (err) {
      alert('Failed to save rule: ' + err.message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Banner */}
      <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h2 className="text-lg font-bold text-white tracking-tight flex items-center space-x-2">
            <Sliders className="w-5 h-5 text-purple-400" />
            <span>Multi-Factor Routing Rules & Policy Engine</span>
          </h2>
          <p className="text-xs text-slate-400 mt-1">
            Configure the mathematical weights governing model selection in Bifrost's PreRequestHook execution phase.
          </p>
        </div>

        <span className="text-xs font-semibold px-3 py-1.5 rounded-lg bg-purple-500/10 text-purple-300 border border-purple-500/20 font-mono">
          Engine: PreRequestHook
        </span>
      </div>

      {savedSuccess && (
        <div className="p-4 rounded-xl bg-emerald-950/40 border border-emerald-600/30 text-emerald-300 text-xs flex items-center space-x-2 animate-fadeIn">
          <CheckCircle2 className="w-4 h-4 text-emerald-400" />
          <span>Routing strategy parameters updated successfully. Applied to all new incoming requests.</span>
        </div>
      )}

      {/* Main Settings Form */}
      <form onSubmit={handleSave} className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left 2 Cols: Sliders & Policy Settings */}
        <div className="lg:col-span-2 p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl space-y-6">
          <div>
            <h3 className="text-sm font-bold text-white mb-2">Preset Routing Strategy</h3>
            <div className="grid grid-cols-3 gap-3">
              {[
                { id: 'cost_optimized', label: 'Cost Optimizer', desc: 'Max savings on simple/creative tasks' },
                { id: 'balanced', label: 'Balanced SLA', desc: 'Optimal trade-off for mixed workloads' },
                { id: 'performance_first', label: 'Low Latency', desc: 'Speed priority with strict SLAs' },
              ].map((p) => (
                <button
                  key={p.id}
                  type="button"
                  onClick={() => handleStrategyPreset(p.id)}
                  className={`p-3.5 rounded-xl border text-left transition-all ${
                    strategy === p.id
                      ? 'bg-purple-600/20 text-white border-purple-500 shadow-sm shadow-purple-950'
                      : 'bg-navy-950 border-navy-800 text-slate-400 hover:text-white hover:bg-navy-850'
                  }`}
                >
                  <div className="text-xs font-bold text-white mb-0.5">{p.label}</div>
                  <div className="text-[10px] text-slate-400">{p.desc}</div>
                </button>
              ))}
            </div>
          </div>

          <div className="space-y-5 pt-4 border-t border-navy-800">
            <h3 className="text-sm font-bold text-white">Optimization Utility Weights</h3>

            {/* Cost Weight Slider */}
            <div className="space-y-2">
              <div className="flex justify-between text-xs">
                <span className="text-slate-300 font-medium flex items-center space-x-1.5">
                  <DollarSign className="w-3.5 h-3.5 text-emerald-400" />
                  <span>Cost Optimization Weight</span>
                </span>
                <span className="font-mono font-bold text-emerald-400">{(costWeight * 100).toFixed(0)}%</span>
              </div>
              <input
                type="range"
                min="0"
                max="1"
                step="0.05"
                value={costWeight}
                onChange={(e) => setCostWeight(parseFloat(e.target.value))}
                className="w-full accent-emerald-500 h-1.5 bg-navy-950 rounded-lg cursor-pointer"
              />
            </div>

            {/* Latency Weight Slider */}
            <div className="space-y-2">
              <div className="flex justify-between text-xs">
                <span className="text-slate-300 font-medium flex items-center space-x-1.5">
                  <Clock className="w-3.5 h-3.5 text-cyan-400" />
                  <span>Latency SLA Weight</span>
                </span>
                <span className="font-mono font-bold text-cyan-400">{(latencyWeight * 100).toFixed(0)}%</span>
              </div>
              <input
                type="range"
                min="0"
                max="1"
                step="0.05"
                value={latencyWeight}
                onChange={(e) => setLatencyWeight(parseFloat(e.target.value))}
                className="w-full accent-cyan-500 h-1.5 bg-navy-950 rounded-lg cursor-pointer"
              />
            </div>

            {/* Reliability Weight Slider */}
            <div className="space-y-2">
              <div className="flex justify-between text-xs">
                <span className="text-slate-300 font-medium flex items-center space-x-1.5">
                  <ShieldCheck className="w-3.5 h-3.5 text-purple-400" />
                  <span>Provider Health & Reliability Weight</span>
                </span>
                <span className="font-mono font-bold text-purple-400">{(reliabilityWeight * 100).toFixed(0)}%</span>
              </div>
              <input
                type="range"
                min="0"
                max="1"
                step="0.05"
                value={reliabilityWeight}
                onChange={(e) => setReliabilityWeight(parseFloat(e.target.value))}
                className="w-full accent-purple-500 h-1.5 bg-navy-950 rounded-lg cursor-pointer"
              />
            </div>

            {/* Complexity Threshold */}
            <div className="space-y-2 pt-2">
              <div className="flex justify-between text-xs">
                <span className="text-slate-300 font-medium">Complexity Frontier Threshold (Score &gt; X routes to Flagship Tier)</span>
                <span className="font-mono font-bold text-white">{complexityThreshold.toFixed(1)} / 10.0</span>
              </div>
              <input
                type="range"
                min="1"
                max="9"
                step="0.5"
                value={complexityThreshold}
                onChange={(e) => setComplexityThreshold(parseFloat(e.target.value))}
                className="w-full accent-purple-500 h-1.5 bg-navy-950 rounded-lg cursor-pointer"
              />
            </div>
          </div>

          <div className="pt-4 border-t border-navy-800 flex justify-end">
            <button
              type="submit"
              disabled={saving}
              className="px-6 py-2.5 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 shadow-md shadow-purple-900/40 transition-all flex items-center space-x-2 disabled:opacity-50"
            >
              <Save className="w-4 h-4" />
              <span>{saving ? 'Applying...' : 'Save & Apply Routing Policy'}</span>
            </button>
          </div>
        </div>

        {/* Right 1 Col: Algorithm Visualizer */}
        <div className="p-6 rounded-2xl bg-gradient-to-b from-navy-900 to-navy-950 border border-purple-600/20 shadow-xl flex flex-col justify-between">
          <div>
            <h3 className="text-sm font-bold text-white mb-2 flex items-center space-x-2">
              <Sparkles className="w-4 h-4 text-purple-400" />
              <span>Optimization Formula</span>
            </h3>
            <p className="text-xs text-slate-400 mb-4 leading-relaxed">
              Every request evaluated in Bifrost's PreRequestHook calculates candidate utility using:
            </p>

            <div className="p-3.5 rounded-xl bg-navy-950 border border-navy-800 font-mono text-[11px] text-purple-300 leading-relaxed mb-4">
              <div>Utility(m) =</div>
              <div className="pl-3 text-emerald-400">+{costWeight.toFixed(2)} × CostScore(m)</div>
              <div className="pl-3 text-cyan-400">+{latencyWeight.toFixed(2)} × LatencyScore(m)</div>
              <div className="pl-3 text-purple-400">+{reliabilityWeight.toFixed(2)} × Health(m)</div>
              <div className="pl-3 text-slate-400">+ TierBonus(m, intent)</div>
            </div>

            <div className="space-y-2 text-xs text-slate-300">
              <div className="flex items-center space-x-2">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
                <span>Simple Q&A &lt; 3.8 ➔ Gemini Flash / GPT-4o-mini</span>
              </div>
              <div className="flex items-center space-x-2">
                <span className="w-1.5 h-1.5 rounded-full bg-cyan-400"></span>
                <span>Coding Intent ➔ Claude 3.5 Sonnet / GPT-4o</span>
              </div>
              <div className="flex items-center space-x-2">
                <span className="w-1.5 h-1.5 rounded-full bg-purple-400"></span>
                <span>Complex Reasoning &ge; 7.0 ➔ Frontier Tier</span>
              </div>
            </div>
          </div>

          <div className="mt-6 pt-4 border-t border-navy-800 text-[11px] text-slate-500">
            Bifrost SmartRoute natively configures the <strong className="text-slate-400">req.Fallbacks</strong> array on every decision.
          </div>
        </div>
      </form>
    </div>
  );
}
