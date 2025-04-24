import websocketService from './websocketService';

// TradingRule API服务类
export class TradingRuleService {
  // 获取所有交易规则信息
  public getTradingRules(): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法发送获取交易规则请求');
      return;
    }
    console.log('发送获取交易规则请求');
    try {
      websocketService.sendMessage('gettradingrules', {});
    } catch (error) {
      console.error('发送获取交易规则请求失败:', error);
    }
  }

  // 添加交易规则
  public addTradingRule(tradingRule: any): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法发送添加交易规则请求');
      return;
    }
    console.log('发送添加交易规则请求', tradingRule);
    try {
      websocketService.sendMessage('addtradingrule', tradingRule);
    } catch (error) {
      console.error('发送添加交易规则请求失败:', error);
    }
  }

  // 更新交易规则
  public updateTradingRule(tradingRule: any): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法发送更新交易规则请求');
      return;
    }
    console.log('发送更新交易规则请求', tradingRule);
    try {
      websocketService.sendMessage('updatetradingrule', tradingRule);
    } catch (error) {
      console.error('发送更新交易规则请求失败:', error);
    }
  }

  // 删除交易规则
  public deleteTradingRule(id: number): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法发送删除交易规则请求');
      return;
    }
    console.log('发送删除交易规则请求', id);
    try {
      websocketService.sendMessage('deletetradingrule', { id });
    } catch (error) {
      console.error('发送删除交易规则请求失败:', error);
    }
  }
}

// 创建单例实例
const tradingRuleService = new TradingRuleService();
export default tradingRuleService;