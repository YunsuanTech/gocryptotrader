import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import Navbar from './components/navbar';
import './App.css';
import './pages/pages.css';
import './components/Navbar.css';
import './styles/table.css';

// 导入页面组件
import LoginPage from './pages/LoginPage';
import AccountsPage from './pages/accountsPage';
import WatchlistPage from './pages/watchlistPage';
import TradingRulesPage from './pages/tradingRulesPage';
const App: React.FC = () => {
  return (
    <div className="app">
      <Navbar />
      <main className="main-content">
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/accounts" element={<AccountsPage />} />
          <Route path="/watchlist" element={<WatchlistPage />} />
          <Route path="/trading-rules" element={<TradingRulesPage />} />
          <Route path="/" element={<Navigate to="/login" replace />} />
        </Routes>
      </main>
    </div>
  );
};

export default App;