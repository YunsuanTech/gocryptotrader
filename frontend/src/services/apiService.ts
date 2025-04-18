import websocketService from './websocketService';
import { AddTokenRequest } from '../types/watchlist';
import { AddTradingRuleRequest, UpdateTradingRuleRequest } from '../types/trading';

// API服务类
export class ApiService {
  // 交易规则相关API
  
  // 获取所有交易规则
  public getTradingRules(limit: number = 0): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法获取交易规则列表');
      return;
    }
    
    const data = { limit };
    websocketService.sendMessage('gettradingrules', data);
  }
  
  // 根据ID获取交易规则
  public getTradingRuleByID(ruleID: number): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法获取交易规则');
      return;
    }
    
    websocketService.sendMessage('gettradingrulebyid', { ruleID });
  }
  
  // 根据代币ID获取交易规则
  public getTradingRulesByTokenID(tokenID: number): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法获取交易规则');
      return;
    }
    
    websocketService.sendMessage('gettradingrulesbytokenid', { tokenID });
  }
  
  // 根据用户地址获取交易规则
  public getTradingRulesByUserAddress(userAddress: string): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法获取交易规则');
      return;
    }
    
    websocketService.sendMessage('gettradingrulesbyuseraddress', { userAddress });
  }
  
  // 根据用户地址和代币ID获取交易规则
  public getTradingRulesByUserAndToken(userAddress: string, tokenID: number): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法获取交易规则');
      return;
    }
    
    websocketService.sendMessage('gettradingrulesbyuserandtoken', { userAddress, tokenID });
  }
  
  // 获取活跃的交易规则
  public getActiveTradingRules(): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法获取活跃交易规则');
      return;
    }
    
    websocketService.sendMessage('getactivetradingrules', {});
  }
  
  // 添加交易规则
  public addTradingRule(ruleRequest: AddTradingRuleRequest): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法添加交易规则');
      return;
    }
    
    websocketService.sendMessage('addtradingrule', ruleRequest);
  }
  
  // 更新交易规则
  public updateTradingRule(ruleID: number, ruleData: UpdateTradingRuleRequest): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法更新交易规则');
      return;
    }
    
    websocketService.sendMessage('updatetradingrule', { 
      ruleId: ruleID,
      ...ruleData
    });
  }
  
  // 删除交易规则
  public deleteTradingRule(ruleID: number): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法删除交易规则');
      return;
    }
    
    websocketService.sendMessage('deletetradingrule', { ruleId: ruleID });
  }
  
  // 代币监视列表相关API
  
  // 获取代币监视列表
  public getWatchlistTokens(network: string = '', limit: number = 0): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法获取监视列表');
      return;
    }
    
    const data = {
      network,
      limit
    };
    
    websocketService.sendMessage('getwatchlisttokens', data);
  }

  // 添加代币到监视列表
  public addWatchlistToken(tokenRequest: AddTokenRequest): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法添加代币到监视列表');
      return;
    }
    
    websocketService.sendMessage('addwatchlisttoken', tokenRequest);
  }

  // 通过ID获取代币
  public getWatchlistTokenByID(tokenID: number): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法获取代币信息');
      return;
    }
    
    websocketService.sendMessage('getwatchlisttokenbyid', { tokenID });
  }

  // 通过地址获取代币
  public getWatchlistTokenByAddress(tokenAddress: string): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法获取代币信息');
      return;
    }
    
    websocketService.sendMessage('getwatchlisttokenbyaddress', { tokenAddress });
  }

  // 通过符号获取代币
  public getWatchlistTokensBySymbol(tokenSymbol: string): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法获取代币信息');
      return;
    }
    
    websocketService.sendMessage('getwatchlisttokensbysymbol', { tokenSymbol });
  }

  // 获取活跃的代币
  public getActiveWatchlistTokens(): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法获取活跃代币');
      return;
    }
    
    websocketService.sendMessage('getactivewatchlisttokens', {});
  }

  // 更新代币信息（通过ID）
  public updateWatchlistToken(tokenID: number, tokenData: Partial<AddTokenRequest>): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法更新代币信息');
      return;
    }
    console.log(tokenData);
    websocketService.sendMessage('updatewatchlisttoken', { 
      tokenId: tokenID, // 修改为后端期望的参数名 tokenId（小写d）
      ...tokenData
    });
  }

  // 更新代币信息（通过地址）
  public updateWatchlistTokenByAddress(tokenAddress: string, tokenData: Partial<AddTokenRequest>): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法更新代币信息');
      return;
    }
    console.log(tokenData);
    websocketService.sendMessage('updatewatchlisttokenbyaddress', { 
      tokenAddress,
      newTokenAddress: tokenData.tokenAddress, // 添加后端期望的newTokenAddress参数
      ...tokenData
    });
  }

  // 删除代币（通过ID）
  public deleteWatchlistToken(tokenID: number): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法删除代币');
      return;
    }
    
    websocketService.sendMessage('deletewatchlisttoken', { tokenID });
  }

  // 删除代币（通过地址）
  public deleteWatchlistTokenByAddress(tokenAddress: string): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法删除代币');
      return;
    }
    
    websocketService.sendMessage('deletewatchlisttokenbyaddress', { tokenAddress });
  }
}

// 导出单例实例
const apiService = new ApiService();
export default apiService;