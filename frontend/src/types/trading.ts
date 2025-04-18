// 交易规则相关的类型定义

// 交易规则信息接口
export interface TradingRule {
  RuleID?: number; // 规则ID (对应后端的RuleID)
  TokenID: number; // 关联的代币ID
  UserAddress: string; // 用户地址
  Direction: string; // 交易方向 (BUY/SELL)
  TriggerPrice: number; // 触发价格
  Quantity: number; // 交易数量
  Slippage: number; // 滑点百分比
  ExpirationTime?: number; // 规则过期时间
  IsEnabled: number; // 是否启用
  OrderType: string; // 订单类型 (MARKET/LIMIT)
  CreatedAt: number; // 创建时间
  LastTriggered?: number; // 上次触发时间
}

// 添加交易规则请求接口
export interface AddTradingRuleRequest {
  tokenId: number; // 关联的代币ID
  userAddress: string; // 用户地址
  direction: string; // 交易方向 (BUY/SELL)
  triggerPrice: number; // 触发价格
  quantity: number; // 交易数量
  slippage?: number; // 滑点百分比
  expirationTime?: number; // 规则过期时间
  isEnabled?: number; // 是否启用
  orderType?: string; // 订单类型 (MARKET/LIMIT)
}

// 更新交易规则请求接口
export interface UpdateTradingRuleRequest {
  tokenId?: number; // 关联的代币ID
  userAddress?: string; // 用户地址
  direction?: string; // 交易方向 (BUY/SELL)
  triggerPrice?: number; // 触发价格
  quantity?: number; // 交易数量
  slippage?: number; // 滑点百分比
  expirationTime?: number; // 规则过期时间
  isEnabled?: number; // 是否启用
  orderType?: string; // 订单类型 (MARKET/LIMIT)
}

// 交易规则响应接口
export interface TradingRuleResponse {
  status: string;
  message: string;
}

// 获取交易规则请求接口
export interface GetTradingRulesRequest {
  limit?: number; // 限制返回数量
  tokenId?: number; // 按代币ID筛选
  userAddress?: string; // 按用户地址筛选
}