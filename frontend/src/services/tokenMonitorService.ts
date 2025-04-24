import websocketService from './websocketService';

// TokenMonitor API服务类
export class TokenMonitorService {
  // 获取所有代币监控信息
  public getTokenMonitors(): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法发送获取代币监控请求');
      return;
    }
    console.log('发送获取代币监控请求');
    try {
      websocketService.sendMessage('gettokenmonitors', {});
    } catch (error) {
      console.error('发送获取代币监控请求失败:', error);
    }
  }

  // 添加代币监控
  public addTokenMonitor(tokenMonitor: any): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法发送添加代币监控请求');
      return;
    }
    console.log('发送添加代币监控请求', tokenMonitor);
    try {
      websocketService.sendMessage('addtokenmonitor', tokenMonitor);
    } catch (error) {
      console.error('发送添加代币监控请求失败:', error);
    }
  }

  // 更新代币监控
  public updateTokenMonitor(tokenMonitor: any): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法发送更新代币监控请求');
      return;
    }
    console.log('发送更新代币监控请求', tokenMonitor);
    try {
      websocketService.sendMessage('updatetokenmonitor', tokenMonitor);
    } catch (error) {
      console.error('发送更新代币监控请求失败:', error);
    }
  }

  // 删除代币监控
  public deleteTokenMonitor(tokenAddress: string): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法发送删除代币监控请求');
      return;
    }
    console.log('发送删除代币监控请求', tokenAddress);
    try {
      websocketService.sendMessage('deletetokenmonitor', { tokenAddress });
    } catch (error) {
      console.error('发送删除代币监控请求失败:', error);
    }
  }
}

// 导出单例实例
const tokenMonitorService = new TokenMonitorService();
export default tokenMonitorService;