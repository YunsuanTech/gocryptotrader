import { BehaviorSubject } from 'rxjs';
import websocketResponseHandlerService from './websocketResponseHandlerService';
import * as CryptoJS from 'crypto-js';

// 认证服务
export class AuthService {
  private isAuthenticatedSubject = new BehaviorSubject<boolean>(false);
  public isAuthenticated$ = this.isAuthenticatedSubject.asObservable();
  private username: string = '';

  constructor() {
    // 监听WebSocket认证消息
    websocketResponseHandlerService.shared.subscribe(message => {
      if (message.event === 'auth') {
        if (!message.error) {
          this.isAuthenticatedSubject.next(true);
          console.log('认证成功');
        } else {
          this.isAuthenticatedSubject.next(false);
          console.log('认证失败:', message.error);
        }
      }
    });
  }

  // 获取当前认证状态
  public get isAuthenticated(): boolean {
    return this.isAuthenticatedSubject.value;
  }

  // 获取当前用户名
  public get currentUsername(): string {
    return this.username;
  }

  // 登录方法
  public login(username: string, password: string): void {
    if (!websocketResponseHandlerService.isConnected) {
      console.error('WebSocket未连接，无法发送认证请求');
      return;
    }
    
    this.username = username;
    const hash = CryptoJS.SHA256(password).toString(CryptoJS.enc.Hex);
    websocketResponseHandlerService.authenticate(username, hash);
  }

  // 登出方法
  public logout(): void {
    this.isAuthenticatedSubject.next(false);
    this.username = '';
  }
}

// 导出单例实例
const authService = new AuthService();
export default authService;