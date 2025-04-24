import React, { useState, useEffect, useRef } from 'react';
import websocketResponseHandlerService from '../services/websocketResponseHandlerService';
import tradingRuleService from '../services/tradingRuleService';
import { TradingRule } from '../types/tradingRule';
import authService from '../services/authService';
import Modal from '../components/Modal';
import '../components/Modal.css';
import './TradingRulePage.css';
import '../styles/pages.css';

const TradingRulePage: React.FC = () => {
  const [tradingRules, setTradingRules] = useState<TradingRule[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [isConnected, setIsConnected] = useState<boolean>(false);
  const [selectedTradingRule, setSelectedTradingRule] = useState<TradingRule | null>(null);
  const [isViewModalOpen, setIsViewModalOpen] = useState<boolean>(false);
  const [isAddModalOpen, setIsAddModalOpen] = useState<boolean>(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState<boolean>(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState<boolean>(false);
  const [formData, setFormData] = useState<Partial<TradingRule>>({});
  const authSent = useRef(false);

  useEffect(() => {
    console.log('TradingRulePage组件挂载');

    if (!authService.isAuthenticated) {
      console.log('用户未认证，重定向到登录页面');
      window.location.href = '/login';
      return;
    }

    const subscription = websocketResponseHandlerService.shared.subscribe({
      next: (message) => {
        if (message.event === 'gettradingrules') {
          if (message.error) {
            setError(`获取交易规则信息失败: ${message.error}`);
            setLoading(false);
          } else if (message.data) {
            const tradingRuleData = Array.isArray(message.data) ? message.data : [];
            setTradingRules(tradingRuleData);
            setLoading(false);
            setError(null);
          }
        } else if (message.event === 'addtradingrule') {
          if (message.error) {
            setError(`添加交易规则失败: ${message.error}`);
          } else {
            setIsAddModalOpen(false);
            tradingRuleService.getTradingRules(); // 刷新列表
          }
        } else if (message.event === 'updatetradingrule') {
          if (message.error) {
            setError(`更新交易规则失败: ${message.error}`);
          } else {
            setIsEditModalOpen(false);
            tradingRuleService.getTradingRules(); // 刷新列表
          }
        } else if (message.event === 'deletetradingrule') {
          if (message.error) {
            setError(`删除交易规则失败: ${message.error}`);
          } else {
            setIsDeleteModalOpen(false);
            tradingRuleService.getTradingRules(); // 刷新列表
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
      tradingRuleService.getTradingRules();
    }

    return () => {
      console.log('TradingRulePage组件卸载');
      subscription.unsubscribe();
      authSent.current = false;
    };
  }, []);

  const handleRefresh = () => {
    setLoading(true);
    if (websocketResponseHandlerService.isConnected) {
      tradingRuleService.getTradingRules();
    } else {
      setError('WebSocket未连接');
      setLoading(false);
    }
  };

  const handleAddTradingRule = () => {
    setFormData({
      ruleName: '',
      ruleType: '',
      buyPrice: 0,
      sellPrice: 0,
      condition: '',
      priority: 0,
      description: ''
    });
    setIsAddModalOpen(true);
  };

  const handleEditTradingRule = (tradingRule: TradingRule) => {
    setSelectedTradingRule(tradingRule);
    setFormData(tradingRule);
    setIsEditModalOpen(true);
  };

  const handleDeleteTradingRule = (tradingRule: TradingRule) => {
    setSelectedTradingRule(tradingRule);
    setIsDeleteModalOpen(true);
  };

  const handleViewDetails = (tradingRule: TradingRule) => {
    setSelectedTradingRule(tradingRule);
    setIsViewModalOpen(true);
  };

  const handleFormChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value, type } = e.target as HTMLInputElement;
    
    let parsedValue: string | number = value;
    
    // 处理数字类型的输入
    if (type === 'number') {
      parsedValue = value === '' ? 0 : Number(value);
    }
    
    setFormData(prev => ({
      ...prev,
      [name]: parsedValue
    }));
  };

  const handleSubmitAdd = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.ruleName || !formData.ruleType) {
      setError('规则名称和类型为必填项');
      return;
    }
    tradingRuleService.addTradingRule(formData);
  };

  const handleSubmitEdit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.ruleName || !formData.ruleType) {
      setError('规则名称和类型为必填项');
      return;
    }
    tradingRuleService.updateTradingRule(formData);
  };

  const handleConfirmDelete = () => {
    if (selectedTradingRule) {
      tradingRuleService.deleteTradingRule(selectedTradingRule.id);
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
        <h1>交易规则</h1>
        <div className="page-actions">
          <button 
            className="btn btn-success" 
            onClick={handleAddTradingRule}
            disabled={!isConnected}
          >
            添加规则
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
          <span>加载交易规则信息中...</span>
        </div>
      ) : tradingRules.length > 0 ? (
        <div className="table-container">
          <table className="data-table trading-rule-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>规则名称</th>
                <th>规则类型</th>
                <th className="numeric-header">买入价格</th>
                <th className="numeric-header">卖出价格</th>
                <th>条件</th>
                <th className="numeric-header">优先级</th>
                <th>描述</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {tradingRules.map((tradingRule) => (
                <tr key={tradingRule.id} className="table-row">
                  <td>{tradingRule.id}</td>
                  <td>{tradingRule.ruleName}</td>
                  <td>{tradingRule.ruleType}</td>
                  <td className="numeric-cell">{formatNumber(tradingRule.buyPrice)}</td>
                  <td className="numeric-cell">{formatNumber(tradingRule.sellPrice)}</td>
                  <td>{tradingRule.condition}</td>
                  <td className="numeric-cell">{formatNumber(tradingRule.priority)}</td>
                  <td title={tradingRule.description}>
                    {tradingRule.description.length > 20 ? `${tradingRule.description.substring(0, 20)}...` : tradingRule.description}
                  </td>
                  <td>
                    <div className="action-buttons">
                      <button 
                        className="btn btn-icon" 
                        onClick={() => handleViewDetails(tradingRule)}
                      >
                        查看
                      </button>
                      <button 
                        className="btn btn-icon edit" 
                        onClick={() => handleEditTradingRule(tradingRule)}
                      >
                        编辑
                      </button>
                      <button 
                        className="btn btn-icon delete" 
                        onClick={() => handleDeleteTradingRule(tradingRule)}
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
        <div className="status-message empty">没有找到交易规则信息</div>
      )}
      
      {/* 查看详情模态框 */}
      <Modal 
        isOpen={isViewModalOpen} 
        onClose={() => setIsViewModalOpen(false)} 
        title="交易规则详细信息"
      >
        {selectedTradingRule && (
          <div className="trading-rule-details">
            <div className="detail-row">
              <span className="detail-label">ID:</span>
              <span className="detail-value">{selectedTradingRule.id}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">规则名称:</span>
              <span className="detail-value">{selectedTradingRule.ruleName}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">规则类型:</span>
              <span className="detail-value">{selectedTradingRule.ruleType}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">买入价格:</span>
              <span className="detail-value highlight">{formatNumber(selectedTradingRule.buyPrice)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">卖出价格:</span>
              <span className="detail-value highlight">{formatNumber(selectedTradingRule.sellPrice)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">条件:</span>
              <span className="detail-value">{selectedTradingRule.condition}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">优先级:</span>
              <span className="detail-value">{formatNumber(selectedTradingRule.priority)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">描述:</span>
              <span className="detail-value">{selectedTradingRule.description}</span>
            </div>
          </div>
        )}
      </Modal>

      {/* 添加交易规则模态框 */}
      <Modal 
        isOpen={isAddModalOpen} 
        onClose={() => setIsAddModalOpen(false)} 
        title="添加交易规则"
      >
        <form onSubmit={handleSubmitAdd} className="trading-rule-form">
          <div className="form-group">
            <label htmlFor="ruleName">规则名称 *</label>
            <input
              type="text"
              id="ruleName"
              name="ruleName"
              value={formData.ruleName || ''}
              onChange={handleFormChange}
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="ruleType">规则类型 *</label>
            <input
              type="text"
              id="ruleType"
              name="ruleType"
              value={formData.ruleType || ''}
              onChange={handleFormChange}
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="buyPrice">买入价格</label>
            <input
              type="number"
              id="buyPrice"
              name="buyPrice"
              value={formData.buyPrice === undefined ? 0 : formData.buyPrice}
              onChange={handleFormChange}
              step="any"
            />
          </div>
          <div className="form-group">
            <label htmlFor="sellPrice">卖出价格</label>
            <input
              type="number"
              id="sellPrice"
              name="sellPrice"
              value={formData.sellPrice === undefined ? 0 : formData.sellPrice}
              onChange={handleFormChange}
              step="any"
            />
          </div>
          <div className="form-group">
            <label htmlFor="condition">条件</label>
            <input
              type="text"
              id="condition"
              name="condition"
              value={formData.condition || ''}
              onChange={handleFormChange}
            />
          </div>
          <div className="form-group">
            <label htmlFor="priority">优先级</label>
            <input
              type="number"
              id="priority"
              name="priority"
              value={formData.priority === undefined ? 0 : formData.priority}
              onChange={handleFormChange}
              step="any"
            />
          </div>
          <div className="form-group">
            <label htmlFor="description">描述</label>
            <textarea
              id="description"
              name="description"
              value={formData.description || ''}
              onChange={handleFormChange}
              rows={3}
            />
          </div>
          <div className="form-actions">
            <button type="submit" className="btn btn-primary">添加</button>
            <button type="button" className="btn btn-secondary" onClick={() => setIsAddModalOpen(false)}>取消</button>
          </div>
        </form>
      </Modal>

      {/* 编辑交易规则模态框 */}
      <Modal 
        isOpen={isEditModalOpen} 
        onClose={() => setIsEditModalOpen(false)} 
        title="编辑交易规则"
      >
        <form onSubmit={handleSubmitEdit} className="trading-rule-form">
          <div className="form-group">
            <label htmlFor="edit-id">ID</label>
            <input
              type="number"
              id="edit-id"
              name="id"
              value={formData.id || 0}
              readOnly
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-ruleName">规则名称 *</label>
            <input
              type="text"
              id="edit-ruleName"
              name="ruleName"
              value={formData.ruleName || ''}
              onChange={handleFormChange}
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-ruleType">规则类型 *</label>
            <input
              type="text"
              id="edit-ruleType"
              name="ruleType"
              value={formData.ruleType || ''}
              onChange={handleFormChange}
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-buyPrice">买入价格</label>
            <input
              type="number"
              id="edit-buyPrice"
              name="buyPrice"
              value={formData.buyPrice === undefined ? 0 : formData.buyPrice}
              onChange={handleFormChange}
              step="any"
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-sellPrice">卖出价格</label>
            <input
              type="number"
              id="edit-sellPrice"
              name="sellPrice"
              value={formData.sellPrice === undefined ? 0 : formData.sellPrice}
              onChange={handleFormChange}
              step="any"
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-condition">条件</label>
            <input
              type="text"
              id="edit-condition"
              name="condition"
              value={formData.condition || ''}
              onChange={handleFormChange}
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-priority">优先级</label>
            <input
              type="number"
              id="edit-priority"
              name="priority"
              value={formData.priority === undefined ? 0 : formData.priority}
              onChange={handleFormChange}
              step="any"
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-description">描述</label>
            <textarea
              id="edit-description"
              name="description"
              value={formData.description || ''}
              onChange={handleFormChange}
              rows={3}
            />
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
        <div className="confirm-dialog">
          <p>确定要删除这条交易规则吗？此操作不可撤销。</p>
          <div className="form-actions">
            <button 
              className="btn btn-danger" 
              onClick={handleConfirmDelete}
            >
              删除
            </button>
            <button 
              className="btn btn-secondary" 
              onClick={() => setIsDeleteModalOpen(false)}
            >
              取消
            </button>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default TradingRulePage;