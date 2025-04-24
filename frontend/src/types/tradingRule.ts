// 交易规则相关的类型定义

// 交易规则信息接口
export interface TradingRule {
  id: number;
  ruleName: string;
  ruleType: string;
  buyPrice: number;
  sellPrice: number;
  condition: string;
  priority: number;
  description: string;
}

// 交易规则列表接口
export interface TradingRuleResponse {
  tradingRules: TradingRule[];
}