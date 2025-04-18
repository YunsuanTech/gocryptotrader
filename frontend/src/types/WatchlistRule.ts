
interface WatchlistRule {
    ID?: number;
    TokenSymbol: string;
    TokenAddress: string;
    Network: string;
    PriceThreshold: number;
    VolumeThreshold: number;
    TimeWindow: number;
    IsActive: number;
    CreationTime: number;
    LastUpdated: number;
  }
  
  interface AddRuleRequest {
    tokenSymbol: string;
    tokenAddress: string;
    network: string;
    priceThreshold: number;
    volumeThreshold: number;
    timeWindow: number;
    isActive: number;
  }
  
  interface UpdateRuleRequest {
    tokenSymbol: string;
    tokenAddress: string;
    network: string;
    priceThreshold: number;
    volumeThreshold: number;
    timeWindow: number;
    isActive: number;
  }