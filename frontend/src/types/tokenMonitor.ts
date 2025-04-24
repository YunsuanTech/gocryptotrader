// 代币监控相关的类型定义

// 代币监控信息接口
export interface TokenMonitor {
  tokenAddress: string;
  tokenName: string;
  price: number | null;
  tokenDecimals: number;
  buyAmount: number | null;
  buyPrice: number | null;
  buyTime: number | null; // Unix 时间戳
  sellPercentage: number | null;
  totalSellPrice: number | null;
  lastSellTime: number | null; // Unix 时间戳
  isMonitoring: number; // 0: false, 1: true
}

// 代币监控列表接口
export interface TokenMonitorResponse {
  tokenMonitors: TokenMonitor[];
}