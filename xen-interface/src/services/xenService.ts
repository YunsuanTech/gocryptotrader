import websocketService from './websocketService';

// Xen记录接口定义
export interface XenRecord {
  slot: number;
  chainName: string;
  count: number;
  days: number;
  executionTime?: Date;
  claimTime?: Date;
  expectedReward?: number;
  ranking: number;
  amp: number;
  eaa: number;
  m?: number;
  status: string;
  txID?: string;
  mintFees?: number;
  claimFees?: number;
}

// 添加Xen记录请求接口
export interface AddXenRequest {
  slot: number;
  chainName: string;
  count: number;
  days: number;
  ranking: number;
  amp: number;
  eaa: number;
  status: string;
  executionTime?: Date;
  claimTime?: Date;
  expectedReward?: number;
  m?: number;
  txID?: string;
  mintFees?: number;
  claimFees?: number;
}

// 更新Xen记录请求接口
export interface UpdateXenRequest {
  slot: number;
  chainName: string;
  count?: number;
  days?: number;
  ranking?: number;
  amp?: number;
  eaa?: number;
  status?: string;
  executionTime?: Date;
  claimTime?: Date;
  expectedReward?: number;
  m?: number;
  txID?: string;
  mintFees?: number;
  claimFees?: number;
}

// Xen服务类
export class XenService {
  // WebSocket API
  
  /**
   * 根据链名获取Xen记录
   * @param chainName 链名称
   */
  public getXensByChainName(chainName: string): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法获取Xen记录');
      return;
    }
    
    console.log('发送获取Xen记录请求，链名:', chainName);
    websocketService.sendMessage('getxensbychainname', { chainName });
  }
  
  /**
   * 根据状态和链名获取Xen记录
   * @param status 状态
   * @param chainName 链名称
   */
  public getXensByStatusAndChain(status: string, chainName: string): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法获取Xen记录');
      return;
    }
    
    console.log('发送获取Xen记录请求，状态:', status, '链名:', chainName);
    websocketService.sendMessage('getxensbystatusandchain', { status, chainName });
  }
  
  /**
   * 添加Xen记录
   * @param xenRequest Xen记录请求
   */
  public addXen(xenRequest: AddXenRequest): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法添加Xen记录');
      return;
    }
    
    console.log('发送添加Xen记录请求:', xenRequest);
    websocketService.sendMessage('addxen', xenRequest);
  }
  
  /**
   * 更新Xen记录
   * @param xenRequest Xen记录更新请求
   */
  public updateXen(xenRequest: UpdateXenRequest): void {
    if (!websocketService.isConnected) {
      console.error('WebSocket未连接，无法更新Xen记录');
      return;
    }
    
    console.log('发送更新Xen记录请求:', xenRequest);
    websocketService.sendMessage('updatexen', xenRequest);
  }
  
  // REST API
  
  /**
   * 根据链名获取Xen记录 (REST API)
   * @param chainName 链名称
   * @returns Promise<XenRecord[]>
   */
  public async fetchXensByChainName(chainName: string): Promise<XenRecord[]> {
    try {
      const response = await fetch(`/xens/chain/${chainName}`);
      if (!response.ok) {
        throw new Error(`获取Xen记录失败: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('获取Xen记录失败:', error);
      throw error;
    }
  }
  
  /**
   * 根据状态和链名获取Xen记录 (REST API)
   * @param status 状态
   * @param chainName 链名称
   * @returns Promise<XenRecord[]>
   */
  public async fetchXensByStatusAndChain(status: string, chainName: string): Promise<XenRecord[]> {
    try {
      const response = await fetch(`/xens/status/${status}/chain/${chainName}`);
      if (!response.ok) {
        throw new Error(`获取Xen记录失败: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('获取Xen记录失败:', error);
      throw error;
    }
  }
  
  /**
   * 添加Xen记录 (REST API)
   * @param xenRequest Xen记录请求
   * @returns Promise<{status: string, message: string}>
   */
  public async fetchAddXen(xenRequest: AddXenRequest): Promise<{status: string, message: string}> {
    try {
      const response = await fetch('/xens/add', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(xenRequest)
      });
      
      if (!response.ok) {
        throw new Error(`添加Xen记录失败: ${response.statusText}`);
      }
      
      return await response.json();
    } catch (error) {
      console.error('添加Xen记录失败:', error);
      throw error;
    }
  }
  
  /**
   * 更新Xen记录 (REST API)
   * @param xenRequest Xen记录更新请求
   * @returns Promise<{status: string, message: string}>
   */
  public async fetchUpdateXen(xenRequest: UpdateXenRequest): Promise<{status: string, message: string}> {
    try {
      const response = await fetch('/xens/update', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(xenRequest)
      });
      
      if (!response.ok) {
        throw new Error(`更新Xen记录失败: ${response.statusText}`);
      }
      
      return await response.json();
    } catch (error) {
      console.error('更新Xen记录失败:', error);
      throw error;
    }
  }
}

// 导出单例实例
const xenService = new XenService();
export default xenService;