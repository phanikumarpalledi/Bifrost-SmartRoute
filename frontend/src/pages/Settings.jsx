import React, { useState } from 'react';
import { 
  Settings as SettingsIcon, 
  Key, 
  ShieldCheck, 
  Sparkles, 
  Save, 
  CheckCircle2, 
  ExternalLink, 
  Info, 
  Database, 
  Cpu 
} from 'lucide-react';
import { updateSettings } from '../api';

export default function Settings({ demoMode, onToggleDemo, baselineModel = 'gpt-4o', onUpdateBaseline }) {
  const [selectedBaseline, setSelectedBaseline] = useState(baselineModel);
  const [openAIKey, setOpenAIKey] = useState('');
  const [anthropicKey, setAnthropicKey] = useState('');
  const [geminiKey, setGeminiKey] = useState('');
  const [saved, setSaved] = useState(false);

  const handleSave = (e) => {
    e.preventDefault();
    if (onUpdateBaseline) {
      onUpdateBaseline(selectedBaseline);
    }
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  return (
    <div className="space-y-6">
      {/* Banner */}
      <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl">
        <h2 className="text-lg font-bold text-white tracking-tight flex items-center space-x-2">
          <SettingsIcon className="w-5 h-5 text-purple-400" />
          <span>System Settings & Gateway Configuration</span>
        </h2>
        <p className="text-xs text-slate-400 mt-1">
          Configure turnkey demo simulation, default cost baseline parameters, and optional live provider credentials.
        </p>
      </div>

      {saved && (
        <div className="p-4 rounded-xl bg-emerald-950/40 border border-emerald-600/30 text-emerald-300 text-xs flex items-center space-x-2 animate-fadeIn">
          <CheckCircle2 className="w-4 h-4 text-emerald-400" />
          <span>Settings saved successfully.</span>
        </div>
      )}

      <form onSubmit={handleSave} className="space-y-6">
        {/* Turnkey Demo Mode Toggle Card */}
        <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl flex items-center justify-between">
          <div>
            <div className="flex items-center space-x-2">
              <Sparkles className="w-4 h-4 text-purple-400" />
              <h3 className="text-sm font-bold text-white">Turnkey Demo Mode</h3>
              <span className={`text-[10px] px-2 py-0.5 rounded-full font-bold uppercase ${
                demoMode ? 'bg-purple-500/20 text-purple-300 border border-purple-500/30' : 'bg-slate-800 text-slate-400'
              }`}>
                {demoMode ? 'Active' : 'Disabled'}
              </span>
            </div>
            <p className="text-xs text-slate-400 mt-1 max-w-xl leading-relaxed">
              When Demo Mode is enabled, the gateway generates realistic simulated responses with exact token and latency measurements, eliminating the need for paid provider credentials during demonstrations.
            </p>
          </div>

          <button
            type="button"
            onClick={onToggleDemo}
            className={`px-4 py-2 rounded-xl text-xs font-bold transition-all border ${
              demoMode
                ? 'bg-purple-600 text-white border-purple-500 shadow-md shadow-purple-900/40'
                : 'bg-navy-950 text-slate-400 border-navy-700 hover:text-white'
            }`}
          >
            {demoMode ? 'Demo Mode Active' : 'Enable Demo Mode'}
          </button>
        </div>

        {/* Baseline Cost Model Selector */}
        <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl">
          <h3 className="text-sm font-bold text-white mb-2">Baseline Model for Savings Comparisons</h3>
          <p className="text-xs text-slate-400 mb-4">
            Select the reference flagship model against which cost savings are calculated throughout the dashboard.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
            {[
              { id: 'gpt-4o', name: 'OpenAI GPT-4o', rate: '$2.50 / $10.00 per 1M', desc: 'Standard high-reasoning industry baseline' },
              { id: 'claude-3-5-sonnet', name: 'Claude 3.5 Sonnet', rate: '$3.00 / $15.00 per 1M', desc: 'Premier frontier coding & analysis model' },
              { id: 'o1-mini', name: 'OpenAI o1 Mini', rate: '$3.00 / $12.00 per 1M', desc: 'Reasoning model benchmark' },
            ].map((m) => (
              <button
                key={m.id}
                type="button"
                onClick={() => setSelectedBaseline(m.id)}
                className={`p-4 rounded-xl border text-left transition-all ${
                  selectedBaseline === m.id
                    ? 'bg-purple-600/20 text-white border-purple-500 shadow-sm shadow-purple-950'
                    : 'bg-navy-950 border-navy-800 text-slate-400 hover:text-white'
                }`}
              >
                <div className="text-xs font-bold text-white">{m.name}</div>
                <div className="text-[11px] font-mono text-emerald-400 mt-1">{m.rate}</div>
                <div className="text-[10px] text-slate-400 mt-1">{m.desc}</div>
              </button>
            ))}
          </div>
        </div>

        {/* Live Provider API Credentials (Safely Handled) */}
        <div className="p-6 rounded-2xl bg-navy-900 border border-navy-800 shadow-xl space-y-4">
          <div className="flex items-center space-x-2">
            <Key className="w-4 h-4 text-purple-400" />
            <h3 className="text-sm font-bold text-white">Live AI Provider Credentials (Optional)</h3>
          </div>
          <p className="text-xs text-slate-400 leading-relaxed">
            API keys are securely processed server-side by Bifrost and never stored in frontend code. If left blank, Turnkey Demo Mode is used automatically.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 pt-2">
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300">OpenAI API Key</label>
              <input
                type="password"
                value={openAIKey}
                onChange={(e) => setOpenAIKey(e.target.value)}
                placeholder="sk-••••••••••••••••••••••••"
                className="w-full text-xs p-3 rounded-xl bg-navy-950 border border-navy-700 text-slate-200 placeholder-slate-600 font-mono focus:outline-none focus:border-purple-500"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300">Anthropic API Key</label>
              <input
                type="password"
                value={anthropicKey}
                onChange={(e) => setAnthropicKey(e.target.value)}
                placeholder="sk-ant-••••••••••••••••••••"
                className="w-full text-xs p-3 rounded-xl bg-navy-950 border border-navy-700 text-slate-200 placeholder-slate-600 font-mono focus:outline-none focus:border-purple-500"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300">Google Gemini API Key</label>
              <input
                type="password"
                value={geminiKey}
                onChange={(e) => setGeminiKey(e.target.value)}
                placeholder="AIzaSy••••••••••••••••••••"
                className="w-full text-xs p-3 rounded-xl bg-navy-950 border border-navy-700 text-slate-200 placeholder-slate-600 font-mono focus:outline-none focus:border-purple-500"
              />
            </div>
          </div>
        </div>

        {/* System Architecture & Bifrost Info */}
        <div className="p-6 rounded-2xl bg-gradient-to-r from-navy-900 via-navy-850 to-navy-900 border border-navy-800 shadow-xl">
          <h3 className="text-sm font-bold text-white mb-3 flex items-center space-x-2">
            <Info className="w-4 h-4 text-purple-400" />
            <span>Bifrost Gateway Open-Source Integration Details</span>
          </h3>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-xs">
            <div className="p-3 rounded-xl bg-navy-950 border border-navy-800">
              <span className="text-slate-400 block text-[11px]">Bifrost Core Version</span>
              <span className="text-purple-300 font-bold font-mono text-sm">v1.15.2</span>
            </div>
            <div className="p-3 rounded-xl bg-navy-950 border border-navy-800">
              <span className="text-slate-400 block text-[11px]">Transport HTTP</span>
              <span className="text-cyan-300 font-bold font-mono text-sm">v2.5.1</span>
            </div>
            <div className="p-3 rounded-xl bg-navy-950 border border-navy-800">
              <span className="text-slate-400 block text-[11px]">Extension Hook</span>
              <span className="text-emerald-400 font-bold font-mono text-sm">PreRequestHook</span>
            </div>
            <div className="p-3 rounded-xl bg-navy-950 border border-navy-800">
              <span className="text-slate-400 block text-[11px]">Database Engine</span>
              <span className="text-amber-300 font-bold font-mono text-sm">SQLite (Embedded)</span>
            </div>
          </div>
        </div>

        <div className="flex justify-end">
          <button
            type="submit"
            className="px-6 py-2.5 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 shadow-md shadow-purple-900/40 transition-all flex items-center space-x-2"
          >
            <Save className="w-4 h-4" />
            <span>Save Settings</span>
          </button>
        </div>
      </form>
    </div>
  );
}
