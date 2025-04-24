import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import Navbar from './components/navbar';
import './App.css';

// 导入页面组件
import LoginPage from './pages/LoginPage';
import AccountsPage from './pages/accountsPage';
import TokenMonitorPage from './pages/TokenMonitorPage';
import TradingRulePage from './pages/TradingRulePage';
const App: React.FC = () => {
  return (
    <div className="app">
      <Navbar />
      <main className="main-content">
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/accounts" element={<AccountsPage />} />
          <Route path="/tokenmonitor" element={<TokenMonitorPage />} />
          <Route path="/trading-rules" element={<TradingRulePage />} />
          <Route path="/" element={<Navigate to="/login" replace />} />
        </Routes>
      </main>
    </div>
  );
};

export default App;