import React from 'react';

export default function MetricCard({ 
  title, 
  value, 
  subtext, 
  icon: Icon, 
  trend, 
  trendPositive, 
  color = 'purple',
  suffix = ''
}) {
  const colorStyles = {
    purple: 'text-purple-400 bg-purple-600/10 border-purple-500/20 shadow-purple-950/40',
    emerald: 'text-emerald-400 bg-emerald-600/10 border-emerald-500/20 shadow-emerald-950/40',
    blue: 'text-cyan-400 bg-cyan-600/10 border-cyan-500/20 shadow-cyan-950/40',
    amber: 'text-amber-400 bg-amber-600/10 border-amber-500/20 shadow-amber-950/40',
    rose: 'text-rose-400 bg-rose-600/10 border-rose-500/20 shadow-rose-950/40',
  };

  const activeStyle = colorStyles[color] || colorStyles.purple;

  return (
    <div className="p-5 rounded-2xl bg-navy-900/90 border border-navy-800 hover:border-purple-600/30 transition-all duration-200 shadow-lg group">
      <div className="flex items-center justify-between mb-3">
        <span className="text-xs font-semibold text-slate-400 tracking-wider uppercase">{title}</span>
        {Icon && (
          <div className={`p-2.5 rounded-xl border ${activeStyle} transition-transform group-hover:scale-105`}>
            <Icon className="w-5 h-5" />
          </div>
        )}
      </div>

      <div className="flex items-baseline space-x-1.5">
        <span className="text-2xl lg:text-3xl font-extrabold text-white tracking-tight">{value}</span>
        {suffix && <span className="text-sm font-semibold text-slate-400">{suffix}</span>}
      </div>

      <div className="mt-2.5 flex items-center justify-between text-xs">
        {subtext && <span className="text-slate-400">{subtext}</span>}
        {trend && (
          <span className={`font-semibold ${trendPositive ? 'text-emerald-400' : 'text-slate-400'}`}>
            {trend}
          </span>
        )}
      </div>
    </div>
  );
}
