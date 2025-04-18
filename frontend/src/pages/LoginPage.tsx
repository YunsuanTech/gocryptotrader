import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import websocketResponseHandlerService from '../services/websocketResponseHandlerService';
import authService from '../services/authService';
import './pages.css';

const LoginPage: React.FC = () => {
  const [username, setUsername] = useState<string>('');
  const [password, setPassword] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [isConnected, setIsConnected] = useState<boolean>(false);
  const navigate = useNavigate();

  useEffect(() => {
    console.log('LoginPage组件挂载');
    
    // 监听认证状态
    const authSubscription = authService.isAuthenticated$.subscribe(isAuthenticated => {
      if (isAuthenticated) {
        setLoading(false);
        console.log('认证成功，跳转到账户页面');
        navigate('/accounts');
      }
    });
    
    // 监听WebSocket消息，处理错误
    const messageSubscription = websocketResponseHandlerService.shared.subscribe({
      next: (message) => {
        if (message.event === 'auth' && message.error) {
          setLoading(false);
          setError(`认证失败: ${message.error}`);
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

    const checkConnection = () => {
      const isConnected = websocketResponseHandlerService.isConnected;
      setIsConnected(isConnected);
    };

    // 立即检查一次
    checkConnection();

    // 设置轮询检测
    const connectionCheck = setInterval(checkConnection, 1000);

    return () => {
      console.log('LoginPage组件卸载');
      authSubscription.unsubscribe();
      messageSubscription.unsubscribe();
      clearInterval(connectionCheck);
    };
  }, [navigate]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!username.trim() || !password.trim()) {
      setError('用户名和密码不能为空');
      return;
    }

    if (!isConnected) {
      setError('WebSocket未连接，请稍后再试');
      return;
    }

    setLoading(true);
    setError(null);
    
    // 使用认证服务进行登录
    authService.login(username, password);
  };

  return (
    <div className="page-container">
      <div className="login-container">
        <div className="login-form-container">
          <h1 className="login-title">GoCryptoTrader</h1>
          <h2 className="login-subtitle">登录</h2>
          
          {error && <div className="error-message">{error}</div>}
          
          <form className="login-form" onSubmit={handleSubmit}>
            <div className="form-group">
              <label htmlFor="username">用户名</label>
              <input
                type="text"
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={loading}
                placeholder="请输入用户名"
                autoComplete="username"
              />
            </div>
            
            <div className="form-group">
              <label htmlFor="password">密码</label>
              <input
                type="password"
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={loading}
                placeholder="请输入密码"
                autoComplete="current-password"
              />
            </div>
            
            <div className="form-actions">
            <div className={`connection-status ${isConnected ? 'connected' : 'disconnected'}`}>
            WebSocket状态: {isConnected ? '已连接' : '未连接'}
            </div>
              <button 
                type="submit" 
                className="btn btn-primary login-button" 
                disabled={loading || !isConnected}
              >
                {loading ? '登录中...' : '登录'}
              </button>
            </div>
          </form>
          

        </div>
      </div>
    </div>
  );
};

export default LoginPage;