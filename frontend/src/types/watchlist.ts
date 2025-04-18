// 代币监视列表相关的类型定义

// 代币信息接口
export interface WatchlistToken {
  ID?: number; // 添加ID字段用于更新和删除操作
  TokenSymbol: string;
  TokenAddress: string;
  Network: string;
  Decimals: number;
  CreationTime: number;
  LastUpdated: number;
  IsActive: number;
}

// 添加代币请求接口
export interface AddTokenRequest {
  tokenSymbol: string;
  tokenAddress: string;
  network: string;
  decimals: number;
  isActive: number;
}

// 更新代币请求接口
export interface UpdateTokenRequest {
  tokenSymbol: string;
  tokenAddress: string;
  network: string;
  decimals: number;
  isActive: number;
}

// 添加/更新代币响应接口
export interface TokenResponse {
  status: string;
  message: string;
}