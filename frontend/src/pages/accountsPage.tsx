import React, { useState, useEffect, useRef } from 'react';
import websocketResponseHandlerService from '../services/websocketResponseHandlerService';
import apiService from '../services/apiService';
import { Account } from '../types/account';
import authService from '../services/authService';
import Modal from '../components/Modal';
import '../components/Modal.css';
import './AccountsPage.css';

const AccountsPage: React.FC = () => {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [isConnected, setIsConnected] = useState<boolean>(false);
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);
  const authSent = useRef(false);

  useEffect(() => {
    console.log('AccountsPage组件挂载');

    if (!authService.isAuthenticated) {
      console.log('用户未认证，重定向到登录页面');
      window.location.href = '/login';
      return;
    }

    const subscription = websocketResponseHandlerService.shared.subscribe({
      next: (message) => {
        if (message.event === 'getaccounts') {
          if (message.error) {
            setError(`获取账户信息失败: ${message.error}`);
            setLoading(false);
          } else if (message.data) {
            const accountData = Array.isArray(message.data) ? message.data : [];
            const processedAccounts = accountData.map(account => ({
              id: account.ID?.toString() || '',
              exchange: account.exchange || '',
              assetType: account.assetType || '',
              name: account.Name || '',
              currencyName: account.CurrencyName || account.ChainName || '',
              totalValue: typeof account.totalValue === 'number' ? account.totalValue : 0,
              hold: typeof account.hold === 'number' ? account.hold : 0,
              free: typeof account.free === 'number' ? account.free : 0,
              frozen: typeof account.frozen === 'number' ? account.frozen : 0,
              lastUpdated: account.lastUpdated || account.UpdatedAt || ''
            }));
            setAccounts(processedAccounts);
            setLoading(false);
            setError(null);
          }
        }
      },
      error: (err) => {
        console.error('WebSocket错误:', err);
        setError('WebSocket连接错误');
        setLoading(false);
      },
      complete: () => {
        console.log('WebSocket连接关闭');
        setIsConnected(false);
      }
    });

    setIsConnected(websocketResponseHandlerService.isConnected);

    if (websocketResponseHandlerService.isConnected && authService.isAuthenticated && loading) {
      apiService.getAccounts();
    }

    return () => {
      console.log('AccountsPage组件卸载');
      subscription.unsubscribe();
      authSent.current = false;
    };
  }, []);

  const handleRefresh = () => {
    setLoading(true);
    if (websocketResponseHandlerService.isConnected) {
      apiService.getAccounts();
    } else {
      setError('WebSocket未连接');
      setLoading(false);
    }
  };

  const formatDateTime = (dateTimeStr: string) => {
    if (!dateTimeStr) return '-';
    try {
      const date = new Date(dateTimeStr);
      return date.toLocaleString();
    } catch (e) {
      return dateTimeStr;
    }
  };

  const formatNumber = (num: number | undefined) => {
    if (num === undefined || num === null || isNaN(Number(num))) {
      return '0.00';
    }
    return Number(num).toLocaleString(undefined, { 
      minimumFractionDigits: 2, 
      maximumFractionDigits: 8 
    });
  };

  return (
    <div className="page-container">
      <div className="page-header">
        <h1>账户信息</h1>
        <div className="page-actions">
          <button 
            className="btn btn-primary" 
            onClick={handleRefresh} 
            disabled={loading || !isConnected}
          >
            {loading ? '加载中...' : '刷新'}
          </button>
          <div className={`connection-status ${isConnected ? 'connected' : 'disconnected'}`}>
            {isConnected ? '已连接' : '未连接'}
          </div>
        </div>
      </div>

      {error && <div className="status-message error">{error}</div>}

      {loading ? (
        <div className="status-message loading">
          <div className="spinner"></div>
          <span>加载账户信息中...</span>
        </div>
      ) : accounts.length > 0 ? (
        <div className="table-container">
          <table className="data-table accounts-table">
            <thead>
              <tr>
                <th>交易所</th>
                <th>资产类型</th>
                <th>账户名称</th>
                <th>货币</th>
                <th className="numeric-header">总价值</th>
                <th className="numeric-header">可用</th>
                <th className="numeric-header">冻结</th>
                <th className="numeric-header">保留</th>
                <th>最后更新</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {accounts.map((account, index) => (
                <tr key={index} className="table-row">
                  <td>{account.exchange}</td>
                  <td>{account.assetType}</td>
                  <td>{account.name}</td>
                  <td>{account.currencyName}</td>
                  <td className="numeric-cell">{formatNumber(account.totalValue)}</td>
                  <td className="numeric-cell">{formatNumber(account.free)}</td>
                  <td className="numeric-cell">{formatNumber(account.frozen)}</td>
                  <td className="numeric-cell">{formatNumber(account.hold)}</td>
                  <td>{formatDateTime(account.lastUpdated)}</td>
                  <td>
                    <button 
                      className="btn btn-icon" 
                      onClick={() => {
                        setSelectedAccount(account);
                        setIsModalOpen(true);
                      }}
                    >
                      查看详情
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="status-message empty">没有找到账户信息</div>
      )}
      
      <Modal 
        isOpen={isModalOpen} 
        onClose={() => setIsModalOpen(false)} 
        title="账户详细信息"
      >
        {selectedAccount && (
          <div className="account-details">
            <div className="detail-row">
              <span className="detail-label">交易所:</span>
              <span className="detail-value">{selectedAccount.exchange}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">资产类型:</span>
              <span className="detail-value">{selectedAccount.assetType}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">账户名称:</span>
              <span className="detail-value">{selectedAccount.name}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">货币:</span>
              <span className="detail-value">{selectedAccount.currencyName}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">总价值:</span>
              <span className="detail-value highlight">{formatNumber(selectedAccount.totalValue)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">可用:</span>
              <span className="detail-value">{formatNumber(selectedAccount.free)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">冻结:</span>
              <span className="detail-value">{formatNumber(selectedAccount.frozen)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">保留:</span>
              <span className="detail-value">{formatNumber(selectedAccount.hold)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">最后更新:</span>
              <span className="detail-value">{formatDateTime(selectedAccount.lastUpdated)}</span>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default AccountsPage;

