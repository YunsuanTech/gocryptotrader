import React, { useState, useEffect, useRef } from 'react';
import websocketResponseHandlerService from '../services/websocketResponseHandlerService';
import { WatchlistToken, AddTokenRequest, UpdateTokenRequest } from '../types/watchlist';
import authService from '../services/authService';
import apiService from '../services/apiService';
import Modal from '../components/Modal';
import '../components/Modal.css';
import './WatchlistPage.css';

const WatchlistPage: React.FC = () => {
  const [tokens, setTokens] = useState<WatchlistToken[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [isConnected, setIsConnected] = useState<boolean>(false);
  const [selectedToken, setSelectedToken] = useState<WatchlistToken | null>(null);
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);
  const [isAddModalOpen, setIsAddModalOpen] = useState<boolean>(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState<boolean>(false);
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState<boolean>(false);
  const [addTokenSuccess, setAddTokenSuccess] = useState<boolean | null>(null);
  const [addTokenMessage, setAddTokenMessage] = useState<string>('');
  const [editTokenSuccess, setEditTokenSuccess] = useState<boolean | null>(null);
  const [editTokenMessage, setEditTokenMessage] = useState<string>('');
  const [deleteTokenSuccess, setDeleteTokenSuccess] = useState<boolean | null>(null);
  const [deleteTokenMessage, setDeleteTokenMessage] = useState<string>('');
  const [networkFilter, setNetworkFilter] = useState<string>('');
  const [tokenToEdit, setTokenToEdit] = useState<WatchlistToken | null>(null);
  const [tokenToDelete, setTokenToDelete] = useState<WatchlistToken | null>(null);
  const authSent = useRef(false);

  // 新增代币表单状态
  const [newToken, setNewToken] = useState<AddTokenRequest>({
    tokenSymbol: '',
    tokenAddress: '',   
    network: '',
    decimals: 18,
    isActive: 1
  });

  // 编辑代币表单状态
  const [editToken, setEditToken] = useState<UpdateTokenRequest>({
    tokenSymbol: '',
    tokenAddress: '',
    network: '',
    decimals: 18,
    isActive: 1
  });

  useEffect(() => {
    console.log('WatchlistPage组件挂载');
    
    // 检查认证状态
    if (!authService.isAuthenticated) {
      console.log('用户未认证，重定向到登录页面');
      window.location.href = '/login';
      return;
    }
    
    const subscription = websocketResponseHandlerService.shared.subscribe({
      next: (message) => {
        if (message.event === 'getwatchlisttokens') {
          if (message.error) {
            setError(`获取代币监视列表失败: ${message.error}`);
            setLoading(false);
          } else if (message.data) {
            const tokenData = Array.isArray(message.data) ? message.data : [];
            setTokens(tokenData);
            setLoading(false);
            setError(null);
          }
        } else if (message.event === 'addwatchlisttoken') {
          if (message.error) {
            setAddTokenSuccess(false);
            setAddTokenMessage(`添加代币失败: ${message.error}`);
          } else {
            setAddTokenSuccess(true);
            setAddTokenMessage('代币添加成功');
            // 重新加载代币列表
            handleRefresh();
            // 重置表单
            setNewToken({
              tokenSymbol: '',
              tokenAddress: '',
              network: '',
              decimals: 18,
              isActive: 1
            });
            // 关闭添加模态框
            setTimeout(() => {
              setIsAddModalOpen(false);
              setAddTokenSuccess(null);
              setAddTokenMessage('');
            }, 2000);
          }
        } else if (message.event === 'updatewatchlisttoken' || message.event === 'updatewatchlisttokenbyaddress') {
          if (message.error) {
            setEditTokenSuccess(false);
            setEditTokenMessage(`更新代币失败: ${message.error}`);
          } else {
            setEditTokenSuccess(true);
            setEditTokenMessage('代币更新成功');
            // 重新加载代币列表
            handleRefresh();
            // 关闭编辑模态框
            setTimeout(() => {
              setIsEditModalOpen(false);
              setEditTokenSuccess(null);
              setEditTokenMessage('');
              setTokenToEdit(null);
            }, 2000);
          }
        } else if (message.event === 'deletewatchlisttoken' || message.event === 'deletewatchlisttokenbyaddress') {
          if (message.error) {
            setDeleteTokenSuccess(false);
            setDeleteTokenMessage(`删除代币失败: ${message.error}`);
          } else {
            setDeleteTokenSuccess(true);
            setDeleteTokenMessage('代币删除成功');
            // 重新加载代币列表
            handleRefresh();
            // 关闭删除确认框
            setTimeout(() => {
              setIsDeleteConfirmOpen(false);
              setDeleteTokenSuccess(null);
              setDeleteTokenMessage('');
              setTokenToDelete(null);
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
    
    // 如果已连接且已认证，则请求代币数据
    if (websocketResponseHandlerService.isConnected && authService.isAuthenticated && loading) {
      apiService.getWatchlistTokens(networkFilter);
    }

    return () => {
      console.log('WatchlistPage组件卸载');
      subscription.unsubscribe();
      authSent.current = false;
    };
  }, [networkFilter]);

  const handleRefresh = () => {
    setLoading(true);
    if (websocketResponseHandlerService.isConnected) {
      apiService.getWatchlistTokens(networkFilter);
    } else {
      setError('WebSocket未连接');
      setLoading(false);
    }
  };

  const handleNetworkFilterChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setNetworkFilter(e.target.value);
    setLoading(true);
  };

  const handleAddToken = () => {
    // 验证表单
    if (!newToken.tokenSymbol || !newToken.tokenAddress || !newToken.network) {
      setAddTokenSuccess(false);
      setAddTokenMessage('请填写所有必填字段');
      return;
    }

    // 发送添加代币请求
    apiService.addWatchlistToken(newToken);
  };

  const handleEditToken = () => {
    // 验证表单
    if (!editToken.tokenSymbol || !editToken.tokenAddress || !editToken.network) {
      setEditTokenSuccess(false);
      setEditTokenMessage('请填写所有必填字段');
      return;
    }

    // 发送更新代币请求
    if (tokenToEdit?.ID) {
      apiService.updateWatchlistToken(tokenToEdit.ID, editToken);
    } else if (tokenToEdit?.TokenAddress) {
      apiService.updateWatchlistTokenByAddress(tokenToEdit.TokenAddress, editToken);
    } else {
      setEditTokenSuccess(false);
      setEditTokenMessage('无法更新代币：缺少ID或地址');
    }
  };

  const handleDeleteToken = () => {
    if (!tokenToDelete) return;

    // 发送删除代币请求
    if (tokenToDelete.ID) {
      apiService.deleteWatchlistToken(tokenToDelete.ID);
    } else if (tokenToDelete.TokenAddress) {
      apiService.deleteWatchlistTokenByAddress(tokenToDelete.TokenAddress);
    } else {
      setDeleteTokenSuccess(false);
      setDeleteTokenMessage('无法删除代币：缺少ID或地址');
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setNewToken(prev => ({
      ...prev,
      [name]: name === 'decimals' || name === 'isActive' ? parseInt(value) : value
    }));
  };

  const handleEditInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setEditToken(prev => ({
      ...prev,
      [name]: name === 'decimals' || name === 'isActive' ? parseInt(value) : value
    }));
  };

  const openEditModal = (token: WatchlistToken) => {
    setTokenToEdit(token);
    setEditToken({
      tokenSymbol: token.TokenSymbol,
      tokenAddress: token.TokenAddress,
      network: token.Network,
      decimals: token.Decimals,
      isActive: token.IsActive
    });
    setIsEditModalOpen(true);
  };

  const openDeleteConfirm = (token: WatchlistToken) => {
    setTokenToDelete(token);
    setIsDeleteConfirmOpen(true);
  };

  const formatDateTime = (timestamp: number) => {
    if (!timestamp) return '-';
    try {
      const date = new Date(timestamp * 1000);
      return date.toLocaleString();
    } catch (e) {
      return timestamp.toString();
    }
  };

  return (
    <div className="page-container">
      <div className="page-header">
        <h1>代币监视列表</h1>
        <div className="page-actions">
          <div className="filter-container">
            <label htmlFor="networkFilter">网络筛选:</label>
            <select
              id="networkFilter"
              value={networkFilter}
              onChange={handleNetworkFilterChange}
              className="filter-select"
            >
              <option value="">全部</option>
              <option value="ethereum">以太坊</option>
              <option value="bsc">币安智能链</option>
              <option value="polygon">Polygon</option>
              <option value="solana">solana</option>
              <option value="arbitrum">Arbitrum</option>
              <option value="optimism">Optimism</option>
            </select>
          </div>
          <button 
            className="add-button" 
            onClick={() => setIsAddModalOpen(true)}
            disabled={!isConnected}
          >
            添加代币
          </button>
          <button 
            className="refresh-button" 
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

      {error && <div className="error-message">{error}</div>}

      {loading ? (
        <div className="loading-indicator">
          <div className="spinner"></div>
          <span>加载代币信息中...</span>
        </div>
      ) : tokens.length > 0 ? (
        <div className="table-container">
          <table className="data-table watchlist-table">
            <thead>
              <tr>
              
                <th>代币符号</th>
                <th>代币地址</th>
                <th>网络</th>
                <th>精度</th>
                <th>创建时间</th>
                <th>最后更新</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {tokens.map((token) => (
                <tr key={token.TokenAddress} className="table-row">
                
                  <td>{token.TokenSymbol}</td>
                  <td className="address-cell">
                    <div className="address-wrapper">
                      {token.TokenAddress ? 
                        <>
                          {token.TokenAddress.substring(0, 8)}...{token.TokenAddress.substring(token.TokenAddress.length - 6)}
                          <button 
                            className="copy-button" 
                            onClick={() => {
                              navigator.clipboard.writeText(token.TokenAddress);
                              alert('地址已复制到剪贴板');
                            }}
                            title="复制地址"
                          >
                            📋
                          </button>
                        </> : 
                        '无地址'
                      }
                    </div>
                  </td>
                  <td>{token.Network}</td>
                  <td>{token.Decimals}</td>
                  <td>{formatDateTime(token.CreationTime)}</td>
                  <td>{formatDateTime(token.LastUpdated)}</td>
                  <td>{token.IsActive ? '活跃' : '非活跃'}</td>
                  <td className="action-buttons">
                    <button 
                      className="view-details-btn"
                      onClick={() => {
                        setSelectedToken(token);
                        setIsModalOpen(true);
                      }}
                    >
                      查看
                    </button>
                    <button 
                      className="edit-btn"
                      onClick={() => openEditModal(token)}
                      disabled={!isConnected}
                    >
                      修改
                    </button>
                    <button 
                      className="delete-btn"
                      onClick={() => openDeleteConfirm(token)}
                      disabled={!isConnected}
                    >
                      删除
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="no-data-message">没有找到代币信息</div>
      )}
      
      {/* 代币详情模态框 */}
      <Modal 
        isOpen={isModalOpen} 
        onClose={() => setIsModalOpen(false)} 
        title="代币详细信息"
      >
        {selectedToken && (
          <div className="token-details">
            <div className="detail-row">
              <div className="detail-label">代币符号:</div>
              <div className="detail-value">{selectedToken.TokenSymbol}</div>
            </div>
            <div className="detail-row">
              <div className="detail-label">代币地址:</div>
              <div className="detail-value address-value">
                {selectedToken.TokenAddress ? (
                  <>
                    {selectedToken.TokenAddress}
                    <button 
                      className="copy-button" 
                      onClick={() => {
                        navigator.clipboard.writeText(selectedToken.TokenAddress);
                        alert('地址已复制到剪贴板');
                      }}
                      title="复制地址"
                    >
                      📋
                    </button>
                  </>
                ) : '无地址'}
              </div>
            </div>
            <div className="detail-row">
              <div className="detail-label">网络:</div>
              <div className="detail-value">{selectedToken.Network}</div>
            </div>
            <div className="detail-row">
              <div className="detail-label">精度:</div>
              <div className="detail-value">{selectedToken.Decimals}</div>
            </div>
            <div className="detail-row">
              <div className="detail-label">创建时间:</div>
              <div className="detail-value">{formatDateTime(selectedToken.CreationTime)}</div>
            </div>
            <div className="detail-row">
              <div className="detail-label">最后更新:</div>
              <div className="detail-value">{formatDateTime(selectedToken.LastUpdated)}</div>
            </div>
            <div className="detail-row">
              <div className="detail-label">状态:</div>
              <div className="detail-value">{selectedToken.IsActive ? '活跃' : '非活跃'}</div>
            </div>
          </div>
        )}
      </Modal>

      {/* 添加代币模态框 */}
      <Modal 
        isOpen={isAddModalOpen} 
        onClose={() => {
          setIsAddModalOpen(false);
          setAddTokenSuccess(null);
          setAddTokenMessage('');
        }} 
        title="添加代币到监视列表"
      >
        <div className="add-token-form">
          {addTokenSuccess !== null && (
            <div className={`message ${addTokenSuccess ? 'success' : 'error'}`}>
              {addTokenMessage}
            </div>
          )}
          <div className="form-group">
            <label htmlFor="tokenSymbol">代币符号 *</label>
            <input
              type="text"
              id="tokenSymbol"
              name="tokenSymbol"
              value={newToken.tokenSymbol}
              onChange={handleInputChange}
              placeholder="例如: ETH"
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="tokenAddress">代币地址 *</label>
            <input
              type="text"
              id="tokenAddress"
              name="tokenAddress"
              value={newToken.tokenAddress}
              onChange={handleInputChange}
              placeholder="例如: 0x..."
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="network">网络 *</label>
            <select
              id="network"
              name="network"
              value={newToken.network}
              onChange={handleInputChange}
              required
            >
              <option value="">选择网络</option>
              <option value="ethereum">以太坊</option>
              <option value="bsc">币安智能链</option>
              <option value="solana">solana</option>
              <option value="polygon">Polygon</option>
              <option value="arbitrum">Arbitrum</option>
              <option value="optimism">Optimism</option>
            </select>
          </div>
          <div className="form-group">
            <label htmlFor="decimals">精度</label>
            <input
              type="number"
              id="decimals"
              name="decimals"
              value={newToken.decimals}
              onChange={handleInputChange}
              min="0"
              max="18"
            />
          </div>
          <div className="form-group">
            <label htmlFor="isActive">状态</label>
            <select
              id="isActive"
              name="isActive"
              value={newToken.isActive}
              onChange={handleInputChange}
            >
              <option value={1}>活跃</option>
              <option value={0}>非活跃</option>
            </select>
          </div>
          <div className="form-actions">
            <button 
              className="cancel-button" 
              onClick={() => {
                setIsAddModalOpen(false);
                setAddTokenSuccess(null);
                setAddTokenMessage('');
              }}
            >
              取消
            </button>
            <button 
              className="submit-button" 
              onClick={handleAddToken}
              disabled={!isConnected}
            >
              添加
            </button>
          </div>
        </div>
      </Modal>

      {/* 编辑代币模态框 */}
      <Modal 
        isOpen={isEditModalOpen} 
        onClose={() => {
          setIsEditModalOpen(false);
          setEditTokenSuccess(null);
          setEditTokenMessage('');
        }} 
        title="修改代币信息"
      >
        <div className="edit-token-form">
          {editTokenSuccess !== null && (
            <div className={`message ${editTokenSuccess ? 'success' : 'error'}`}>
              {editTokenMessage}
            </div>
          )}
          <div className="form-group">
            <label htmlFor="edit-tokenSymbol">代币符号 *</label>
            <input
              type="text"
              id="edit-tokenSymbol"
              name="tokenSymbol"
              value={editToken.tokenSymbol}
              onChange={handleEditInputChange}
              placeholder="例如: ETH"
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-tokenAddress">代币地址 *</label>
            <input
              type="text"
              id="edit-tokenAddress"
              name="tokenAddress"
              value={editToken.tokenAddress}
              onChange={handleEditInputChange}
              placeholder="例如: 0x..."
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-network">网络 *</label>
            <select
              id="edit-network"
              name="network"
              value={editToken.network}
              onChange={handleEditInputChange}
              required
            >
              <option value="">选择网络</option>
              <option value="ethereum">以太坊</option>
              <option value="bsc">币安智能链</option>
              <option value="solana">solana</option>
              <option value="polygon">Polygon</option>
              <option value="arbitrum">Arbitrum</option>
              <option value="optimism">Optimism</option>
            </select>
          </div>
          <div className="form-group">
            <label htmlFor="edit-decimals">精度</label>
            <input
              type="number"
              id="edit-decimals"
              name="decimals"
              value={editToken.decimals}
              onChange={handleEditInputChange}
              min="0"
              max="18"
            />
          </div>
          <div className="form-group">
            <label htmlFor="edit-isActive">状态</label>
            <select
              id="edit-isActive"
              name="isActive"
              value={editToken.isActive}
              onChange={handleEditInputChange}
            >
              <option value={1}>活跃</option>
              <option value={0}>非活跃</option>
            </select>
          </div>
          <div className="form-actions">
            <button 
              className="cancel-button" 
              onClick={() => {
                setIsEditModalOpen(false);
                setEditTokenSuccess(null);
                setEditTokenMessage('');
              }}
            >
              取消
            </button>
            <button 
              className="submit-button" 
              onClick={handleEditToken}
              disabled={!isConnected}
            >
              保存
            </button>
          </div>
        </div>
      </Modal>

      {/* 删除确认模态框 */}
      <Modal 
        isOpen={isDeleteConfirmOpen} 
        onClose={() => {
          setIsDeleteConfirmOpen(false);
          setDeleteTokenSuccess(null);
          setDeleteTokenMessage('');
        }} 
        title="确认删除"
      >
        <div className="delete-confirm">
          {deleteTokenSuccess !== null ? (
            <div className={`message ${deleteTokenSuccess ? 'success' : 'error'}`}>
              {deleteTokenMessage}
            </div>
          ) : (
            <>
              <p>您确定要删除以下代币吗？</p>
              {tokenToDelete && (
                <div className="token-info">
                  <p><strong>代币符号:</strong> {tokenToDelete.TokenSymbol}</p>
                  <p><strong>代币地址:</strong> {tokenToDelete.TokenAddress}</p>
                  <p><strong>网络:</strong> {tokenToDelete.Network}</p>
                </div>
              )}
              <div className="form-actions">
                <button 
                  className="cancel-button" 
                  onClick={() => {
                    setIsDeleteConfirmOpen(false);
                    setTokenToDelete(null);
                  }}
                >
                  取消
                </button>
                <button 
                  className="delete-button" 
                  onClick={handleDeleteToken}
                  disabled={!isConnected}
                >
                  确认删除
                </button>
              </div>
            </>
          )}
        </div>
      </Modal>
    </div>
  );
};

export default WatchlistPage;