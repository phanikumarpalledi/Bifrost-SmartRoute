import React, { useState, useEffect } from 'react';
import Sidebar from './components/Sidebar';
import Header from './components/Header';
import Overview from './pages/Overview';
import SmartRouter from './pages/SmartRouter';
import Providers from './pages/Providers';
import ModelPerformance from './pages/ModelPerformance';
import CostAnalytics from './pages/CostAnalytics';
import RequestLogs from './pages/RequestLogs';
import RoutingRules from './pages/RoutingRules';
import Settings from './pages/Settings';
import { 
  fetchMetrics, 
  fetchCostOverTime, 
  fetchProviderPerformances, 
  fetchModelUsage, 
  fetchLogs, 
  fetchFailovers, 
  fetchProviders, 
  fetchRules, 
  fetchSettings, 
  updateSettings 
} from './api';

export default function App() {
  const [activeTab, setActiveTab] = useState('overview');
  const [tabParams, setTabParams] = useState({});
  const [metrics, setMetrics] = useState(null);
  const [costOverTime, setCostOverTime] = useState([]);
  const [providerPerformances, setProviderPerformances] = useState([]);
  const [modelUsage, setModelUsage] = useState([]);
  const [logs, setLogs] = useState([]);
  const [failovers, setFailovers] = useState([]);
  const [providersData, setProvidersData] = useState(null);
  const [rules, setRules] = useState([]);
  const [demoMode, setDemoMode] = useState(true);
  const [baselineModel, setBaselineModel] = useState('gpt-4o');
  const [refreshing, setRefreshing] = useState(false);

  const loadData = async () => {
    setRefreshing(true);
    try {
      const [
        m,
        cost,
        perfs,
        usage,
        l,
        f,
        p,
        r,
        s,
      ] = await Promise.all([
        fetchMetrics().catch(() => null),
        fetchCostOverTime().catch(() => []),
        fetchProviderPerformances().catch(() => []),
        fetchModelUsage().catch(() => []),
        fetchLogs(60).catch(() => []),
        fetchFailovers().catch(() => []),
        fetchProviders().catch(() => null),
        fetchRules().catch(() => []),
        fetchSettings().catch(() => null),
      ]);

      if (m) setMetrics(m);
      if (cost) setCostOverTime(cost);
      if (perfs) setProviderPerformances(perfs);
      if (usage) setModelUsage(usage);
      if (l) setLogs(l);
      if (f) setFailovers(f);
      if (p) setProvidersData(p);
      if (r) setRules(r);
      if (s) {
        if (s.demo_mode !== undefined) setDemoMode(s.demo_mode);
        if (s.baseline_model) setBaselineModel(s.baseline_model);
      }
    } catch (err) {
      console.error('Failed to load dashboard data:', err);
    } finally {
      setRefreshing(false);
    }
  };

  useEffect(() => {
    loadData();
    // Auto-poll every 15 seconds
    const interval = setInterval(loadData, 15000);
    return () => clearInterval(interval);
  }, []);

  const handleToggleDemo = async () => {
    const nextMode = !demoMode;
    setDemoMode(nextMode);
    try {
      await updateSettings({ demo_mode: nextMode });
      loadData();
    } catch (err) {
      console.error('Failed to update demo mode:', err);
    }
  };

  const handleUpdateBaseline = async (model) => {
    setBaselineModel(model);
    try {
      await updateSettings({ baseline_model: model });
      loadData();
    } catch (err) {
      console.error('Failed to update baseline model:', err);
    }
  };

  const handleNavigate = (tabId, params = {}) => {
    setTabParams(params);
    setActiveTab(tabId);
  };

  const getPageInfo = () => {
    switch (activeTab) {
      case 'overview':
        return { title: 'System Overview & Telemetry', subtitle: 'Global AI gateway traffic, cost efficiency, and latency health' };
      case 'smart-router':
        return { title: 'Smart Router Testing Playground', subtitle: 'Evaluate prompts against multi-factor complexity and routing policies' };
      case 'providers':
        return { title: 'AI Providers & Circuit Breakers', subtitle: 'Fleet status, outage simulation, and automated failover control' };
      case 'performance':
        return { title: 'Model Performance & Latency', subtitle: 'P50 and P95 latency profiles, throughput metrics, and SLA adherence' };
      case 'cost-analytics':
        return { title: 'Cost Optimization Analytics', subtitle: 'Granular spending, token volume, and savings comparison vs baseline' };
      case 'logs':
        return { title: 'Request Audit Logs', subtitle: 'Comprehensive request history, token metrics, and routing rationale' };
      case 'rules':
        return { title: 'Routing Rules & Policy Config', subtitle: 'Tweak optimization weights (Cost vs Latency vs Reliability)' };
      case 'settings':
        return { title: 'Gateway Settings & Info', subtitle: 'Demo mode toggle, reference baseline, and Bifrost integration status' };
      default:
        return { title: 'Dashboard', subtitle: '' };
    }
  };

  const pageInfo = getPageInfo();

  return (
    <div className="flex min-h-screen bg-navy-950 text-slate-100 font-sans">
      {/* Sidebar */}
      <Sidebar 
        activeTab={activeTab} 
        setActiveTab={(tab) => {
          setTabParams({});
          setActiveTab(tab);
        }} 
        metrics={metrics} 
      />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-w-0">
        <Header
          title={pageInfo.title}
          subtitle={pageInfo.subtitle}
          demoMode={demoMode}
          onToggleDemo={handleToggleDemo}
          baselineModel={baselineModel}
          onRefresh={loadData}
          refreshing={refreshing}
        />

        <main className="p-8 flex-1 max-w-7xl w-full mx-auto">
          {activeTab === 'overview' && (
            <Overview
              metrics={metrics}
              costOverTime={costOverTime}
              providers={providersData}
              modelUsage={modelUsage}
              onNavigate={handleNavigate}
            />
          )}

          {activeTab === 'smart-router' && (
            <SmartRouter
              initialPrompt={tabParams.prompt || ''}
              models={providersData?.models || []}
            />
          )}

          {activeTab === 'providers' && (
            <Providers
              providersData={providersData}
              failovers={failovers}
              onRefresh={loadData}
            />
          )}

          {activeTab === 'performance' && (
            <ModelPerformance
              providerPerformances={providerPerformances}
              modelUsage={modelUsage}
            />
          )}

          {activeTab === 'cost-analytics' && (
            <CostAnalytics
              metrics={metrics}
              modelUsage={modelUsage}
              baselineModel={baselineModel}
              onUpdateBaseline={handleUpdateBaseline}
            />
          )}

          {activeTab === 'logs' && (
            <RequestLogs
              logs={logs}
              onFilterChange={(prov, stat) => {
                fetchLogs(60, prov, stat).then((res) => setLogs(res));
              }}
            />
          )}

          {activeTab === 'rules' && (
            <RoutingRules
              rules={rules}
              onRefresh={loadData}
            />
          )}

          {activeTab === 'settings' && (
            <Settings
              demoMode={demoMode}
              onToggleDemo={handleToggleDemo}
              baselineModel={baselineModel}
              onUpdateBaseline={handleUpdateBaseline}
            />
          )}
        </main>
      </div>
    </div>
  );
}
