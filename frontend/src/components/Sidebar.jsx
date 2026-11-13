import React from 'react';
import { 
  LayoutDashboard, 
  Cpu, 
  Server, 
  Zap, 
  TrendingDown, 
  ListFilter, 
  Sliders, 
  Settings, 
  ShieldCheck, 
  ExternalLink 
} from 'lucide-react';

export default function Sidebar({ activeTab, setActiveTab, metrics }) {
  const navItems = [
    { id: 'overview', label: 'Overview', icon: LayoutDashboard },
    { id: 'smart-router', label: 'Smart Router', icon: Cpu, badge: 'Live' },
    { id: 'providers', label: 'AI Providers', icon: Server },
    { id: 'performance', label: 'Model Performance', icon: Zap },
    { id: 'cost-analytics', label: 'Cost Analytics', icon: TrendingDown },
    { id: 'logs', label: 'Request Logs', icon: ListFilter },
    { id: 'rules', label: 'Routing Rules', icon: Sliders },
    { id: 'settings', label: 'Settings', icon: Settings },
  ];

  return (
    <aside className="w-64 bg-navy-900 border-r border-navy-800 flex flex-col justify-between h-screen sticky top-0 select-none z-20">
      <div>
        {/* Brand Header */}
        <div className="p-5 border-b border-navy-800/80 flex items-center space-x-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-purple-700 via-purple-600 to-indigo-500 flex items-center justify-center shadow-lg shadow-purple-600/30">
            <Cpu className="w-6 h-6 text-white" />
          </div>
          <div>
            <div className="flex items-center space-x-1.5">
              <span className="font-extrabold text-base tracking-tight text-white">BIFROST</span>
              <span className="text-xs font-semibold px-1.5 py-0.5 rounded bg-purple-500/20 text-purple-300 border border-purple-500/30">ROUTE</span>
            </div>
            <p className="text-[11px] text-slate-400 font-medium">AI Routing & Cost Engine</p>
          </div>
        </div>

        {/* Navigation List */}
        <nav className="p-3 space-y-1">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = activeTab === item.id;
            return (
              <button
                key={item.id}
                onClick={() => setActiveTab(item.id)}
                className={`w-full flex items-center justify-between px-3.5 py-2.5 rounded-xl font-medium text-sm transition-all duration-150 ${
                  isActive
                    ? 'bg-purple-600/20 text-purple-200 border border-purple-500/30 shadow-sm shadow-purple-950'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-navy-800/60'
                }`}
              >
                <div className="flex items-center space-x-3">
                  <Icon className={`w-4 h-4 ${isActive ? 'text-purple-400' : 'text-slate-400'}`} />
                  <span>{item.label}</span>
                </div>
                {item.badge && (
                  <span className="text-[10px] font-bold px-1.5 py-0.5 rounded-full bg-purple-500/30 text-purple-300 uppercase tracking-wider">
                    {item.badge}
                  </span>
                )}
              </button>
            );
          })}
        </nav>
      </div>

      {/* Gateway Engine Status Footer */}
      <div className="p-4 border-t border-navy-800/80 bg-navy-950/40">
        <div className="p-3 rounded-xl bg-navy-850/80 border border-navy-700/60">
          <div className="flex items-center justify-between mb-2">
            <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">Gateway Core</span>
            <div className="flex items-center space-x-1.5">
              <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
              <span className="text-xs font-semibold text-emerald-400">Online</span>
            </div>
          </div>
          <div className="text-xs text-slate-300 space-y-1">
            <div className="flex justify-between">
              <span className="text-slate-500">Port:</span>
              <span className="font-mono text-purple-300">8080</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-500">Extension:</span>
              <span className="text-slate-300">PreRequestHook</span>
            </div>
          </div>
        </div>
        
        <a 
          href="https://github.com/maximhq/bifrost" 
          target="_blank" 
          rel="noreferrer"
          className="mt-3 flex items-center justify-center space-x-1.5 text-xs text-slate-500 hover:text-purple-400 transition-colors py-1"
        >
          <span>maximhq/bifrost Gateway</span>
          <ExternalLink className="w-3 h-3" />
        </a>
      </div>
    </aside>
  );
}
