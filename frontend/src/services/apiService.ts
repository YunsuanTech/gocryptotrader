import websocketService from './websocketService';

// API服务类
export class ApiService {
  // 获取账户信息
  public getAccounts(): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法发送获取账户请求');
      return;
    }
    console.log('发送获取账户请求');
    try {
      websocketService.sendMessage('getaccounts', {});
    } catch (error) {
      console.error('发送获取账户请求失败:', error);
    }
  }
}

// 导出单例实例
const apiService = new ApiService();
export default apiService;