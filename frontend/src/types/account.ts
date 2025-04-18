// 账户信息相关的类型定义

// 账户信息接口
export interface Account {
  id: string;
  exchange: string;
  assetType: string;
  name: string;
  currencyName: string;
  totalValue: number;
  hold: number;
  free: number;
  frozen: number;
  lastUpdated: string;
}

// 账户列表接口
export interface AccountsResponse {
  accounts: Account[];
}