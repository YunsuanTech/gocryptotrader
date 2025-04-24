import React, { useState, useEffect, useRef } from 'react';
import websocketResponseHandlerService from '../services/websocketResponseHandlerService';
import tokenMonitorService from '../services/tokenMonitorService';
import { TokenMonitor } from '../types/tokenMonitor';
import authService from '../services/authService';
import Modal from '../components/Modal';
import '../components/Modal.css';
import './TokenMonitorPage.css';
import '../styles/pages.css';

const TokenMonitorPage: React.FC = () => {
  const [tokenMonitors, setTokenMonitors] = useState<TokenMonitor[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [isConnected, setIsConnected] = useState<boolean>(false);
  const [selectedTokenMonitor, setSelectedTokenMonitor] = useState<TokenMonitor | null>(null);
  const [isViewModalOpen, setIsViewModalOpen] = useState<boolean>(false);
  const [isAddModalOpen, setIsAddModalOpen] = useState<boolean>(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState<boolean>(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState<boolean>(false);
  const [formData, setFormData] = useState<Partial<TokenMonitor>>({});
  const authSent = useRef(false);

  useEffect(() => {
    console.log('TokenMonitorPage组件挂载');

    if (!authService.isAuthenticated) {
      console.log('用户未认证，重定向到登录页面');
      window.location.href = '/login';
      return;
    }

    const subscription = websocketResponseHandlerService.shared.subscribe({
      next: (message) => {
        if (message.event === 'gettokenmonitors') {
          if (message.error) {
            setError(`获取代币监控信息失败: ${message.error}`);
            setLoading(false);
          } else if (message.data) {
            const tokenMonitorData = Array.isArray(message.data) ? message.data : [];
            setTokenMonitors(tokenMonitorData);
            setLoading(false);
            setError(null);
          }
        } else if (message.event === 'addtokenmonitor') {
          if (message.error) {
            setError(`添加代币监控失败: ${message.error}`);
          } else {
            setIsAddModalOpen(false);
            tokenMonitorService.getTokenMonitors(); // 刷新列表
          }
        } else if (message.event === 'updatetokenmonitor') {
          if (message.error) {
            setError(`更新代币监控失败: ${message.error}`);
          } else {
            setIsEditModalOpen(false);
            tokenMonitorService.getTokenMonitors(); // 刷新列表
          }
        } else if (message.event === 'deletetokenmonitor') {
          if (message.error) {
            setError(`删除代币监控失败: ${message.error}`);
          } else {
            setIsDeleteModalOpen(false);
            tokenMonitorService.getTokenMonitors(); // 刷新列表
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
      tokenMonitorService.getTokenMonitors();
    }

    return () => {
      console.log('TokenMonitorPage组件卸载');
      subscription.unsubscribe();
      authSent.current = false;
    };
  }, []);

  const handleRefresh = () => {
    setLoading(true);
    if (websocketResponseHandlerService.isConnected) {
      tokenMonitorService.getTokenMonitors();
    } else {
      setError('WebSocket未连接');
      setLoading(false);
    }
  };

  const handleAddTokenMonitor = () => {
    setFormData({
      tokenAddress: '',
      tokenName: '',
      price: null,
      tokenDecimals: 18,
      buyAmount: null,
      buyPrice: null,
      buyTime: null,
      sellPercentage: null,
      totalSellPrice: null,
      lastSellTime: null,
      isMonitoring: 1
    });
    setIsAddModalOpen(true);
  };

  const handleEditTokenMonitor = (tokenMonitor: TokenMonitor) => {
    setSelectedTokenMonitor(tokenMonitor);
    setFormData(tokenMonitor);
    setIsEditModalOpen(true);
  };

  const handleDeleteTokenMonitor = (tokenMonitor: TokenMonitor) => {
    setSelectedTokenMonitor(tokenMonitor);
    setIsDeleteModalOpen(true);
  };

  const handleViewDetails = (tokenMonitor: TokenMonitor) => {
    setSelectedTokenMonitor(tokenMonitor);
    setIsViewModalOpen(true);
  };

  const handleFormChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target as HTMLInputElement;
    
    let parsedValue: string | number | null = value;
    
    // 处理数字类型的输入
    if (type === 'number') {
      parsedValue = value === '' ? null : Number(value);
    }
    
    setFormData(prev => ({
      ...prev,
      [name]: parsedValue
    }));
  };

  const handleSubmitAdd = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.tokenAddress || !formData.tokenName) {
      setError('代币地址和名称为必填项');
      return;
    }
    tokenMonitorService.addTokenMonitor(formData);
  };

  const handleSubmitEdit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.tokenAddress || !formData.tokenName) {
      setError('代币地址和名称为必填项');
      return;
    }
    tokenMonitorService.updateTokenMonitor(formData);
  };

  const handleConfirmDelete = () => {
    if (selectedTokenMonitor) {
      tokenMonitorService.deleteTokenMonitor(selectedTokenMonitor.tokenAddress);
    }
  };

  const formatDateTime = (timestamp: number | null) => {
    if (!timestamp) return '-';
    try {
      const date = new Date(timestamp * 1000); // 转换Unix时间戳为毫秒
      return date.toLocaleString();
    } catch (e) {
      return '-';
    }
  };

  const formatNumber = (num: number | null | undefined) => {
    if (num === undefined || num === null || isNaN(Number(num))) {
      return '-';
    }
    return Number(num).toLocaleString(undefined, { 
      minimumFractionDigits: 2, 
      maximumFractionDigits: 8 
    });
  };

  return (
    <div className="page-container">
      <div className="page-header">
        <h1>代币监控</h1>
        <div className="page-actions">
          <button 
            className="btn btn-success" 
            onClick={handleAddTokenMonitor}
            disabled={!isConnected}
          >
            添加代币
          </button>
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
          <span>加载代币监控信息中...</span>
        </div>
      ) : tokenMonitors.length > 0 ? (
        <div className="table-container">
          <table className="data-table token-monitor-table">
            <thead>
              <tr>
                <th>代币地址</th>
                <th>代币名称</th>
                <th className="numeric-header">价格</th>
                <th>精度</th>
                <th className="numeric-header">买入数量</th>
                <th className="numeric-header">买入价格</th>
                <th>买入时间</th>
                <th className="numeric-header">卖出百分比</th>
                <th className="numeric-header">总卖出价格</th>
                <th>最后卖出时间</th>
                <th>监控状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {tokenMonitors.map((tokenMonitor, index) => (
                <tr key={index} className="table-row">
                  <td title={tokenMonitor.tokenAddress}>
                    {tokenMonitor.tokenAddress.substring(0, 10)}...
                  </td>
                  <td>{tokenMonitor.tokenName}</td>
                  <td className="numeric-cell">{formatNumber(tokenMonitor.price)}</td>
                  <td>{tokenMonitor.tokenDecimals}</td>
                  <td className="numeric-cell">{formatNumber(tokenMonitor.buyAmount)}</td>
                  <td className="numeric-cell">{formatNumber(tokenMonitor.buyPrice)}</td>
                  <td>{formatDateTime(tokenMonitor.buyTime)}</td>
                  <td className="numeric-cell">{tokenMonitor.sellPercentage ? `${tokenMonitor.sellPercentage}%` : '-'}</td>
                  <td className="numeric-cell">{formatNumber(tokenMonitor.totalSellPrice)}</td>
                  <td>{formatDateTime(tokenMonitor.lastSellTime)}</td>
                  <td>{tokenMonitor.isMonitoring === 1 ? '监控中' : '已停止'}</td>
                  <td>
                    <div className="action-buttons">
                      <button 
                        className="btn btn-icon" 
                        onClick={() => handleViewDetails(tokenMonitor)}
                      >
                        查看
                      </button>
                      <button 
                        className="btn btn-icon edit" 
                        onClick={() => handleEditTokenMonitor(tokenMonitor)}
                      >
                        编辑
                      </button>
                      <button 
                        className="btn btn-icon delete" 
                        onClick={() => handleDeleteTokenMonitor(tokenMonitor)}
                      >
                        删除
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="status-message empty">没有找到代币监控信息</div>
      )}
      
      {/* 查看详情模态框 */}
      <Modal 
        isOpen={isViewModalOpen} 
        onClose={() => setIsViewModalOpen(false)} 
        title="代币详细信息"
      >
        {selectedTokenMonitor && (
          <div className="token-details">
            <div className="detail-row">
              <span className="detail-label">代币地址:</span>
              <span className="detail-value">{selectedTokenMonitor.tokenAddress}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">代币名称:</span>
              <span className="detail-value">{selectedTokenMonitor.tokenName}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">当前价格:</span>
              <span className="detail-value highlight">{formatNumber(selectedTokenMonitor.price)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">代币精度:</span>
              <span className="detail-value">{selectedTokenMonitor.tokenDecimals}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">买入数量:</span>
              <span className="detail-value">{formatNumber(selectedTokenMonitor.buyAmount)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">买入价格:</span>
              <span className="detail-value">{formatNumber(selectedTokenMonitor.buyPrice)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">买入时间:</span>
              <span className="detail-value">{formatDateTime(selectedTokenMonitor.buyTime)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">卖出百分比:</span>
              <span className="detail-value">{selectedTokenMonitor.sellPercentage ? `${selectedTokenMonitor.sellPercentage}%` : '-'}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">总卖出价格:</span>
              <span className="detail-value">{formatNumber(selectedTokenMonitor.totalSellPrice)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">最后卖出时间:</span>
              <span className="detail-value">{formatDateTime(selectedTokenMonitor.lastSellTime)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">监控状态:</span>
              <span className="detail-value">{selectedTokenMonitor.isMonitoring === 1 ? '监控中' : '已停止'}</span>
            </div>
          </div>
        )}
      </Modal>

      {/* 添加代币模态框 */}
      <Modal 
        isOpen={isAddModalOpen} 
        onClose={() => setIsAddModalOpen(false)} 
        title="添加代币监控"
      >
        <form onSubmit={handleSubmitAdd} className="token-form">
          <div className="form-group">
            <label htmlFor="tokenAddress">代币地址 *</label>
            <input
              type="text"
              id="tokenAddress"
              name="tokenAddress"
              value={formData.tokenAddress || ''}
              onChange={handleFormChange}
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="tokenName">代币名称 *</label>
            <input
              type="text"
              id="tokenName"
              name="tokenName"
              value={formData.tokenName || ''}
              onChange={handleFormChange}
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="price">当前价格</label>
            <input
              type="number"
              id="price"
              name="price"
              value={formData.price === null ? '' : formData.price}
              onChange={handleFormChange}
              step="any"
            />
          </div>
          <div className="form-group">
            <label htmlFor="tokenDecimals">代币精度 *</label>
            <input
              type="number"
              id="tokenDecimals"
              name="tokenDecimals"
              value={formData.tokenDecimals || 18}
              onChange={handleFormChange}
              required
              min="0"
            />
          </div>
          <div className="form-group">
            <label htmlFor="buyAmount">买入数量</label>
            <input
              type="number"
              id="buyAmount"
              name="buyAmount"
              value={formData.buyAmount === null ? '' : formData.buyAmount}
              onChange={handleFormChange}
              step="any"
            />
          </div>
          <div className="form-group">
            <label htmlFor="buyPrice">买入价格</label>
            <input
              type="number"
              id="buyPrice"
              name="buyPrice"
              value={formData.buyPrice === null ? '' : formData.buyPrice}
              onChange={handleFormChange}
              step="any"
            />
          </div>
          <div className="form-group">
            <label htmlFor="sellPercentage">卖出百分比</label>
            <input
              type="number"
              id="sellPercentage"
              name="sellPercentage"
              value={formData.sellPercentage === null ? '' : formData.sellPercentage}
              onChange={handleFormChange}
              step="any"
            />
          </div>
          <div className="form-group">
            <label htmlFor="isMonitoring">监控状态</label>
            <select
              id="isMonitoring"
              name="isMonitoring"
              value={formData.isMonitoring === undefined ? 1 : formData.isMonitoring}
              onChange={handleFormChange}
            >
              <option value={1}>监控中</option>
              <option value={0}>已停止</option>
            </select>
          </div>
          <div className="form-actions">
            <button type="submit" className="btn btn-primary">添加</button>
            <button type="button" className="btn btn-secondary" onClick={() => setIsAddModalOpen(false)}>取消</button>
          </div>
        </form>
      </Modal>

      {/* 编辑代币模态框 */}
      <Modal 
        isOpen={isEditModalOpen} 
        onClose={() => setIsEditModalOpen(false)} 
        title="编辑代币监控"
      >
        <form onSubmit={handleSubmitEdit} className="token-form">
          <div className="form-group">
            <label htmlFor="edit-tokenAddress">代币地址 *</label>
            <input
              type="text"
              id="edit-tokenAddress"
              name="tokenAddress"
              value={formData.tokenAddress || ''}
              onChange={handleFormChange}
              required
              readOnly // 主键不可修改
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-tokenName">代币名称 *</label>
            <input
              type="text"
              id="edit-tokenName"
              name="tokenName"
              value={formData.tokenName || ''}
              onChange={handleFormChange}
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-price">当前价格</label>
            <input
              type="number"
              id="edit-price"
              name="price"
              value={formData.price === null ? '' : formData.price}
              onChange={handleFormChange}
              step="any"
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-tokenDecimals">代币精度 *</label>
            <input
              type="number"
              id="edit-tokenDecimals"
              name="tokenDecimals"
              value={formData.tokenDecimals || 18}
              onChange={handleFormChange}
              required
              min="0"
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-buyAmount">买入数量</label>
            <input
              type="number"
              id="edit-buyAmount"
              name="buyAmount"
              value={formData.buyAmount === null ? '' : formData.buyAmount}
              onChange={handleFormChange}
              step="any"
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-buyPrice">买入价格</label>
            <input
              type="number"
              id="edit-buyPrice"
              name="buyPrice"
              value={formData.buyPrice === null ? '' : formData.buyPrice}
              onChange={handleFormChange}
              step="any"
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-sellPercentage">卖出百分比</label>
            <input
              type="number"
              id="edit-sellPercentage"
              name="sellPercentage"
              value={formData.sellPercentage === null ? '' : formData.sellPercentage}
              onChange={handleFormChange}
              step="any"
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-isMonitoring">监控状态</label>
            <select
              id="edit-isMonitoring"
              name="isMonitoring"
              value={formData.isMonitoring === undefined ? 1 : formData.isMonitoring}
              onChange={handleFormChange}
            >
              <option value={1}>监控中</option>
              <option value={0}>已停止</option>
            </select>
          </div>
          <div className="form-actions">
            <button type="submit" className="btn btn-primary">保存</button>
            <button type="button" className="btn btn-secondary" onClick={() => setIsEditModalOpen(false)}>取消</button>
          </div>
        </form>
      </Modal>

      {/* 删除确认模态框 */}
      <Modal 
        isOpen={isDeleteModalOpen} 
        onClose={() => setIsDeleteModalOpen(false)} 
        title="确认删除"
      >
        {selectedTokenMonitor && (
          <div className="delete-confirmation">
            <p>确定要删除以下代币监控吗？此操作不可撤销。</p>
            <div className="token-info">
              <p><strong>代币地址:</strong> {selectedTokenMonitor.tokenAddress}</p>
              <p><strong>代币名称:</strong> {selectedTokenMonitor.tokenName}</p>
            </div>
            <div className="form-actions">
              <button 
                className="btn btn-danger" 
                onClick={handleConfirmDelete}
              >
                确认删除
              </button>
              <button 
                className="btn btn-secondary" 
                onClick={() => setIsDeleteModalOpen(false)}
              >
                取消
              </button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default TokenMonitorPage;