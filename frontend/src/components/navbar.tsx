import React, { useState, useEffect } from 'react';
import { NavLink, useNavigate, useLocation } from 'react-router-dom';
import authService from '../services/authService';
import '../styles/navbar.css';

const Navbar: React.FC = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [username, setUsername] = useState('');
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    // 监听认证状态变化
    const subscription = authService.isAuthenticated$.subscribe(authenticated => {
      setIsAuthenticated(authenticated);
      if (authenticated) {
        setUsername(authService.currentUsername);
      } else {
        setUsername('');
      }
    });

    // 初始化状态
    setIsAuthenticated(authService.isAuthenticated);
    setUsername(authService.currentUsername);

    return () => {
      subscription.unsubscribe();
    };
  }, []);
  
  const handleLogout = () => {
    authService.logout();
    navigate('/login');
  };

  // 检查是否在登录页面
  const isLoginPage = location.pathname === '/login';

  return (
    <nav className="navbar">
      <div className="navbar-brand">
        <span>GoCryptoTrader</span>
      </div>
      {!isLoginPage && (
        <div className="navbar-menu">
          <NavLink to="/tokenmonitor" className={({ isActive }) => isActive ? 'active' : ''}>
            代币监视列表
          </NavLink>
          <NavLink to="/trading-rules" className={({ isActive }) => isActive ? 'active' : ''}>
            交易规则
          </NavLink>
          <NavLink to="/accounts" className={({ isActive }) => isActive ? 'active' : ''}>
            账户管理
          </NavLink>
        </div>
      )}
      <div className="navbar-end">
        {isAuthenticated && !isLoginPage && (
          <div className="user-info">
            <span className="username">{username}</span>
            <button className="logout-button" onClick={handleLogout}>登出</button>
          </div>
        )}
      </div>
    </nav>
  );
};

export default Navbar;