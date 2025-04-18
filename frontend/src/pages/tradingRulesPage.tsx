import React, { useState, useEffect, useRef } from 'react';
import websocketResponseHandlerService from '../services/websocketResponseHandlerService';
import { TradingRule, AddTradingRuleRequest, UpdateTradingRuleRequest } from '../types/trading';
import { WatchlistToken } from '../types/watchlist';
import authService from '../services/authService';
import apiService from '../services/apiService';
import Modal from '../components/Modal';
import '../components/Modal.css';
import './pages.css';
import './tradingrulespage.css';

const TradingRulesPage: React.FC = () => {
  const [rules, setRules] = useState<TradingRule[]>([]);
  const [tokens, setTokens] = useState<WatchlistToken[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [isConnected, setIsConnected] = useState<boolean>(false);
  const [selectedRule, setSelectedRule] = useState<TradingRule | null>(null);
  const [isAddModalOpen, setIsAddModalOpen] = useState<boolean>(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState<boolean>(false);
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState<boolean>(false);
  const [addRuleSuccess, setAddRuleSuccess] = useState<boolean | null>(null);
  const [addRuleMessage, setAddRuleMessage] = useState<string>('');
  const [editRuleSuccess, setEditRuleSuccess] = useState<boolean | null>(null);
  const [editRuleMessage, setEditRuleMessage] = useState<string>('');
  const [deleteRuleSuccess, setDeleteRuleSuccess] = useState<boolean | null>(null);
  const [deleteRuleMessage, setDeleteRuleMessage] = useState<string>('');
  const [tokenFilter, setTokenFilter] = useState<number | ''>('');
  const [userAddressFilter, setUserAddressFilter] = useState<string>('');
  const [ruleToEdit, setRuleToEdit] = useState<TradingRule | null>(null);
  const [ruleToDelete, setRuleToDelete] = useState<TradingRule | null>(null);
  const authSent = useRef(false);

  // 新增交易规则表单状态
  const [newRule, setNewRule] = useState<AddTradingRuleRequest>({
    tokenId: 0,
    userAddress: '',
    direction: 'BUY',
    triggerPrice: 0,
    quantity: 0,
    slippage: 1.0,
    expirationTime: undefined,
    isEnabled: 1,
    orderType: 'MARKET'
  });

  // 编辑交易规则表单状态
  const [editRule, setEditRule] = useState<UpdateTradingRuleRequest>({
    tokenId: 0,
    userAddress: '',
    direction: 'BUY',
    triggerPrice: 0,
    quantity: 0,
    slippage: 1.0,
    expirationTime: undefined,
    isEnabled: 1,
    orderType: 'MARKET'
  });

  useEffect(() => {
    console.log('TradingRulesPage组件挂载');
    
    // 检查认证状态
    if (!authService.isAuthenticated) {
      console.log('用户未认证，重定向到登录页面');
      window.location.href = '/login';
      return;
    }
    
    const subscription = websocketResponseHandlerService.shared.subscribe({
      next: (message) => {
        if (message.event === 'gettradingrules') {
          if (message.error) {
            setError(`获取交易规则列表失败: ${message.error}`);
            setLoading(false);
          } else if (message.data) {
            const rulesData = Array.isArray(message.data) ? message.data : [];
            setRules(rulesData);
            setLoading(false);
            setError(null);
          }
        } else if (message.event === 'getwatchlisttokens') {
          if (message.error) {
            console.error(`获取代币列表失败: ${message.error}`);
          } else if (message.data) {
            const tokenData = Array.isArray(message.data) ? message.data : [];
            setTokens(tokenData);
          }
        } else if (message.event === 'addtradingrule') {
          if (message.error) {
            setAddRuleSuccess(false);
            setAddRuleMessage(`添加交易规则失败: ${message.error}`);
          } else {
            setAddRuleSuccess(true);
            setAddRuleMessage('交易规则添加成功');
            // 重新加载规则列表
            handleRefresh();
            // 重置表单
            setNewRule({
              tokenId: 0,
              userAddress: '',
              direction: 'BUY',
              triggerPrice: 0,
              quantity: 0,
              slippage: 1.0,
              expirationTime: undefined,
              isEnabled: 1,
              orderType: 'MARKET'
            });
            // 关闭添加模态框
            setTimeout(() => {
              setIsAddModalOpen(false);
              setAddRuleSuccess(null);
              setAddRuleMessage('');
            }, 2000);
          }
        } else if (message.event === 'updatetradingrule') {
          if (message.error) {
            setEditRuleSuccess(false);
            setEditRuleMessage(`更新交易规则失败: ${message.error}`);
          } else {
            setEditRuleSuccess(true);
            setEditRuleMessage('交易规则更新成功');
            // 重新加载规则列表
            handleRefresh();
            // 关闭编辑模态框
            setTimeout(() => {
              setIsEditModalOpen(false);
              setEditRuleSuccess(null);
              setEditRuleMessage('');
              setRuleToEdit(null);
            }, 2000);
          }
        } else if (message.event === 'deletetradingrule') {
          if (message.error) {
            setDeleteRuleSuccess(false);
            setDeleteRuleMessage(`删除交易规则失败: ${message.error}`);
          } else {
            setDeleteRuleSuccess(true);
            setDeleteRuleMessage('交易规则删除成功');
            // 重新加载规则列表
            handleRefresh();
            // 关闭删除确认框
            setTimeout(() => {
              setIsDeleteConfirmOpen(false);
              setDeleteRuleSuccess(null);
              setDeleteRuleMessage('');
              setRuleToDelete(null);
            }, 2000);
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

    // 初始化连接状态
    setIsConnected(websocketResponseHandlerService.isConnected);
    
    // 如果已连接且已认证，则请求交易规则数据和代币数据
    if (websocketResponseHandlerService.isConnected && authService.isAuthenticated && loading) {
      apiService.getTradingRules();
      apiService.getWatchlistTokens();
    }

    return () => {
      console.log('TradingRulesPage组件卸载');
      subscription.unsubscribe();
      authSent.current = false;
    };
  }, []);

  const handleRefresh = () => {
    setLoading(true);
    if (websocketResponseHandlerService.isConnected) {
      apiService.getTradingRules();
    } else {
      setError('WebSocket未连接');
      setLoading(false);
    }
  };

  const handleTokenFilterChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    setTokenFilter(value === '' ? '' : parseInt(value));
    setLoading(true);
    if (value === '') {
      apiService.getTradingRules();
    } else {
      apiService.getTradingRulesByTokenID(parseInt(value));
    }
  };

  const handleUserAddressFilterChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setUserAddressFilter(value);
    if (value.trim() === '') {
      apiService.getTradingRules();
    } else {
      apiService.getTradingRulesByUserAddress(value);
    }
  };

  const handleAddRule = () => {
    // 验证表单
    if ( !newRule.userAddress || !newRule.triggerPrice || !newRule.quantity) {
      setAddRuleSuccess(false);
      setAddRuleMessage('请填写所有必填字段');
      return;
    }

    // 发送添加交易规则请求
    apiService.addTradingRule(newRule);
  };

  const handleEditRule = () => {
    // 验证表单
    if ( !editRule.userAddress || !editRule.triggerPrice || !editRule.quantity) {
      setEditRuleSuccess(false);
      setEditRuleMessage('请填写所有必填字段');
      return;
    }

    // 发送更新交易规则请求
    console.log(ruleToEdit)
    if (ruleToEdit?.RuleID !== undefined) {
      apiService.updateTradingRule(ruleToEdit.RuleID, editRule);
    } else {
      setEditRuleSuccess(false);
      setEditRuleMessage('无法更新交易规则：缺少ID');
    }
  };

  const handleDeleteRule = () => {
    if (ruleToDelete?.RuleID) {
      apiService.deleteTradingRule(ruleToDelete.RuleID);
    } else {
      setDeleteRuleSuccess(false);
      setDeleteRuleMessage('无法删除交易规则：缺少ID');
    }
  };

  const openEditModal = (rule: TradingRule) => {
    setRuleToEdit(rule);
    setEditRule({
      tokenId: rule.TokenID,
      userAddress: rule.UserAddress,
      direction: rule.Direction,
      triggerPrice: rule.TriggerPrice,
      quantity: rule.Quantity,
      slippage: rule.Slippage,
      expirationTime: rule.ExpirationTime,
      isEnabled: rule.IsEnabled,
      orderType: rule.OrderType
    });
    setIsEditModalOpen(true);
  };

  const openDeleteConfirm = (rule: TradingRule) => {
    setRuleToDelete(rule);
    setIsDeleteConfirmOpen(true);
  };

  const getTokenName = (tokenId: number) => {
    const token = tokens.find(t => t.ID === tokenId);
    return token ? token.TokenSymbol : `未知代币(ID:${tokenId})`;
  };

  const formatDate = (timestamp: number | undefined) => {
    if (!timestamp) return '未设置';
    return new Date(timestamp * 1000).toLocaleString();
  };

  return (
    <div className="page-container trading-rules-page">
      <div className="page-header">
        <h1>交易规则管理</h1>
        <div className="connection-status-wrapper">
          <div className={`connection-status ${isConnected ? 'connected' : 'disconnected'}`}>
            {isConnected ? '已连接' : '未连接'}
          </div>
        </div>
      </div>
      
      <div className="controls-container">
        <div className="filter-container">
          <div className="filter-item">
            <label>按代币筛选：</label>
            <select value={tokenFilter} onChange={handleTokenFilterChange}>
              <option value="">全部代币</option>
              {tokens.map(token => (
                <option key={token.ID} value={token.ID}>{token.TokenSymbol}</option>
              ))}
            </select>
          </div>
          
          <div className="filter-item">
            <label>按用户地址筛选：</label>
            <input 
              type="text" 
              value={userAddressFilter} 
              onChange={handleUserAddressFilterChange} 
              placeholder="输入用户地址"
            />
          </div>
        </div>
        
        <div className="action-buttons">
          <button className="btn refresh-button" onClick={handleRefresh} disabled={loading}>
            {loading ? '加载中...' : '刷新'}
          </button>
          <button className="btn add-button" onClick={() => setIsAddModalOpen(true)}>
            <span>+</span> 添加交易规则
          </button>
        </div>
      </div>
      
      {error && <div className="error-message">{error}</div>}
      
      {loading ? (
        <div className="loading-indicator">
          <div className="spinner"></div>
          <span>加载交易规则中...</span>
        </div>
      ) : (
        <div className="table-container">
          <table className="data-table">
            <thead>
              <tr>
                <th>代币</th>
                <th>用户地址</th>
                <th>方向</th>
                <th>触发价格</th>
                <th>数量</th>
                <th>滑点</th>
                <th>订单类型</th>
                <th>过期时间</th>
                <th>状态</th>
                <th>创建时间</th>
                <th>上次触发</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {rules.length === 0 ? (
                <tr>
                  <td colSpan={13} className="no-data">暂无交易规则</td>
                </tr>
              ) : (
                rules.map(rule => (
                  <tr>
                  
                    <td>{getTokenName(rule.TokenID)}</td>
                    <td className="address-cell">{rule.UserAddress}</td>
                    <td className={rule.Direction === 'BUY' ? 'buy-direction' : 'sell-direction'}>
                      {rule.Direction}
                    </td>
                    <td>{rule.TriggerPrice}</td>
                    <td>{rule.Quantity}</td>
                    <td>{rule.Slippage}%</td>
                    <td>{rule.OrderType}</td>
                    <td>{formatDate(rule.ExpirationTime)}</td>
                    <td className={rule.IsEnabled ? 'status-active' : 'status-inactive'}>
                      {rule.IsEnabled ? '启用' : '禁用'}
                    </td>
                    <td>{formatDate(rule.CreatedAt)}</td>
                    <td>{formatDate(rule.LastTriggered)}</td>
                    <td className="action-cell">
                      <button className="btn edit-button" onClick={() => openEditModal(rule)}>编辑</button>
                      <button className="btn delete-button" onClick={() => openDeleteConfirm(rule)}>删除</button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}
      
      {/* 添加交易规则模态框 */}
      <Modal isOpen={isAddModalOpen} onClose={() => setIsAddModalOpen(false)} title="添加交易规则">
        <div className="form-container">
          {addRuleSuccess === true && (
            <div className="success-message">{addRuleMessage}</div>
          )}
          {addRuleSuccess === false && (
            <div className="error-message">{addRuleMessage}</div>
          )}
          
          <div className="form-group">
            <label>代币：</label>
            <select 
              value={newRule.tokenId === 0 ? '0' : (newRule.tokenId || '')} 
              onChange={(e) => setNewRule({...newRule, tokenId: parseInt(e.target.value)})}
              required
            >
              <option value="">选择代币</option>
              {tokens.map(token => (
                <option key={token.ID} value={token.ID}>{token.TokenSymbol}</option>
              ))}
            </select>
          </div>
          
          <div className="form-group">
            <label>用户地址：</label>
            <input 
              type="text" 
              value={newRule.userAddress} 
              onChange={(e) => setNewRule({...newRule, userAddress: e.target.value})}
              placeholder="输入用户钱包地址"
              required
            />
          </div>
          
          <div className="form-row">
            <div className="form-group form-group-half">
              <label>交易方向：</label>
              <select 
                value={newRule.direction} 
                onChange={(e) => setNewRule({...newRule, direction: e.target.value})}
              >
                <option value="BUY">买入</option>
                <option value="SELL">卖出</option>
              </select>
            </div>
            
            <div className="form-group form-group-half">
              <label>订单类型：</label>
              <select 
                value={newRule.orderType || 'MARKET'} 
                onChange={(e) => setNewRule({...newRule, orderType: e.target.value})}
              >
                <option value="MARKET">市价单</option>
                <option value="LIMIT">限价单</option>
              </select>
            </div>
          </div>
          
          <div className="form-row">
            <div className="form-group form-group-half">
              <label>触发价格：</label>
              <input 
                type="number" 
                value={newRule.triggerPrice || ''} 
                onChange={(e) => setNewRule({...newRule, triggerPrice: parseFloat(e.target.value)})}
                step="0.000001"
                min="0"
                required
              />
            </div>
            
            <div className="form-group form-group-half">
              <label>数量：</label>
              <input 
                type="number" 
                value={newRule.quantity || ''} 
                onChange={(e) => setNewRule({...newRule, quantity: parseFloat(e.target.value)})}
                step="0.000001"
                min="0"
                required
              />
            </div>
          </div>
          
          <div className="form-row">
            <div className="form-group form-group-half">
              <label>滑点(%)：</label>
              <input 
                type="number" 
                value={newRule.slippage || ''} 
                onChange={(e) => setNewRule({...newRule, slippage: parseFloat(e.target.value)})}
                step="0.1"
                min="0"
                max="100"
              />
            </div>
            
            <div className="form-group form-group-half">
              <label>状态：</label>
              <select 
                value={newRule.isEnabled} 
                onChange={(e) => setNewRule({...newRule, isEnabled: parseInt(e.target.value)})}
              >
                <option value="1">启用</option>
                <option value="0">禁用</option>
              </select>
            </div>
          </div>
          
          <div className="form-group">
            <label>过期时间：</label>
            <input 
              type="datetime-local" 
              value={newRule.expirationTime ? new Date(newRule.expirationTime * 1000).toISOString().slice(0, 16) : ''} 
              onChange={(e) => {
                const date = e.target.value ? new Date(e.target.value).getTime() / 1000 : undefined;
                setNewRule({...newRule, expirationTime: date});
              }}
            />
          </div>
          
          <div className="form-actions">
            <button className="btn cancel-button" onClick={() => setIsAddModalOpen(false)}>取消</button>
            <button className="btn submit-button" onClick={handleAddRule}>添加</button>
          </div>
        </div>
      </Modal>
      
      {/* 编辑交易规则模态框 */}
      <Modal isOpen={isEditModalOpen} onClose={() => setIsEditModalOpen(false)} title="编辑交易规则">
        <div className="form-container">
          {editRuleSuccess === true && (
            <div className="success-message">{editRuleMessage}</div>
          )}
          {editRuleSuccess === false && (
            <div className="error-message">{editRuleMessage}</div>
          )}
          
          <div className="form-group">
            <label>代币：</label>
            <select 
              value={editRule.tokenId === 0 ? '0' : (editRule.tokenId || '')} 
              onChange={(e) => setEditRule({...editRule, tokenId: parseInt(e.target.value)})}
              required
            >
              <option value="">选择代币</option>
              {tokens.map(token => (
                <option key={token.ID} value={token.ID}>{token.TokenSymbol}</option>
              ))}
            </select>
          </div>
          
          <div className="form-group">
            <label>用户地址：</label>
            <input 
              type="text" 
              value={editRule.userAddress} 
              onChange={(e) => setEditRule({...editRule, userAddress: e.target.value})}
              placeholder="输入用户钱包地址"
              required
            />
          </div>
          
          <div className="form-row">
            <div className="form-group form-group-half">
              <label>交易方向：</label>
              <select 
                value={editRule.direction} 
                onChange={(e) => setEditRule({...editRule, direction: e.target.value})}
              >
                <option value="BUY">买入</option>
                <option value="SELL">卖出</option>
              </select>
            </div>
            
            <div className="form-group form-group-half">
              <label>订单类型：</label>
              <select 
                value={editRule.orderType || 'MARKET'} 
                onChange={(e) => setEditRule({...editRule, orderType: e.target.value})}
              >
                <option value="MARKET">市价单</option>
                <option value="LIMIT">限价单</option>
              </select>
            </div>
          </div>
          
          <div className="form-row">
            <div className="form-group form-group-half">
              <label>触发价格：</label>
              <input 
                type="number" 
                value={editRule.triggerPrice || ''} 
                onChange={(e) => setEditRule({...editRule, triggerPrice: parseFloat(e.target.value)})}
                step="0.000001"
                min="0"
                required
              />
            </div>
            
            <div className="form-group form-group-half">
              <label>数量：</label>
              <input 
                type="number" 
                value={editRule.quantity || ''} 
                onChange={(e) => setEditRule({...editRule, quantity: parseFloat(e.target.value)})}
                step="0.000001"
                min="0"
                required
              />
            </div>
          </div>
          
          <div className="form-row">
            <div className="form-group form-group-half">
              <label>滑点(%)：</label>
              <input 
                type="number" 
                value={editRule.slippage || ''} 
                onChange={(e) => setEditRule({...editRule, slippage: parseFloat(e.target.value)})}
                step="0.1"
                min="0"
                max="100"
              />
            </div>
            
            <div className="form-group form-group-half">
              <label>状态：</label>
              <select 
                value={editRule.isEnabled} 
                onChange={(e) => setEditRule({...editRule, isEnabled: parseInt(e.target.value)})}
              >
                <option value="1">启用</option>
                <option value="0">禁用</option>
              </select>
            </div>
          </div>
          
          <div className="form-group">
            <label>过期时间：</label>
            <input 
              type="datetime-local" 
              value={editRule.expirationTime ? new Date(editRule.expirationTime * 1000).toISOString().slice(0, 16) : ''} 
              onChange={(e) => {
                const date = e.target.value ? new Date(e.target.value).getTime() / 1000 : undefined;
                setEditRule({...editRule, expirationTime: date});
              }}
            />
          </div>
          
          <div className="form-actions">
            <button className="btn cancel-button" onClick={() => setIsEditModalOpen(false)}>取消</button>
            <button className="btn submit-button" onClick={handleEditRule}>保存</button>
          </div>
        </div>
      </Modal>
      
      {/* 删除确认模态框 */}
      <Modal isOpen={isDeleteConfirmOpen} onClose={() => setIsDeleteConfirmOpen(false)} title="确认删除">
        <div className="confirm-container">
          {deleteRuleSuccess === true && (
            <div className="success-message">{deleteRuleMessage}</div>
          )}
          {deleteRuleSuccess === false && (
            <div className="error-message">{deleteRuleMessage}</div>
          )}
          
          {!deleteRuleSuccess && (
            <>
              <div className="delete-warning-icon">⚠️</div>
              <p>确定要删除以下交易规则吗？</p>
              {ruleToDelete && (
                <div className="confirm-details">
                  <p><strong>ID:</strong> {ruleToDelete.RuleID}</p>
                  <p><strong>代币:</strong> {getTokenName(ruleToDelete.TokenID)}</p>
                  <p><strong>方向:</strong> <span className={ruleToDelete.Direction === 'BUY' ? 'buy-direction' : 'sell-direction'}>{ruleToDelete.Direction}</span></p>
                  <p><strong>触发价格:</strong> {ruleToDelete.TriggerPrice}</p>
                  <p><strong>数量:</strong> {ruleToDelete.Quantity}</p>
                </div>
              )}
              <p className="delete-warning-text">此操作不可撤销！</p>
              <div className="confirm-actions delete-actions">
                <button className="btn cancel-button" onClick={() => setIsDeleteConfirmOpen(false)}>取消</button>
                <button className="btn delete-confirm-button" onClick={handleDeleteRule}>确认删除</button>
              </div>
            </>
          )}
        </div>
      </Modal>
    </div>
  );
};

export default TradingRulesPage;